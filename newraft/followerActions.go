package newraft


/* ---- METHODS REACTING TO INCOMING EVENTS -------- */
func (n *Node)HandleFollowerAction(mesage Message) []Message{

	switch mesage.Type{
	
	case MsgAppendEntries: 
		payload, ok:= mesage.Payload.(AppendEntriesEvent)
		if !ok{
			panic("casting error on append entries ")
		}
		return n.HandleAppendEntries(payload)

	case MsgRequestVote: 
		payload, ok:= mesage.Payload.(RequestVoteEvent)
		if !ok{
			panic("casting error on SendRequestVoteResponse ")
		}
		return n.HandleRequestVote(payload)

	case MsgHeartbeatTimeout: 
		return n.handleHeartbeatTimeout()

	case MsgNewEntry: 
		return []Message{
			{
				SenderId: n.Id,
				ReceiverId: "CLIENT",
				Term: n.CurrTerm,
				Type: MsgLeaderCheck,
				Payload: 	LeaderCheck{LeaderId: n.CurrentLeader}},
			
		}
	}



	return nil
} 


func (n *Node) HandleAppendEntries(requestEvent AppendEntriesEvent) []Message{

	//this is a trick just for the simulated stuff
	if n.SimulatorFields!=nil{
		n.SimulatorFields.HeartbeatTimeoutCounter = n.HeartbeatTimeout
	}
	//TODO: aca deberia reiniciar el timer o ticket de timeout, o algo (incluso podria ir en el mensaje de reiniciar para fire and forger)
	n.CurrentLeader = requestEvent.LeaderId

	messages:= []Message{}
	responseMessage:= Message{
		SenderId: n.Id,
		Term: n.CurrTerm,
		ReceiverId: requestEvent.LeaderId,
		Type: MsgAppendEntriesResponse,

	}

	responseEvent:= AppendEntriesResponseEvent{
		Term: n.CurrTerm,
		Succes: false,
	}


	followerLastIndex:= len(n.Log) - 1 

	//if log does not contain an entry on the prevLogIndex, automatic error, this makes sure we have something in the same index  
	if requestEvent.PrevLogIndex > followerLastIndex{
		responseMessage.Payload = responseEvent
		return append(messages, responseMessage)

	}

	
	//if the Term does not match, we have a problem
	if requestEvent.PrevLogTerm != n.Log[requestEvent.PrevLogIndex].Term{
		responseMessage.Payload = responseEvent
		return append(messages, responseMessage)

	}


	//HAPPY PATH
	responseEvent.Succes = true

	for index, entry:= range requestEvent.Entries{
		if entry.Index > len(n.Log) - 1{
			n.Log = append(n.Log, requestEvent.Entries[index:]...)
			break
		}
	}
	//TODO: SAVE IN STORAGE porque debe estar ya appendeado



	responseEvent.MatchIndex = len(n.Log)-1

	if requestEvent.LeaderCommitIndex > n.CommitIndex{
		n.CommitIndex = min(requestEvent.LeaderCommitIndex, len(n.Log) -1)
		/*TODO:  aca tambien deberia ir esto If commitIndex > lastApplied: increment lastApplied, apply
log[lastApplied] to state machine (§5.3), que signficaria que deberiamos hacer o aplciar los datos a la state machine*/
	}


	responseMessage.Payload = responseEvent

	return append(messages, responseMessage)

}


func (n *Node) HandleRequestVote(requestEvent RequestVoteEvent)[]Message{

	messages:= []Message{}
	responseMessage:= Message{
		SenderId: n.Id,
		Term: n.CurrTerm,
		ReceiverId: requestEvent.CandidateId,
		Type: MsgRequestVoteResponse,

	}

	responseEventPayload:= RequestVoteResponseEvent{
		Term: n.CurrTerm,
		VoteGranted: false,
	}

	//NOTE: RULES TO NOT Grant any vote  
	if requestEvent.Term < n.CurrTerm{
		responseMessage.Payload = responseEventPayload
		return append(messages, responseMessage)
	}

	//MEans that i ALREADY voted for someone in this TERM
	if len(n.VotedFor)>0{
		responseMessage.Payload = responseEventPayload
		return append(messages, responseMessage)
	}


	followerLastLogIndex := len(n.Log) -1 
	followerLastLogTerm := n.Log[followerLastLogIndex].Term
	if requestEvent.LastLogTerm < followerLastLogTerm{
		responseMessage.Payload = responseEventPayload
		return append(messages, responseMessage)
	}


	if requestEvent.LastLogTerm == followerLastLogIndex{
		if requestEvent.LastLogIndex < followerLastLogIndex{
		responseMessage.Payload = responseEventPayload
		return append(messages, responseMessage)
		}

	}


	/* ----HAPPY PATH --- */
	n.VotedFor = requestEvent.CandidateId
	//TODO: this shouold be stored in storage

	responseEventPayload.VoteGranted=true
	responseMessage.Payload = responseEventPayload
	return append(messages, responseMessage)

}



//Starts an election as soon as the transition ocurres
func (n *Node) handleHeartbeatTimeout()[]Message{
	n.updateTerm(n.CurrTerm+1)
	//TODO: update term on storage
	return n.transitionRole(CANDIDATE)
}



func (n *Node) checkEntryExists(index int) bool{

	if index > len(n.Log) - 1{
		return false
	}

	return true
}

func (n *Node) checkEntryTerm(index int) bool{

	if index > len(n.Log) - 1{
		return false
	}

	return true
}

