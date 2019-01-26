package observer

type LogKey string

const (
	LogKeyRole      LogKey = "role"
	LogKeyHost             = "host"
	LogKeyReqID            = "request_id"
	LogKeyTimestamp        = "ts"
	LogKeyMsg              = "msg"
)
