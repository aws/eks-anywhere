package task

import (
	"context"
	"time"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Task is a logical unit of work - meant to be implemented by each Task
type Task interface {
	Run(ctx context.Context, commandContext *CommandContext) Task
	Name() string
	Checkpoint() *CompletedTask
	Restore(ctx context.Context, commandContext *CommandContext, completedTask *CompletedTask) (Task, error)
}

// Command context maintains the mutable and shared entities
type CommandContext struct {
	Bootstrapper       interfaces.Bootstrapper
	Provider           providers.Provider
	ClusterManager     interfaces.ClusterManager
	AddonManager       interfaces.AddonManager
	Validations        interfaces.Validator
	Writer             filewriter.FileWriter
	EksdInstaller      interfaces.EksdInstaller
	PackageInstaller   interfaces.PackageInstaller
	EksdUpgrader       interfaces.EksdUpgrader
	CAPIManager        interfaces.CAPIManager
	ClusterSpec        *cluster.Spec
	CurrentClusterSpec *cluster.Spec
	UpgradeChangeDiff  *types.ChangeDiff
	BootstrapCluster   *types.Cluster
	ManagementCluster  *types.Cluster
	WorkloadCluster    *types.Cluster
	Profiler           *Profiler
	OriginalError      error
}

func (c *CommandContext) SetError(err error) {
	if c.OriginalError == nil {
		c.OriginalError = err
	}
}

type Profiler struct {
	metrics map[string]map[string]time.Duration
	starts  map[string]map[string]time.Time
}

// profiler for a Task
func (pp *Profiler) SetStartTask(taskName string) {
	pp.SetStart(taskName, taskName)
}

// this can be used to profile sub tasks
func (pp *Profiler) SetStart(taskName string, msg string) {
	if _, ok := pp.starts[taskName]; !ok {
		pp.starts[taskName] = map[string]time.Time{}
	}
	pp.starts[taskName][msg] = time.Now()
}

// needs to be called after setStart
func (pp *Profiler) MarkDoneTask(taskName string) {
	pp.MarkDone(taskName, taskName)
}

// this can be used to profile sub tasks
func (pp *Profiler) MarkDone(taskName string, msg string) {
	if _, ok := pp.metrics[taskName]; !ok {
		pp.metrics[taskName] = map[string]time.Duration{}
	}
	if start, ok := pp.starts[taskName][msg]; ok {
		pp.metrics[taskName][msg] = time.Since(start)
	}
}

// get Metrics
func (pp *Profiler) Metrics() map[string]map[string]time.Duration {
	return pp.metrics
}

// debug logs for task metric
func (pp *Profiler) logProfileSummary(taskName string) {
	if durationMap, ok := pp.metrics[taskName]; ok {
		for k, v := range durationMap {
			if k != taskName {
				logger.V(4).Info("Subtask finished", "task_name", taskName, "subtask_name", k, "duration", v)
			}
		}
		if totalTaskDuration, ok := durationMap[taskName]; ok {
			logger.V(4).Info("Task finished", "task_name", taskName, "duration", totalTaskDuration)
			logger.V(4).Info("----------------------------------")
		}
	}
}

// Manages Task execution
type taskRunner struct {
	task   Task
	writer filewriter.FileWriter
}

// executes Task
func (pr *taskRunner) RunTask(ctx context.Context, commandContext *CommandContext) error {
	commandContext.Profiler = &Profiler{
		metrics: make(map[string]map[string]time.Duration),
		starts:  make(map[string]map[string]time.Time),
	}
	task := pr.task
	start := time.Now()
	defer taskRunnerFinalBlock(start)
	for task != nil {
		logger.V(4).Info("Task start", "task_name", task.Name())
		commandContext.Profiler.SetStartTask(task.Name())
		nextTask := task.Run(ctx, commandContext)
		commandContext.Profiler.MarkDoneTask(task.Name())
		commandContext.Profiler.logProfileSummary(task.Name())
		task = nextTask
	}
	return commandContext.OriginalError
}

func taskRunnerFinalBlock(startTime time.Time) {
	logger.V(4).Info("Tasks completed", "duration", time.Since(startTime))
}

func NewTaskRunner(task Task, writer filewriter.FileWriter) *taskRunner {
	return &taskRunner{
		task:   task,
		writer: writer,
	}
}

type TaskCheckpoint interface{}

type CheckpointInfo struct {
	CompletedTasks map[string]*CompletedTask `json:"completedTasks"`
}

type CompletedTask struct {
	Checkpoint TaskCheckpoint `json:"checkpoint"`
}
