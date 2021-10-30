package mgboot

import "github.com/gofiber/fiber/v2"

type HtmlResponse struct {
	contents string
}

func NewHtmlResponse(contents string) HtmlResponse {
	return HtmlResponse{contents: contents}
}

func (p HtmlResponse) GetContentType() string {
	return fiber.MIMETextHTMLCharsetUTF8
}

func (p HtmlResponse) GetContents() (int, string) {
	return 200, p.contents
}
