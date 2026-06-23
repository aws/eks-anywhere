# Removing Unsupported etcd 3.5 Flags for Kubernetes 1.36

## Problem

Kubernetes 1.36 ships kubeadm that generates the etcd static pod manifest with flags
only supported by etcd 3.6+:

- `--feature-gates`
- `--watch-progress-notify-interval`

When using EKS Distro's etcd v3.5.x image with Kubernetes 1.36, the etcd pod crash-loops:

```
flag provided but not defined: -feature-gates
flag provided but not defined: -watch-progress-notify-interval
```

These flags cannot be suppressed through kubeadm's `extraArgs` configuration (extraArgs
can only add flags, not remove built-in ones).

## Execution Timeline

Understanding the CAPI/kubeadm execution timeline is critical:

```
1. preKubeadmCommands  (runs BEFORE kubeadm)
2. kubeadm init        (generates etcd manifest, starts etcd, WAITS for etcd health)
3. postKubeadmCommands (runs AFTER kubeadm init succeeds)
```

Since etcd crashes immediately with the unsupported flags, `kubeadm init` never completes
(times out after ~5 minutes), and `postKubeadmCommands` never executes.

**Key constraint: `postKubeadmCommands` cannot be used to fix this issue.**

## Approach Comparison

### Option A: Background Watcher in preKubeadmCommands (Chosen)

**How it works:**

A background process is spawned in `preKubeadmCommands` that polls for the etcd manifest
file. Once kubeadm writes it, the watcher immediately removes the unsupported flags via
`sed`. kubelet detects the manifest change and restarts etcd with the corrected command.

```yaml
preKubeadmCommands:
- |
  (while [ ! -f /etc/kubernetes/manifests/etcd.yaml ]; do sleep 0.5; done
  sed -i '/--feature-gates/d' /etc/kubernetes/manifests/etcd.yaml
  sed -i '/--watch-progress-notify-interval/d' /etc/kubernetes/manifests/etcd.yaml) &
```

**Behavior:**

1. preKubeadmCommands spawns background watcher, returns immediately
2. kubeadm starts, writes etcd manifest to `/etc/kubernetes/manifests/etcd.yaml`
3. kubelet may start etcd with bad flags (etcd crashes)
4. Background watcher detects file (within ~0.5s), removes flags
5. kubelet detects manifest change, restarts etcd with clean flags
6. etcd starts successfully, kubeadm health check passes
7. kubeadm init completes

**Stability analysis:**

- Race window: ~0.5s between manifest write and patch. etcd may crash-loop 1-2 times.
- kubeadm timeout: 5 minutes. Patch applies within <1s. Large safety margin.
- kubelet restart: kubelet watches manifest directory via inotify. Manifest change
  triggers pod recreation within seconds.
- Deterministic outcome: etcd will always recover before the timeout.

**Pros:**
- Simple implementation (3 lines of shell)
- No dependency on kubeadm internals or patch file format
- Works regardless of what flags kubeadm adds in the future (as long as the flag names
  match the sed patterns)
- Same pattern could be extended for other manifest patches

**Cons:**
- Brief etcd crash-loop (1-2 restarts, <5s total)
- Race condition exists (though inconsequential in practice)
- Background process could theoretically be killed before completing (extremely unlikely)

### Option B: kubeadm Patches Directory

**How it would work:**

kubeadm supports a `patches.directory` configuration that applies patches to generated
manifests before writing them to disk. Patch files follow the naming convention:

- `etcd0+strategic.yaml` — strategic merge patch
- `etcd0+merge.yaml` — JSON merge patch
- `etcd0+json.json` — JSON Patch (RFC 6902)

**Why it doesn't work for this case:**

1. **Strategic merge patch**: The Pod spec's `containers[].command` field uses "replace"
   merge strategy (atomic list). You cannot selectively remove items — you must provide
   the entire command array. The command includes runtime values (listen URLs with the
   node's IP, data directory paths, peer URLs) that are unknown at template render time.

2. **JSON merge patch**: Same limitation. Cannot selectively remove array elements.

3. **JSON Patch (RFC 6902)**: Could use `remove` operations on specific array indices.
   However, the indices of `--feature-gates` and `--watch-progress-notify-interval` in
   the command array depend on kubeadm's internal logic and could change between patch
   versions. This makes index-based removal fragile.

4. **Dynamic generation in preKubeadmCommands**: Could run
   `kubeadm init phase etcd local --dry-run` to discover the command, strip the flags,
   and write a patch file. However:
   - `kubeadm init phase etcd local --dry-run` may not be available or may have
     side effects depending on the kubeadm version
   - The generated patch must exactly match the target schema
   - Requires complex shell scripting to parse YAML, modify it, and write valid JSON/YAML

**Pros (if it could work):**
- No crash-loop at all — manifest is correct from the start
- kubelet never sees invalid flags
- Cleaner semantically (kubeadm's own mechanism)

**Cons:**
- Cannot selectively remove command array elements with strategic/JSON merge patches
- JSON Patch requires knowing array indices (fragile across versions)
- Dynamic generation adds complexity and kubeadm version dependencies
- Fundamentally incompatible with the "remove a flag" use case

### Option C: Use etcd 3.6 (PR #10813 approach for Kind cluster)

**How it works:**

For the Kind bootstrap cluster, use the upstream etcd 3.6 image that kubeadm 1.36
expects. The flags are valid for etcd 3.6, so no patching needed.

**Pros:**
- No workaround needed — uses the intended etcd version
- No crash-loop, no race conditions
- Forward-compatible with future kubeadm changes

**Cons:**
- Requires EKS Distro to ship etcd 3.6 for K8s 1.36 bundles
- May introduce behavioral changes between etcd 3.5 and 3.6
- Not viable if the requirement is specifically to use etcd 3.5

## Implementation

### Files Modified

- `pkg/providers/docker/config/template-cp.yaml` — Docker (CAPD) control plane template
- `pkg/providers/vsphere/config/template-cp.yaml` — vSphere control plane template
- `pkg/providers/tinkerbell/config/template-cp.yaml` — Tinkerbell control plane template
- `pkg/providers/nutanix/config/cp-template.yaml` — Nutanix control plane template

### Condition

The background watcher is only added for Kubernetes >= 1.36 (stacked etcd only, not
external etcd):

```yaml
{{- if (ge (atoi $kube_minor_version) 36) }}
```

For Bottlerocket format (vSphere), the condition is excluded since Bottlerocket manages
etcd differently.

### Scope

This patch only applies to **stacked etcd** (etcd running as a static pod on control
plane nodes). External etcd uses etcdadm which has its own flag management.

## Testing

The CP templates are embedded into the `eks-anywhere-cluster-controller` binary via
`//go:embed`. Simply running e2e tests against the source tree won't pick up template
changes — you must rebuild the controller image.

### Build and Deploy Custom Controller

1. **Build the controller image locally:**

```bash
make build-cluster-controller
```

This produces an OCI tarball at `/tmp/eks-anywhere-cluster-controller.tar`.

2. **Load and push to a test registry:**

```bash
docker load -i /tmp/eks-anywhere-cluster-controller.tar
docker tag <image_id> <your-registry>/eks-anywhere-cluster-controller:<your-tag>
docker push <your-registry>/eks-anywhere-cluster-controller:<your-tag>
```

3. **Update the local bundle manifest** to point to your custom controller image:

Edit `bin/local-bundle-release.yaml` and replace the controller image URI with your
custom build's URI and tag.

4. **Run the e2e test with bundles override:**

```bash
T_BUNDLES_OVERRIDE=true go test -v -timeout 60m \
  -run TestDockerKubernetes136SimpleFlow ./test/e2e/ -tags e2e
```

### Quick Validation (without full e2e rebuild)

If you have an existing management cluster and want to test incrementally:

1. Rebuild just the controller binary:
   ```bash
   make eks-a-cluster-controller
   ```

2. Rebuild and redeploy the controller pod in the management cluster with the new binary.

3. Create a K8s 1.36 workload cluster.

4. Validate on the control plane node (e.g., `docker exec` into the CAPD container):
   ```bash
   # Verify the unsupported flags are gone
   grep -E "feature-gates|watch-progress-notify" /etc/kubernetes/manifests/etcd.yaml
   # Should return nothing (exit code 1)

   # Verify etcd is running and healthy
   crictl ps | grep etcd
   # Should show etcd container in Running state without repeated restarts
   ```

### What to Look For

- **Success**: etcd pod starts without crash-loops, `kubectl get pods -n kube-system`
  shows etcd in `Running` state with 0 restarts (or 1-2 restarts if the race window
  was hit before the background patcher applied).
- **Failure indicators**:
  - etcd stuck in `CrashLoopBackOff` with `flag provided but not defined` in logs
  - kubeadm init timeout (5 minutes) — means the background watcher didn't patch in time
  - Background process not spawned — check `ps aux | grep etcd.yaml` on the node

### Note on Bootstrap (Kind) Cluster

This change only fixes the **workload cluster**. The Kind bootstrap cluster requires a
separate fix (using upstream etcd 3.6 in `kind.yaml`) — see PR #10813's `kind.yaml`
changes. Both fixes are needed together for a full K8s 1.36 Docker e2e flow.

## Future Considerations

- When EKS Distro ships etcd 3.6 for K8s 1.36, this workaround can be removed
- If kubeadm adds additional etcd 3.6-only flags in future versions, the sed patterns
  need to be updated
- The condition should be bounded (e.g., `(and (ge minor 36) (lt minor N))`) once etcd
  3.6 is adopted
