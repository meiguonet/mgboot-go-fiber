package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/enum/RegexConst"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"github.com/meiguonet/mgboot-go-fiber/enum/JwtVerifyErrno"
	"strings"
)

func MidJwtAuth(settingsKey string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if AppConf.GetBoolean("logging.logMiddlewareRun") {
			RuntimeLogger().Info("middleware run: mgboot.MidJwtAuth")
		}

		if settingsKey == "" {
			return ctx.Next()
		}

		settings := GetJwtSettings(settingsKey)

		if settings == nil {
			return ctx.Next()
		}

		token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
		token = stringx.RegexReplace(token, RegexConst.SpaceSep, " ")

		if strings.Contains(token, " ") {
			token = stringx.SubstringAfter(token, " ")
		}

		if token == "" {
			return NewJwtAuthError(JwtVerifyErrno.NotFound)
		}

		errno := VerifyJsonWebToken(token, settings)

		if errno < 0 {
			return NewJwtAuthError(errno)
		}

		return ctx.Next()
	}
}
