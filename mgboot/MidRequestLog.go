package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"strings"
	"time"
)

func MidRequestLog() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRequestLog")
		}

		ctx.Locals("ExecStart", time.Now())

		if !RequestLogEnabled() {
			return ctx.Next()
		}

		logger := RequestLogLogger()
		sb := strings.Builder{}
		sb.WriteString(ctx.Method())
		sb.WriteString(" ")
		sb.WriteString(GetRequestUrl(ctx, true))
		sb.WriteString(" from ")
		sb.WriteString(GetClientIp(ctx))
		logger.Info(sb.String())

		if LogRequestBody() {
			rawBody := GetRawBody(ctx)

			if len(rawBody) > 0 {
				logger.Debugf(string(rawBody))
			}
		}

		return ctx.Next()
	}
}
