package logx

import "sync"

type writer struct {
	appenders []appender
}

func (w *writer) Write(buf []byte) (int, error) {
	if len(w.appenders) < 1 {
		return len(buf), nil
	}

	if len(w.appenders) == 1 {
		return w.appenders[0].Write(buf)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(w.appenders))

	for _, a := range w.appenders {
		go func(a appender) {
			defer wg.Done()
			_, _ = a.Write(buf)
		}(a)
	}

	return len(buf), nil
}
