package mgboot

type ErrorHandler interface {
	GetErrorName() string
	MatchError(err error) bool
	HandleError(err error) ResponsePayload
}
