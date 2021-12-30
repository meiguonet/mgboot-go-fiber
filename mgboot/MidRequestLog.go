package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"strings"
	"time"
)

func MidRequestLog() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Locals("ExecStart", time.Now())

		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRequestLog")
		}

		if !RequestLogEnabled() {
			return ctx.Next()
		}

		req := NewRequest(ctx)
		logger := RequestLogLogger()
		sb := strings.Builder{}
		sb.WriteString(ctx.Method())
		sb.WriteString(" ")
		sb.WriteString(req.GetRequestUrl(true))
		sb.WriteString(" from ")
		sb.WriteString(req.GetClientIp())
		logger.Info(sb.String())

		if LogRequestBody() {
			rawBody := req.GetRawBody()

			if len(rawBody) > 0 {
				logger.Debugf(string(rawBody))
			}
		}

		return ctx.Next()
	}
}
