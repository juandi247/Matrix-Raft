package newraft

import "fmt"

func (n *Node) printEventInformation(msgType MessageType) {
	eventString:=""

	switch (msgType){

	case MsgNewEntry: 
		eventString= "New entry from outside"

	//---received by FOLLOWER-----
	case MsgAppendEntries:
		eventString= "Append Entries from a leadeer"
		
	case 	MsgRequestVote:
		eventString= "Request Vote from a Candidate"

	case 	MsgSendAppendEntriesTimeout:
		eventString= "AppendEntirsTimeout (im a leader, sending new append entries)"


	//---received by CANDIDATE-----
	case 	MsgRequestVoteResponse:
		eventString= " Request vote Response from a follower (i am candidate)"

	case 	MsgElectionTimeout :
		eventString= "Election timeout (im a candidate, going to start a new one)"


	//---received by LEADER-----

	case 	MsgAppendEntriesResponse:
		eventString= "APpend entries resopnse from a follower (i am leader)"

	case 	MsgHeartbeatTimeout:
		eventString="Heartbeatimeout (im a follower to start a election)"
	}

	fmt.Printf("Id: %s ProcessEvent: %s CurrentTerm: %v  \n\n", n.Id, eventString, n.CurrTerm)

}
