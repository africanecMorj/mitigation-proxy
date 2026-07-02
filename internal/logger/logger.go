package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

var counter uint64

type Entry struct {
	ID      uint64                 `json:"id"`
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

func NextID() uint64 {
	return atomic.AddUint64(&counter, 1)
}

func Log(level, msg string, fields map[string]interface{}) {
	entry := Entry{
		ID:      NextID(),
		Time:    time.Now().UTC().Format(time.RFC3339Nano),
		Level:   level,
		Message: msg,
		Fields:  fields,
	}

	data, _ := json.Marshal(entry)

	fmt.Fprintln(os.Stdout, string(data))
}

func Info(msg string, fields map[string]interface{}) {
	Log("INFO", msg, fields)
}

func Error(msg string, fields map[string]interface{}) {
	Log("ERROR", msg, fields)
}

var connectionID uint64

func NextConnectionID() uint64 {
	return atomic.AddUint64(&connectionID, 1)
}