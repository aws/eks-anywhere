---
toc_hide: true
---

* **Standalone clusters**: If you are only running a single EKS Anywhere cluster, you can deploy a standalone cluster. This deployment type runs the EKS Anywhere management components on the same cluster that runs workloads. Standalone clusters must be managed with the `eksctl` CLI. A standalone cluster is effectively a management cluster, but in this deployment type, only manages itself.

* **Management cluster with separate workload clusters**: If you plan to deploy multiple EKS Anywhere clusters, it's recommended to deploy a management cluster with separate workload clusters. With this deployment type, the EKS Anywhere management components are only run on the management cluster, and the management cluster can be used to perform cluster lifecycle operations on a fleet of workload clusters. The management cluster must be managed with the `eksctl` CLI, whereas workload clusters can be managed with the `eksctl` CLI or with Kubernetes API-compatible clients such as `kubectl`, GitOps, or Terraform.