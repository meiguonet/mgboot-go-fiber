package taskx

import (
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"time"
)

type retryPolicy struct {
	failTimes     int
	retryAttempts int
	retryInterval time.Duration
}

func NewRetryPolicy(settings map[string]interface{}) *retryPolicy {
	var retryInterval time.Duration

	if d1, ok := settings["retryInterval"].(time.Duration); ok {
		retryInterval = d1
	} else if n1, ok := settings["retryInterval"].(int64); ok {
		retryInterval = time.Duration(n1) * time.Millisecond
	} else if n1, ok := settings["retryInterval"].(int64); ok {
		retryInterval = time.Duration(n1) * time.Millisecond
	} else if n1, ok := settings["retryInterval"].(int); ok {
		retryInterval = time.Duration(n1) * time.Second
	} else if s1, ok := settings["retryInterval"].(string); ok && s1 != "" {
		retryInterval = castx.ToDuration(s1)
	}

	return &retryPolicy{
		failTimes:     castx.ToInt(settings["failTimes"]),
		retryAttempts: castx.ToInt(settings["retryAttempts"]),
		retryInterval: retryInterval,
	}
}
