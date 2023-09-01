---
title: "Customize HookOS for EKS Anywhere on Bare Metal"
linkTitle: "Customize HookOS"
weight: 30
aliases:
    /docs/reference/baremetal/bare-custom-hookos/
description: >
  Customizing HookOS for EKS Anywhere on Bare Metal
---

To initially network boot bare metal machines used in EKS Anywhere clusters, Tinkerbell acquires a kernel and initial ramdisk that is referred to as the HookOS.
A default HookOS is provided when you create an EKS Anywhere cluster.
However, there may be cases where you want to override the default HookOS, such as to add drivers required to boot your particular type of hardware.

The following procedure describes how to get the Tinkerbell stack’s Hook/Linuxkit OS built locally.
For more information on Tinkerbell’s Hook Installation Environment, see the [Tinkerbell Hook repo](https://github.com/tinkerbell/hook).

1. Clone the hook repo or your fork of that repo:

    ```bash
    git clone https://github.com/tinkerbell/hook.git
    cd hook/
    ```

1. Pull down the commit that EKS Anywhere is tracking for Hook:

    ```bash
    git checkout -b <new-branch> 03a67729d895635fe3c612e4feca3400b9336cc9
    ```

    >**_NOTE_**: This commit number can be obtained from the [EKS-A build tooling repo](https://github.com/aws/eks-anywhere-build-tooling/blob/main/projects/tinkerbell/hook/GIT_TAG).
    >

1. Make changes shown in the following `diff` in the `Makefile` located in the root of the repo using your favorite editor. 

    ```bash
    diff --git a/Makefile b/Makefile
    index e7fd844..8e87c78 100644
    --- a/Makefile
    +++ b/Makefile
    @@ -2,7 +2,7 @@
     ### !!NOTE!!
     # If this is changed then a fresh output dir is required (`git clean -fxd` or just `rm -rf out`)
     # Handling this better shows some of make's suckiness compared to newer build tools (redo, tup ...) where the command lines to tools invoked isn't tracked by make
    -ORG := quay.io/tinkerbell
    +ORG := localhost:5000/tinkerbell
     # makes sure there's no trailing / so we can just add them in the recipes which looks nicer
     ORG := $(shell echo "${ORG}" | sed 's|/*$$||')

     ```

    Changes above change the ORG variable to use a local registry (`localhost:5000`) 

1. Make changes shown in the following `diff` in the `rules.mk` located in the root of the repo using your favorite editor.

    ```bash
    diff --git a/rules.mk b/rules.mk
    index b2c5133..64e32b1 100644
    --- a/rules.mk
    +++ b/rules.mk
    @@ -22,7 +22,7 @@ ifeq ($(ARCH),aarch64)
     ARCH = arm64
     endif
 
    -arches := amd64 arm64
    +arches := amd64
     modes := rel dbg
 
     hook-bootkit-deps := $(wildcard hook-bootkit/*)
    @@ -87,13 +87,12 @@ push-hook-bootkit push-hook-docker:
            docker buildx build --platform $$platforms --push -t $(ORG)/$(container):$T $(container)
 
     .PHONY: dist
    -dist: out/$T/rel/amd64/hook.tar out/$T/rel/arm64/hook.tar ## Build tarballs for distribution
    +dist: out/$T/rel/amd64/hook.tar ## Build tarballs for distribution
     dbg-dist: out/$T/dbg/$(ARCH)/hook.tar ## Build debug enabled tarball
     dist dbg-dist:
            for f in $^; do
            case $$f in
            *amd64*) arch=x86_64 ;;
     -      *arm64*) arch=aarch64 ;;
            *) echo unknown arch && exit 1;;
            esac
            d=$$(dirname $$(dirname $$f))

    ```

    Above changes are for the `docker build` command to only build for the immediately required platform (amd64 in this case) to save time.


1. Modify the `hook.yaml` file located in the root of the repo with the following changes:

    ```bash
    diff --git a/hook.yaml b/hook.yaml
    
     index 0c5d789..b51b35e 100644
    
     net: host
    --- a/hook.yaml
    +++ b/hook.yaml
    @@ -1,5 +1,5 @@
     kernel:
    - image: quay.io/tinkerbell/hook-kernel:5.10.85 (http://quay.io/tinkerbell/hook-kernel:5.10.85)
    + image: localhost:5000/tinkerbell/hook-kernel:5.10.85
     cmdline: "console=tty0 console=ttyS0 console=ttyAMA0 console=ttysclp0"
     init:
     - linuxkit/init:v0.8
    @@ -42,7 +42,7 @@ services:
     binds:
     - /var/run:/var/run
     - name: docker
    - image: quay.io/tinkerbell/hook-docker:0.0 (http://quay.io/tinkerbell/hook-docker:0.0)
    + image: localhost:5000/tinkerbell/hook-docker:0.0
     capabilities:
     - all
     net: host
    @@ -64,7 +64,7 @@ services:
     - /var/run/docker
     - /var/run/worker
     - name: bootkit
    - image: quay.io/tinkerbell/hook-bootkit:0.0 (http://quay.io/tinkerbell/hook-bootkit:0.0)
    + image: localhost:5000/tinkerbell/hook-bootkit:0.0
     capabilities:
     - all
    ```

    The changes above are for using local registry (localhost:5000) for hook-docker, hook-bootkit, and hook-kernel. 

    >**_NOTE_**: You may also need to modify the `hook.yaml` file if you want to add or change components that are used to build up the image. So far, for example, we have needed to change versions of `init` and `getty` and inject SSH keys. Take a look at the [LinuxKit Examples](https://github.com/linuxkit/linuxkit/tree/master/examples) site for examples.
    >

1. Make any planned custom modifications to the files under `hook`, if you are only making changes to `bootkit` or `tink-docker`.
    
    
1. If you are modifying the kernel, such as to change kernel config parameters to add or modify drivers, follow these steps:

    * Change into kernel directory and make a local image for amd64 architecture:

    ```bash
    cd kernel; make kconfig_amd64
    ```

    * Run the image

    ```bash
    docker run --rm -ti -v $(pwd):/src:z quay.io/tinkerbell/kconfig
    ```

    * You can now navigate to the source code and run the UI for configuring the kernel:

    ```bash
    cd linux-5-10
    make menuconfig
    ```

    * Once you have changed the necessary kernel configuration parameters, copy the new configuration:

    ```bash
    cp .config /src/config-5.10.x-x86_64
    ```

    Exit out of container into the repo’s kernel directory and run make:

    ```bash
    /linux-5.10.85 # exit
    user1 % make
    ```

1. Install Linuxkit based on instructions from the [LinuxKit](https://github.com/linuxkit/linuxkit) page.
    
    
1. Ensure that the `linuxkit` tool is in your PATH:

    ```bash
    export PATH=$PATH:/home/tink/linuxkit/bin
    ```

1. Start a local registry:

    ```bash
    docker run -d -p 5000:5000 --name registry registry:2
    ```

1. Compile by running the following in the root of the repo:

    ```bash
    make dist  
    ```
1. Artifacts will be put under the `dist` directory in the repo’s root:

    ```bash
    ./initramfs-aarch64
    ./initramfs-x86_64
    ./vmlinuz-aarch64
    ./vmlinuz-x86_64
    ```

1. To use the kernel (`vmlinuz`) and initial ram disk (`initramfs`) when you build your cluster, see the description of the `hookImagesURLPath` field in your Bare Metal configuration file.
