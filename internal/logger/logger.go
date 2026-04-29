package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBold   = "\033[1m"
)

type formatter struct{}

func (f *formatter) Format(entry *log.Entry) ([]byte, error) {
	var buf bytes.Buffer

	ts := entry.Time.Format("01:00:00")
	fmt.Fprintf(&buf, "%s%s%s ", colorGray, ts, colorReset)

	levelColor, levelTag := levelStyle(entry.Level)
	fmt.Fprintf(&buf, "%s%s[%s]%s ", colorBold, levelColor, levelTag, colorReset)

	fmt.Fprintf(&buf, "%s", entry.Message)

	for k, v := range entry.Data {
		fmt.Fprintf(&buf, " %s%s%s=%v", colorCyan, k, colorReset, v)
	}

	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func levelStyle(level log.Level) (color, tag string) {
	switch level {
	case log.DebugLevel:
		return colorCyan, "DEBUG"
	case log.InfoLevel:
		return colorGreen, "INFO "
	case log.WarnLevel:
		return colorYellow, "WARN "
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		return colorRed, "ERROR"
	default:
		return colorReset, "?????"
	}
}

func Setup(debug, silent bool) {
	log.SetFormatter(&formatter{})
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	if silent {
		log.SetOutput(io.Discard)
	} else {
		log.SetOutput(os.Stdout)
	}
}
