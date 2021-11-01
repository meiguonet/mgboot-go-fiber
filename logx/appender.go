package logx

type appender interface {
	GetAppenderName() string
	Write(buf []byte) (int, error)
}
