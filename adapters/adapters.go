package adapters

import "simba/newraft"

type Runner interface {
	Start()
	Stop()
}

type TransportAdapter interface {
	SendMessage([]newraft.Message)
}

type TimeAdapter interface {
	Now() int64
	Advance(int64)
	Sleep(int64)
}

type StorageAdapter interface {
	appendEntryLog()
}
