package mgboot

type ResponsePayload interface {
	GetContentType() string
	GetContents() (int, string)
}
