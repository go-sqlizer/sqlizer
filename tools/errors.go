package tools

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

func CaptureError(err *error, msg ...interface{}) {
	if r := recover(); r != nil {
		*err = Error(msg...)
		if os.Getenv("DB_DEBUG") == "true" {
			fmt.Println(err, r, "\n", string(debug.Stack()))
		}
	}
}

func Error(msg ...interface{}) error {
	var err []string
	err = append(err, "sqlizer:")

	for _, v := range msg {
		err = append(err, reflect.TypeOf(v).String())

	}

	return errors.New(strings.Join(err, " "))
}
