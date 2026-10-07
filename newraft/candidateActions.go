package newraft


/* ---- METHODS REACTING TO INCOMING EVENTS -------- */
func (n *Node)HandleCandidateAction(mesage Message) []Message{

	switch mesage.Type{
	
	case MsgRequestVoteResponse: 
		payload, ok:= mesage.Payload.(RequestVoteResponseEvent)
		if !ok{
			panic("casting error on append entries ")
		}
		return n.HandleResponseVote(payload, mesage.SenderId)

	case MsgElectionTimeout:
		return n.handleElectionTimeout()	
	}


	return nil
} 



func (n *Node) sendRequestVote() []Message{
	messages:= []Message{}

	for _ , votersId:= range n.OtherNodesId{
		lastLogIndex:= len(n.Log)-1
		lastLogTerm:= n.Log[lastLogIndex].Term

		eventPayload:= RequestVoteEvent{
			Term: n.CurrTerm,
			CandidateId: n.Id,
			LastLogIndex: lastLogIndex,
			LastLogTerm: lastLogTerm,
		}


		messages = append(messages, Message{
			SenderId: n.Id,
			Term: n.CurrTerm,
			ReceiverId: votersId,
			Type: MsgRequestVote,
			Payload: eventPayload,
		})

	}

	return messages
}


func (n *Node) HandleResponseVote(eventResponse RequestVoteResponseEvent, voterId string) []Message{
	/*checkear  */

	if eventResponse.VoteGranted{
		n.VotesGranted[voterId] = 1
	}

	//we havent REached our goal of votes YET
	if minQuorumValue(n.VotesGranted) != 1{
		return nil
	}


	//we reached the minimum Votes to be leader, we convert to leader, and return here the mesage
	return n.transitionRole(LEADER)
}


/* STarts a new election when the current one Timesout*/
func (n *Node) handleElectionTimeout() []Message{
	n.updateTerm(n.CurrTerm + 1)
	//TODO: update it storage
	return n.transitionRole(CANDIDATE)

}
