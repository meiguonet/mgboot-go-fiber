package taskx

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/meiguonet/mgboot-go-common/enum/DatetimeFormat"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/errorx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-dal/poolx"
	"github.com/meiguonet/mgboot-go-fiber/cachex"
	"github.com/meiguonet/mgboot-go-fiber/mgboot"
	"sync"
	"time"
)

type redismqDelayableTaskHandler struct {
}

func (h *redismqDelayableTaskHandler) Run() {
	defer func() {
		if r := recover(); r != nil {
			var err error

			if ex, ok := r.(error); ok {
				err = ex
			} else {
				err = fmt.Errorf("%v", r)
			}

			mgboot.RuntimeLogger().Error(errorx.Stacktrace(err))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := poolx.GetRedisConnection(ctx)

	if err != nil {
		return
	}

	defer conn.Close()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	cacheKey := cachex.CacheKeyRedismqDelayable()
	payloads, _ := redis.Strings(conn.Do("ZRANGEBYSCORE", cacheKey, now.Unix() - 60, now.Unix() + 5))

	if len(payloads) < 1 {
		return
	}

	entries := make([]string, 0)
	payloadsToRemove := make([]interface{}, 0)

	for _, payload := range payloads {
		if payload == "" {
			continue
		}

		map1 := jsonx.MapFrom(payload)
		runAt, err := time.ParseInLocation(DatetimeFormat.Full, castx.ToString(map1["runAt"]), loc)

		if err != nil {
			payloadsToRemove = append(payloadsToRemove, payload)
			continue
		}

		if now.Unix() < runAt.Unix() {
			continue
		}

		entries = append(entries, payload)
		payloadsToRemove = append(payloadsToRemove, payload)
	}

	if len(payloadsToRemove) > 0 {
		payloadsToRemove = append([]interface{}{cacheKey}, payloadsToRemove...)
		_, _ = conn.Do("ZREM", payloadsToRemove...)
	}

	if len(entries) < 1 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(entries))

	for _, payload := range entries {
		go func(payload string) {
			defer wg.Done()
			RunMqTask(payload)
		}(payload)
	}

	wg.Wait()
}
