package logs

// Logger interface represents the logs methods you want to use in your application.
type Logger interface {
	Info(args ...any)
	Error(args ...any)
	WithFields(fields map[string]any) Logger
}
