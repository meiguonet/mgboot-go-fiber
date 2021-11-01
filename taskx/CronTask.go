package taskx

import "github.com/robfig/cron/v3"

type CronTask interface {
	cron.Job
	GetTaskName() string
	GetSpec() string
}
