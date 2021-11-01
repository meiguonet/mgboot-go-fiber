package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-dal/ratelimiter"
)

func MidRateLimit() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var req *Request

		if r, ok := ctx.Locals("request").(*Request); ok {
			req = r
		}

		if req == nil || req.RateLimitSettings() == nil {
			return ctx.Next()
		}

		settings := req.RateLimitSettings()
		id := req.HandlerFuncName()

		if settings.LimitByIp() {
			id += "@" + req.GetClientIp(ctx)
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
