package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// Task is a logical unit of work - meant to be implemented by each Task.
type Task interface {
	Run(ctx context.Context, commandContext *CommandContext) Task
	Name() string
	Checkpoint() *CompletedTask
	Restore(ctx context.Context, commandContext *CommandContext, completedTask *CompletedTask) (Task, error)
}

// Command context maintains the mutable and shared entities.
type CommandContext struct {
	ClientFactory         interfaces.ClientFactory
	Bootstrapper          interfaces.Bootstrapper
	Provider              providers.Provider
	ClusterManager        interfaces.ClusterManager
	GitOpsManager         interfaces.GitOpsManager
	Validations           interfaces.Validator
	Writer                filewriter.FileWriter
	EksdInstaller         interfaces.EksdInstaller
	PackageInstaller      interfaces.PackageInstaller
	EksdUpgrader          interfaces.EksdUpgrader
	ClusterUpgrader       interfaces.ClusterUpgrader
	ClusterCreator        interfaces.ClusterCreator
	CAPIManager           interfaces.CAPIManager
	ClusterSpec           *cluster.Spec
	CurrentClusterSpec    *cluster.Spec
	UpgradeChangeDiff     *types.ChangeDiff
	BootstrapCluster      *types.Cluster
	ManagementCluster     *types.Cluster
	WorkloadCluster       *types.Cluster
	Profiler              *Profiler
	OriginalError         error
	BackupClusterStateDir string
	ForceCleanup          bool
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

// profiler for a Task.
func (pp *Profiler) SetStartTask(taskName string) {
	pp.SetStart(taskName, taskName)
}

// this can be used to profile sub tasks.
func (pp *Profiler) SetStart(taskName string, msg string) {
	if _, ok := pp.starts[taskName]; !ok {
		pp.starts[taskName] = map[string]time.Time{}
	}
	pp.starts[taskName][msg] = time.Now()
}

// needs to be called after setStart.
func (pp *Profiler) MarkDoneTask(taskName string) {
	pp.MarkDone(taskName, taskName)
}

// this can be used to profile sub tasks.
func (pp *Profiler) MarkDone(taskName string, msg string) {
	if _, ok := pp.metrics[taskName]; !ok {
		pp.metrics[taskName] = map[string]time.Duration{}
	}
	if start, ok := pp.starts[taskName][msg]; ok {
		pp.metrics[taskName][msg] = time.Since(start)
	}
}

// get Metrics.
func (pp *Profiler) Metrics() map[string]map[string]time.Duration {
	return pp.metrics
}

// debug logs for task metric.
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

// Manages Task execution.
type taskRunner struct {
	task           Task
	writer         filewriter.FileWriter
	withCheckpoint bool
}

type TaskRunnerOpt func(*taskRunner)

func WithCheckpointFile() TaskRunnerOpt {
	return func(t *taskRunner) {
		logger.V(4).Info("Checkpoint feature enabled")
		t.withCheckpoint = true
	}
}

func (tr *taskRunner) RunTask(ctx context.Context, commandContext *CommandContext) error {
	checkpointFileName := fmt.Sprintf("%s-checkpoint.yaml", commandContext.ClusterSpec.Cluster.Name)
	var checkpointInfo CheckpointInfo
	var err error

	commandContext.BackupClusterStateDir = fmt.Sprintf("%s-backup-%s", commandContext.ClusterSpec.Cluster.Name, time.Now().Format("2006-01-02T15_04_05"))
	commandContext.Profiler = &Profiler{
		metrics: make(map[string]map[string]time.Duration),
		starts:  make(map[string]map[string]time.Time),
	}
	task := tr.task
	start := time.Now()
	defer taskRunnerFinalBlock(start)

	checkpointInfo, err = tr.setupCheckpointInfo(commandContext, checkpointFileName)
	if err != nil {
		return err
	}

	for task != nil {
		if completedTask, ok := checkpointInfo.CompletedTasks[task.Name()]; ok {
			logger.V(4).Info("Restoring task", "task_name", task.Name())
			nextTask, err := task.Restore(ctx, commandContext, completedTask)
			if err != nil {
				return fmt.Errorf("restoring checkpoint info: %v", err)
			}
			task = nextTask
			continue
		}
		logger.V(4).Info("Task start", "task_name", task.Name())
		commandContext.Profiler.SetStartTask(task.Name())
		nextTask := task.Run(ctx, commandContext)
		commandContext.Profiler.MarkDoneTask(task.Name())
		commandContext.Profiler.logProfileSummary(task.Name())
		if commandContext.OriginalError == nil {
			checkpointInfo.taskCompleted(task.Name(), task.Checkpoint())
		}
		task = nextTask
	}
	if commandContext.OriginalError != nil {
		if err := tr.saveCheckpoint(checkpointInfo, checkpointFileName); err != nil {
			return err
		}
	}
	return commandContext.OriginalError
}

func taskRunnerFinalBlock(startTime time.Time) {
	logger.V(4).Info("Tasks completed", "duration", time.Since(startTime))
}

func NewTaskRunner(task Task, writer filewriter.FileWriter, opts ...TaskRunnerOpt) *taskRunner {
	t := &taskRunner{
		task:   task,
		writer: writer,
	}

	for _, o := range opts {
		o(t)
	}
	return t
}

func (tr *taskRunner) saveCheckpoint(checkpointInfo CheckpointInfo, filename string) error {
	logger.V(4).Info("Saving checkpoint", "file", filename)
	content, err := yaml.Marshal(checkpointInfo)
	if err != nil {
		return fmt.Errorf("saving task runner checkpoint: %v\n", err)
	}

	if _, err = tr.writer.Write(filename, content); err != nil {
		return fmt.Errorf("saving task runner checkpoint: %v\n", err)
	}
	return nil
}

func (tr *taskRunner) setupCheckpointInfo(commandContext *CommandContext, checkpointFileName string) (CheckpointInfo, error) {
	checkpointInfo := newCheckpointInfo()
	if tr.withCheckpoint {
		checkpointFilePath := filepath.Join(commandContext.Writer.TempDir(), checkpointFileName)
		if _, err := os.Stat(checkpointFilePath); err == nil {
			checkpointFile, err := readCheckpointFile(checkpointFilePath)
			if err != nil {
				return checkpointInfo, err
			}
			checkpointInfo.CompletedTasks = checkpointFile.CompletedTasks
		}
	}
	return checkpointInfo, nil
}

type TaskCheckpoint interface{}

type CheckpointInfo struct {
	CompletedTasks map[string]*CompletedTask `json:"completedTasks"`
}

type CompletedTask struct {
	Checkpoint TaskCheckpoint `json:"checkpoint"`
}

func newCheckpointInfo() CheckpointInfo {
	return CheckpointInfo{
		CompletedTasks: make(map[string]*CompletedTask),
	}
}

func (c CheckpointInfo) taskCompleted(name string, completedTask *CompletedTask) {
	c.CompletedTasks[name] = completedTask
}

func readCheckpointFile(file string) (*CheckpointInfo, error) {
	logger.V(4).Info("Reading checkpoint", "file", file)
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed reading checkpoint file: %v\n", err)
	}
	checkpointInfo := &CheckpointInfo{}
	err = yaml.Unmarshal(content, checkpointInfo)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling checkpoint: %v\n", err)
	}

	return checkpointInfo, nil
}

/*
	UnmarshalTaskCheckpoint marshals the received task checkpoint (type interface{}) then unmarshalls it into the desired type

specified in the Restore() method.
When reading from a yaml file, there isn't a direct way in Go to do a type conversion from interface{} to the desired type.
We use interface{} because the TaskCheckpoint type will vary depending on what's needed for a specific task. The known workaround
for this is to marshal & unmarshal it into the checkpoint type.
*/
func UnmarshalTaskCheckpoint(taskCheckpoint TaskCheckpoint, config TaskCheckpoint) error {
	checkpointYaml, err := yaml.Marshal(taskCheckpoint)
	if err != nil {
		return nil
	}
	return yaml.Unmarshal(checkpointYaml, config)
}
