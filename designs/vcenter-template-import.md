# Import vcenter template before create

## Introduction

**Problem:** If a user doesn't have a template defined in the cluster spec, the create command 
imports the latest available OVA for the specific version of the CLI. However, there is no way
to import this OVA before running create without manually importing the OVA into a template
in vcenter and specifying that template name in the cluster spec, even though we have that 
automation logic in the CLI already. 

### Goals and Objectives

As an EKS Anywhere user:

* I want to have the ability to pre-import the template before I run create
* I want to specify the template I want to import in a simple manner
* I want the create command to know to use the pre-imported template


## Overview of Solution

With this feature, a user can specify to import the latest OVA template before create by
running a command. They would specify the cluster spec as input so that there is nothing 
new that they need to know in terms of flags or different configuration. We would be introducing new
subcommands as follows:

```
eksctl anywhere import vsphere template -f <cluster_spec.yaml>
```


### Solution Details

With this feature, users will be able to import the latest OVA available based on the Kubernetes version
and OS family defined in the cluster spec. We would introduce a new 
`import` subcommand. Since this feature is specific to vsphere, we will introduce a `vsphere` 
subcommand to give flexibility of the types of resources that we can import under 
vsphere. This also gives us flexibility to import provider-agnostic resources under the `import`
subcommand. Since this is importing the OVA into a vcenter template, we will be calling
the final subcommand `template` based on the resource name. 

Currently for the `create` subcommand, users can specify either a template if they already
manually imported the OVA into a template in vcenter, or not specify the template field or leave the
field empty if they want it imported via a generated name that is based on the Kubernetes
version, OS, and hash of the build. 

With this command, if a user has the template name defined, we will check to see if a template of that 
name already exists and the appropriate tags are set. If the template does not exist, we will import the template under
that name with the appropriate tags. If the template exists but the tags don't match the bundle 
that the CLI is configured with, we will error saying that the appropriate tags are not set, allowing
the user to decide if they want to manually fix the tags or delete the template first before running
the command again. 

If a user does not have the template name defined, then we will do what we already do today in the `create` command, 
which is importing the template under the name that we have predefined and setting the appropriate tags 
for the user. 

For both cases, we will have output saying that the template has been imported under the 
specific name set in vcenter. If the user did not have the template name defined, the `create`
cluster will still use the template that was imported before because the name would match.
It would only change if the user ran `import`, upgraded the CLI (or the bundle got updated), and then 
ran `create`.

### Alternatives

We could also solve this by not passing in the cluster spec file and just take in all the values
as flags, but having a consistent cluster spec driven approach for cluster related actions
will give the user a less confusing experience when trying to interact with all the commands.

We could also introduce a `setup` command instead to cover any other pre-create operations
we want to support, such as creating folders or anything else that might come up in the
future, but it would be better to target the solutions and have a `setup` command reuse
the actions from here when we see that it would be something that the users can benefit from.

### Side Effects

With these changes, the `import-images` command that we have currently feels out of place because
of how we generally structure the subcommands. With these changes, I also propose deprecating the 
`import-images` command in favor of having an `images` subcommand under `import`. This shows an 
example of why the `import` command is flexible here having provider-specific and 
provider-agnostic changes.

Current:
```
eksctl anywhere import-images -f <cluster_spec.yaml>
```

Proposed changes:
```
eksctl anywhere import images -f <cluster_spec.yaml>
```