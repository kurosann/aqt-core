package task

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kurosann/aqt-core/safe"
)

type Status string

const (
	StatusPending     Status = "pending"
	StatusRunning     Status = "running"
	StatusCompleted   Status = "completed"
	StatusRollingBack Status = "rollingBack"
	StatusRolledBack  Status = "rolledBack"
	StatusFailed      Status = "failed"
	StatusTimeout     Status = "timeout"
	StatusCanceled    Status = "canceled"
)

type TaskInfo struct {
	Name string
	Task TaskWorker
}

func NewTaskInfo(name string, task TaskWorker) *TaskInfo {
	return &TaskInfo{Name: name, Task: task}
}

type TaskRecorder interface {
	// RecordTask 记录任务执行结果
	// taskID 为0时创建并生成任务ID,不允许失败
	RecordTask(taskID int, result TaskExecResult) (id int)
}
type TaskExecResult struct {
	Status          Status
	Order           int
	StepName        string
	StepID          int
	ProcessPosition int
	Err             error
	Payload         []byte
	Result          []byte
}
type Recorder interface {
	StepRecorder
	TaskRecorder
}

type TaskManager struct {
	TaskWorkers map[string]*TaskInfo
	mu          sync.RWMutex
	Recorder    Recorder
}

var _ Recorder = &DefaultRecorder{}

type DefaultRecorder struct {
}

func (d *DefaultRecorder) RecordTask(taskID int, result TaskExecResult) (id int) {
	return 0
}
func (d *DefaultRecorder) RecordStep(stepID int, result StepExecResult) (id int) {
	return 0
}

var (
	globalTaskManager *TaskManager
	once              sync.Once
)

func InitGlobalTaskManager(recorder Recorder) {
	once.Do(func() {
		globalTaskManager = NewTaskManager(recorder)
	})
}

func GetGlobalTaskManager() *TaskManager {
	return globalTaskManager
}

// NewTaskManager 创建任务管理器
func NewTaskManager(recorder Recorder) *TaskManager {
	if recorder == nil {
		recorder = &DefaultRecorder{}
	}
	return &TaskManager{
		TaskWorkers: make(map[string]*TaskInfo),
		Recorder:    recorder,
	}
}

// RegisterTask 注册任务
func (tm *TaskManager) RegisterTask(task *TaskInfo) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.TaskWorkers[task.Name] = task
}

// RunTask 执行任务
func (tm *TaskManager) RunTask(ctx context.Context, taskName string, payload []byte) (result []byte, err error) {
	tm.mu.RLock()
	task, ok := tm.TaskWorkers[taskName]
	defer tm.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("task %s not found", taskName)
	}
	steps := task.Task.Arrange(ctx, payload)
	taskWorker := &TaskWorkerProcessor{
		Steps:    steps,
		Recorder: tm.Recorder,
	}
	taskResult := taskWorker.Process(ctx, payload)
	if taskResult.Err != nil {
		return taskResult.Result, taskResult.Err
	}
	return taskResult.Result, nil
}

// TaskWorkerProcessor 任务执行器
type TaskWorkerProcessor struct {
	Steps    []StepInfo
	Recorder Recorder
}

// Process 执行任务
func (tp *TaskWorkerProcessor) Process(ctx context.Context, payload []byte) (result TaskExecResult) {
	taskID := tp.Recorder.RecordTask(0, TaskExecResult{Status: StatusRunning})

	// 记录已执行的步骤和它们的结果
	type executedStep struct {
		Step       StepInfo
		ExecResult StepExecResult
	}
	executedSteps := make([]executedStep, 0)
	stepResult := make([]byte, 0)
	defer func() {
		if result.Err != nil {
			rollbackErrs := make([]error, 0)
			// 如果执行失败，则逆向回滚已执行的步骤
			for i := len(executedSteps) - 1; i >= 0; i-- {
				step := executedSteps[i]
				tp.Record(step.Step, StepExecResult{Status: StatusRollingBack})
				// 全部执行过的步骤进行rollback
				if err := safe.TryE(func() error {
					return step.Step.StepWorker.Rollback(ctx, payload, step.ExecResult.Result)
				}); err != nil {
					rollbackErrs = append(rollbackErrs, fmt.Errorf("rollback %d failed: %w", i+1, err))
				}
				tp.Record(step.Step, StepExecResult{Status: StatusRolledBack, Err: step.ExecResult.Err, Result: step.ExecResult.Result})
			}
			tp.Recorder.RecordTask(taskID, TaskExecResult{Status: result.Status, Result: stepResult})
			result.Err = errors.Join(append([]error{result.Err}, rollbackErrs...)...)
		}
		tp.Recorder.RecordTask(taskID, result)
	}()

	for i, step := range tp.Steps {
		step.TaskID = taskID
		select {
		case <-ctx.Done():
			result.Err = context.Cause(ctx)
			result.Status = StatusTimeout
			return result
		default:
			step.Order = i + 1
			step.StepID, result = tp.Record(step, StepExecResult{Status: StatusRunning})

			// 使用channel来处理Execute的超时
			execDone := make(chan StepExecResult, 1)
			var stepExecResult StepExecResult

			wg := safe.NewWaitGroup()
			wg.Go(func() {
				defer close(execDone)
				execDone <- step.StepWorker.Execute(ctx, payload, stepResult)
			})

			// 等待执行完成或context取消
			select {
			case <-ctx.Done():
				stepExecResult = StepExecResult{
					Status: StatusTimeout,
					Err:    ctx.Err(),
				}
			case stepExecResult = <-execDone:
				if err := wg.WaitAndRecover(); err != nil {
					stepExecResult = StepExecResult{
						Status: StatusFailed,
						Err:    fmt.Errorf("步骤执行panic: %w", err),
					}
				}
				stepResult = stepExecResult.Result
			}

			if stepExecResult.Status == StatusFailed ||
				stepExecResult.Status == StatusTimeout {
				step.StepID, result = tp.Record(step, stepExecResult)
				result.Err = stepExecResult.Err
				result.Status = stepExecResult.Status
			}
			// 只有pending状态需要轮询
			if stepExecResult.Status == StatusPending {
				done := make(chan struct{})
				wg := safe.NewWaitGroup()
				wg.Go(func() {
					defer close(done)
					stepExecResult = tp.PollStatus(ctx, step, payload, stepResult)
					if stepExecResult.Result != nil {
						stepResult = stepExecResult.Result
					}
				})

				select {
				case <-ctx.Done():
					// done优先级高于ctx.Done()
					select {
					case <-done:
					default:
						result.Err = context.Cause(ctx)
						result.Status = StatusTimeout
					}
				case <-done:
					err := wg.WaitAndRecover()
					if err != nil {
						stepExecResult = StepExecResult{Status: StatusFailed, Err: err}
					}
				}
			}
			// 记录完整的执行结果
			executedSteps = append(executedSteps, executedStep{
				Step:       step,
				ExecResult: stepExecResult,
			})
			if result.Status == StatusFailed ||
				result.Status == StatusTimeout {
				return
			}
			step.StepID, result = tp.Record(step, stepExecResult)
			result.Result = stepResult
		}
	}
	return result
}

// PollStatus 轮询步骤状态
func (tp *TaskWorkerProcessor) PollStatus(ctx context.Context, step StepInfo, payload []byte, stepResult []byte) (result StepExecResult) {
	for {
		select {
		case <-ctx.Done():
			return StepExecResult{Status: StatusTimeout, Err: context.Cause(ctx)}
		default:
			var monitorResult StepPollStatusResult
			if err := safe.Try(func() {
				monitorResult = step.StepWorker.PollStatus(ctx, payload, stepResult)
			}); err != nil {
				monitorResult = StepPollStatusResult{
					Status: StatusFailed,
					Err:    fmt.Errorf("状态轮询panic: %w", err),
				}
			}

			stepExecResult := StepExecResult{
				StepID:          step.StepID,
				Order:           step.Order,
				StepName:        step.Name,
				TaskID:          step.TaskID,
				Status:          monitorResult.Status,
				ProcessPosition: monitorResult.ProcessPosition,
				Result:          monitorResult.Result,
				Err:             monitorResult.Err,
			}

			if monitorResult.Status == StatusPending {
				tp.Record(step, stepExecResult)
				continue
			}
			return stepExecResult
		}
	}
}

// Record 记录步骤执行结果
func (tp *TaskWorkerProcessor) Record(step StepInfo, result StepExecResult) (stepID int, taskResult TaskExecResult) {
	result.StepID = step.StepID
	result.Order = step.Order
	result.StepName = step.Name
	stepID = tp.Recorder.RecordStep(step.StepID, result)
	taskResult = TaskExecResult{
		Status:          result.Status,
		Order:           step.Order,
		StepName:        step.Name,
		StepID:          stepID,
		ProcessPosition: result.ProcessPosition,
		Err:             result.Err,
		Result:          result.Result,
	}
	tp.Recorder.RecordTask(step.TaskID, taskResult)
	return stepID, taskResult
}

// TaskWorker 任务执行器
type TaskWorker interface {
	// Arrange 步骤编排
	Arrange(ctx context.Context, payload []byte) (steps []StepInfo)
}
