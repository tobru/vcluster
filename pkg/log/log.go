package log

type Logger interface {
	Infof(format string, a ...interface{})
	Warn(args ...interface{})
}
