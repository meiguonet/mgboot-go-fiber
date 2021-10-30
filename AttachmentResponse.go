package mgboot

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/meiguonet/mgboot-go-common/util/mimex"
	"io/ioutil"
)

type AttachmentResponse struct {
	buf                []byte
	mimeType           string
	attachmentFileName string
}

func NewAttachmentResponseFromFile(fpath, attachmentFileName string, mimeType ...string) AttachmentResponse {
	buf, _ := ioutil.ReadFile(fpath)
	var _mimeType string

	if len(mimeType) > 0 {
		_mimeType = mimeType[0]
	}

	if _mimeType == "" {
		_mimeType = mimex.GetMimeType(buf)
	}

	return AttachmentResponse{
		buf:                buf,
		mimeType:           _mimeType,
		attachmentFileName: attachmentFileName,
	}
}

func NewAttachmentResponseFromBuffer(buf []byte, attachmentFileName string, mimeType ...string) AttachmentResponse {
	var _mimeType string

	if len(mimeType) > 0 {
		_mimeType = mimeType[0]
	}

	if _mimeType == "" {
		_mimeType = mimex.GetMimeType(buf)
	}

	return AttachmentResponse{
		buf:                buf,
		mimeType:           _mimeType,
		attachmentFileName: attachmentFileName,
	}
}

func (p AttachmentResponse) GetContentType() string {
	if p.mimeType == "" {
		return fiber.MIMEOctetStream
	}

	return p.mimeType
}

func (p AttachmentResponse) GetContents() (int, string) {
	if len(p.buf) < 1 || p.attachmentFileName == "" {
		return 400, ""
	}
	
	return 200, ""
}

func (p AttachmentResponse) Buffer() []byte {
	return p.buf
}

func (p AttachmentResponse) AddSpecifyHeaders(ctx *fiber.Ctx) {
	disposition := fmt.Sprintf(`attachment; filename="%s"`, p.attachmentFileName)
	ctx.Set(fiber.HeaderContentType, p.GetContentType())
	ctx.Set(fiber.HeaderContentLength, fmt.Sprintf("%d", len(p.buf)))
	ctx.Set(fiber.HeaderTransferEncoding, "binary")
	ctx.Set(fiber.HeaderContentDisposition, disposition)
	ctx.Set(fiber.HeaderCacheControl, "no-cache, no-store, max-age=0, must-revalidate")
	ctx.Set(fiber.HeaderPragma, "public")
}
