# Support Custom DNS Resolver Configuration 


###Problem
Currently, if the customer has a custom vm image (ova, etc.) that has a custom resolv.conf file, there is no way for the customer to overwrite the default resolv.conf file with the one from the custom image.

###Overview of solution

**Allow customers to provide a custom DNS resolver file**

This can be achieved by specifying the path to the custom DNS resolver file in the cluster spec under `clusterNetwork`:

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
name: cluster-name
spec:
clusterNetwork:
    cni: cilium
    dns:
      resolvConf:
        path: /etc/my-custom-resolv.conf                              
    pods:                                      
      cidrBlocks:                              
      - 192.168.0.0/16                         
    services:                                  
      cidrBlocks:                              
      - 10.96.0.0/12                           
controlPlaneConfiguration:   
.
.          
```

If defined, we will pull this from the cluster spec & it will be added to `kubeletExtraArgs`:
```
kubeletExtraArgs:
  cloud-provider: external
  resolv-conf: /etc/my-custom-resolv.conf
  tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```
If it's not defined, we will use the default (`/etc/resolv.conf`), and it will not be added to `kubeletExtraArgs`. This is the current behavior.

---

**Future considerations:**

In the future, a potential addition to this is to allow the customer to define the content of the DNS resolver file within the cluster spec file, along with the path to the file to store it.
From there, we will create the DNS resolver file based on those specifications and add it to `kubeletExtraArgs`.

Example in the cluster spec:
```
dns:
  resolvConf:
    path: /etc/my-custom-resolv.conf
    content: "my-dns-resolver-content"
```

---
*Things to note:*
* If the `path` is set to the empty string, this will override the default and effectively disable DNS lookups.
  * To avoid this, if the customer sets the path to an empty string, it will be ignored.
    It will proceed to use the default `/etc/resolv.conf` file and it will NOT be added to `kubeletExtraArgs`.
* If specifying a pre-existing custom DNS resolver file, the referenced file must be present on the host filesystem.

### Documentation

We need to add `clusterNetwork.dns.resolvConf.path` as an optional configuration for the cluster spec in our EKS-A docs.