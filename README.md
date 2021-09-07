## Amazon EKS Anywhere

Amazon EKS Anywhere is a new deployment option for Amazon EKS that enables you to easily create and operate Kubernetes clusters on-premises, on your own virtual machines. It brings a consistent AWS management experience to your data center, building on the strengths of [Amazon EKS Distro](https://github.com/aws/eks-distro), the same distribution of Kubernetes that powers EKS on AWS. Its goal is to include full lifecycle management of multiple Kubernetes clusters that are capable of operating completely independently of any AWS services.

Here are the steps to create a cluster using the EKS Anywhere CLI:

1. Ensure that you have all the binaries installed as mentioned in the `pkg/executables` folder of this repository
2. Run `eksctl anywhere generate clusterconfig <cluster-name> -p <provider> > eksa-cluster.yaml` to create a config template. Choose between VSphere or Docker(for development purposes).
3. Run `eksctl anywhere create cluster -f eksa-cluster.yaml` to create your cluster.

Then if you want to delete your cluster, use the `delete cluster` command to delete.

### Testing

Refer [this](https://github.com/aws/eks-anywhere/tree/main/test/e2e) doc for running e2e tests locally.

### Docker Support

EKS Anywhere supports cluster management with docker infrastructure.  
**NOTE**: The Docker support is not designed for production use and is intended for development environments only.

### Logging

EKS-A supports verbosity flag (-v), by default this is set to 0. Increasing this value increases the log volume.

## Releases

Full documentation for releases can be found on [https://anywhere.eks.amazonaws.com](https://anywhere.eks.amazonaws.com).

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the [Apache-2.0 License](LICENSE).
