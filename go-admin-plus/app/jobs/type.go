package jobs

import (
	"context"

	"github.com/robfig/cron/v3"
)

type Job interface {
	Run()
	addJob(*cron.Cron) (int, error)
}

type JobExec interface {
	Exec(arg interface{}) error
}

// ContextJobExec is the cancellable expand-phase extension for executable jobs.
// Implementations should return ctx.Err when cancellation interrupts work.
type ContextJobExec interface {
	ExecContext(context.Context, interface{}) error
}

func CallExec(e JobExec, arg interface{}) error {
	return e.Exec(arg)
}
