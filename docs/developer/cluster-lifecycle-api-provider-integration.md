# Cluster Lifecycle API provider integration guide

> `THIS IS A WORK IN PROGRESS`

## Context
> Note: if you haven't yet, reading the [original design doc](https://github.com/aws/eks-anywhere/blob/main/designs/full-cluster-lifecycle-api.md) is a must to execute this integration.

The main reconciliation logic is ran in the cluster controller. This watches all eks-a CRs and reacts to their changes. However, currently the cluster controller itself only performs some config/setup (controller watches, patch the `Cluster` CR at the end of the reconcile method, set owner refs, etc.) and reconciles deleted `Cluster`s. The main reconciliation logic is delegated to the `ProviderClusterReconciler`'s, one per provider.

`ProviderClusterReconciler` has to handle all the actions needed to reconcile the workload cluster to the state defined in the eks-a CRs. It has complete control over how this is implemented and there are no restrictions imposed in this regard. That said, there are 3 main tasks generic tasks that every `ProviderClusterReconciler` has to perform, in this order:
1. Reconcile the control plane to the eks-a CR specifications
1. Reconcile the workload cluster worker nodes to the eks-a CR specifications
1. Reconcile the CNI to the eks-a CR specifications

Each provider can add any other additional steps, in order: before, after or even in between of these 3.

The CNI reconciliation should be executed by using the Cilium reconciler (and other in the future), orchestrated by the main [CNI Reconciler](https://github.com/aws/eks-anywhere/blob/main/pkg/networking/reconciler/reconciler.go). These are designed to work in a reconciliation loop provided a stable Control Plane and a client to interact with.

## Requirements

* Implement the [`ProviderClusterReconciler` interface](https://github.com/aws/eks-anywhere/blob/main/pkg/controller/clusters/registry.go#L12)
* Build and register the new reconciler in the [controller factory](https://github.com/aws/eks-anywhere/blob/main/controllers/factory.go#L170)
* Add the necessary watches to the [cluster controller](https://github.com/aws/eks-anywhere/blob/main/controllers/cluster_controller.go#L54)
* Update [cluster controller RBAC](https://github.com/aws/eks-anywhere/blob/main/controllers/cluster_controller.go#L94) accordingly. You might need permissions to watch the provider eks-a CRDs or read/create/update/delete the provider infra CAPI CRDs.
* If they haven't been added yet, add the infra CAPX API for your provider to the [scheme](https://github.com/aws/eks-anywhere/blob/main/manager/main.go#L45).
* Add your provider specific API structs to [`cluster.Config`](https://github.com/aws/eks-anywhere/blob/main/pkg/cluster/config.go#L10). Add them as [child objects](https://github.com/aws/eks-anywhere/blob/main/pkg/cluster/config.go#L88) if necessary.
* Implement the proper [client processors](https://github.com/aws/eks-anywhere/blob/main/pkg/cluster/client_builder.go#L18) for your API structs and register them in the [default `ConfigClientBuilder`](https://github.com/aws/eks-anywhere/blob/main/pkg/cluster/config.go#L10)
* Add validation webhooks for all provider specific CRDs
* Add mutation webhooks for all provider specific CRDs when defaults are needed
* If context aware defaults or validations (those that require any kind of API or disk call) are needed, they should be implemented in a controller for that CRD ([example](https://github.com/aws/eks-anywhere/blob/main/controllers/snow_machineconfig_controller.go)). For any extra validations, the `ProviderClusterReconciler` will need to check if the spec of the eks-a provider CRs is valid before starting the reconciliation loop. In case they are not, bubble that error to the `Cluster.status`.

## Recommendations

### Generating CAPI specs from EKS-A objects

#### Templating vs API structs

Currently, most providers make use of `go` templates to generate `yaml` documents with all the CAPI objects based on a particular eks-a spec. This has resulted in quite complex templates, with a lot of duplication between providers. While it's possible to fix the duplication issue by using reusable template fragments, and reconciling yaml documents from the controller shouldn't be an issue (using for example [this](https://github.com/aws/eks-anywhere/blob/main/pkg/controller/serverside/reconcile.go#L14)), we recommend moving away from this and using `go` API structs directly.

A common set of generator helpers are available in the [`clusterapi`](https://github.com/aws/eks-anywhere/blob/main/pkg/clusterapi/apibuilder.go) package. We encourage all providers implementers to enrich the package with anything not provider specific so everyone can take advantage of it.

#### Immutable objects

The biggest challenge when generating CAPI specs is the generation of new objects for immutable CRDs. The current pattern the CLI implements is: retrieving the current state of eks-a CRs and based on the diff with the new desired spec and determine if such immutable objects would change after executing the conversion logic or not. This comparison is mostly and implicit version of the reverse transformation logic, one that tries to reverse the eks-a -> CAPI conversion logic.

Unfortunately this is not possible in the controller world since there is no old and new version of objects, only current. To solve this, the original cluster implements a reverse logic, trying to infer the past state of eks-a objects from the current state of CAPI objects. Then it plugs into the main CLI logic, running the previously mentioned comparison logic. This is extremely prone to errors and very hard to maintain.

We encourage all providers to follow the reverse algorithm: instead of comparing the old and new inputs and inferring if the outputs would differ, compare directly the old and new output. This is a way simpler process: retrieve the current status of the immutable CAPI object, run the conversion logic to generate that same object based on the current EKS-A spec and just compare them.

When used in combination with API structs, the comparison step becomes trivial most of the times, being able to take advantage of `equality.Semantic.DeepDerivative` provided by the `apimachinery` module. Examples of this can be found in the [Snow](https://github.com/aws/eks-anywhere/blob/main/pkg/providers/snow/objects.go) provider, which reuses the same code/logic for both the controller and the CLI.


## Development

### Tooling

The are some packages that can be useful when implementing a `ProviderClusterReconciler`. Their use is encouraged but not enforced.

* [`PhaseRunner`](https://github.com/aws/eks-anywhere/blob/main/pkg/controller/runner.go). Useful for running reconciliation logic that can divided in phases where each phase can abort the execution by either erroring out or requesting a reconcile request requeue (for example, to wait some time and so it check again for a particular event).
* [Helpers to retrieve CAPI objects](https://github.com/aws/eks-anywhere/blob/main/pkg/controller/clusterapi.go)
* [Helpers to use server-side apply](https://github.com/aws/eks-anywhere/tree/main/pkg/controller/serverside)
* [Helpers to extract data from CAPI objects](https://github.com/aws/eks-anywhere/blob/main/pkg/controller/clusters/clusterapi.go)
* [Build cluster.Spec from a client using `cluster.BuildSpec`](https://github.com/aws/eks-anywhere/blob/main/pkg/cluster/fetch.go#L159)
* [`ObjectApplier`](https://github.com/aws/eks-anywhere/blob/main/pkg/controller/serverside/applier.go#L16): helpful when generating the CAPI spec as API structs (`client.Object`s), allowing to easily reconcile them using server-side apply.
