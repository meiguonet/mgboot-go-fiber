package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-dal/ratelimiter"
)

func MidRateLimit() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRateLimit")
		}

		req := NewRequest(ctx)
		id := FindMatchedHandlerName(req)

		if id == "" {
			return ctx.Next()
		}

		settings := FindMatchedRateLimitSettings(req)

		if settings == nil {
			return ctx.Next()
		}

		if settings.LimitByIp() {
			id += "@" + req.GetClientIp()
		}

		opts := ratelimiter.NewRatelimiterOptions(RatelimiterLuaFile(), RatelimiterCacheDir())
		limiter := ratelimiter.NewRatelimiter(id, settings.total, settings.duration, opts)
		result := limiter.GetLimit()
		remaining := castx.ToInt(result["remaining"])

		if remaining >= 0 {
			return ctx.Next()
		}

		return NewRateLimitError(result)
	}
}
