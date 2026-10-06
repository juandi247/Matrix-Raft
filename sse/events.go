package sse

import (
	"simba/newraft"
)


type SseEvent interface {
	GetEventType() SimulatorEventType
}


type SimulatorEventType int
const (
	TickAdvance SimulatorEventType =  iota
	NodeCrashed 
	NodeBackToLife 	
	HeartbeatTimeout
	NodeStateUpdate
	SimulationMessagePushed //pushed means it was pushed into the network event array
	SimulationMessageDelivered //message was delivered to the node
)




type TickAdvanceEvent struct {
	Tick int
}

func NewTickAdvanceEvent(tick int) TickAdvanceEvent{
	return TickAdvanceEvent{
		Tick: tick,
	}
}



type NodeCrashedEvent struct {
	NodeId string
}

func NewNodeCrashedEvent(nodeId string) NodeCrashedEvent{
	return NodeCrashedEvent{
		NodeId:nodeId,
	}
}




type NodeBackToLifeEvent struct {
	NodeId string
}




type HeartbeatTimeoutEvent struct {
	NodeId                  string
	HeartbeatTimeoutCounter int
}

func NewHeartbeatTimeoutEvent(nodeId string, timeoutCounter int) HeartbeatTimeoutEvent{
	return HeartbeatTimeoutEvent{
		NodeId: nodeId,
		HeartbeatTimeoutCounter: timeoutCounter,
	}
}




type NodeStateUpdateEvent struct {
	NodeId string
	Term   int
	Role newraft.Role
	Log    []newraft.Entry

	CommitIndex                      int
	SimulatorHeartBeatTimeoutCounter int
	NextIndex map[string]int 
	MatchIndex map[string]int 
}


func NewNodeStateUpdateEvent(node *newraft.Node) NodeStateUpdateEvent{
	return NodeStateUpdateEvent{
		NodeId: node.Id,
		Term: node.CurrTerm,
		Role: node.CurrentRole,
		Log: node.Log,
		CommitIndex  :node.CommitIndex,             
		SimulatorHeartBeatTimeoutCounter: node.SimulatorFields.HeartbeatTimeoutCounter,
		NextIndex: node.NextIndex, 
		MatchIndex: node.MatchIindex,
	}
}







type SimMessage struct {
	Index int
	Id int
	DeliveryTick int
	Message      newraft.Message
}

// TODO: seria un array de mensajes con sus datos, o mejor un solo dato de mensaje? nose
type SimulationMessagePushedEvent struct {
	Messages []SimMessage 
}

func NewSimulationMessagesPushed(msgSlice []SimMessage) SimulationMessagePushedEvent{
	return SimulationMessagePushedEvent{
		Messages: msgSlice,
	}
}




type SimulationMessageDeliveredEvent struct {
	MessageId int
}

func NewSimulationMessageDelivered(messageId int) SimulationMessageDeliveredEvent{
	return SimulationMessageDeliveredEvent{
		MessageId: messageId,
	}
}




//boilerplate to implement the SSE interface
func (t TickAdvanceEvent) GetEventType() SimulatorEventType{ return TickAdvance}
func (t NodeCrashedEvent) GetEventType() SimulatorEventType{ return NodeCrashed}
func (t HeartbeatTimeoutEvent) GetEventType() SimulatorEventType{ return HeartbeatTimeout}
func (t NodeBackToLifeEvent) GetEventType() SimulatorEventType{ return NodeBackToLife}
func (t NodeStateUpdateEvent) GetEventType() SimulatorEventType{ return NodeStateUpdate}
func (t SimulationMessagePushedEvent) GetEventType() SimulatorEventType{ return SimulationMessagePushed}
func (t SimulationMessageDeliveredEvent) GetEventType() SimulatorEventType{ return SimulationMessageDelivered}


