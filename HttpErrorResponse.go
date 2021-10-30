package mgboot

type HttpErrorResponse struct {
	statusCode int
}

func NewHttpErrorResponse(statusCode int) HttpErrorResponse {
	return HttpErrorResponse{statusCode: statusCode}
}

func (p HttpErrorResponse) GetContentType() string {
	return ""
}

func (p HttpErrorResponse) GetContents() (int, string) {
	return p.statusCode, ""
}
