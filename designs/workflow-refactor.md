# Workflow Refactor

> **Status:** RFC

> **Authors:** Chris Doherty, Guillermo Gaston, Joey Wang, Eric Wollesen, Jacob Weinstock, Vivek Koppuru

## Glossary

* **Cluster Management Components:** deployable components required for provisioning and managing Kubernetes clusters.
* **Workflows:** orchestration logic for cluster management.
* **Provider:** an abstraction used by EKS-A to facilitate cluster provisioning.
* **Management Cluster:** a cluster containing the deployable components required for provisioning a Kubernetes cluster with a particular provider.
* **Bootstrap Cluster:** a temporary cluster hosted on an administrative machine that facilitates interactions with management clusters.
* **Workload Cluster:** a cluster used strictly for running customer workloads.
* **Hook:** a point during a workflow at which arbitrary functionality can be executed.

## Introduction

EKS Anywhere (EKS-A) has a desire to support numerous providers maintained by both AWS and third parties. Provider maintainers must satisfy the [provider interface](https://github.com/aws/eks-anywhere/blob/main/pkg/providers/provider.go#L12) to provide behavior required by various components throughout the code base. The expectations for provider behavior are difficult to rationalize for several reasons:

* Poorly documented methods.
* No clear relationship to other parts of the code base.
* Parameters that are bundled in packages with no clear scope (for example, `types`).
* Hard to pin down behavioral expectations.

The ambiguity invoked in the provider interface forces provider developers to reverse engineer the code base to establish expectations. This typically leads developers to need to understand 3 core components:

1. *Parsing, defaulting and validating* logic that occurs in various parts of the code base.
2. *Workflows* responsible for orchestrating the installation and upgrade of generic and provider specific cluster management components; orchestrating cluster create, upgrade and delete operations; and triggering the installation of non-cluster provisioning related activities such as AWS IAM Auth.
3. *Cluster Manager* that has a wide range of concerns including cluster create, upgrade and delete logic; network installation; AWS IAM management; OIDC management; GitOps management.

Reverse engineering is time consuming and can lead to incorrect assumptions that result in incorrectly implemented behavior.

### Tenets

1. **Clarity**: clear and concise expectations produce a better developer experience.
2. **Modular**: modular code that strives for single responsibility is easier to delete when requirements inevitably change.
3. **Complimentary**: engineers have put time and effort into writing designs and implementing code so new code must take into consideration in-flight work and concrete future work.

### Goals and objectives

* Reduce the reliance by provider developers on reverse engineering workflows to understand expected behavior.
* Disambiguate the installation and upgrade of provider components from cluster creation, upgrade and deletion.
* Reduce the scope of workflows to provider component management (installation and upgrade) and cluster management (create, upgrade and delete).
* Provide a mechanism to hook into and extend workflows with additional tasks during cluster creation, upgrade and deletion.
* Reduce the surface area of the provider interface.

### Statement of scope

#### In Scope

* Refactoring of workflow tasks to have a single responsibility.
* Refactoring of cluster management components such that behavior can be invoked atomically in the context of a task. Behavior not capable of being broken into atomic actions complementary to tasks should be clearly documented against tasks.
* Refactoring of parsing, defaulting and validating input specific to providers.
* Refactoring of provider interfaces to align with workflow expectations (will require adjustment of provider implementations to satisfy the interfaces).
* Documenting workflow tasks using native go docs so that provider developers have a clear understanding of the context in which they are executed.

#### Out of Scope

* Holistic refactoring of the cluster management component. The cluster management component has a hard to pin down scope of concerns that are heavily leveraged by workflows and it would take a broader effort to break them apart and organize that likely isn’t required for these proposed changes.
* A plug-able provider interface that can be used to develop providers in a separate repository to EKS-A (off-tree). Off-tree provider development requires much broader thought such as how our API should be structured to enable API extension, how to version the APIs and ensure they function together, and how to build and distribute artifacts specific to a given provider.
* Refactoring the mechanisms for workflow checkpoint-recovery. The interfaces for checkpoint-recovery interaction with workflows may require adjusting but the fundamentals of the design should not change.

## Proposal

The following sections describe smaller proposals that aim to better define concerns for workflows and providers, provide mechanisms for hooking into or extending workflows, and reduce cognitive load when understanding workflow execution. Some sections work together with others and some are more independent.

These proposals also invokes a fundamental change in how we think about providers. Specifically, providers have historically been modeled as data providers for centralized orchestration logic. The data provider approach is built on assumptions around how providers are composed with a desire to treat everything the same; this has since been proven false. This proposal seeks to change providers to take on greater functional responsibility such as installation of provider specific components to an environment. This sets into motion an expectation that providers will handle their specifics with an orchestration layer telling them when to run.

### Separation of Management and Workload workflows

We propose separating management and workload cluster workflows and constraining tasks to only those required for cluster management operations. Additional tasks supported by EKS-A will be moved to an extensible post-workflow hook that we can continually build on to add more features separate from the minimal cluster provisioning tasks.

Having implemented this change, there is an expectation that workflow maintainers follow a documented set of guidelines on the granularity of tasks and what tasks are appropriate for featuring explicitly in a workflow.

#### Management cluster workflows

Management cluster workflow will comprise of the following tasks.

##### Create

1. **Pre Create** executes pre workflow hooks.
2. **Create Bootstrap Cluster** produces a bootstrap cluster that is used to provision the permanent management cluster. It should populate the workflow context with information to contact the bootstrap cluster.
3. **Install Bootstrap Components** install CAPI components, EKS-A components, and provider specific components on a bootstrap cluster. Providers are responsible for installing the components they require.
4. **Create Management Cluster** creates the cluster by asking the provider to convert EKS-A APIs to APIs understood by cluster management components resulting in a functional cluster.
5. **Install Networking** install the network configuration to support cluster communication.
6. **Install Management Components** install CAPI components, EKS-A components, and provider specific components on a bootstrap cluster. Providers are responsible for installing the components they require.
7. **Pivot to Management** moves objects from the bootstrap cluster to the management cluster. This includes executing  `clusterctl move` .
8. **Install Cluster Configuration** installs the cluster configuration supplied by the user into the management cluster.
9. **Delete Bootstrap Cluster** removes the bootstrap cluster from the admin machine.
10. **Post Create** executes post workflow hooks.


Behavior currently invoked as part of management cluster create that will be changed includes:

* **Install Resources on Management** is used by the Tinkerbell provider only to install arbitrary resources on the the management cluster and will be invoked with a *Post Pivot To Management* hook.
* **Install Addon Manager** used to install GitOps will be moved to a post workflow activity instead of an explicit task.
* **Install Curated Packages** used to install components necessary for curated package installation will be invoked as a post workflow activity instead of an explicit task.
* **Write Cluster Config**writes the cluster configuration to non-volatile storage and will be featured as a post-workflow hook. This is to better align with the scope of workflows in that writing configuration has little to do with cluster provisioning.

##### Upgrade

1. **Pre Upgrade** executes pre workflow hooks.
2. **Upgrade Management Components** upgrades core components such as Cluster API and generic EKS-A components and invokes the provider to upgrade its components.
3. **Upgrade Networking** upgrades the networking implementation in the management cluster.
4. **Install Management Components** (see *Install Management Components* in *Create*).
5. **Create Bootstrap Cluster** (see *Create Bootstrap Cluster* in *Create*).
6. **Install Bootstrap Components** (see *Install Bootstrap Components* in *Create*).
7. **Pivot to Bootstrap** is the inverse of *Pivot to Management* (see *Create*).
8. **Upgrade Cluster** upgrades the management cluster’s by invoking the provider upgrade cluster behavior.
9. **Pivot to Management** (see *Pivot to Management* in *Create*).
10. **Delete Bootstrap Cluster** (see *Delete* *Bootstrap Components* in *Create*).
11. **Write Cluster Config** writes an updated cluster configuration file to disk. This is necessary for upgrade operations as we may upgrade the EKS-A cluster configuration objects and they need to be made available to the user.
12. **Post Upgrade** executes post workflow hooks.


Behavior currently invoked as part of management cluster upgrade that will be changed includes:

* **Setup and Validate** will execute prior to workflow construction. This compliments the desire to construct workflows before running them where it is necessary to ensure configuration is valid before executing decision logic.
* **Ensure Etcd CAPI Components Exist** is replaced by a generic *Install Management Components*. The *Install Management Components* task is expected to be idempotent.
* **Upgrade Core Components** is split into *Upgrade Management Components* and *Upgrade Networking.* Non-cluster provisioning specific items will be moved to hooks.
* EKSD upgrade may need splitting into its own task as noted under the *Create* workflow.
* **Pause EKSA and Flux** is currently invoked when an upgrade is needed as defined by providers. We need to determine how this fits into workflows, i.e. should it be a hook or a generic *Pause* and *Resume* task with special handling of some kind.

##### Delete

Deletion of a management cluster entails creation of the bootstrap cluster, pivoting of resources onto the bootstrap cluster and deletion of resources to invoke CAPI de-provisioning of nodes.

#### Workload cluster workflows

##### Create

1. **Pre Create** executes pre workflow hooks.
2. **Create Workload Cluster** creates the cluster by asking the provider to convert EKS-A APIs to APIs understood by cluster management components resulting in a functional cluster.
3. **Install Networking** install the network configuration to support cluster communication.
4. **Install EKSA Components** installs EKS-A types used to identify properties of the management cluster in subsequent operations.
5. **Post Create** executes post workflow hooks.

##### Upgrade

1. **Pre Upgrade** executes pre workflow hooks.
2. **Upgrade Networking** upgrades the networking implementation in the management cluster.
3. **Upgrade EKSA Components** upgrades EKS-A specific components installed during the create *Install EKSA Components* task.
4. **Upgrade Cluster** upgrades the management cluster’s by invoking the provider upgrade cluster behavior.
5. **Write Cluster Config** writes an updated cluster configuration file to disk. This is necessary for upgrade operations as we may upgrade the EKS-A cluster configuration objects and they need to be made available to the user.
6. **Post Upgrade** executes post workflow hooks.

##### Delete

Deletion of a workload cluster entails deletion of resources on the management cluster to invoke CAPI de-provisioning of nodes.

#### Impact

* Having separate workflows creates lower cognitive load as both workflow maintainers and provider developers have to hold less context mentally.
* The removal of tasks unrelated to producing a functional cluster further reduces cognitive load.
* It allows each type of workflow set to define the interface they expect from providers independently of other workflows. For example management workflows may require bootstrap related cluster management component installation methods that are only consumed by management workflows making it clear to a provider the context under which the method is called so they can make any necessary adjustments.
* A document and clear set of guidelines for workflow maintainers on what granularity of task and types of task should feature explicitly in a workflow.

### Separate workflow construction logic from task behavior

Tasks are currently written to know whether or not they should run at the time they are invoked. They then return the next task to be run creating a directed acyclic graph (DAG) of tasks.

Instead of having tasks dictate the ordering, we will consolidate what tasks run in construction logic as part of the `management.NewCreateWorkflow()` set of functions. This will yield a single location for understanding the conditions for running a task, separate the concern of constructing a set of tasks from the behavior of a task, reduce the need to pass information in the context between tasks, and enable explicit declaration of task dependencies that can be injected into the `New*()` funcs.

Furthermore, any future rollback capabilities can be easily introduced as a set of inverse functions for each task.

### Move the responsibility of provider installation to the provider

In the context of management cluster creation, there are 2 concerns provider implementers need to be aware of: (1) installation and upgrade of components and the upgrade of the Kubernetes cluster itself (including scaling). Currently the cluster manager installs the infrastructure. This is based on the assumption that providers only ever need to install a CAPI infrastructure provider. As we’ve introduced more providers we’ve observed a need to install more components.

We will remove the responsibility of installing all necessary components specific to a provider to providers. The installation will occur during the *Install Bootstrap Components* and *Install Management Components* outlined earlier in this document. Providers will be expected to install and upgrade their infrastructure provider along with any other components.

#### Impact

* A clearer delineation in responsibilities for installing provider specific cluster management components (for example the Tink stack).
* A clear opportunity for providers to install components they require for provisioning when creating or upgrading clusters.
* Removal of the `EnvMap` and, possibly, `Version()` methods on the provider interface. These methods, when taken at face value, are examples of ambiguous methods that require reverse engineering to understand how they’re used.

### Workflow hooks

There are a handful of behaviors on the provider interface that suggest they’re run pre or post a workflow task or activity within a task. The unclear expectations on pre/post tasks creates ambiguity and, consequently, hard to rationalize expectations for provider developers and hard for workflow maintainers to know where to instrument pre and post capabilities.

We will introduce hooks that provider implementers can use to execute arbitrary functionality around workflow tasks. Constructs needing to hook into workflows will receive a construct they can use to register handlers. Some hooks will require data unavailable until runtime such as a client configuration for contacting the bootstrap cluster API server. This will be made available via an injected context using well documented helper functions.

#### Example implementation from a hook implementers perspective

```go
package hook

// Handler is the signature of handler functions invoked before or after a
// particular task.
type Handler func(context.Context) error

// Binder defines methods for binding before or after a given task.
type Binder interface {
  // Before registers a Handler to be invoked before Task is executed.
  Before(Task, Handler) error

  // After registers a Handler to be invoked after a Task is executed.
  After(Task, Handler) error
}
```

```go
package foo

import (
  "github.com/aws/eks-anywhere/pkg/workflow/hook"
  "github.com/aws/eks-anywhere/pkg/workflow/management"
  "github.com/aws/eks-anywhere/pkg/workflow/workload"
)

// Foo is a data structure needing to hook into workflow runs.
type Foo struct {
     // Embed a nooping hook registrar to avoid empty method definitions
    // in providers.
    workload.NoopHookRegistrar
}

// BindCreateWorkflowHooks is called from the management.NewCreateWorkflow()
// func extending an opportunity to hook into a management cluster creation
// workflow
func (Foo) BindCreateWorkflowHooks(binder hook.Binder) {
   binder.Before(management.CreateBootstrapCluster, func(context.Context) error {
    // Run before create bootstrap cluster logic
  })

  binder.After(management.PostCreateWorkflow, func(context.Context) error {
    // Run after a create workflow finishes.
  })
}
```

#### Impact

* The removal of ambiguous pre-post behavior on the provider interface.
* A clear way to arbitrarily hook into workflow execution from a providers or any other component within the application.

### Contexts for each type of workflow (management and workload)

In the current implementation of workflows we pass 2 contexts between tasks. The first is a generic `context.Context` type and the second is a `task.Context`.

Generally, a code path should have a single context of execution. We will change to passing a single context between tasks thats specific to management and workload clusters and the operation under execution.

```go
package management

// CreateContext is an example context passed between tasks featuring in
// a management cluster creation workflow.
type CreateContext struct {
  context.Context

  // BootstrapCluster is a client configuration used to talk to the
  // API server of the bootstrap cluster.
  BootstrapCluster cluster.ClientConfig

  // Additional management cluster creation objects appropriate for
  // a context.

  ...
}
```

#### Impact

* Tasks and hooks can establish exactly what context they’re executed under and take the required action.
* A clear set of dependencies in a given context that provider developers can tap into.

### Introduce a workflow error handling channel

The majority of workflow tasks contain control flow that results in a separate diagnostics task being executed  when an error occurs. This convolutes tasks and adds to the cognitive load for maintainers.

We will change tasks to return an error. Should an error occur during task execution we will trigger a pre-configured error handling task/channel isolating the error handling from task implementation.

#### Impact

* Easier to understand error handling semantics for workflow maintainers.
* A clear error handling path during workflow execution.
* A hookable error handling process.

### Package organization

The following packages constitute the primary packages for workflow orchestration. Other packages may contribute to workflow orchestration but will likely have less relevance with respect to understandability of workflows.

#### `/pkg/workflow`

Renamed from `/pkg/workflows`. A package for any shared functions or types for both management and workload workflows such as the noop hook implementation. This package will contain root documentation detailing scope of concern and how subpackages are structured.

#### `/pkg/workflow/management`

Workflows, tasks and contexts for creating, upgrading and deleting management clusters.

#### `/pkg/workflow/workload`

Workflows, tasks and contexts for creating, upgrading and deleting workload clusters.

#### `/pkg/workflow/hook`

Generic hook related functionality leveraged by workflow creation functions.
