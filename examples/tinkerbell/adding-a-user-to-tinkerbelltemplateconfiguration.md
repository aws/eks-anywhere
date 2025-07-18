# Add a user to the `TinkerbellTempalteConfiguration`

The `create-user` action creates a user in the root file system called `tinkerbell` with password `tinkerbell`.

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellTemplateConfig
metadata:
  name: template-name
spec:
  template:
    global_timeout: 6000
    id: ""
    name: template-name
    tasks:
    - actions:
      // Append to existing actions.
      - environment:
          BLOCK_DEVICE: <block device path> # E.g. /dev/sda1
          FS_TYPE: ext4
          CHROOT: y
          DEFAULT_INTERPRETER: "/bin/sh -c"
          CMD_LINE: "useradd --password $(openssl passwd -1 tinkerbell) --shell /bin/bash --create-home --groups sudo tinkerbell"
        image: public.ecr.aws/l0g8r8j6/tinkerbell/actions/cexec:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-v0.0.0-dev-build.2301
        name: "create-user"
        timeout: 90
```

