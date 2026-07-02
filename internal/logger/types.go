package logger

type Logger interface {

	Info(
		msg string,
		fields map[string]interface{},
	)

	Debug(
		msg string,
		fields map[string]interface{},
	)

	Error(
		msg string,
		fields map[string]interface{},
	)
}