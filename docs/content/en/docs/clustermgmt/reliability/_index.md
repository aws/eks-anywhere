---
title: "Reliability"
linkTitle: "Reliability"
weight: 50
description: >
  Features and best practices for reliable cluster operations
---

This section covers features and best practices for maintaining reliable EKS Anywhere cluster operations and upgrades.

## Reliability Features

EKS Anywhere includes mechanisms to help ensure operational reliability:

- **Admission Webhook Protection**: Prevents custom admission webhooks from interfering with system operations and cluster upgrades
- **Cluster Validation**: Pre-flight checks identify potential issues before cluster creation or upgrades
- **Support Bundle Collection**: Diagnostic data collection for troubleshooting

## Best Practices

For reliable cluster operations:

1. Enable admission webhook protection for production clusters
2. Keep clusters up-to-date with the latest EKS Anywhere releases
3. Monitor cluster health and resource utilization
4. Test upgrades in non-production environments before production
5. Maintain backup and disaster recovery procedures

## Related Documentation

- [Admission Webhook Protection]({{< relref "./admission-webhook-protection.md" >}})
- [Cluster Upgrades]({{< relref "../cluster-upgrades" >}})
- [Security Best Practices]({{< relref "../security/best-practices.md" >}})
- [Troubleshooting]({{< relref "../../troubleshooting" >}})
