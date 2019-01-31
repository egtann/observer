package observer

const (
	LogKeyRole      = "role"
	LogKeyHost      = "host"
	LogKeyReqID     = "req_id"
	LogKeyTimestamp = "ts"
	LogKeyMsg       = "msg"
)

type Logger interface {
	Printf(msg string, vals ...interface{})
}

// Start recording an event.
func Start(l Logger, event string) {
	l.Printf("start: %s", event)
}

// End event recording.
func End(l Logger, event string) {
	l.Printf("end: %s", event)
}
