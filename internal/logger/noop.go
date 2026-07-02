package logger

type NoopLogger struct{}

func NewNoop(_ bool) Logger {
	return &NoopLogger{}
}

func (_ *NoopLogger) Info(_ string,_ map[string]interface{},) {}


func (_ *NoopLogger) Error(_ string, _ map[string]interface{},) {}

func (_ *NoopLogger) Debug(_ string, _ map[string]interface{},) {}

