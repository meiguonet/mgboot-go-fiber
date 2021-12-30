package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/util/slicex"
	"strings"
	"time"
)

func MidRawBodyInjector() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidRawBodyInjector")
		}

		ctx.Locals("ExecStart", time.Now())
		method := ctx.Method()
		methods := []string{"POST", "PUT", "PATCH", "DELETE"}

		if !slicex.InStringSlice(method, methods) {
			return ctx.Next()
		}

		contentType := strings.ToLower(ctx.Get(fiber.HeaderContentType))
		isJson := strings.Contains(contentType, fiber.MIMEApplicationJSON)
		isXml1 := strings.Contains(contentType, fiber.MIMEApplicationXML)
		isXml2 := strings.Contains(contentType, fiber.MIMETextXML)

		if !isJson && !isXml1 && !isXml2 {
			return ctx.Next()
		}

		if len(ctx.Body()) > 0 {
			buf := make([]byte, 0, 64 * 1024 * 1024)
			n1 := copy(buf, ctx.Body())

			if n1 > 0 {
				ctx.Locals("requestRawBody", buf[:n1])
			}
		}

		return ctx.Next()
	}
}
