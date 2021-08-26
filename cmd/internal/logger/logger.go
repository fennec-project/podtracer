package logger

type Level string

const (
	debug Level="DEBUG"
	info Level = "INFO"
)

type Logger struct {
	Level Level
}

func (logger *Logger) Log(format string, a  ...interface{}) {

	if Logger.

}  