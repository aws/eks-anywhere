# Supporting Provisioner-set metrics-bind-addr for CAPC (and possibly other infrastructure provider) Managers

## Introduction

**Problem:**

A need has been expressed for the scraping of CAPC Manager metrics, primarily to
facilitate alerting on reconciliation failures. As do most KubeBuilder-based
controllers, the CAPC manager offers metrics on an interface/port defined by its
*--metrics-bind-addr* command-line parameter.  For security purposes This
parameter defaults to *localhost:8080*, making the port only visible to other
containers in the manager's pod (sidecars).  To expose this port from the pod
one needs to bind it to the pod's NIC interface, typically by setting it to
*0.0.0.0:8080* or just *:8080* (since the actual IP address that will be assigned to the pod is not easily known).

CAPC users can easily add this argument to an overridden infrastructure-components.yaml.
EKS-A users, however, have no available mechanism for doing so to the
infrastructure-components.yaml that EKS-A will use.

This design proposes a simple mechanism for supporting this.  The scope of the proposed change is CAPC,
though CAPV, CAPD and CAPT are known to support the same parameter (and other providers likely do too.

### Tenets

* ****Simple:**** simple to use, simple to understand, simple to maintain
* ****Minimally Impactful:**** to the EKS-A code base.
* ****Secure by Default:**** if unused defaults to the most secure option

### Goals and Objectives

As a Kubernetes administrator I want to be able to expose the CAPC Manager's
metrics port from the pod running it so I can scrape metrics from it.

### Statement of Scope

**In scope**

* Add support for overriding the default CAPC manager's metrics-bind-addr at
cluster provisioning time.**

**Not in scope**

* Doing this for all EKS-A providers.

**Future scope**

* Implement this for the other EKS-A providers.


## Overview of Solution

We propose to
* Modify the CAPC infrastructure-components.yaml to define argument *--metrics-bind-addr* for the CAPC Manager deployment, from an optional environment variable which defaults to the same setting as the CAPC manager uses for this parameter.

* Modify function EnvMap() in cloudstack.go, which sets up the environment that *clusterctl init* will run under, to read the value of a new environment
variable (if defined by the provisioner in their *eksctl anywhere create cluster* session) into the map that it creates.

## Solution Details

### CAPC infrastructure-components.yaml change

The deployment spec for the CAPC manager will be modified to define a *--metrics-bind-address*
for the deployment's pod template's manager container, resolved with an environment
variable substitution defaulting to the original secure default used by the manager:

```
apiVersion: apps/v1
kind: Deployment
metadata:
...
  name: capc-controller-manager
  namespace: capc-system
spec:
...
    spec:
      containers:
      - name: manager
        args:
        - --leader-elect
        - --metrics-bind-addr=${CAPC_MANAGER_METRICS_BIND_ADDR:-localhost:8080}
        command:
        - /manager
        image: localhost:5000/cluster-api-provider-cloudstack:v0.4.5
...
```

### EKS-A change to expose this environment variable, if defined in the provisioner's environment at the time of *eksctl anywhere cluster create*:

cloudstack.go:

#### Current:
```
var requiredEnvs = []string{decoder.CloudStackCloudConfigB64SecretKey}

...
func (p *cloudstackProvider) EnvMap(spec *cluster.Spec) (map[string]string, error) {

	var x = spec.CloudStackDatacenter.Name

	envMap := make(map[string]string)
	for _, key := range requiredEnvs {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			return envMap, fmt.Errorf("warning required env not set %s", key)
		}
	}
	return envMap, nil
}
```

#### Proposed:
```
var requiredEnvs = []string{decoder.CloudStackCloudConfigB64SecretKey}
var optionalEnvs = []string{decoder.CloudStackMetricsBindAddr}  // METRICS_BIND_ADDR

...
func (p *cloudstackProvider) EnvMap(spec *cluster.Spec) (map[string]string, error) {

	var x = spec.CloudStackDatacenter.Name

	envMap := make(map[string]string)
	for _, key := range requiredEnvs {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			return envMap, fmt.Errorf("warning required env not set %s", key)
		}
	}
    for _, key := range optionalEnvs {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		}
	}
	return envMap, nil
}
```

#### Alternative 1:

As an alternative to the provisioner passing the desired value for the metrics-bind-addr
through an environment variable, a new parameter for this could be added to the CloudStackDatacenter struct,
as an optional field defaulting to the secure value.

As shown above, the *config.spec* is available to EnvMap() as a param, so EnvMap() could obtain the value of this new parameter from it, placing it in the environment map it's creating for running clusterctl init.  The new parameter would not be used elsewhere
during provisioning.

This would persist the override in the EKS-A Cluster Config for future reference/audit.  It would modify the
CloudStack Datacenter struct, albeit in a non-breaking way.


### Backwards Compatibility

The proposed solution is completely backward compatible, as it makes no
non-defaulting interface changes.  When not specified EKS-A/CAPC provision
a cluster with the same metric-bind-addr as currently used.

The Alternative Solution 1 proposes a non-breaking interface change.

## User Experience


## Security

The parameter defaults to the secure option, exposing the port only within the pod.
The parameter can be overridden by the cluster provisioner, who is making numerous
security-impacting decisions about their clusters already.

## Testing

The new code will be covered by unit and e2e tests, and the e2e framework will be extended to support cluster creation across multiple Cloudstack API endpoints.

The following e2e test will be added:

* provision a Cloudstack cluster with metrics-bind-addr overridden.  Confirm that the resulting cluster's
CAPC manager pod is configured with this address.
