package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"github.com/meiguonet/mgboot-go-dal/ratelimiter"
	"strings"
	"time"
)

func MidRateLimit(handlerName string, settings interface{}) fiber.Handler {
	var total int
	var duration time.Duration
	var limitByIp bool

	if map1, ok := settings.(map[string]interface{}); ok && len(map1) > 0 {
		total = castx.ToInt(map1["total"])

		if d1, ok := map1["duration"].(time.Duration); ok && d1 > 0 {
			duration = d1
		} else if n1, err := castx.ToInt64E(map1["duration"]); err == nil && n1 > 0 {
			duration = time.Duration(n1) * time.Millisecond
		}

		limitByIp = castx.ToBool(map1["limitByIp"])
	} else if s1, ok := settings.(string); ok && s1 != "" {
		s1 = strings.ReplaceAll(s1, "[syh]", `"`)
		map1 := jsonx.MapFrom(s1)

		if len(map1) > 0 {
			total = castx.ToInt(map1["total"])

			if d1, ok := map1["duration"].(time.Duration); ok && d1 > 0 {
				duration = d1
			} else if n1, err := castx.ToInt64E(map1["duration"]); err == nil && n1 > 0 {
				duration = time.Duration(n1) * time.Millisecond
			}

			limitByIp = castx.ToBool(map1["limitByIp"])
		}
	}

	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRateLimit")
		}

		if handlerName == "" || total < 1 || duration < 1 {
			return ctx.Next()
		}

		req := NewRequest(ctx)
		id := handlerName

		if limitByIp {
			id += "@" + req.GetClientIp()
		}

		opts := ratelimiter.NewRatelimiterOptions(RatelimiterLuaFile(), RatelimiterCacheDir())
		limiter := ratelimiter.NewRatelimiter(id, total, duration, opts)
		result := limiter.GetLimit()
		remaining := castx.ToInt(result["remaining"])

		if remaining >= 0 {
			return ctx.Next()
		}

		return NewRateLimitError(result)
	}
}
