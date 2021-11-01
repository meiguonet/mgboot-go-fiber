package taskx

type cronJob struct {
	taskName string
}

func NewCronJob(taskName string) *cronJob {
	return &cronJob{taskName: taskName}
}

func (job *cronJob) Run() {
	RunCronTask(job.taskName)
}
