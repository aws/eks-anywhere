---
title: "Harbor Configuration"
linkTitle: "Harbor"
weight: 130
aliases:
    /docs/reference/packagespec/harbor/
    /docs/tasks/packages/harbor/
description: >
---

[Harbor](https://goharbor.io/) is an open source trusted cloud native registry project that stores, signs, and scans content. Harbor extends the open source Docker Distribution by adding the functionalities usually required by users such as security, identity and management. Having a registry closer to the build and run environment can improve the image transfer efficiency. Harbor supports replication of images between registries, and also offers advanced security features such as user management, access control and activity auditing. For EKS Anywhere deployments, common use cases for Harbor include:

- Supporting [Airgapped]({{< relref "../../getting-started/airgapped/" >}}) environments.
- Running a [registry mirror]({{< relref "../../getting-started/optional/registrymirror/" >}}) that is closer to the build and run environment to improve the image transfer efficiency.
- Following any company policies around image locality.

For additional Harbor use cases see [Harbor use cases]({{< relref "harboruse" >}}).

{{< content "../best_practice.md" >}}

### Configuration options for Harbor
