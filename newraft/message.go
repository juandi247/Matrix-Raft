package newraft


type MessageType int 

const(
	MsgNewEntry MessageType = iota
	MsgAppendEntries 
	MsgAppendEntriesResponse

	MsgRequestVote
	MsgRequestVoteResponse

	MsgElectionTimeout 
	MsgHeartbeatTimeout
	MsgSendAppendEntriesTimeout
)


type Message struct{
	SenderId string
	ReceiverId string
	Type MessageType
	Term int
	Payload EventPayload

}


type EventPayload interface{
	isEventPayload()
}


type FrontEndEntry struct{
	Magic string 
	//TODO : this entry will be later modified for some struct of instructions, but for now meh
	Entry string
}

/*-------------APPEND ENTRIES RPCS ------------------ */
type AppendEntriesEvent struct{
	Term int
	LeaderId string

	PrevLogIndex int
	PrevLogTerm int
	Entries []Entry
	LeaderCommitIndex int

}
type AppendEntriesResponseEvent struct{
	Term int
	Succes bool
	MatchIndex int //NEW data para poder decirle al lider hasta donde aplicar el matchindex, porque esto es un fire and forget
}


/*-------------VOTING RPCS ------------------ */
type RequestVoteEvent struct{
	Term int
	CandidateId string
	LastLogIndex int
	LastLogTerm int
}

type RequestVoteResponseEvent struct{
	Term int 
	VoteGranted bool
}


/* ------ Timeout ------------*/
type TimeoutEvent struct{

}

//Boilerplate because go deos not have taggd unions, so every event must be with this interface
func (a FrontEndEntry) isEventPayload(){}
func (a AppendEntriesEvent) isEventPayload(){}
func (a AppendEntriesResponseEvent) isEventPayload(){}
func (a RequestVoteEvent) isEventPayload(){}
func (a RequestVoteResponseEvent) isEventPayload(){}
func (a TimeoutEvent) isEventPayload(){}
