## Bottlerocket

Bottlerocket vends its VMware variant OVAs using a secure distribution tool called tuftool. Please follow instructions down below to
download Bottlerocket OVA.
1. Install Rust and Cargo
```
curl https://sh.rustup.rs -sSf | sh
```
2. Install tuftool using Cargo
```
CARGO_NET_GIT_FETCH_WITH_CLI=true cargo install --force tuftool
```
3. Download the root role tuftool will use to download the OVA
```
curl -O "https://cache.bottlerocket.aws/root.json"
sha512sum -c <<<"90393204232a1ad6b0a45528b1f7df1a3e37493b1e05b1c149f081849a292c8dafb4ea5f7ee17bcc664e35f66e37e4cfa4aae9de7a2a28aa31ae6ac3d9bea4d5  root.json"
```
4. Export the desired Kubernetes Version. EKS-A currently supports 1.21 and 1.20
```
export KUBEVERSION="1.21"
```
5. Download the OVA
```
OVA="bottlerocket-vmware-k8s-${KUBEVERSION}-x86_64-v1.2.0.ova"

tuftool download . --target-name "${OVA}" \
   --root ./root.json \
   --metadata-url "https://updates.bottlerocket.aws/2020-07-07/vmware-k8s-${KUBEVERSION}/x86_64/" \
   --targets-url "https://updates.bottlerocket.aws/targets/"
```

Bottlerocket Tags

OS Family - `os:bottlerocket`

EKS-D Release

1.21 - `eksdRelease:kubernetes-1-21-eks-4`

1.20 - `eksdRelease:kubernetes-1-20-eks-6`

## Ubuntu with Kubernetes 1.21

* https://eks-anywhere-beta.s3.amazonaws.com/0.4.0/ova/ubuntu-v1.21.2-eks-d-1-21-4-eks-a-0.4.0-amd64.ova
* `os:ubuntu`
* `eksdRelease:kubernetes-1-21-eks-4`

## Ubuntu with Kubernetes 1.20

* https://eks-anywhere-beta.s3.amazonaws.com/0.4.0/ova/ubuntu-v1.20.7-eks-d-1-20-6-eks-a-0.4.0-amd64.ova
* `os:ubuntu`
* `eksdRelease:kubernetes-1-20-eks-6`