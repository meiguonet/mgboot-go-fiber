package mgboot

import "github.com/gofiber/fiber/v2"

type XmlResponse struct {
	contents string
}

func NewXmlResponse(contents string) XmlResponse {
	return XmlResponse{contents: contents}
}

func (p XmlResponse) GetContentType() string {
	return fiber.MIMETextXMLCharsetUTF8
}

func (p XmlResponse) GetContents() (int, string) {
	return 200, p.contents
}
