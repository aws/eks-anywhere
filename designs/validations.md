# EKS-A API validations
## Problem statement

The current state of the CLI flow is undesirable. Validations and defaults are invoked from multiple places and at different stages with some validations contain hidden dependencies. This creates an inconsistent user experience (error logging, etc.) and, more importantly, a significant maintenance issue. Given a big part of EKS-A's added value is its validations, this topic is of tremendous importance for the project.

There have been previous efforts on this area, however, they only focused on the **type** of validation and **where** to put the validation code, not about **when** or **how** to run them. Although this set the foundations for a better separation of validations between webhooks and controllers, it failed to solve the CLI inconsistency and coupling problems. Moreover, it didn't align well with the team's intent of keeping provider specific validations in their own package.

This document attempts to solve this issue by proposing a wholistic approach for defaulting and validating both CLI and API operations.

### Previous work

https://github.com/aws/eks-anywhere/pull/1343

> ## Validations
> 
> There are two types of validations:
> - Context agnostic validations: there are no external dependencies, having the API object is enough to validate it: mutually exclusive fields, string format, empty fields…
>    - These should live in the pkg/api/version package
 >   - These should be the ones called in the validation webhook
> - Context aware validations: these are dependent on other entities like clients or even other API objects
>   - These could leave in different places, I’m not making a decision right now: providers, validations package… This is probably worth a separate discussion.
>   - Try to keep these modular and reusable (eg. Don’t make them part of the providers struct’s api)
>   - If they only need other API objects from the same package, they can live in the api package
>   - Controller: these shouldn’t be run in a webhook and must be part of the main reconciliation loop

### Current state
Validations have grown organically over time. However, the natural evolution of requirements and usecases without taking the time to refactor previous implementations, has resulted in an incohesive array of frameworks, ideas and patterns. This doc doesn't intend to collect or represent them all in detail, but let this list be an example of the lack of cohesion and variety of usecases as well as a data to drive the solution:
* `pkg/validations`
	* It holds validation of many different domains (feature gates, certificates, docker daemon, local disk files, command line args, etc.), all implemented as standalone functions.
	* It also offers some validation "utils/framework", like a validation runner, reporting, error types, etc.
	* The validations are imported from validation compositions as well as directly from the commands in `cmd`.
* `pkg/validations/createvalidations`
	* This implements both *cluster create* validations (as standalone functions) as well as a "validation composition" that allows to build and run "preflight" validations for the *create* operation.
	* This uses `validations.Validation` and `validations.ValidationResult` but not the `validations.Runner`.
	* The compiled validations (`createvalidations.CreateValidations`) is used in the `cluster create` command, injected as a dependency to the corresponding workflow and called from the `SetAndValidateTask` task (more on this later).
* `pkg/validations/upgradevalidations`
	* This implements both *cluster upgrade* validations (as standalone functions) as well as a "validation composition" that allows to build and run "preflight" validations for the *upgrade* operation.
	* Strangely, this uses the `validations.ValidationResult` but not the `validations.Runner` and works slightly different than  `CreateValidations`.
	* The compiled validations (`upgradevalidations.UpgradeValidations`) is used in the `cluster upgrade` command, injected as a dependency to the corresponding workflow and called from the `setupAndValidateTasks` task (more on this later).
* `pkg/validations/createcluster`
	* This is a validations composition. It attempts to group all validations for `create cluster` (the ones in `pkg/createcluster`, provider validations, docker executable validations, kubeconfig path, `cluster.Config` validations and Flux validations).
	* It uses `validations.Runner` to register, run the validations sequentially and report the validations result with `validations.ValidationResult`.
	* It's a step in the good direction in terms of pattern and semantics vs other previous ideas but it might still be limited in scope.
* `pkg/api/v1alpha1`
	* Most eks-a API objects implement the `Validate` method, which is supposed to run only static data validations (or "context unaware"). Unfortunately most of these methods return at the first error as opposed to run them all and aggregate the errors.
	* Most eks-a API objects implement the `SetDefaults` method.
	* Most webhooks are also implemented in this package. They are composed of a combination of `object.Validate` calls and field immutability checks (for `ValidateUpgrade`). For defaults, they call `SetDefaults`.
* `pkg/cluster`
	* This package exports `cluster.Config` and allows to set defaults and validate it with `cluster.SetDefaults` and `cluster.Validate`. These internally call a `ConfigManager`, which allows to register logic dynamically to parse from yaml, set defaults and validate a `cluster.Config`.
	* Validations in the package either call `.Validate` in the API objects or implement static validations but at the `Config` level (eg. validations that need multiple API objects).
	* `ConfigManager`'s defaults also call the `SetDefaults` method in API objects.
* `pkg/gitops/flux`
	* Exports validations through `Validations` method, which returns a slice of `validations.Validation`, hence designed to be used in conjunction with `validations.Runner`, etc. These are run from the create cluster workflow.
* `pkg/providers`
	* All providers implement `SetupAndValidateCreateCluster` and `SetupAndValidateUpgradeCluster`. They are called from the *create* and *upgrade* workflows.
	* These methods mix API defaults, validations and internal state changes (equivalent to an `Init` method).
	* The internal implementation of these method variates from provider to provider. Some of them have a `Validator` and a `Defaulter` and Snow, for example, uses the `cluster.ConfigManager` to organize defaults and validations.
* `pkg/workflows`
	* Compose and run validations and defaults with `validations.Runner` from a workflow task.
* `test/framework`
	* `ClusterValidator` is yet another validator runner. This is one centered in validating the state of a cluster against a given cluster specification by using a k8s client to inspect the cluster API.
* Special note is due for commands in `cmd`, since these are the entities with more knowledge and they are a good example of the fragmentation addressed in this doc. Currently, they handle different types of validations and at different stages:
	* Independently run validations, like:
		* `FileExists` for the cluster config yaml.
		* Parsing and Cluster config API validation (`GetAndValidateClusterConfig`) 
		* Tinkerbell CLI flags.
		* Registry mirror auth.
		* Validations to `cluster.Config` with `cluster.Validate`. Defaults are applied by just calling `cluster.NewSpecFromClusterConfig`.
	* "Grouped" validations in `commonValidations`
		* Docker daemon config
		* File exists, parsing and Cluster config API validation (`GetAndValidateClusterConfig`). This is duplicated.
		* Kubeconfig file and format validation.
	* Delegated validations to the workflows through `interfaces.Validator`.

## Goals
* Define a standard for types of validation and defaults.
* Define where validations and defaults should be defined within the code base.
* Define a standard for when and where to run validations and defaults for both CLI commands and API requests.
* Define what validations and defaults should and should not do (scope).
* Propose a mechanism to run defaults from a single place.
	* This needs to allow for defaults defined in different packages.
* Propose a mechanism to run validations from a single place.
	* This needs to allow for validations defined in different packages.
	* This should allow to run all validations and collect all errors, as opposed to return after the first error.
* Propose an execution plan to move all existing validations to such model.

## Proposal

### Kinds
Validations and defaults can be classified in:
* Static data: they only require pure API objects and have no other dependencies.
* *Context aware*: they need to read information from the context (env vars, files, external APIs, etc.). Generally the depend on clients/adapters and a `context.Context`.

Although this doc focuses on API validations, it's worth mentioning *CLI input* validation. This entails command arguments, flags and env vars. In general, these should be fast and have no dependencies. However, it's possible config files (provided through a path) might need to be parsed in order to run conditional validations (for example, depending on the provider specified in the cluster config, extra flags or env vars might be needed).

### Implementation (*where*)
Static data validations/defaults that only require one single API object **should** live in the `Validate` and `SetDefaults` api struct methods, thus in the `pkg/api/v1**` package. (*Note:* I'm more than open to relax this constraint if there is a good reason to, but given this is one of the only things that seem to resemble some kind consistency with regards to this whole topic, it seems worth to maintain it).

Immutability validations have been treated differently (for some unknown reason) in the CLI than in webhooks. These checks **should** be implemented in the `api` package and **should not** be duplicated in two places. If the same checks are needed in both webhook and CLI, implement as standalone functions and reuse in both places. In the CLI, just read the object with a client and pass both old and new like the webhook does.

The rest of validations and defaults (static data for a group of API structs and the context aware ones) **can** live anywhere, as long the dependency directions make sense, with a few constraints:
* Context aware validation **cannot** live in `pkg/api/v1**`. This is to keep this package dependencies at a minimum.
* **Avoid** packages like `validations`. Don't group validations for what they are, group them by what they validate (their context). For example, for docker validations, don't use `validations.ValidateDockerDesktopVersion`, think about `docker.ValidateDesktopVersion` in `pkg/docker`.
* **Avoid** the `cmd` package. These tend to end up being untested and lack structure.

Webhooks that only require static data validations/defaults **should** live in the `pkg/api/v1**` (where all present webhooks live today). If they have any other dependency the **cannot** live in the `api` package. (*Note*: we could possibly define a `webhooks` folder elsewhere if we think that could avoid problems in the future, leaving this up for discussion)

### Scope (*what*)
#### Defaults
* Defaults should not **modify** data provided by the user, they can only **add** information. Any transformation needed should be moved to runtime operations and never get reflected back on the API objects.
* They should have not secondary effects on anything else except the provided API object/s, including the internal state of the entity running the validation.
* API objects should only be modified by the defaulting logic. This guarantees they are invariants during the CLI execution, allowing for workflow reentry/re-execution (think command fails and on the next execution we continue from where it stopped).

#### Validations
* They must not have secondary effects on such as modifying the object under validation.
* API struct `Validate` methods should return a [`field.ErrorList`](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/validation/field#ErrorList). This facilitates using it from webhooks and ensures the API user facing error is as complete as possible.

### Calling (*when*)
#### Defaults
 * CLI
	* At the beginning of the command flow, in most cases before any other business logic.
	* Always before validations.
	* Defaults will be grouped into a single usecase per CLI flow: one for create cluster, another one for upgrade, etc. The individual defaulting logic can be reused by importing it and composing it into multiple usecases. These will live in `pkg/cli`, and be named after its flow: `CreateClusterDefaulter`, `UpgradeClusterDefaulter`, etc. 
	* This doc doesn't prescribe which entity should trigger the defaults (command in `cmd` package, workflow, etc.). But they do need to be run before the API objects get injected in any other entity as a dependency (think *service* objects). Given the approach the provider's refactor project has taken to inject the `cluster.Spec` and any other non dynamic data (this is, known at the beginning of the command execution) as state in the service objects, `cmd`  seems like the obvious place to do this since this is where service objects are constructed.
* API
	* Webhook: static data defaults (only depend on the given object) and sometimes, defaults that only require kubernetes API calls (to the same cluster where the object lives). Remember that only the given object should be altered from the webhook.
	* Controllers: any other defaults requiring extra information, calls to external systems or simply *too slow* for a webhook.

#### Validations
*  CLI
	* At the beginning of the command flow, in most cases before any other business logic.
	* Always after defaults.
	* Validations will be grouped into a single usecase per CLI flow: one for create cluster, another one for upgrade, etc. The individual validation logic can be reused by importing it and composing it into multiple usecases. These will live in `pkg/cli`, and be named after its flow: `CreateClusterValidator`, `UpgradeClusterValidator`, etc. 
	* This doc does not prescribe the entity invoking validations (command in `cmd` package, workflow, etc.). The decision is left for the provider refactor team to make later.
* API
	* Webhook: static data validations (only depend on the given object) and sometimes, validations that only require kubernetes API calls (to the same cluster where the object lives).
		* When using a client to read from the API server, be very careful. Webhooks are supposed to be very fast, remember they are synchronous.
			* Only run *efficient* queries (avoid full scans) and always think about the number of results you expect to get (depending on the type, the difference can be of orders of magnitude).
			* If your design performance is not very obvious, benchmark it.
	* Controllers: any other validations requiring extra information, calls to external systems or simply *too slow* for a webhook.

### Composing validations
This section describes the validation framework and how we will be using it. Requirements:
* Allow to compose and run validations implemented in different places.
* Make implementing validations as easy as possible.
* Design mainly for the CLI use-case. Optimize defaults for this scenario and possibly allow custom configuration for others.

This approach is similar to `cluster.ConfigManager`, and some other approaches, in that it takes a two step process where validations are registered before they are run.

```go
// Runner allows to compose and run validations.
type Runner struct
```

```go
runner := validation.NewRunner()
runner.Register(myValidation)
runner.Register(otherValidations...)

err := runner.RunAll(ctx, spec)
```

#### Validation signature
Requirements:
* They might need to contact external systems. In that case they should make use of a `context.Context`.
* They might need access to the full specification of an EKS-A cluster. In that case, they should use `cluster.Spec`.
* Validations should be able to, at least, return an error if the input is not valid.

This boils down to take a context and a cluster spec as input and return some type of error as output. We could implement it with functions or an interface, but given that we want to make registering validations as easy as possible and keep the framework API surface as simple as possible, we opt out for a function:

```go
type Validation func(ctx context.Context, spec *cluster.Spec) error
```

As seen above, the validation `Runner` allows to register these `Validation` funtions independtly. And as a separate step run them all, returning an aggregation of all errors.

#### Composability
We want the framework to allow, not only to register validations implemented in different places, but also validations that are composed of multiple validations. This, although it introduces some complexity, it offers a lot of flexibility to registrars to aggregate validations in different ways and it facilitates reusing existing validations (that tend to be already aggregated in one big method).

##### Aggregating errors
For validations to be registered as one handler but still be considered as different in the runner result, they should return an `errors.Aggregate`. The framework will understand this and will flatten the result. This allows for multiple levels of nesting.

##### Concurrency
Most *preflight* validations we currently run in the CLI can be run concurrently as they are not dependent on each other. For validations that make external calls this can be slow, concurrently executing validations may provide a speed improvement.

One could argue that, given validations are composable, concurrency can be left for registrars/validation implementers. However, in the past this approach has led to concurrency never been implemented.

We will take advantage of validations being composable and will make the default model concurrent. All handlers registered through `Register` will run concurrently.

If there is a subset of validations that need to be sequential, we will compose them into one handler that iterates and aggregates the results via a `Sequentially` helper.

```go
// Sequentially composes a set of validations into one which will run them sequentially and in order.
func Sequentially(validations ...Validation) Validation
```

It can be tempting to allow to compose concurrent validations inside a sequential collection. Although appealing, this could significantly increase the complexity of the runner. Since we can't find a place where this is currently needed, the framework will only run concurrently the first level validations. If that ever becomes a requirement, we will default to wrap another runner inside a handler. Given all returned aggregated errors are flatten, this should work without changes to the framework.

#### Logging
Some of the current attempts at a validation runner/composer log the success of validations. This requires providing a name for the validation. In addition, some of the provider validations log both success and failure of validations (using ✅  and ❌).

Trying to handle this at the framework level has several drawbacks:
* It requires naming validations, which increases the framework surface area, increases the entry level to register a validation and composability more difficult.
* It limits logging granularity when a validation function is composed of multiple sequential validations.
* It couples CLI output with validation functions expanding their scope of concern.

For these reasons, the framework won't provide any logging functionality.
	* For debug logging (we can generalize this to anything above verbosity `0`), validations should use their own logger. This should be injected at construction time as any other dependency. This still gives control to the program to decide which logs are presented to the user and when, by passing different loggers.
	* For information more tailored to feedback for the end user (like progress in slow validations), it will be left up to the framework callers to implement this. Based how logging is currently used to communicate this progress, the information is usually presented by "grouping" validations and not at the individual validation level. So the *usecases*, where validations are grouped and given semantic value, will be responsible to define these different groups and print logs in between them as necessary. This can be easily accomplished by composing multiple runners. Given a traditional logger might not be the best way to implement a presentation layer of this kind, this approach also allows to replace the logger with something else (in the future) without having to refactor any of the validation logic.

#### Validation error meta information
The existing `validations.ValidationResult` allows handlers to return a `Remediation` message. Given we want to just use the `error` interface to make it easy to register validations and that `Remediation` is empty for most `ValidationResult`s , we chose to make this optional. Validations wanting to provide a possible remediation for a failure, can return an `error` that also implements the `Remediable` interface.

```go
type Remediable type {
	Remediation() string
}
```

#### Retrying
Existing validation runners, like `framework.ClusterValidator`, incorporate retries into the runner itself. That might seem sensible, given a lot of validations might want to rely on retries to bypass transient errors. However, not only this would complicate the API to register validations but the retry logic will always be limited by the API itself (which is only addressable by making the API more complex).

In this case we would leave this out of the framework and let each validation handler decide how to deal with retries.

#### Framework implementation
Given the framework logic is very much independent from the object being validated (`cluster.Spec`), we propose to use generics so the same framework can be used for different types. The complexity added is minimal and even in the case where the framework grows in a direction where generics don't make sense, moving back to a concrete type would be trivial. There are already places in the codebase where we envision make use of this:

```go
// Validation is the logic for a validation of a type O.
type Validation[O any] func(ctx context.Context, obj O) error
```

In order to be able to prevent validations from modifying `obj` and detecting when if that happens, we add a constraint to `O` that allows us to `DeepCopy` the object so we can compare after running all validations:

```go
type Validatable[O] interface {
	DeepCopy() O
}

// Validation is the logic for a validation of a type O.
type Validation[O Validatable[O]] func(ctx context.Context, obj O) error
```

#### Framework usage and *use-cases*

This framework is the glue that will give form to the validation usecases. Example:

```go
type CreateClusterValidator struct {
	runner *validation.Runner
}

type Validation func(ctx context.Context, spec *cluster.Spec) (error)

type ProviderValidator interface {
	CreateClusterValidations() []Validation
}

func NewCreateClusterValidator(provider ProviderValidatonRegistrar) CreateClusterValidator {
	r := validation.NewRunner[*cluster.Spec]()
	r.Register(
		docker.ValidateMemory,
		cluster.ValidateConfig,
	)
	r.Register(provider.CreateClusterValidations()...)
	r.Register(gitops.ValidateRepositoryAccess)
}

func (v CreateClusterValidator) Run(ctx context.Context, spec *cluster.Spec) error {
	return v.runner.RunAll(ctx, spec)
}
```

### Composing defaults
#### Defaults signature
The requirements for defaults are quite similar to the ones for validations:
* Take a `context.Context`
* Based on `cluster.Spec`
* Be able to return an error if, for example, an external call fails.

In addition, and following good practices for functions that are supposed to update a given object:
* Don't rely on updating a given pointer value but take a `cluster.Spec` as input and return a `cluster.Spec` as output, with the necessary updates.

```go
type Default func(ctx context.Context, spec cluster.Spec) (cluster.Spec, error)
```

##### Problems with this approach

Defaults should only be set in the API objects configurable by the user. For example, the `Bundles` (which compiles all eks-a dependencies) shouldn't be modified. However, all of these are part of the `cluster.Spec`, giving defaulting logic the ability to modify them. We could enforce this by separating them and providing a `cluster.Config` (all user inputable API objects) and a `Bundles` (or something else that includes the rest of the cluster specification).

#### Framework implementation
Following the same logic as for validations, we propose the use of generics:

```go
// Default is the logic for a default for a type O. It should return a value of O
// wether it updates it or not. When there is an error, return the zero value of O
// and the error.
type Defaulter[O any] func(ctx context.Context, obj O) (O, error)
```

#### Concurrency
All defaults will be sequential given that they need to update the `cluster.Spec` and changes need to be propagated. However, defaults shouldn't rely on order. If such dependency exists between two or more defaults, then make them one single default.

Default handlers can choose to implement concurrency but they are responsible for handling data races.

#### Framework usage and *use-cases*

This framework is the glue that will give form to the default usecases. Example:

```go
type CreateClusterDefaulter struct {
	runner *defaulting.Runner
}

type Default func(ctx context.Context, spec *cluster.Spec) (*cluster.Spec, error)

type ProviderDefaulter interface {
	CreateClusterDefaults() []Default
}

func NewCreateClusterDefaulter(provider ProviderDefaulter) CreateClusterDefaulter {
	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		cluster.DefaultConfig,
	)
	r.Register(provider.CreateClusterDefaults()...)
	r.Register(gitops.DefaultRepositoryName)
}

func (v CreateClusterDefaulter) Run(ctx context.Context, spec *cluster.Spec) (*cluster.Spec, error) {
	return v.runner.RunAll(ctx, spec)
}
```

### Package structure
* The validation runner framework will go in `pkg/validation`.
* The defaults runner framework will go in `pkg/defaulting`.
* The composed validations and defaults for `create cluster` command will go in `pkg/cli`, implemented in `CreateClusterValidator` and `CreateClusterDefaulter`.
* The composed validations and defaults for `upgrade cluster` command will go in `pkg/cli` implemented in `UpgradeClusterValidator` and `UpgradeClusterDefaulter`.
* All validations in `pkg/validations`, `pkg/validations/createvalidations` and `pkg/validations/upgradevalidations` should be moved to their appropriate packages.
*  `pkg/validations/createcluster` is replaced by `cli.CreateClusterValidator`.
* Validations and defaults in `pkg/cluster` might stay there but won't be called by `cluster.Validate` and `cluster.SetDefaults`, which will be removed.
* Validation in `test/framework` should be probably be moved into proper packages. This will be low priority unless such validations need to be reused from the CLI controller.

### `errors` package
We propose the addition of our own `errors` package (in `pkg/errors`) that will implement the `Aggregate` error functionality. Although we will mirror the functionality from `apimachinery` so we don't reinvent the wheel, this will give us more flexibility to adapt to changes in Go error aggregation and handling (possibly go 1.20).

In addition, we plan on using this package to mirror the `Wrap` functionality from `github.com/pkg/errors`. This will allow us to only import one `errors` package from our code.

## Implementation plan
* Get this proposal approved
* Move validations and defaults out of workflows
* Split providers defaults and validations into two methods for both upgrade and delete.
* Update providers to return a slice of validations and defaults. The first iteration will probably be a wrapper around existing ones.
* Compose validations and defaults in `pkg/validation/*cluster` and `pkg/default/*cluster`
* Invoke the new aggregated defaults and validations that use the framework.
* Move validations from `pkg/validations/**` to the proper packages.
* Cleanup: delete unused framework/runners, `cluster.ConfigManager`, etc. This will probably need multiple rounds.
* Split providers validations and defaults into different Handlers that can be run independently.
* Update validations for `{apiobject}.Validate` to not return on first error and return a `field.ErrorList`.
