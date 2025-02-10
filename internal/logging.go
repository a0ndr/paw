package internal

import (
	"fmt"
	"os"
	"runtime"
)

type Logger struct {
	Debug bool
}

func (l *Logger) Print(str string) {
	caller := ""
	if l.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("(%s:%d) ", file, no)
		}
	}
	fmt.Printf("%s%s", caller, str)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.Print(fmt.Sprintf(format, args...))
}

func (l *Logger) Println(str string) {
	l.Print(str + "\n")
}

func (l *Logger) Error(str string) {
	caller := ""
	if l.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("(%s:%d) ", file, no)
		}
	}
	_, err := fmt.Fprintf(os.Stderr, "%s%s", caller, str)
	if err != nil {
		l.Print(str)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(code int, str string) {
	l.Error(str)
	os.Exit(code)
}

func (l *Logger) Fatalf(code int, format string, args ...interface{}) {
	l.Fatal(code, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.Debug {
		l.Printf(format+"\n", args...)
	}
}
