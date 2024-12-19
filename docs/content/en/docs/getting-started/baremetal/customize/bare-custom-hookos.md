---
title: "Customize HookOS for EKS Anywhere on Bare Metal"
linkTitle: "Customize HookOS"
weight: 30
aliases:
    /docs/reference/baremetal/bare-custom-hookos/
description: >
  Customizing HookOS for EKS Anywhere on Bare Metal
---

To network boot bare metal machines in EKS Anywhere clusters, machines acquire a kernel and initial ramdisk that is referred to as HookOS.
A default HookOS is provided when you create an EKS Anywhere cluster.
However, there may be cases where you want and/or need to customize the default HookOS, such as to add drivers required to boot your particular type of hardware.

The following procedure describes how to customize and build HookOS.
For more information on Tinkerbell’s HookOS Installation Environment, see the [Tinkerbell Hook repo](https://github.com/tinkerbell/hook).

## System requirements

- `>= 2G memory`
- `>= 4 CPU cores` # the more cores the better for kernel building.
- `>= 20G disk space`

## Dependencies

Be sure to install all the following dependencies.

- `jq`
- `envsubst`
- `pigz`
- `docker`
- `curl`
- `bash` >= 4.4
- `git`
- `findutils`

1. Clone the Hook repo or your fork of that repo:

    ```bash
    git clone https://github.com/tinkerbell/hook.git
    cd hook/
    ```

1. Run the Linux kernel [menuconfig](https://en.wikipedia.org/wiki/Menuconfig) TUI and configuring the kernel as needed. Be sure to save the config before exiting.
The result of this step will be a modified kernel configuration file (`./kernel/configs/generic-6.6.y-x86_64`).

    ```bash
    ./build.sh kernel-config hook-latest-lts-amd64
    ```

1. Build the kernel container image. The result of this step will be a container image. Use `docker images quay.io/tinkerbell/hook-kernel` to see it.

    ```bash
    ./build.sh kernel hook-latest-lts-amd64
    ```

1. Add the embedded Action images. This creates the file, `images.txt`, in the `images/hook-embedded` directory and runs the script, `images/hook-embedded/pull-images.sh`, to pull and embedded the images in the HookOS initramfs.
The result of this step will be a populated images file: `images/hook-embedded/images.txt` and a Docker directory cache of images: `images/hook-embedded/images/`.

    ```bash
    BUNDLE_URL=$(eksctl anywhere version | grep "https://anywhere-assets.eks.amazonaws.com/releases/bundles" | tr -d ' ' | cut -d":" -f2,3)
    IMAGES=$(curl -s $BUNDLE_URL | grep "public.ecr.aws/eks-anywhere/tinkerbell/hub/\|public.ecr.aws/eks-anywhere/tinkerbell/tink/tink-worker" | sort | uniq | tr -d ' ' | cut -d":" -f2,3)
    images_file="images/hook-embedded/images.txt"
    rm "$images_file"
    while read -r image; do
      action_name=$(basename "$image" | cut -d":" -f1)
      echo "$image 127.0.0.1/embedded/$action_name" >> "$images_file"
    done <<< "$IMAGES"
    (cd images/hook-embedded; ./pull-images.sh)
    ```

1. Build the HookOS kernel and initramfs artifacts. The `sudo` command is needed as the image embedding step uses Docker-in-Docker (DinD) which changes file ownerships to the root user.
The result of this step will be the kernel and initramfs. These files are located at `./out/hook/vmlinuz-latest-lts-x86_64` and `./out/hook/initramfs-latest-lts-x86_64` respectively.

    ```bash
    sudo ./build.sh linuxkit hook-latest-lts-amd64
    ```

    **Note:** If you did not customize the kernel configuration, you can use the latest upstream built kernel by setting the `USE_LATEST_BUILT_KERNEL` to `yes`. Run this command instead of the previous one.

    ```bash
    sudo ./build.sh linuxkit hook-latest-lts-amd64 USE_LATEST_BUILT_KERNEL=yes
    ```

1. Rename the kernel and initramfs files to `vmlinuz-x86_64` and `initramfs-x86_64` respectively.

    ```bash
    mv ./out/hook/vmlinuz-latest-lts-x86_64 ./out/hook/vmlinuz-x86_64
    mv ./out/hook/initramfs-latest-lts-x86_64 ./out/hook/initramfs-x86_64
    ```

1. To use the kernel (`vmlinuz-x86_64`) and initial ram disk (`initramfs-x86_64`) when you build your EKS Anywhere cluster, see the description of the [`hookImagesURLPath`]({{< relref "../bare-spec#hookimagesurlpath-optional" >}}) field in your cluster configuration file.
