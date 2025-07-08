---
title: "Monitoring Certificate Expiration"
linkTitle: "Monitoring Certificate Expiration"
weight: 10
description: >
  How to monitor certificate expiration in EKS Anywhere clusters
---

There are a few ways to check certificate expiration dates depending on your cluster's accessibility.

#### Using ClusterCertificateInfo (When cluster is accessible)

When your cluster is accessible via kubectl, use the ClusterCertificateInfo field in the cluster status:

```bash
kubectl get cluster <cluster-name> -o json | jq '.status.clusterCertificateInfo'
```

This will output a list of objects containing the machine name and the number of days until the certificate expires:

```json
[
  {
    "machine": "my-cluster-control-plane-abc123",
    "expiresInDays": 300
  },
  {
    "machine": "my-cluster-control-plane-def456",
    "expiresInDays": 300
  },
  {
    "machine": "my-cluster-etcd-ghi789",
    "expiresInDays": 300
  }
]
```

The `ClusterCertificateInfo` field contains a list of machine objects with the following properties:

- `machine`: The name of the machine (control plane or external etcd)
- `expiresInDays`: The number of days until the certificate expires

This information is updated periodically as part of the cluster status reconciliation process. The certificate expiration check is performed by connecting to each machine's API server (port 6443 for control plane) or etcd server (port 2379 for external etcd) and retrieving the certificate expiration date.


#### Using OpenSSL for Direct Certificate Inspection (When cluster is not accessible)

When the cluster is not accessbile, you can directly check certificates using opessl:

```bash
# The expiry time of api-server certificate on control plane node
echo | openssl s_client -connect ${CONTROL_PLANE_IP}:6443 2>/dev/null | openssl x509 -noout -dates

# The expiry time of certificate used by your external etcd server, if you configured one
echo | openssl s_client -connect ${EXTERNAL_ETCD_IP}:2379 2>/dev/null | openssl x509 -noout -dates
```

### Monitoring Best Practices

1. **Regular Checks**: Periodically check certificate expiration to ensure you have enough time to plan for certificate renewal.
2. **Set Up Alerts**: Consider setting up alerts to notify you when certificates are approaching expiration (e.g., 30 days before expiration).
3. **Proactive Renewal**: If certificates are approaching expiration (less than 30 days), plan for certificate renewal using one of the methods described below.

### Certificate Renewal

EKS Anywhere automatically rotates certificates when new machines are rolled out during cluster lifecycle operations such as `upgrade` (ie. EKS Anywhere version upgrades or Kubernetes version upgrades where nodes actually roll out). If you upgrade your cluster at least once a year, as recommended, manual rotation of cluster certificates will not be necessary.

If you need to manually renew certificates, you can use one of the following methods:

- [Manual steps to renew certificates]({{< relref "manual-steps-renew-certs.md" >}}) - Step-by-step manual process
- [Script to renew certificates]({{< relref "script-renew-certs.md" >}}) - Automated approach using a script
