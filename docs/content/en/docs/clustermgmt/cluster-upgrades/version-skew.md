---
toc_hide: true
---

There are a few dimensions of versioning to consider in your EKS Anywhere deployments:

- **Management cluster to workload cluster**: Management clusters can be at most 1 EKS Anywhere minor version greater than the EKS Anywhere version of workload clusters. Workload clusters cannot have an EKS Anywhere version greater than management clusters.
- **Management component to workload component**: Management components can be at most 1 EKS Anywhere minor version greater than the EKS Anywhere version of workload components (as of EKS Anywhere version v0.19)
- **EKS Anywhere version upgrades**: Skipping EKS Anywhere minor versions during upgrade is not supported (v0.17 to v0.19). We recommend you upgrade one EKS Anywhere minor version at a time (v0.17 to v0.18 to v0.19).
- **Kubernetes version upgrades**: Skipping Kubernetes minor versions during upgrade is not supported (v1.26 to v1.28). You must upgrade one Kubernetes minor version at a time (v1.26 to v1.27 to v1.28).
- **Kubernetes control plane and worker nodes**: As of Kubernetes v1.28, worker nodes can be up to 3 minor versions lower than the Kubernetes control plane minor version. In earlier Kubernetes versions, worker nodes could be up to 2 minor versions lower than the Kubernetes control plane minor version.