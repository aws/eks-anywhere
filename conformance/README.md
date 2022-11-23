## Amazon EKS Anywhere Conformance

This document needs to be expanded, but this is something

1. Fork the conformance repository https://github.com/cncf/k8s-conformance
1. Update the version of Kubernetes in the Makefile e.g.: https://github.com/aws/eks-anywhere/pull/1790
1. Create a cluster
1. Run the conformance tests using the Makefile
1. Create a pull request from the branch created in your k8s-conformance repository. e.g.: https://github.com/cncf/k8s-conformance/pull/1872
1. After the pull request is merged, update our README from these images https://github.com/cncf/artwork/blob/master/examples/other.md#certified-kubernetes-logos e.g.: https://github.com/aws/eks-anywhere/pull/1807

