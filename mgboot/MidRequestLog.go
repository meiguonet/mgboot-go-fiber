package mgboot

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
)

func MidRequestLog() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRequestLog")
		}

		if !RequestLogEnabled() {
			return ctx.Next()
		}

		var req *Request

		if r, ok := ctx.Locals("request").(*Request); ok {
			req = r
		}

		if req == nil {
			return ctx.Next()
		}

		logger := RequestLogLogger()

		msg := fmt.Sprintf(
			"%s %s from %s",
			ctx.Method(),
			req.GetRequestUrl(ctx, true),
			req.GetClientIp(ctx),
		)

		logger.Info(msg)

		if LogRequestBody() {
			rawBody := req.GetRawBody(ctx)

			if len(rawBody) > 0 {
				logger.Debugf(string(rawBody))
			}
		}

		return ctx.Next()
	}
}
