# Add proxy config to the `TinkerbellTempalteConfiguration`

The `configure-containerd-proxy` action configures proxy for containerd on all the Kubernetes nodes. You can add the following section in your cluster config to 
setup proxy on the nodes:

```yaml
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
          CONTENTS: |
            [Service]
            Environment:"HTTP_PROXY=<HTTP-PROXY-IP:PORT>"
            Environment:"HTTPS_PROXY=<HTTPS-PROXY-IP:PORT>"
            Environment:"NO_PROXY=<Comma-separated-list-of-no-proxies>"
          DEST_DISK: <block device path> # E.g. /dev/sda2
          DEST_PATH: /etc/systemd/system/containerd.service.d/http-proxy.conf
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: writefile:v1.0.0
        name: configure-containerd-proxy
        timeout: 90
```

Note the you have to provide values for `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` in the config above.
