package internal

import (
	"fmt"
	"github.com/fatih/color"
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
	caller := ""
	if l.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("(%s:%d) ", file, no)
		}
	}
	fmt.Printf("%s%s", caller, fmt.Sprintf(format, args...))
}

func (l *Logger) Println(str string) {
	caller := ""
	if l.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("(%s:%d) ", file, no)
		}
	}
	fmt.Printf("%s%s", caller, str+"\n")
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
	caller := ""
	if l.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("(%s:%d) ", file, no)
		}
	}
	_, err := fmt.Fprintf(os.Stderr, "%s%s", caller, fmt.Sprintf(format, args...))
	if err != nil {
		l.Print(fmt.Sprintf(format, args...))
	}
}

func (l *Logger) Fatal(code int, str string) {
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
	os.Exit(code)
}

func (l *Logger) Fatalf(code int, format string, args ...interface{}) {
	caller := ""
	if l.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			caller = fmt.Sprintf("(%s:%d) ", file, no)
		}
	}
	_, err := fmt.Fprintf(os.Stderr, "%s%s", caller, fmt.Sprintf(format, args...))
	if err != nil {
		l.Print(fmt.Sprintf(format, args...))
	}
	os.Exit(code)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.Debug {
		caller := ""
		if l.Debug {
			_, file, no, ok := runtime.Caller(1)
			if ok {
				caller = fmt.Sprintf("(%s:%d) ", file, no)
			}
		}
		fmt.Printf("%s%s\n", caller, fmt.Sprintf(format, args...))
	}
}

func Logf(format string, args ...interface{}) {
	_, _ = color.New(color.FgYellow).Add(color.Bold).Print("==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}

func Log2f(format string, args ...interface{}) {
	_, _ = color.New(color.FgCyan).Add(color.Bold).Print("  ==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}

func Errorf(format string, args ...interface{}) {
	_, _ = color.New(color.FgRed).Add(color.Bold).Print("==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}

func Error2f(format string, args ...interface{}) {
	_, _ = color.New(color.FgRed).Add(color.Bold).Print("  ==> ")
	_, _ = color.New(color.Bold).Printf(format+"\n", args...)
}
