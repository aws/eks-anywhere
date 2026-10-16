---
title: "Boot modes for EKS Anywhere on Bare Metal"
linkTitle: "Boot Modes"
weight: 30
aliases:
    /docs/reference/baremetal/bare-metal-boot-modes/
description: >
  OS installation boot modes for EKS Anywhere on Bare Metal
---

In order to install an Operating System on a machine, the machine needs to boot into the EKS Anywhere Operating System Installation Environment (HookOS). This is accomplished in one of two ways. The first is via the network, also known as PXE boot. This is the default mode. The second is via a virtual CD/DVD device. In EKS Anywhere this is also known as ISO boot.

## Network Boot

This is the default boot method used for machines in EKS Anywhere. This method requires the EKS Anywhere Bootstrap and Management Clusters to have layer 2 network access to the machines being booted. The network boot process hands out the initial bootloader location via DHCP. In EKS Anywhere this is the iPXE bootloader. The iPXE bootloader then downloads the EKS Anywhere OSIE image and boots into it. The following is a simplified sequence diagram of the network boot process:

For IPv6, configure `dhcpv6Config.derivedAddressPool` for temporary boot addresses and provide Router Advertisements for the installed operating system's SLAAC address and default route. Follow the [IPv6 network boot requirements]({{< relref "../bare-prereq/#ipv6-network-boot-requirements" >}}).

<!-- This is the code to generate this image
```sequence
participant EKS Anywhere
participant BMC
participant Machine
participant File Server

EKS Anywhere -> BMC: set next boot device pxe
Machine -> EKS Anywhere: get network boot program location via DHCP
Machine -> EKS Anywhere: load iPXE
Machine -> File Server: load HookOS
```
--->

![network-boot](/images/eksa-baremetal-net-boot.png)

### Cluster Spec Configuration - Network Boot

As this is the default method, no additional configuration is required. Follow the installation instructions as normal.

IPv6 network boot additionally requires the [DHCPv6 configuration]({{< relref "../bare-spec/#dhcpv6config-required-for-ipv6-network-boot" >}}).

## ISO Boot

The IPv6 release supports network boot. IPv6 ISO boot qualification remains pending; the IPv6 URL example below explains configuration and does not establish ISO support.

This method does not require the EKS Anywhere Bootstrap and Management Clusters to have layer 2 network access to the machines being booted. It does not use DHCP or require layer 2 network access. Instead, the EKS Anywhere HookOS image is provided as a CD/DVD ISO file by the Tinkerbell stack. The HookOS ISO is then attached to the machine via the BMC as a virtual CD/DVD device. The machine is then booted into the virtual CD/DVD device and HookOS boots up. The following is a simplified sequence diagram of the ISO boot process:

<!-- This is the code to generate this image
```sequence
participant EKS Anywhere
participant BMC
participant Machine
participant File Server

EKS Anywhere -> BMC: mount HookOS ISO as virtual CD/DVD device
EKS Anywhere -> BMC: set next boot device virtual CD/DVD device
Machine -> EKS Anywhere: load HookOS
EKS Anywhere -> File Server: patch and serve HookOS
```
-->

![iso-boot](/images/eksa-baremetal-iso-boot.png)

### Cluster Spec Configuration - ISO Boot

To configure the ISO boot method, use these fields:

- Required: `TinkerbellDatacenterConfig.spec.isoBoot` - Set this field to `true` to enable the ISO boot mode.
- Optional: `TinkerbellDatacenterConfig.spec.hookIsoURL` - The source HookOS image that Smee downloads. When omitted, Smee uses the image from the bundle.
- Optional: `TinkerbellDatacenterConfig.spec.isoBootURL` - An absolute HTTP or HTTPS URL for the Smee ISO handler that the BMC can reach. When omitted, EKS Anywhere generates a URL for the current provisioning phase.

{{% alert title="Important" color="warning" %}}
In order to use the ISO boot mode all of the following must be true:
- The BMC info for all machines must be provided in the <a href="{{< relref "../bare-preparation/#prepare-hardware-inventory" >}}">hardware.csv</a> file.
- All BMCs must have virtual media mounting capabilities with remote HTTP(S) support.
- All BMCs must have <a href="https://www.dmtf.org/standards/redfish">Redfish</a> enabled and Redfish must have virtual media mounting capabilities.
{{% /alert %}}

```yaml
spec:
  tinkerbellIP: "192.0.2.10"
  isoBoot: true
  isoBootURL: "https://192.0.2.20/iso/hook.iso"
  hookIsoURL: "http://192.0.2.30/hook-source.iso"
```

Here, `192.0.2.20` represents a stable gateway or load balancer provided by your network that forwards requests to Smee during bootstrap and after the management move. The override is used unchanged in both phases, including later provisioning and rollouts. A URL available only from the temporary administrative bootstrap stack is insufficient. The source image server at `192.0.2.30` is configured independently.

For an IPv6 Tinkerbell cluster, the handler suffix must be `/iso6/hook.iso`, including when the BMC reaches the gateway over IPv4:

```yaml
spec:
  tinkerbellIP: "2001:db8:10::20"
  isoBoot: true
  isoBootURL: "http://192.0.2.20/iso6/hook.iso"
  hookIsoURL: "http://192.0.2.30/hook-source.iso"
```

All addresses in these examples are documentation values. CAPT inserts the machine's MAC address before `hook.iso`, so the gateway must forward the resulting `/iso/<mac>/hook.iso` or `/iso6/<mac>/hook.iso` path. See [`isoBootURL`]({{< relref "../bare-spec/#isobooturl-optional" >}}). IPv6 ISO qualification remains pending.

It is highly recommended that the HookOS ISO file is available locally in your environment and its location is specified in the `hookIsoURL` field. The default location for the HookOS ISO file is hosted in the cloud by AWS. This means that the Tinkerbell stack, and more specifically the Smee service, will need internet access to this location and Smee will need to pull from this location every time a machine is booted. This can lead to slow boot times and potential boot failures if the internet connection is slow, constrained, or lost.

Run the following commmand to download the HookOS ISO file:

```bash
BUNDLE_URL=$(eksctl anywhere version | grep "https://anywhere-assets.eks.amazonaws.com/releases/bundles" | tr -d ' ' | cut -d":" -f2,3)
IMAGE=$(curl -SsL $BUNDLE_URL | grep -E 'uri: .*hook-x86_64-efi-initrd.iso' | uniq | tr -d ' ' | cut -d":" -f2,3)
wget $IMAGE
```

Make this file available via a web server and put the full URL where this ISO is downloadable in the `hookIsoURL` field.

