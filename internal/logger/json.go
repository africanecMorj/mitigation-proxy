package logger

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONLogger struct{
	Dev bool
}

func New(debug bool) Logger {
	return &JSONLogger{
		Dev: debug,
	}
}

func (j *JSONLogger) Info(
	msg string,
	fields map[string]interface{},
) {
	j.write("INFO", msg, fields)
}


func (j *JSONLogger) Error(
	msg string,
	fields map[string]interface{},
) {
	j.write("ERROR", msg, fields)
}

func (j *JSONLogger) Debug(
	msg string,
	fields map[string]interface{},	
) {
	if j.Dev {
		j.write("DEBUG", msg, fields)
	}
}

func (j *JSONLogger) write(
	level string,
	msg string,
	fields map[string]interface{},
) {

	entry := map[string]interface{}{
		"level": level,
		"message": msg,
		"fields": fields,
	}

	b, _ := json.Marshal(entry)

	fmt.Fprintln(os.Stdout, string(b))
}

