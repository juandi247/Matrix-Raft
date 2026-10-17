package newraft

import "fmt"

/* ---- METHODS REACTING TO INCOMING EVENTS -------- */
func (n *Node)HandleLeaderAction(mesage Message) []Message{

	switch mesage.Type{
	
	case MsgNewEntry: 
		payload, ok:= mesage.Payload.(FrontEndEntry)
		if !ok{
			panic("casting error on newEntry ")
		}
		return n.HandleNewEntry(payload.Entry)

	case MsgAppendEntriesResponse: 
		payload, ok:= mesage.Payload.(AppendEntriesResponseEvent)
		if !ok{
			panic("casting error on  AppendEntriesResponseEvent")
		}
		return n.HandleAppendEntriesResponse(payload, mesage.SenderId)

	case MsgSendAppendEntriesTimeout:
		return n.handleSendAppendEntriesTimeout()
	}

	return nil
} 

//LEADER ONLY
func (n *Node) HandleNewEntry(data string)[]Message{

	switch (n.CurrentRole){
	
	//he should report to the other side, here is he leader
	case FOLLOWER: 
	panic("esto no ebderia ocurr")
	//he should tell them, hey for security resons we are not ready, we are on an election, safety first
	case CANDIDATE: 
	panic("esto no ebderia ocurr tampoco")

	case LEADER:
		idx:= len(n.Log)
		n.Log = append(n.Log, Entry{ Term: n.CurrTerm, Value: data, Index: idx})

		return append(n.SendAppendEntries(),Message{
				SenderId: n.Id,
				ReceiverId: "CLIENT",
				Type: MsgLeaderCheck,
				Payload: 	LeaderCheck{LeaderId: n.CurrentLeader}})

	default: 
		panic("invalid state of the leader")
	}

}



func (n *Node) buildAppendEntry(nextIndex int, followerId string) AppendEntriesEvent{
		prevLogIndex:= nextIndex - 1
	if prevLogIndex == -1{
		panic("pvlogInedx es -1, paniccc")
	}
		if prevLogIndex >= len(n.Log){
		fmt.Printf("followerId: %s \n", followerId)
		fmt.Printf("prevlogIndex: %d, logLength: %v \n", prevLogIndex, len(n.Log))
		panic("prevlog index is equal to the length of the log")

		}


		prevLogTerm:= n.Log[prevLogIndex].Term
		
		entries:= []Entry{}

		/* if the length of the log - 1, is bigger that the prevLogIndex, means that there is new entries, if not we would be out of bounds.
		Example
		Log[Null]

		NextIndex = 1
		PrevLogIndex = 0 

 		We would send from Index 0 + 1, but in this case Log[1:] is out of bounds
		*/
		if len(n.Log) - 1 > prevLogIndex{
			entries = append(entries, n.Log[prevLogIndex+1:]...)
		}

		return AppendEntriesEvent{
			Term: n.CurrTerm,
			LeaderId: n.Id,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm: prevLogTerm,
			Entries: entries,
			LeaderCommitIndex: n.CommitIndex,

		}


}
func (n *Node) SendAppendEntries() []Message{

	if n.CurrentRole != LEADER{
		panic("trying to send an append entry from a non leader node")
	}

	messages:= []Message{}




	for _ , followerId := range n.OtherNodesId{
		nextIndex:= n.NextIndex[followerId]
		messages = append(messages, Message{
			SenderId: n.Id,
			ReceiverId: followerId ,
			Term: n.CurrTerm,
			Type: MsgAppendEntries,
			Payload: n.buildAppendEntry(nextIndex, followerId),

		})

	}

	return messages

}

//this method is ONLY for the leader, since is just n evnet loop, and there is no direct RPC that waits for the resopnse from AppendEntries follower respose same thing should be on the RPC from start election

func (n *Node) HandleAppendEntriesResponse(requestEvent AppendEntriesResponseEvent, followerId string) []Message{

	if !requestEvent.Succes{
		messages := []Message{}

		n.NextIndex[followerId] --

		//try again, with a lower NEXT index, to update it quickly
		return append(messages, Message{
			SenderId: n.Id,
			Term: n.CurrTerm,
			ReceiverId: followerId ,
			Type: MsgAppendEntries,
			Payload: n.buildAppendEntry(n.NextIndex[followerId], followerId),

		})
	}



	n.MatchIindex[followerId] = max(n.MatchIindex[followerId], requestEvent.MatchIndex)
	n.NextIndex[followerId] = n.MatchIindex[followerId]+1

	n.CommitIndex = max(n.CommitIndex, minQuorumValue(n.MatchIindex, Quorum))


	//TODO: checkear last applied, pero hasta ahi estamos BIEN MELOS
	return nil
}



//triggered when the timeout of the leader ocurs, to send the append entries to the followers
func (n *Node) handleSendAppendEntriesTimeout() []Message{
	return n.SendAppendEntries()
}
