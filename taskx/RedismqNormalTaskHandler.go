package taskx

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/meiguonet/mgboot-go-common/util/errorx"
	"github.com/meiguonet/mgboot-go-dal/poolx"
	"github.com/meiguonet/mgboot-go-fiber/cachex"
	"github.com/meiguonet/mgboot-go-fiber/mgboot"
	"sync"
	"time"
)

type redismqNormalTaskHandler struct {
}

func (h *redismqNormalTaskHandler) Run() {
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

	cacheKey := cachex.CacheKeyRedismqNormal()
	payloads := make([]string, 0)

	for {
		if len(payloads) >= 10 {
			break
		}

		conn, err := poolx.GetRedisConnection()

		if err != nil {
			break
		}

		payload, _ := redis.String(conn.Do("LPOP", cacheKey))
		conn.Close()

		if payload != "" {
			payloads = append(payloads, payload)
		}

		time.Sleep(50 * time.Millisecond)
	}

	n1 := len(payloads)

	if n1 < 1 {
		return
	}

	if n1 == 1 {
		RunMqTask(payloads[0])
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(n1)

	for _, payload := range payloads {
		go func(wg *sync.WaitGroup, payload string) {
			defer wg.Done()
			RunMqTask(payload)
		}(wg, payload)
	}

	wg.Wait()
}
