#!/usr/bin/env python3
"""Set kubeadm's node IP from the installed OS's preferred IPv6 address."""
import ipaddress
import json
from pathlib import Path
import subprocess
import sys
import time


def select_node_ip(interfaces, route, excluded):
    candidates = set()
    for interface in interfaces:
        if interface["ifname"] != route.get("dev"):
            continue
        for address in interface.get("addr_info", []):
            if address.get("scope") != "global":
                continue
            if any(address.get(flag) for flag in
                   ("deprecated", "tentative", "dadfailed", "temporary")):
                continue
            if address.get("preferred_life_time") == 0:
                continue
            ip = ipaddress.ip_address(address["local"])
            if ip.version == 6 and not ip.is_link_local and ip not in excluded:
                candidates.add(ip)
    source = route.get("prefsrc") or route.get("src")
    if source and ipaddress.ip_address(source) in candidates:
        return source
    if len(candidates) > 1:
        raise RuntimeError("Multiple preferred provisioning IPv6 addresses: " +
                           ", ".join(str(ip) for ip in sorted(candidates)))
    return str(next(iter(candidates))) if candidates else None


def provisioning_route(targets):
    for target in targets:
        result = subprocess.run(
            ["ip", "-j", "-6", "route", "get", target],
            capture_output=True, text=True)
        routes = json.loads(result.stdout) if result.returncode == 0 else []
        for route in routes:
            if route.get("type") != "local" and route.get("dev") not in (None, "lo"):
                return route
    routes = json.loads(subprocess.check_output(
        ["ip", "-j", "-6", "route", "show", "default"], text=True))
    if routes:
        return min(routes, key=lambda route: route.get("metric", 0))
    return {}


def set_kubeadm_node_ip(node_ip, config_dir=Path("/run/kubeadm")):
    import yaml

    changed = False
    for name in ("kubeadm.yaml", "kubeadm-join-config.yaml"):
        path = config_dir / name
        if not path.exists():
            continue
        documents = list(yaml.safe_load_all(path.read_text()))
        found = False
        for document in documents:
            if not document or document.get("kind") not in (
                    "InitConfiguration", "JoinConfiguration"):
                continue
            registration = document.setdefault("nodeRegistration", {})
            args = registration.get("kubeletExtraArgs")
            if args is None:
                args = ({} if document.get("apiVersion") == "kubeadm.k8s.io/v1beta3"
                        else [])
            if isinstance(args, dict):
                args["node-ip"] = node_ip
            else:
                args = [arg for arg in args if arg["name"] != "node-ip"]
                args.append({"name": "node-ip", "value": node_ip})
            registration["kubeletExtraArgs"] = args
            # Kubeadm selects API/etcd addresses independently of kubelet node-ip.
            if document["kind"] == "InitConfiguration":
                endpoint = document.setdefault("localAPIEndpoint", {})
                endpoint["advertiseAddress"] = node_ip
            elif "controlPlane" in document:
                control_plane = document["controlPlane"]
                if control_plane is None:
                    control_plane = document["controlPlane"] = {}
                endpoint = control_plane.setdefault("localAPIEndpoint", {})
                endpoint["advertiseAddress"] = node_ip
            found = True
        if found:
            path.write_text(yaml.safe_dump_all(documents, sort_keys=False))
            changed = True
    if not changed:
        raise RuntimeError("No kubeadm init/join configuration found")


def main():
    arguments = sys.argv[1:]
    print_only = "--print" in arguments
    arguments = [argument for argument in arguments if argument != "--print"]
    excluded = {ipaddress.ip_address(argument) for argument in arguments}
    for attempt in range(60):
        interfaces = json.loads(subprocess.check_output(
            ["ip", "-j", "-6", "address", "show", "scope", "global"], text=True))
        node_ip = select_node_ip(interfaces, provisioning_route(arguments), excluded)
        if node_ip:
            if not print_only:
                set_kubeadm_node_ip(node_ip)
            print(node_ip)
            return
        time.sleep(1)
    raise RuntimeError("No preferred provisioning IPv6 address became available")


if __name__ == "__main__":
    main()
