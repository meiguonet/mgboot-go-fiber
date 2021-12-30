package mgboot

import (
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/util/jsonx"
	"strings"
)

type JsonResponse struct {
	payload interface{}
}

func NewJsonResponse(payload interface{}) JsonResponse {
	return JsonResponse{payload: payload}
}

func (p JsonResponse) GetContentType() string {
	return fiber.MIMEApplicationJSONCharsetUTF8
}

func (p JsonResponse) GetContents() (statusCode int, contents string) {
	statusCode = 200

	if s1, ok := p.payload.(string); ok {
		s1 = strings.TrimSpace(s1)

		if p.isJson(s1) {
			contents = s1
			return
		}
	}

	opts := jsonx.NewToJsonOption().HandleTimeField().StripZeroTimePart()
	contents = strings.TrimSpace(jsonx.ToJson(p.payload, opts))

	if !p.isJson(contents) {
		contents = "{}"
	}

	return
}

func (p JsonResponse) isJson(contents string) bool {
	var flag bool

	if strings.HasPrefix(contents, "{") && strings.HasSuffix(contents, "}") {
		flag = true
	} else if strings.HasPrefix(contents, "[") && strings.HasSuffix(contents, "]") {
		flag = true
	}

	return flag
}
