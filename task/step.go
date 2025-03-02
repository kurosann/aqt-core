package task

import "context"

// StepExecResult 步骤执行结果
type StepExecResult struct {
	Status          Status
	Order           int
	StepName        string
	TaskID          int
	StepID          int
	ProcessPosition int
	Err             error
	Result          []byte
}

// StepExecCompleted 步骤执行完成
func StepExecCompleted(result []byte) StepExecResult {
	return StepExecResult{
		Status: StatusCompleted,
		Result: result,
	}
}

// StepExecPending 步骤执行中
func StepExecPending(result []byte) StepExecResult {
	return StepExecResult{
		Status: StatusPending,
		Result: result,
	}
}

// StepExecFailed 步骤执行失败
func StepExecFailed(err error) StepExecResult {
	return StepExecResult{
		Status: StatusFailed,
		Err:    err,
		Result: make([]byte, 0),
	}
}

// StepExecFailedWithResult 步骤执行失败
func StepExecFailedWithResult(err error, result []byte) StepExecResult {
	return StepExecResult{
		Status: StatusFailed,
		Result: result,
		Err:    err,
	}
}

// StepPollStatusResult 步骤轮询状态结果
type StepPollStatusResult struct {
	Status          Status
	ProcessPosition int
	Err             error
	Result          []byte
}

// StepPollCompleted 步骤轮询完成
func StepPollCompleted(result []byte) StepPollStatusResult {
	return StepPollStatusResult{
		Status:          StatusCompleted,
		ProcessPosition: 100,
		Result:          result,
	}
}

// StepPollFailed 步骤轮询失败
func StepPollFailed(err error) StepPollStatusResult {
	return StepPollStatusResult{
		Status: StatusFailed,
		Err:    err,
	}
}

// StepPollPending 步骤轮询中
// processPosition 进度百分比,范围0-100
func StepPollPending(processPosition int) StepPollStatusResult {
	return StepPollStatusResult{
		Status:          StatusPending,
		ProcessPosition: processPosition,
	}
}

// StepWorker 步骤执行器
type StepWorker interface {
	// Execute 执行步骤
	Execute(ctx context.Context, payload, preStepResult []byte) (execResult StepExecResult)
	// PollStatus 轮询步骤状态
	PollStatus(ctx context.Context, payload, stepResult []byte) (execResult StepPollStatusResult)
	// Rollback	回滚步骤
	Rollback(ctx context.Context, payload, stepResult []byte) (err error)
}

// StepInfo 步骤信息
type StepInfo struct {
	TaskID     int
	StepID     int
	Order      int
	Name       string
	StepWorker StepWorker
}

func NewStepInfo(name string, stepWorker StepWorker) StepInfo {
	return StepInfo{
		Name:       name,
		StepWorker: stepWorker,
	}
}

// StepRecorder 步骤记录器
type StepRecorder interface {
	// RecordStep 记录步骤执行结果
	// stepID 为0时创建并生成步骤ID,不允许失败
	RecordStep(stepID int, result StepExecResult) (id int)
}
