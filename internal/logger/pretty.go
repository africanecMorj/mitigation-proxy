package logger

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
)

type PrettyLogger struct {
	counter uint64
	Dev bool
}

func NewPretty(debug bool) Logger {
	return &PrettyLogger{
		Dev: debug,		
	}
}

func (l *PrettyLogger) Info(
	msg string,
	fields map[string]interface{},
) {
	l.print("INFO", msg, fields)
}

func (l *PrettyLogger) Debug(
	msg string,
	fields map[string]interface{},	
) {
	if l.Dev {
		l.print("DEBUG", msg, fields)
	}
}

func (l *PrettyLogger) Error(
	msg string,
	fields map[string]interface{},
) {
	l.print("ERROR", msg, fields)
}

func (l *PrettyLogger) print(
	level string,
	msg string,
	fields map[string]interface{},
) {

	now := time.Now().Format("15:04:05.000")

	fmt.Printf(
		"%s %s %s\n",
		text.FgHiBlack.Sprintf(now),
		text.FgGreen.Sprintf(level),
		msg,
	)

	for k, v := range fields {
		fmt.Printf(
			"    %s=%v\n",
			text.FgYellow.Sprintf(k),
			v,
		)
	}

	fmt.Println()
}