package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
)

func MidOptionsReq() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidOptionsReq")
		}

		if ctx.Method() != "OPTIONS" {
			return ctx.Next()
		}

		AddCorsSupport(ctx)
		AddPoweredBy(ctx)
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		ctx.SendString(`{"code":200}`)
		return nil
	}
}
