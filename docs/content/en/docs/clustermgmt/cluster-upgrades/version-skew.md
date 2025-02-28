---
toc_hide: true
---

There are a few dimensions of versioning to consider in your EKS Anywhere deployments:

- **Management clusters to workload clusters**: Management clusters can be at most 1 EKS Anywhere minor version greater than the EKS Anywhere version of workload clusters. Workload clusters cannot have an EKS Anywhere version greater than management clusters.
- **Management components to cluster components**: Management components can be at most 1 EKS Anywhere minor version greater than the EKS Anywhere version of cluster components.
- **EKS Anywhere version upgrades**: Skipping EKS Anywhere minor versions during upgrade is not supported (`v0.20.x` to `v0.22.x`). We recommend you upgrade one EKS Anywhere minor version at a time (`v0.20.x` to `v0.21.x` to `v0.22.x`).
- **Kubernetes version upgrades**: Skipping Kubernetes minor versions during upgrade is not supported (`v1.30.x` to `v1.32.x`). You must upgrade one Kubernetes minor version at a time (`v1.30.x` to `v1.31.x` to `v1.32.x`).
- **Kubernetes control plane and worker nodes**: As of Kubernetes v1.28, worker nodes can be up to 3 minor versions lower than the Kubernetes control plane minor version. In earlier Kubernetes versions, worker nodes could be up to 2 minor versions lower than the Kubernetes control plane minor version.