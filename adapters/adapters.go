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
	Now() uint64
	Advance(uint64)
	Sleep(uint64)
}

type StorageAdapter interface {
	appendEntryLog()
}
