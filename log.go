package main

type LogKey string

const (
	LogKeyRole      LogKey = "role"
	LogKeyHost             = "host"
	LogKeyReqID            = "req_id"
	LogKeyTimestamp        = "ts"
	LogKeyMsg              = "msg"
)
