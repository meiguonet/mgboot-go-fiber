package taskx

import (
	"fmt"
	"github.com/meiguonet/mgboot-go-common/enum/DatetimeFormat"
	"github.com/meiguonet/mgboot-go-common/logx"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-dal/poolx"
	"github.com/meiguonet/mgboot-go-fiber/cachex"
	"github.com/meiguonet/mgboot-go-fiber/mgboot"
	"github.com/robfig/cron/v3"
	"strings"
	"time"
)

type fnFindMqTask func(name string) Task
var mqTaskFinder fnFindMqTask
var cronTaskLogEnabled bool
var cronTaskLogger logx.Logger
var mqTaskLogEnabled bool
var mqTaskLogger logx.Logger
var cronTasks = make([]CronTask, 0)

func SetMqTaskFinder(fn fnFindMqTask) {
	mqTaskFinder = fn
}

func CronTaskLogEnabled(flag ...bool) bool {
	if len(flag) > 0 {
		cronTaskLogEnabled = flag[0]
	}

	return cronTaskLogEnabled
}

func CronTaskLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		cronTaskLogger = logger[0]
	}

	_logger := cronTaskLogger

	if _logger == nil {
		_logger = mgboot.NewNoopLogger()
	}

	return _logger
}

func MqTaskLogEnabled(flag ...bool) bool {
	if len(flag) > 0 {
		mqTaskLogEnabled = flag[0]
	}

	return mqTaskLogEnabled
}

func MqTaskLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		mqTaskLogger = logger[0]
	}

	_logger := mqTaskLogger

	if _logger == nil {
		_logger = mgboot.NewNoopLogger()
	}

	return _logger
}

func RunCronTask(taskName string) {
	var task CronTask

	for _, ct := range cronTasks {
		if ct.GetTaskName() == taskName {
			task = ct
			break
		}
	}

	if task == nil {
		return
	}

	if cronTaskLogEnabled {
		CronTaskLogger().Info("run cron task: " + taskName)
	}

	task.Run()
}

func RunMqTask(payload string) {
	if mqTaskFinder == nil {
		return
	}

	map1 := jsonx.MapFrom(payload)
	taskName := castx.ToString(map1["taskName"])

	if taskName == "" || mqTaskFinder == nil {
		return
	}

	task := mqTaskFinder(taskName)

	if task == nil {
		return
	}

	taskParams := castx.ToStringMap(map1["taskParams"])

	if len(taskParams) > 0 {
		task.SetParams(taskParams)
	}

	runAt := castx.ToString(map1["runAt"])
	var taskType string

	if runAt != "" {
		taskType = "delayable"
	} else {
		taskType = "normal"
	}

	if mqTaskLogEnabled {
		sb := []string{
			fmt.Sprintf("run %s task: %s", taskType, taskName),
		}

		if runAt != "" {
			sb = append(sb, ", scheduled at: " + runAt)
		}

		if len(taskParams) > 0 {
			sb = append(sb, ", task params: " + jsonx.ToJson(taskParams))
		}

		MqTaskLogger().Info(strings.Join(sb, ""))
	}

	success := task.Run()

	if mqTaskLogEnabled {
		sb := make([]string, 0)

		if success {
			sb = append(sb, "success ")
		} else {
			sb = append(sb, "fail ")
		}

		sb = append(sb, fmt.Sprintf("to run %s task: %s", taskType, taskName))

		if runAt != "" {
			sb = append(sb, ", scheduled at: " + runAt)
		}

		if len(taskParams) > 0 {
			sb = append(sb, ", task params: " + jsonx.ToJson(taskParams))
		}

		MqTaskLogger().Info(strings.Join(sb, ""))
	}

	if success {
		return
	}

	retryAttempts := castx.ToInt(map1["retryAttempts"])

	if retryAttempts < 1 {
		return
	}

	retryInterval := castx.ToInt64(map1["retryInterval"])

	if retryInterval < 1 {
		return
	}

	failTimes := castx.ToInt(map1["failTimes"], 0) + 1

	if failTimes > retryAttempts {
		return
	}

	retryDuration := time.Duration(retryInterval) * time.Millisecond

	policy := NewRetryPolicy(map[string]interface{}{
		"failTimes":     failTimes,
		"retryAttempts": retryAttempts,
		"retryInterval": retryDuration,
	})

	PublishDelayable(task, retryDuration, policy)
}

func WithCronTasks(task CronTask) {
	entries := make([]CronTask, 0)
	var added bool

	for _, ct := range cronTasks {
		if ct.GetTaskName() == task.GetTaskName() {
			entries = append(entries, task)
			added = true
			continue
		}

		entries = append(entries, ct)
	}

	if !added {
		entries = append(entries, task)
	}

	cronTasks = entries
}

func Publish(task Task, policy ...*retryPolicy) {
	var rp *retryPolicy

	if len(policy) > 0 {
		rp = policy[0]
	}

	payload := map[string]interface{}{
		"taskName": task.GetTaskName(),
	}

	if len(task.GetTaskParams()) > 0 {
		payload["taskParams"] = task.GetTaskParams()
	}

	if rp != nil {
		payload["failTimes"] = rp.failTimes
		payload["retryAttempts"] = rp.retryAttempts
		payload["retryInterval"] = rp.retryInterval.Milliseconds()
	}

	conn, err := poolx.GetRedisConnection()

	if err != nil {
		return
	}

	defer conn.Close()
	_, err = conn.Do("RPUSH", cachex.CacheKeyRedismqNormal(), jsonx.ToJson(payload))

	if !mqTaskLogEnabled {
		return
	}

	sb := make([]string, 0)

	if err == nil {
		sb = append(sb, "success ")
	} else {
		sb = append(sb, "fail ")
	}

	sb = append(sb, "to publish normal task: " + task.GetTaskName())

	if len(task.GetTaskParams()) > 0 {
		sb = append(sb, ", task params: " + jsonx.ToJson(task.GetTaskParams()))
	}

	MqTaskLogger().Info(strings.Join(sb, ""))
}

func PublishDelayable(task Task, runAfter time.Duration, policy ...*retryPolicy) {
	var rp *retryPolicy

	if len(policy) > 0 {
		rp = policy[0]
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	runAt := time.Now().In(loc).Add(runAfter)

	payload := map[string]interface{}{
		"taskName": task.GetTaskName(),
		"runAt":    runAt.Format(DatetimeFormat.Full),
	}

	if len(task.GetTaskParams()) > 0 {
		payload["taskParams"] = task.GetTaskParams()
	}

	if rp != nil {
		payload["failTimes"] = rp.failTimes
		payload["retryAttempts"] = rp.retryAttempts
		payload["retryInterval"] = rp.retryInterval.Milliseconds()
	}

	conn, err := poolx.GetRedisConnection()

	if err != nil {
		return
	}

	defer conn.Close()
	_, err = conn.Do("ZADD", cachex.CacheKeyRedismqDelayable(), runAt.Unix(), jsonx.ToJson(payload))

	if !mqTaskLogEnabled {
		return
	}

	sb := make([]string, 0)

	if err == nil {
		sb = append(sb, "success ")
	} else {
		sb = append(sb, "fail ")
	}

	msg := fmt.Sprintf(
		"to publish delayable task: %s, scheduled at: %s",
		task.GetTaskName(),
		runAt.Format(DatetimeFormat.Full),
	)

	sb = append(sb, msg)

	if len(task.GetTaskParams()) > 0 {
		sb = append(sb, ", task params: " + jsonx.ToJson(task.GetTaskParams()))
	}

	MqTaskLogger().Info(strings.Join(sb, ""))
}

func HandleCronTasks(crond *cron.Cron) {
	for _, task := range cronTasks {
		crond.AddJob(task.GetSpec(), NewCronJob(task.GetTaskName()))
	}
}

func HandleRedismqNormalTasks(crond *cron.Cron) {
	crond.AddJob("@every 1s", &redismqNormalTaskHandler{})
}

func HandleRedismqDelayableTasks(crond *cron.Cron) {
	crond.AddJob("@every 1s", &redismqDelayableTaskHandler{})
}
