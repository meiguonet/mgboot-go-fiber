package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"github.com/meiguonet/mgboot-go-fiber/enum/JwtVerifyErrno"
	"regexp"
	"strings"
)

func MidJwtAuth() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var req *Request

		if r, ok := ctx.Locals("request").(*Request); ok {
			req = r
		}

		if req == nil || req.JwtAuthSettings() == nil {
			return ctx.Next()
		}

		settings := GetJwtSettings(req.JwtAuthSettings().Key())

		if settings == nil {
			return ctx.Next()
		}

		token := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
		re := regexp.MustCompile(`[\x20\t]+`)
		token = re.ReplaceAllString(token, " ")

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
