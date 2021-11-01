package taskx

type Task interface {
	GetTaskName() string
	SetParams(params map[string]interface{})
	GetTaskParams() map[string]interface{}
	Run() bool
}
