---
title: "Ubuntu RTOS"
linkTitle: "Ubuntu RTOS"
weight: 55
description: >
  Ubuntu RTOS Artifacts distributed by EKS Anywhere
---

Starting from EKS Anywhere version v0.20.0, Ubuntu ProRealTime bare metal images will be included in every release of EKS Anywhere. Real-time Ubuntu images contain a real-time Linux kernel which includes the `PREEMPT_RT patchset` to increase predictability and determinism of job execution and response times. These images are designed for workloads with mission-critical latency and security requirements.

These images will be available to use only for specific customers whose AWS accounts have been allowlisted for RTOS image access. If you are a customer whose account is part of the allowlist, you can access the images through the following ways.

1. Install the latest release of the AWS CLI version 2. For more information, refer to the [installion docs](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) in the AWS Command Line Interface User Guide.

2. Configure your AWS CLI to use the credentials of the AWS account which has access to the RTOS artifacts.

3. List the `eks-anywhere-rtos-artifacts` S3 bucket to see the available releases of the Ubuntu RTOS images.
    ```bash
    aws s3 ls s3://eks-anywhere-rtos-artifacts/releases/canonical/
    ```
    This will provide a list of releases ordered by release date, for example,
    ```
    PRE 20240606/
    PRE 20240614/
    PRE 20240624/
    ```

4. Once you have decided the Ubuntu release you wish to use for your EKS Anywhere Tinkerbell cluster, you can download the corresponding Ubuntu image as follows.
    ```bash
    aws s3 cp s3://eks-anywhere-rtos-artifacts/releases/canonical/20240624/artifacts/rtos/1-29/ . --recursive --exclude "*" --include "ubuntu*"
    ```
    The images are also pushed to a `latest` folder which always points to the latest available release of the Ubuntu real-time image.

5. You can now host the downloaded image locally and provide the location as the [`osImageURL`]({{< ref "../getting-started/baremetal/bare-spec#osimageurl-optional" >}}) value in your cluster config.