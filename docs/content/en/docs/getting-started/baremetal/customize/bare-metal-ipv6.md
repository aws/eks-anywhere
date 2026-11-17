---
title: "IPv6 networking for EKS Anywhere on Bare Metal"
linkTitle: "IPv6 Networking"
weight: 40
description: >
  Configure IPv6 bare metal provisioning with SLAAC.
---

Bare Metal IPv6 clusters use SLAAC (Stateless Address Autoconfiguration) for
installed Ubuntu and RHEL node addresses. The control plane endpoint,
Tinkerbell IP, pod CIDR and service CIDR must all use IPv6. Dual-stack clusters
and changing a cluster's IP family during an upgrade are not supported.

## Network prerequisites

- Provide Router Advertisements with `M=0`, `O=1` and `A=1`, an on-link
  provisioning prefix, and a nonzero router lifetime. EKS Anywhere does not
  configure the router. Networks using stateful DHCPv6 (`M=1`) or disabling
  SLAAC (`A=0`) are not supported.
- Make the administrative machine, provisioning endpoints, DNS resolver and
  OS image server reachable over IPv6. Reserve separate IPv6 addresses for
  the control plane endpoint and Tinkerbell services.
- Use a trusted provisioning network. Metadata is selected by the machine's
  MAC address, and its user-data can contain bootstrap credentials. Preserve
  the requesting machine's source address when forwarding metadata traffic.
- Keep BMC connectivity available during creation and subsequent lifecycle
  operations. IPv4 BMC endpoints can be used. If they require an IPv4 address
  or route on the node hosting Tinkerbell, provide it through your host or
  network configuration; IPv6 provisioning does not allocate those values.
- Synchronize machine clocks before creating the cluster.

Network boot requires layer 2 connectivity between machines and the
provisioning stack, plus UEFI IPv6 boot support. ISO boot uses routable
connectivity and BMC virtual media. Router Advertisements are required for
the installed operating system in both modes. See [Boot Modes]({{< relref
"bare-metal-boot-modes" >}}) for the BMC requirements.

## Configure the cluster

Generate the [Bare Metal cluster configuration]({{< relref
"../baremetal-getstarted" >}}), then change its networking fields as shown
below. Retain the generated machine references, node counts, Kubernetes
version and OS image settings. Replace the documentation addresses with
values from your network.

```yaml
# Cluster.spec
clusterNetwork:
  cniConfig:
    cilium: {}
  pods:
    cidrBlocks:
    - "2001:db8:100::/56"
  services:
    cidrBlocks:
    - "2001:db8:200::/108"
controlPlaneConfiguration:
  endpoint:
    host: "2001:db8:66::100"
```

For network boot, configure the temporary firmware address pool:

```yaml
# TinkerbellDatacenterConfig.spec
tinkerbellIP: "2001:db8:66::101"
dhcpv6Config:
  derivedAddressPool: "2001:db8:66::/64"
```

Use the provisioning network's prefix. The firmware boot address is temporary;
HookOS and the installed operating system use SLAAC. See
[`dhcpv6Config`]({{< relref "../bare-spec/#dhcpv6config-required-for-ipv6-network-boot" >}})
for accepted prefixes.

For ISO boot, use the same IPv6 cluster settings and set `isoBoot: true`.
The temporary pool is not required:

```yaml
# TinkerbellDatacenterConfig.spec
tinkerbellIP: "2001:db8:66::101"
isoBoot: true
```

If BMCs cannot reach the generated ISO URL, configure
[`isoBootURL`]({{< relref "../bare-spec/#isobooturl-optional" >}}) with a stable
BMC-reachable Smee handler ending in `/iso6/hook.iso`. The media transport can
use IPv4 even when the machines are provisioned over IPv6. `hookIsoURL`
selects the source image; `isoBootURL` selects the handler served to BMCs.

## Prepare the hardware inventory

Keep the existing CSV columns. Leave `ip_address`, `netmask` and `gateway`
empty; provide the provisioning NIC MAC, installation disk, labels and at
least one reachable IPv6 nameserver. The default route comes from the router.
See [Prepare hardware inventory]({{< relref
"../bare-preparation/#prepare-hardware-inventory" >}}).

Use that CSV with the normal `eksctl anywhere create cluster --hardware-csv`
command. When generating Hardware YAML separately for kubectl or GitOps, use:

```bash
eksctl anywhere generate hardware \
  --hardware-csv hardware.csv --ipv6-slaac > hardware.yaml
```

For custom provisioning templates, regenerate the default template from the
IPv6 cluster configuration before applying your changes. Keep its SLAAC
network configuration, metadata URL and node address setup.

## Verify the addresses

After creation, use the workload cluster's kubeconfig:

```bash
kubectl get nodes -o wide
kubectl -n eksa-system get machines.cluster.x-k8s.io -o wide
kubectl -n eksa-system get tinkerbellmachines -o yaml
```

Confirm that the Node and machine status report the installed OS address.
The temporary firmware address and the Tinkerbell service IP are not node
addresses. On the server, check `ip -6 address`, `ip -6 route`, DNS resolution
and clock synchronization.
