## Amazon EKS Anywhere

Amazon EKS Anywhere is a new deployment option for Amazon EKS that enables you to create and operate Kubernetes clusters on-premises with your own virtual machines easily.
It brings a consistent AWS management experience to your datacenter, building on the strengths of [Amazon EKS Distro](https://github.com/aws/eks-distro), the same distribution of Kubernetes that powers EKS on AWS.
Its goal is to include full lifecycle management of multiple Kubernetes clusters that are capable of operating completely independently of any AWS services.

Here are the steps for [getting started](https://anywhere.eks.amazonaws.com/docs/getting-started/) with EKS Anywhere.
Full documentation for releases can be found on [https://anywhere.eks.amazonaws.com](https://anywhere.eks.amazonaws.com/).

## Development

The EKS Anywhere is tested using
[Prow](https://github.com/kubernetes/test-infra/tree/master/prow), the Kubernetes CI system.
EKS operates an installation of Prow, which is visible at [https://prow.eks.amazonaws.com/](https://prow.eks.amazonaws.com/).
Please read our [CONTRIBUTING](CONTRIBUTING.md) guide before making a pull request.
Refer to our [end to end guide](https://github.com/aws/eks-anywhere/tree/main/test/e2e) to run E2E tests locally.

## Security

If you discover a potential security issue in this project, or think you may
have discovered a security issue, we ask that you notify AWS Security via our
[vulnerability reporting page](http://aws.amazon.com/security/vulnerability-reporting/).
Please do **not** create a public GitHub issue.

## License

This project is licensed under the [Apache-2.0 License](LICENSE).
