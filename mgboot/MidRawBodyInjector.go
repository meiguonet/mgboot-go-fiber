package mgboot

import (
	"bufio"
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/AppConf"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/slicex"
	"io"
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

		reader := bufio.NewReader(bytes.NewReader([]byte{}))

		if err := ctx.Request().ReadLimitBody(reader, 64 * 1024 * 1024); err != nil && err != io.EOF {
			return ctx.Next()
		}

		buf := make([]byte, 0, 64 * 1024 * 1024)

		if n1, err := reader.WriteTo(bytes.NewBuffer(buf)); err == nil && n1 > 0 {
			n2 := castx.ToInt(n1)
			ctx.Locals("requestRawBody", buf[:n2])
		}

		return ctx.Next()
	}
}