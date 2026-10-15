package server

import (
	"simba/simulator"

	"github.com/google/uuid"
)

type SimulatorSessionManager struct {
	simulatorSessions map[uuid.UUID]*simulator.SimulationRunner
}

func (s *SimulatorSessionManager) cleanupSession(uid uuid.UUID) {
	delete(s.simulatorSessions, uid)
}
