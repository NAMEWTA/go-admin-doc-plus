package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"go-admin/internal/platform/coordination"
	"go-admin/internal/platform/database"
)

const taskSavepoint = "scheduler_task_effect"

type ExecutorConfig struct {
	Owner       string
	BatchSize   int
	TaskTimeout time.Duration
	Clock       Clock
	Observer    Observer
}

type ExecuteResult struct {
	Triggered int
	Succeeded int
	Failed    int
}

type Executor struct {
	db         *database.Database
	repository repository
	registry   *Registry
	config     ExecutorConfig
}

type transactionExecutor interface {
	WithinTx(context.Context, func(context.Context, database.Tx) error) error
}

func NewExecutor(db *database.Database, registry *Registry, config ExecutorConfig) (*Executor, error) {
	if db == nil || registry == nil || !ownerPattern(config.Owner) || config.BatchSize < 1 || config.BatchSize > 100 || config.TaskTimeout <= 0 || config.TaskTimeout > time.Hour || config.Clock == nil {
		return nil, errors.New("scheduler executor config is invalid")
	}
	if _, err := utcNow(config.Clock); err != nil {
		return nil, errors.New("scheduler executor clock is invalid")
	}
	return &Executor{db: db, repository: repository{dialect: db.Dialect()}, registry: registry, config: config}, nil
}

// RunOnce consumes only a lease acquired by the runtime composition root. Scheduler never acquires
// an independent lease, so Outbox and Scheduler share the same DB, owner, and cancellation group.
func (e *Executor) RunOnce(ctx context.Context, lease *coordination.Lease) (ExecuteResult, error) {
	if e == nil || lease == nil {
		return ExecuteResult{}, errors.New("scheduler coordination lease is required")
	}
	if err := lease.Authorize(e.db, e.config.Owner); err != nil {
		e.observeLease(err, 0)
		return ExecuteResult{}, err
	}
	return e.runOnce(ctx, lease)
}

func (e *Executor) runOnce(ctx context.Context, transactions transactionExecutor) (ExecuteResult, error) {
	if e == nil || transactions == nil {
		return ExecuteResult{}, errors.New("scheduler transaction executor is required")
	}
	result := ExecuteResult{}
	for result.Triggered < e.config.BatchSize {
		now, err := utcNow(e.config.Clock)
		if err != nil {
			e.observe(OutcomeDependency, result.Triggered, false)
			return result, ErrInternal
		}
		executed := false
		failed := false
		err = transactions.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
			var err error
			executed, failed, err = e.executeDue(ctx, tx, now)
			return err
		})
		if err != nil {
			if errors.Is(err, coordination.ErrLeaseLost) || errors.Is(err, coordination.ErrNotLeader) {
				e.observeLease(err, result.Triggered)
				return result, err
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return result, err
			}
			e.observe(OutcomeDependency, result.Triggered, false)
			return result, ErrInternal
		}
		if !executed {
			outcome := OutcomeIdle
			if result.Triggered > 0 {
				outcome = OutcomeExecuted
				if result.Failed > 0 {
					outcome = OutcomeFailed
				}
			}
			e.observe(outcome, result.Triggered, false)
			return result, nil
		}
		result.Triggered++
		if failed {
			result.Failed++
		} else {
			result.Succeeded++
		}
	}
	outcome := OutcomeExecuted
	if result.Failed > 0 {
		outcome = OutcomeFailed
	}
	e.observe(outcome, result.Triggered, false)
	return result, nil
}

func (e *Executor) executeDue(ctx context.Context, tx database.Tx, now time.Time) (bool, bool, error) {
	record, err := e.repository.dueDefinition(ctx, tx, now)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	schedule, err := unmarshalSchedule(record.ScheduleJSON)
	if err != nil {
		return false, false, err
	}
	if !record.NextRunAt.Valid {
		return false, false, ErrInternal
	}
	started, err := utcNow(e.config.Clock)
	if err != nil {
		return false, false, err
	}
	status, errorCode, err := e.runTask(ctx, tx, record)
	if err != nil {
		return false, false, err
	}
	finished, err := utcNow(e.config.Clock)
	if err != nil {
		return false, false, err
	}
	if started.Before(record.NextRunAt.Time.UTC()) || finished.Before(started) {
		return false, false, ErrInternal
	}
	next, ok := nextOccurrence(schedule, finished)
	if !ok {
		return false, false, ErrInternal
	}
	execution := Execution{
		ID:                 uuid.NewString(),
		DefinitionID:       record.ID,
		DefinitionRevision: record.Revision,
		TaskType:           record.TaskType,
		ScheduledFor:       record.NextRunAt.Time.UTC(),
		StartedAt:          started,
		FinishedAt:         finished,
		Status:             status,
		ErrorCode:          errorCode,
		ExecutorOwner:      e.config.Owner,
	}
	if err := e.repository.insertExecution(ctx, tx, execution); err != nil {
		return false, false, err
	}
	if err := e.repository.advanceDefinition(ctx, tx, record.ID, next, finished); err != nil {
		return false, false, err
	}
	return true, status == ExecutionFailed, nil
}

func (e *Executor) runTask(ctx context.Context, tx database.Tx, record definitionRecord) (ExecutionStatus, string, error) {
	task, exists := e.registry.task(record.TaskType)
	if !exists {
		return ExecutionFailed, "task_unregistered", nil
	}
	if _, err := task.normalize(record.ParametersJSON); err != nil {
		return ExecutionFailed, "parameters_invalid", nil
	}
	if _, err := tx.ExecContext(ctx, `SAVEPOINT `+taskSavepoint); err != nil {
		return "", "", err
	}
	taskContext, cancel := context.WithTimeout(ctx, e.config.TaskTimeout)
	err := safelyRun(taskContext, task, tx, record.ParametersJSON)
	taskContextErr := taskContext.Err()
	cancel()
	if taskContextErr == nil {
		if afterCancel := taskContext.Err(); errors.Is(afterCancel, context.DeadlineExceeded) {
			taskContextErr = afterCancel
		}
	}
	if ctx.Err() != nil {
		return "", "", ctx.Err()
	}
	if err == nil && taskContextErr == nil {
		if _, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT `+taskSavepoint); releaseErr != nil {
			return "", "", releaseErr
		}
		return ExecutionSucceeded, "", nil
	}
	if _, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT `+taskSavepoint); rollbackErr != nil {
		return "", "", rollbackErr
	}
	if _, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT `+taskSavepoint); releaseErr != nil {
		return "", "", releaseErr
	}
	var taskFailure TaskFailure
	switch {
	case errors.Is(taskContextErr, context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded):
		return ExecutionFailed, "task_timeout", nil
	case errors.As(err, &taskFailure) && errorCodePattern.MatchString(taskFailure.Code):
		return ExecutionFailed, taskFailure.Code, nil
	default:
		return ExecutionFailed, "task_failed", nil
	}
}

func safelyRun(ctx context.Context, task registeredTask, tx database.Tx, parameters []byte) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("scheduler task panicked")
		}
	}()
	return task.run(ctx, tx, parameters)
}

type RunOptions struct {
	PollInterval   time.Duration
	FailureBackoff time.Duration
	Wait           func(context.Context, time.Duration) error
}

func (e *Executor) Run(ctx context.Context, lease *coordination.Lease, options RunOptions) error {
	if options.PollInterval <= 0 || options.FailureBackoff <= 0 {
		return errors.New("scheduler loop options are invalid")
	}
	wait := options.Wait
	if wait == nil {
		wait = waitFor
	}
	for {
		_, err := e.RunOnce(ctx, lease)
		if errors.Is(err, coordination.ErrLeaseLost) || errors.Is(err, coordination.ErrNotLeader) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		delay := options.PollInterval
		if err != nil {
			delay = options.FailureBackoff
		}
		if err := wait(ctx, delay); err != nil {
			return err
		}
	}
}

func waitFor(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (e *Executor) observe(outcome string, triggered int, lost bool) {
	if e.config.Observer != nil {
		e.config.Observer.Observe(Observation{Outcome: outcome, Triggered: triggered, ActiveExecutor: true, LostLock: lost})
	}
}

func (e *Executor) observeLease(err error, triggered int) {
	if errors.Is(err, coordination.ErrLeaseLost) || errors.Is(err, coordination.ErrNotLeader) {
		e.observe(OutcomeLeaseLost, triggered, true)
	}
}

func ownerPattern(owner string) bool {
	if owner == "" || len(owner) > 255 {
		return false
	}
	for index, value := range owner {
		if index == 0 && !asciiAlphaNumeric(value) || index > 0 && !(asciiAlphaNumeric(value) || value == '.' || value == '_' || value == ':' || value == '/' || value == '-') {
			return false
		}
	}
	return true
}

func asciiAlphaNumeric(value rune) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z' || value >= '0' && value <= '9'
}
