# Rerun CLI commands after failure

## Introduction
**Problem:**
Most of the errors returned by the CLI cluster commands are either transient or fixable with manual intervention.
However, the CLI commands are not idempotent.
This makes impossible to rerun commands after they have already failed, even if the root cause of the issues has already been solved.
This makes the experience very disruptive, specially for the upgrade cluster command, which could leave the cluster in an irrecoverable state without the necessary manual cleanup.
This "cleanup" is not documented and it heavily dependents on the internal CLI's implementation.
For certain scenarios, it is not even possible, requiring users to destroy and recreate their clusters from scratch.

## Solution
The proposed solution is to implement a "checkpoint" capability.
This will serve as the first step to a larger story of improving the customer experience with upgrading clusters and troubleshooting in general.
The CLI will keep a registry of all completed steps for a workflow and the necessary data to restore its state.
If the `upgrade` command fails and the error is manually fixable, the user can simply rerun the same command after implementing the fix.
When a command is rerun after an error, the CLI will skip the completed steps, restoring the state of the program and proceeding with the upgrade from there.
The checkpoint data would be stored in the `<clusterName>/generated` folder as a file named `<clusterName>-checkpoint.yaml`.

Example checkpoint file:

```yaml
completedTasks:
  bootstrap-cluster-init:
    checkpoint: 
      ExistingManagement: false
      KubeconfigFile: test/generated/test.kind.kubeconfig
      Name: test
  ensure-etcd-capi-components-exist:
    checkpoint: null
  install-capi:
    checkpoint: null
  pause-controllers-reconcile:
    checkpoint: null
  setup-and-validate:
    checkpoint: null
  update-secrets:
    checkpoint: null
  upgrade-core-components:
    checkpoint:
      components: []
  upgrade-needed:
    checkpoint: null
```

---
We will add two methods, `Restore()` and `Checkpoint()`.
The logic in both methods will vary depending on what is needed from each task.

```go
type Task interface {
  Run(ctx context.Context, commandContext *CommandContext) Task
  Name() string
  Checkpoint() TaskCheckpoint
  Restore(ctx context.Context, commandContext *CommandContext, checkpoint TaskCheckpoint) (Task, error)
}
```

The `Checkpoint()` method will return the information necessary to successfully restore the task if the operation fails.
This information will then be retained in the checkpoint file.

The `Restore()` method will use the information from the checkpoint file to restore the state.
Again, this will vary depending on what's needed for that specific task.
Once the restoration succeeds for a task, we will return the next task (essentially skipping the task that was restored).

 ---
When a task finishes with no errors, we will add it as a completed task in the checkpoint file with the necessary information (if any) to allow the command to rerun successfully.

```go
if commandContext.OriginalError == nil {
    // add task to completedTasks in checkpoint file
    checkpointInfo.taskCompleted(task.Name(), task.Checkpoint())
}
```

When the user re-runs the command, we will check if the task exists as a completed task in the checkpoint file.
If so, we will restore any necessary information from the checkpoint file and set the next task.

```go
for task != nil {
    if completedTask, ok := checkpointInfo.CompletedTasks[task.Name()]; ok {
        // Restore any necessary information from the checkpoint file & set task to the next task
        nextTask, err := task.Restore(ctx, commandContext, completedTask)
        ...
        continue
    }
    logger.V(4).Info("Task start", "task_name", task.Name())
    ...
    ...
}
```

### Other notable changes:
* If the CLI fails to install CAPI objects to the bootstrap cluster, the CLI currently deletes the bootstrap cluster.
  This will obviously cause problems if we were to rerun the command at that point.
  Instead of deleting the bootstrap cluster if `InstallCAPITask` fails, it will proceed to `CollectDiagnosticsTask`.
  From there, the user can re-run the command to attempt installing the CAPI components on the previously created bootstrap cluster.


* If the CLI fails when moving CAPI management, things start to get a bit tricky.
  We could move CAPI management back to the source cluster if it fails, but this operation can be risky & potentially leave the cluster in an undesirable state.
  For now, we will retry moving the CAPI objects if the task fails the first time.
  A deeper dive would be required to decide on a better way to roll back those changes if moving CAPI fails.
  An idea for a future improvement is to backup the CAPI objects associated with the move using the `clusterctl backup` command. See [here](https://github.com/kubernetes-sigs/cluster-api/blob/main/cmd/clusterctl/cmd/backup.go).


* The only task that we should always run is the `setup-and-validate` task.
  Skipping this task would leave a blind spot for misconfigurations from the user input.

### Future Work

* Rolling back changes if moving CAPI management fails
* Add more validations to ensure each task's purpose was successful & left the cluster in a healthy state before skipping
* Building on top of this framework to further improve user experience