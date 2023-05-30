---
title: "Custom DHCP"
weight: 40
aliases:
    /docs/reference/vsphere/vsphere-dhcp/
description: >
  Create a custom DHCP configuration for your vSphere deployment
---

If your vSphere deployment is not configured with DHCP, you may want to run your own DHCP server.
It may be necessary to turn off DHCP snooping on your switch to get DHCP working across VM servers.
If you are running your administration machine in vSphere, it would most likely be easiest to run the DHCP server on that machine.
This example is for Ubuntu.
 
# Install
Install DHCP server
```
sudo apt-get install isc-dhcp-server
```
 
# Configure /etc/dhcp/dhcpd.conf

Update the ip address range, subnet, mask, etc to suite your configuration similar to this:

```
default-lease-time 600;
max-lease-time 7200;
 
ddns-update-style none;
 
authoritative;
 
subnet 10.8.105.0 netmask 255.255.255.0 {
range 10.8.105.9  10.8.105.41;
option subnet-mask 255.255.255.0;
option routers 10.8.105.1;
 option domain-name-servers 147.149.1.69;
}
```
 
# Configure /etc/default/isc-dhcp-server

Add the main NIC device interface to this file, such as eth0 (this example uses ens160).
 
```
INTERFACESv4="ens160"
```
 
# Restart DHCP

```
service isc-dhcp-server restart
```
 
# Verify your configuration

This example assumes the `ens160` interface:
```
tcpdump -ni ens160 port 67 -vvvv
 
tcpdump: listening on ens160, link-type EN10MB (Ethernet), capture size 262144 bytes
09:13:54.297704 IP (tos 0xc0, ttl 64, id 40258, offset 0, flags [DF], proto UDP (17), length 327)
    10.8.105.12.68 > 10.8.105.5.67: [udp sum ok] BOOTP/DHCP, Request from 00:50:56:90:56:cf, length 299, xid 0xf7a5aac5, secs 50310, Flags [none] (0x0000)
          Client-IP 10.8.105.12
          Client-Ethernet-Address 00:50:56:90:56:cf
          Vendor-rfc1048 Extensions
            Magic Cookie 0x63825363
            DHCP-Message Option 53, length 1: Request
            Client-ID Option 61, length 19: hardware-type 255, 2d:1a:a1:33:00:02:00:00:ab:11:f2:c8:ef:ba:aa:5a:2f:33
            Parameter-Request Option 55, length 11:
              Subnet-Mask, Default-Gateway, Hostname, Domain-Name
              Domain-Name-Server, MTU, Static-Route, Classless-Static-Route
              Option 119, NTP, Option 120
            MSZ Option 57, length 2: 576
            Hostname Option 12, length 15: "prod-etcd-m8ctd"
            END Option 255, length 0
09:13:54.299762 IP (tos 0x0, ttl 64, id 56218, offset 0, flags [DF], proto UDP (17), length 328)
    10.8.105.5.67 > 10.8.105.12.68: [bad udp cksum 0xe766 -> 0x502f!] BOOTP/DHCP, Reply, length 300, xid 0xf7a5aac5, secs 50310, Flags [none] (0x0000)
          Client-IP 10.8.105.12
          Your-IP 10.8.105.12
          Server-IP 10.8.105.5
          Client-Ethernet-Address 00:50:56:90:56:cf
          Vendor-rfc1048 Extensions
            Magic Cookie 0x63825363
            DHCP-Message Option 53, length 1: ACK
            Server-ID Option 54, length 4: 10.8.105.5
            Lease-Time Option 51, length 4: 600
            Subnet-Mask Option 1, length 4: 255.255.255.0
            Default-Gateway Option 3, length 4: 10.8.105.1
            Domain-Name-Server Option 6, length 4: 147.149.1.69
            END Option 255, length 0
            PAD Option 0, length 0, occurs 26
```
 
