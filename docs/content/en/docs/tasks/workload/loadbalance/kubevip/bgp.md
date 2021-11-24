---
title: "Kube-Vip BGP Mode"
linkTitle: "Kube-Vip BGP Mode"
weight: 30
date: 2017-01-05
description: >
  How to set up kube-vip for Service-type Load Balancer in BGP mode
---

<!-- overview -->

In BGP mode, kube-vip will assign the Virtual IP to all running Pods.
All nodes, therefore, will advertise the VIP address.

<!-- body -->


### Prerequisites
- BGP-capable network switch connected to EKS-A cluster
- Vendor-specific BGP configuration on switch

Required BGP settings on network vendor equipment are described in [BGP Configuration on Network Switch Side]({{< ref "docs/tasks/workload/loadbalance/kubevip/bgp.md#bgp-configuration-on-network-switch-side" >}}) section below.


## Setting up Kube-Vip for Service-type Load Balancer in BGP mode

1. Create a configMap to specify the IP range for load balancer.
You can use either a CIDR block or an IP range

    ```bash
    CIDR=192.168.0.0/24 # Use your CIDR range here
    kubectl create configmap --namespace kube-system kubevip --from-literal cidr-global=${CIDR}
    ```
    ```bash
    IP_START=192.168.0.0  # Use the starting IP in your range
    IP_END=192.168.0.255  # Use the ending IP in your range
    kubectl create configmap --namespace kube-system kubevip --from-literal range-global=${IP_START}-${IP_END}
    ```

1. Deploy kubevip-cloud-provider

    ```bash
    kubectl apply -f https://kube-vip.io/manifests/controller.yaml
    ```

1. Create ClusterRoles and RoleBindings for kube-vip Daemonset

    ```bash
    kubectl apply -f https://kube-vip.io/manifests/rbac.yaml
    ```

1. Create the kube-vip DaemonSet

    ```bash
    alias kube-vip="docker run --network host --rm plndr/kube-vip:latest"
    kube-vip manifest daemonset \
        --interface lo \
        --localAS <AS#> \
        --sourceIF <src interface> \
        --services \
        --inCluster \
        --bgp \
        --bgppeers <bgp-peer1>:<peerAS>::<bgp-multiphop-true-false>,<bgp-peer2>:<peerAS>::<bgp-multihop-true-false> | kubectl apply -f -
    ``` 
    
    Explanation of the options provided above to kube-vip for manifest generation:

    ```
	--interface — This interface needs to be set to the loopback in order to suppress ARP responses from worker nodes that get the LoadBalancer VIP assigned
	--localAS — Local Autonomous System ID
	--sourceIF — source interface on the worker node which will be used to communicate BGP with the switch
	--services — Service Type LoadBalancer (not Control Plane)
	--inCluster — Defaults to looking inside the Pod for the token
	--bgp — Enables BGP peering from kube-vip
	--bgppeers — Comma separated list of BGP peers in the format <address:AS:password:multihop>
    ```
  
    Below is an example DaemonSet creation command. 

    ```bash
    kube-vip manifest daemonset \
        --interface $INTERFACE \
        --localAS 65200 \
        --sourceIF eth0 \
        --services \
        --inCluster \
        --bgp \
        --bgppeers 10.69.20.2:65000::false,10.69.20.3:65000::false
    ```

    Below is the manifest generated with these example values.

    ```yaml
    apiVersion: apps/v1
    kind: DaemonSet
    metadata:
      creationTimestamp: null
      name: kube-vip-ds
      namespace: kube-system
    spec:
      selector:
        matchLabels:
          name: kube-vip-ds
      template:
        metadata:
          creationTimestamp: null
          labels:
            name: kube-vip-ds
        spec:
          containers:
          - args:
            - manager
            env:
            - name: vip_arp
              value: "false"
            - name: vip_interface
              value: lo
            - name: port
              value: "6443"
            - name: vip_cidr
              value: "32"
            - name: svc_enable
              value: "true"
            - name: cp_enable
              value: "false"
            - name: vip_startleader
              value: "false"
            - name: vip_addpeerstolb
              value: "true"
            - name: vip_localpeer
              value: docker-desktop:192.168.65.3:10000
            - name: bgp_enable
              value: "true"
            - name: bgp_routerid
            - name: bgp_source_if
              value: eth0
            - name: bgp_as
              value: "65200"
            - name: bgp_peeraddress
            - name: bgp_peerpass
            - name: bgp_peeras
              value: "65000"
            - name: bgp_peers
              value: 10.69.20.2:65000::false,10.69.20.3:65000::false
            - name: bgp_routerinterface
              value: eth0
            - name: vip_address
            image: ghcr.io/kube-vip/kube-vip:v0.3.7
            imagePullPolicy: Always
            name: kube-vip
            resources: {}
            securityContext:
              capabilities:
                add:
                - NET_ADMIN
                - NET_RAW
                - SYS_TIME
          hostNetwork: true
          serviceAccountName: kube-vip
      updateStrategy: {}
    status:
      currentNumberScheduled: 0
      desiredNumberScheduled: 0
      numberMisscheduled: 0
      numberReady: 0
   ```
1. Manually add the following to the manifest file as shown in the example above

   ```yaml
   - name: bgp_routerinterface
     value: eth0
   ```

1. Deploy the [Hello EKS Anywhere]({{< relref "../../test-app" >}}) test application.

    ```bash
    kubectl apply -f https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml
    ```

1. Expose the hello-eks-a service

    ```bash
    kubectl expose deployment hello-eks-a --port=80 --type=LoadBalancer --name=hello-eks-a-lb
    ```

1. Describe the service to get the IP. The external IP will be the one in CIDR range specified in step 4

    ```bash
    EXTERNAL_IP=$(kubectl get svc hello-eks-a-lb -o jsonpath='{.spec.externalIP}')
    ```

1. Ensure the load balancer is working by curl'ing the IP you got in step 8

    ```bash
    curl ${EXTERNAL_IP}
    ```   
 
You should see something like this in the output

```
   ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

   Thank you for using

   ███████╗██╗  ██╗███████╗                                             
   ██╔════╝██║ ██╔╝██╔════╝                                             
   █████╗  █████╔╝ ███████╗                                             
   ██╔══╝  ██╔═██╗ ╚════██║                                             
   ███████╗██║  ██╗███████║                                             
   ╚══════╝╚═╝  ╚═╝╚══════╝                                             
                                                                     
    █████╗ ███╗   ██╗██╗   ██╗██╗    ██╗██╗  ██╗███████╗██████╗ ███████╗
   ██╔══██╗████╗  ██║╚██╗ ██╔╝██║    ██║██║  ██║██╔════╝██╔══██╗██╔════╝
   ███████║██╔██╗ ██║ ╚████╔╝ ██║ █╗ ██║███████║█████╗  ██████╔╝█████╗  
   ██╔══██║██║╚██╗██║  ╚██╔╝  ██║███╗██║██╔══██║██╔══╝  ██╔══██╗██╔══╝  
   ██║  ██║██║ ╚████║   ██║   ╚███╔███╔╝██║  ██║███████╗██║  ██║███████╗
   ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝    ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝
                                                                     
   You have successfully deployed the hello-eks-a pod hello-eks-a-c5b9bc9d8-fx2fr

   For more information check out
   https://anywhere.eks.amazonaws.com

   ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

   ```


## BGP Configuration on Network Switch Side

BGP configuration will vary depending upon network vendor equipment and local network environment. Listed below are the basic conceptual configuration steps for BGP operation. Included with each step is a sample configuration from a Cisco Switch (Cisco Nexus 9000) running in NX-OS mode. You will need to find similar steps in your network vendor equipment's manual for BGP configuration on your specific switch.


1. Configure BGP local AS, router ID, and timers

   ```
   router bgp 65000
     router-id 10.69.5.1
     timers bgp 15 45
     log-neighbor-changes
   ```


2. Configure BGP neighbors

   BGP neighbors can be configured individually or as a subnet

   a. Individual BGP neighbors

      Determine the IP addresses of each of the EKS-A nodes via VMWare console or DHCP server allocation.  
      In the example below, node IP addresses are 10.69.20.165, 10.69.20.167, and 10.69.20.170.  
      Note that remote-as is the AS used as the bgp_as value in the generated example manifest above.

      ```
      neighbor 10.69.20.165
          remote-as 65200
          address-family ipv4 unicast
            soft-reconfiguration inbound always
      neighbor 10.69.20.167
          remote-as 65200
          address-family ipv4 unicast
            soft-reconfiguration inbound always
      neighbor 10.69.20.170
          remote-as 65200
          address-family ipv4 unicast
            soft-reconfiguration inbound always
      ```

   b. Subnet-based BGP neighbors

      Determine the subnet address and netmask of the EKS-A nodes.
      In this example the EKS-A nodes are on 10.69.20.0/24 subnet.
      Note that remote-as is the AS used as the bgp_as value in the generated example manifest above.

      ```
      neighbor 10.69.20.0/24
          remote-as 65200
          address-family ipv4 unicast
            soft-reconfiguration inbound always
      ```


3. Verify bgp neighbors are established with each node

   ```
   switch% show ip bgp summary
   information for VRF default, address family IPv4 Unicast
   BGP router identifier 10.69.5.1, local AS number 65000
   BGP table version is 181, IPv4 Unicast config peers 7, capable peers 7
   32 network entries and 63 paths using 11528 bytes of memory
   BGP attribute entries [16/2752], BGP AS path entries [6/48]
   BGP community entries [0/0], BGP clusterlist entries [0/0]
   3 received paths for inbound soft reconfiguration
   3 identical, 0 modified, 0 filtered received paths using 0 bytes

   Neighbor        V    AS MsgRcvd MsgSent   TblVer  InQ OutQ Up/Down  State/PfxRcd
   10.69.20.165    4 65200   34283   34276      181    0    0    5d20h 1
   10.69.20.167    4 65200   34543   34531      181    0    0    5d20h 1
   10.69.20.170    4 65200   34542   34530      181    0    0    5d20h 1
   ```


4. Verify routes learned from EKS-A cluster match the external IP address assigned by kube-vip LoadBalancer configuration

   In the example below, 10.35.10.13 is the external kube-vip LoadBalancer IP

   ```
   switch% show ip bgp neighbors 10.69.20.165 received-routes

   Peer 10.69.20.165 routes for address family IPv4 Unicast:
   BGP table version is 181, Local Router ID is 10.69.5.1
   Status: s-suppressed, x-deleted, S-stale, d-dampened, h-history, *-valid, >-best
   Path type: i-internal, e-external, c-confed, l-local, a-aggregate, r-redist, I-injected
   Origin codes: i - IGP, e - EGP, ? - incomplete, | - multipath, & - backup, 2 - best2

      Network            Next Hop            Metric     LocPrf     Weight Path
   *>e10.35.10.13/32     10.69.20.165                                   0 65200 i
   ```
