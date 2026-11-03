"""Deterministic tests of the installed node-IP helper; no network or pip setup."""
import contextlib
import copy
import importlib.util
import io
import ipaddress
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import types
import unittest
from unittest.mock import Mock, patch

sys.dont_write_bytecode = True
path = Path(os.environ.get("EKSA_NODE_IP_HELPER",
                           Path(__file__).with_name("ipv6-node-ip.py")))
spec = importlib.util.spec_from_file_location("node_ip_helper", path)
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)

try:
    import yaml as real_yaml
except ImportError:
    real_yaml = None


def address(value="2001:db8::20", **fields):
    return {"local": value, "scope": "global",
            "preferred_life_time": 3600, **fields}


class AddressSelectionTests(unittest.TestCase):
    def test_preferred_source_and_single_candidate(self):
        interfaces = [{"ifname": "dummy0", "addr_info": [address()]}]
        for source in ({}, {"prefsrc": "2001:db8::20"},
                       {"src": "2001:db8::20"}):
            with self.subTest(source=source):
                selected = helper.select_node_ip(
                    interfaces, {"dev": "dummy0", **source}, set())
                self.assertEqual(selected, "2001:db8::20")

    def test_ineligible_addresses(self):
        cases = [
            address(scope="link"),
            address("fe80::20"),
            address("192.0.2.20"),
            address(deprecated=True),
            address(tentative=True),
            address(dadfailed=True),
            address(temporary=True),
            address(preferred_life_time=0),
        ]
        for entry in cases:
            with self.subTest(entry=entry):
                self.assertIsNone(helper.select_node_ip(
                    [{"ifname": "dummy0", "addr_info": [entry]}],
                    {"dev": "dummy0", "prefsrc": entry["local"]}, set()))

    def test_excluded_vip_and_wrong_interface(self):
        interfaces = [{"ifname": "dummy0", "addr_info": [address()]}]
        excluded = {ipaddress.ip_address("2001:db8::20")}
        self.assertIsNone(helper.select_node_ip(
            interfaces, {"dev": "dummy0", "prefsrc": "2001:db8::20"},
            excluded))
        self.assertIsNone(helper.select_node_ip(
            interfaces, {"dev": "other0"}, set()))


class RouteTests(unittest.TestCase):
    def test_skips_local_route_then_uses_next_target(self):
        results = [
            subprocess.CompletedProcess([], 0, json.dumps(
                [{"type": "local", "dev": "lo"}])),
            subprocess.CompletedProcess([], 0, json.dumps(
                [{"dev": "dummy0", "prefsrc": "2001:db8::20"}])),
        ]
        with patch.object(helper.subprocess, "run", side_effect=results) as run:
            route = helper.provisioning_route(["2001:db8::11", "2001:db8::10"])
        self.assertEqual(route["dev"], "dummy0")
        self.assertEqual(run.call_args_list[1].args[0],
                         ["ip", "-j", "-6", "route", "get", "2001:db8::10"])

    def test_failed_target_falls_back_to_default(self):
        result = subprocess.CompletedProcess([], 1, "", "no route")
        with patch.object(helper.subprocess, "run", return_value=result), \
                patch.object(helper.subprocess, "check_output",
                             return_value='[{"dev":"dummy0","metric":100}]'):
            self.assertEqual(helper.provisioning_route(["2001:db8::11"]),
                             {"dev": "dummy0", "metric": 100})

    def test_missing_default_route(self):
        with patch.object(helper.subprocess, "check_output", return_value="[]"):
            self.assertEqual(helper.provisioning_route([]), {})


class KubeadmDocumentTests(unittest.TestCase):
    def test_init_control_plane_join_and_worker_documents(self):
        # Mock only the YAML serialization boundary. All field updates execute
        # production code, including on CI hosts without PyYAML installed.
        yaml_adapter = types.SimpleNamespace(
            safe_load_all=Mock(side_effect=json.loads),
            safe_dump_all=Mock(side_effect=lambda docs, **kwargs: json.dumps(docs)))
        for args in ({}, [], {"provider-id": "dummy", "node-ip": "2001:db8::99"},
                     [{"name": "provider-id", "value": "dummy"},
                      {"name": "node-ip", "value": "2001:db8::99"},
                      {"name": "node-ip", "value": "2001:db8::98"}]):
            for kind, control_plane in (
                    ("InitConfiguration", "absent"),
                    ("JoinConfiguration", {"localAPIEndpoint": {
                        "advertiseAddress": "2001:db8::99", "bindPort": 7443},
                        "certificateKey": "dummy-key"}),
                    ("JoinConfiguration", {}),
                    ("JoinConfiguration", None),
                    ("JoinConfiguration", "absent")):
                with self.subTest(args=args, kind=kind, control_plane=control_plane), \
                        tempfile.TemporaryDirectory() as directory:
                    document = {
                        "apiVersion": ("kubeadm.k8s.io/v1beta3" if isinstance(args, dict)
                                       else "kubeadm.k8s.io/v1beta4"), "kind": kind,
                        "nodeRegistration": {"name": "dummy-node",
                                             "kubeletExtraArgs": copy.deepcopy(args)},
                        "patches": {"directory": "/dummy/patches"},
                    }
                    if kind == "InitConfiguration":
                        document["localAPIEndpoint"] = {
                            "advertiseAddress": "2001:db8::99", "bindPort": 7443}
                    elif control_plane != "absent":
                        document["controlPlane"] = copy.deepcopy(control_plane)
                    cluster_document = {"kind": "ClusterConfiguration",
                                        "networking": {"podSubnet": "2001:db8:100::/48"}}
                    filename = ("kubeadm.yaml" if kind == "InitConfiguration"
                                else "kubeadm-join-config.yaml")
                    path = Path(directory) / filename
                    path.write_text(json.dumps([None, cluster_document, document]))
                    with patch.dict(sys.modules, {"yaml": yaml_adapter}):
                        helper.set_kubeadm_node_ip("2001:db8::20", Path(directory))
                    documents = json.loads(path.read_text())
                    self.assertIsNone(documents[0])
                    self.assertEqual(documents[1], cluster_document)
                    updated = documents[2]
                    self.assertEqual(updated["patches"], document["patches"])
                    self.assertEqual(updated["nodeRegistration"]["name"], "dummy-node")
                    registration_args = updated["nodeRegistration"]["kubeletExtraArgs"]
                    if isinstance(registration_args, dict):
                        expected = {**args, "node-ip": "2001:db8::20"}
                        self.assertEqual(registration_args,
                                         expected)
                    else:
                        expected = ([{"name": "provider-id", "value": "dummy"}]
                                    if args else [])
                        expected.append({"name": "node-ip", "value": "2001:db8::20"})
                        self.assertEqual(registration_args,
                                         expected)
                    if kind == "InitConfiguration":
                        endpoint = updated["localAPIEndpoint"]
                        self.assertEqual(endpoint["bindPort"], 7443)
                    elif control_plane != "absent":
                        self.assertNotIn("localAPIEndpoint", updated)
                        endpoint = updated["controlPlane"]["localAPIEndpoint"]
                        if control_plane:
                            self.assertEqual(endpoint["bindPort"], 7443)
                            self.assertEqual(updated["controlPlane"]["certificateKey"], "dummy-key")
                    else:
                        self.assertNotIn("controlPlane", updated)
                        self.assertNotIn("localAPIEndpoint", updated)
                        continue
                    self.assertEqual(endpoint["advertiseAddress"], "2001:db8::20")

    def test_missing_configuration(self):
        yaml_adapter = types.SimpleNamespace()
        with tempfile.TemporaryDirectory() as directory, \
                patch.dict(sys.modules, {"yaml": yaml_adapter}):
            with self.assertRaisesRegex(RuntimeError, "No kubeadm init/join"):
                helper.set_kubeadm_node_ip("2001:db8::20", Path(directory))

    def test_omitted_args_keep_kubeadm_api_format(self):
        yaml_adapter = types.SimpleNamespace(
            safe_load_all=json.loads,
            safe_dump_all=lambda docs, **kwargs: json.dumps(docs))
        for version, expected_type in (("v1beta3", dict), ("v1beta4", list)):
            with self.subTest(version=version), tempfile.TemporaryDirectory() as directory:
                path = Path(directory) / "kubeadm-join-config.yaml"
                path.write_text(json.dumps([{
                    "apiVersion": "kubeadm.k8s.io/" + version,
                    "kind": "JoinConfiguration"}]))
                with patch.dict(sys.modules, {"yaml": yaml_adapter}):
                    helper.set_kubeadm_node_ip("2001:db8::20", Path(directory))
                doc = json.loads(path.read_text())[0]
                self.assertIsInstance(doc["nodeRegistration"]["kubeletExtraArgs"], expected_type)
                self.assertNotIn("controlPlane", doc)

    @unittest.skipUnless(real_yaml, "PyYAML is optional for the serialization integration test")
    def test_real_yaml_round_trip(self):
        with tempfile.TemporaryDirectory() as directory:
            directory = Path(directory)
            init = {"apiVersion": "kubeadm.k8s.io/v1beta3", "kind": "InitConfiguration",
                    "nodeRegistration": {"kubeletExtraArgs": {"provider-id": "dummy"}},
                    "localAPIEndpoint": {"bindPort": 7443}}
            join = {"apiVersion": "kubeadm.k8s.io/v1beta4", "kind": "JoinConfiguration",
                    "nodeRegistration": {"kubeletExtraArgs": [
                        {"name": "provider-id", "value": "dummy"}]},
                    "controlPlane": {"localAPIEndpoint": {"bindPort": 7443}}}
            cluster = {"kind": "ClusterConfiguration", "networking": {
                "podSubnet": "2001:db8:100::/48"}}
            (directory / "kubeadm.yaml").write_text(real_yaml.safe_dump_all([cluster, init]))
            (directory / "kubeadm-join-config.yaml").write_text(real_yaml.safe_dump(join))
            helper.set_kubeadm_node_ip("2001:db8::20", directory)
            init_documents = list(real_yaml.safe_load_all(
                (directory / "kubeadm.yaml").read_text()))
            join_document = real_yaml.safe_load(
                (directory / "kubeadm-join-config.yaml").read_text())
            self.assertEqual(init_documents[0], cluster)
            self.assertEqual(init_documents[1]["localAPIEndpoint"],
                             {"bindPort": 7443, "advertiseAddress": "2001:db8::20"})
            self.assertEqual(join_document["controlPlane"]["localAPIEndpoint"],
                             {"bindPort": 7443, "advertiseAddress": "2001:db8::20"})
            self.assertEqual(init_documents[1]["nodeRegistration"]["kubeletExtraArgs"]["node-ip"],
                             "2001:db8::20")
            self.assertIn({"name": "node-ip", "value": "2001:db8::20"},
                          join_document["nodeRegistration"]["kubeletExtraArgs"])


class MainTests(unittest.TestCase):
    def test_selected_address_updates_config_and_prints(self):
        interfaces = [{"ifname": "dummy0", "addr_info": [address()]}]
        with patch.object(sys, "argv", ["helper", "2001:db8::11"]), \
                patch.object(helper.subprocess, "check_output",
                             return_value=json.dumps(interfaces)), \
                patch.object(helper, "provisioning_route",
                             return_value={"dev": "dummy0", "prefsrc": "2001:db8::20"}), \
                patch.object(helper, "set_kubeadm_node_ip") as write, \
                contextlib.redirect_stdout(io.StringIO()) as output:
            helper.main()
        write.assert_called_once_with("2001:db8::20")
        self.assertEqual(output.getvalue(), "2001:db8::20\n")

    def test_print_only_does_not_write(self):
        interfaces = [{"ifname": "dummy0", "addr_info": [address()]}]
        with patch.object(sys, "argv", ["helper", "--print", "2001:db8::11"]), \
                patch.object(helper.subprocess, "check_output",
                             return_value=json.dumps(interfaces)), \
                patch.object(helper, "provisioning_route", return_value={"dev": "dummy0"}), \
                patch.object(helper, "set_kubeadm_node_ip") as write, \
                contextlib.redirect_stdout(io.StringIO()):
            helper.main()
        write.assert_not_called()

    def test_address_wait_exhaustion_fails_without_writing(self):
        with patch.object(sys, "argv", ["helper", "2001:db8::11"]), \
                patch.object(helper.subprocess, "check_output", return_value="[]") as read, \
                patch.object(helper, "provisioning_route", return_value={}), \
                patch.object(helper.time, "sleep") as sleep, \
                patch.object(helper, "set_kubeadm_node_ip") as write:
            with self.assertRaisesRegex(RuntimeError, "No preferred provisioning"):
                helper.main()
        self.assertEqual(read.call_count, 60)
        self.assertEqual(sleep.call_count, 60)
        write.assert_not_called()

    def test_command_and_config_errors_propagate(self):
        interfaces = [{"ifname": "dummy0", "addr_info": [address()]}]
        for command_error in (True, False):
            with self.subTest(command_error=command_error), \
                    patch.object(sys, "argv", ["helper", "2001:db8::11"]), \
                    patch.object(helper.subprocess, "check_output",
                                 side_effect=(subprocess.CalledProcessError(1, ["ip"])
                                              if command_error else None),
                                 return_value=json.dumps(interfaces)), \
                    patch.object(helper, "provisioning_route", return_value={"dev": "dummy0"}), \
                    patch.object(helper, "set_kubeadm_node_ip",
                                 side_effect=RuntimeError("missing config")):
                expected = subprocess.CalledProcessError if command_error else RuntimeError
                with self.assertRaises(expected):
                    helper.main()


if __name__ == "__main__":
    unittest.main(verbosity=2)
