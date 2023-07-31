---
title: "License cluster"
linkTitle: "License cluster"
weight: 20
date: 2021-12-10
aliases:
    /docs/tasks/cluster/cluster-license/
description: >
  How to license your cluster.
---

If you are are licensing an existing cluster, apply the following secret to your cluster (replacing `my-license-here` with your license):

   ```bash
   kubectl apply -f - <<EOF 
   apiVersion: v1
   kind: Secret
   metadata:
     name: eksa-license
     namespace: eksa-system
   stringData:
     license: "my-license-here"
   type: Opaque
   EOF
   ```
