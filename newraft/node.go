package newraft

import (
	"fmt"
	"slices"
)

type Role int

const (
	FOLLOWER Role = iota
	CANDIDATE
	LEADER
)

const TotalNodes = 5
const Quorum = TotalNodes/2 + 1

type Entry struct {
	Term  int
	Value string
	Index int
}

type Node struct {
	Id            string
	CurrentRole   Role
	OtherNodesId  []string //id of other nodes
	CurrentLeader string   //id of leader

	//timeouts
	HeartbeatTimeout         int //refers to the timeout to START an election because of a non recevied heartbeat
	ElectionTimeout          int //election timeout refers to the timout inside an election when the node is cnandidate
	SendAppendEntriesTimeout int

	//Must storage state
	CurrTerm int
	VotedFor string  //if id is EMPTY means i havent voted YET
	Log      []Entry //starts allways index = 1

	/*--VOLATILE DATA ------ */

	//All nodes volatile data
	CommitIndex int //starts always as 0, monotinc
	LastApplied int //starts always as 0, monotinc

	//leader data
	NextIndex   map[string]int //start each value as lastLogIndex + 1 (for this case could be the length of the Log,which always is lastlogindex +1 (becasue it started with index 1))
	MatchIindex map[string]int //Starts each value  with 0

	//candidate Data
	VotesGranted map[string]int



	SimulatorFields *SimulatorFields
}


//ONLY used in the simulator, but easy to put them here. 
type SimulatorFields struct{
	HeartbeatTimeoutCounter int //follower
	ElectionTimeoutCounter int  //candidate
	SendAppendEntriesTimeoutCounter int //leader
	Alive bool
}






/* LAST LOG INDEX del leader o de mi current user, +1 , y claro esto deberia estar es en el propio */
/*toma el last logIndex  */
func (n *Node) initializeLeaderIndexes() {

	lastLogIndex := len(n.Log) - 1

	for _, followerId := range n.OtherNodesId {
		n.NextIndex[followerId] = lastLogIndex + 1
		n.MatchIindex[followerId] = 0
	}

}


// updates term and updates the voted for, saves both in STORAGE
func (n *Node) updateTerm(newTerm int) {
	if newTerm <= n.CurrTerm {
		panic("something is wrong updating the term, recevied the new term same or smaller than current term")
	}
	n.CurrTerm = newTerm
	//TODO: save in storage
}

func (n *Node) HandleEvent(message Message) []Message {

	//Ignore the message, we can also send back to him, to update the term, This will also check that is not a timeout message (needed because timeouts send term also)
	if message.Term < n.CurrTerm && !isTimeoutMessage(message.Type){
		fmt.Println("mensaje ignorado por TERM")
		return nil
	}

	//if the message term is bigger, we convert to follower and update the TERM
	if message.Term > n.CurrTerm {
		n.updateTerm(message.Term)
		n.transitionRole(FOLLOWER)
	}

	switch n.CurrentRole {

	case FOLLOWER:
		return n.HandleFollowerAction(message)

	case CANDIDATE:
		return n.HandleCandidateAction(message)

	case LEADER:
		return n.HandleLeaderAction(message)

	default:
		panic("Not valid role wtf")

	}
}

func isTimeoutMessage(msgType MessageType) bool {
	if msgType == MsgSendAppendEntriesTimeout || msgType == MsgHeartbeatTimeout || msgType== MsgElectionTimeout || msgType == MsgNewEntry{
		return true
	}
	return false

}

func (n *Node) transitionRole(target Role) []Message {
	//aca podriamos poner un restart de todos los timeouts, porque no importa mucho la verdad 

	switch target {

	case FOLLOWER:
		n.VotedFor = ""
		n.CurrentRole = FOLLOWER
		//TODO: save them in storeage

	case CANDIDATE:
		if n.CurrentRole == LEADER {
			panic("Assertion, can nnot pass to candidate, from LEADER")
		}

		n.CurrentRole = CANDIDATE
		n.VotedFor = n.Id
		for key, _ := range n.VotesGranted{
			n.VotesGranted[key]=0
		}

		return n.sendRequestVote()

	case LEADER:
		if n.CurrentRole == FOLLOWER {
			panic("Assertion, can not pass to leader, from FOLLOWER")
		}

		n.initializeLeaderIndexes()
		n.CurrentRole = LEADER
		n.CurrentLeader = n.Id
		return n.SendAppendEntries()
		

	default:
		panic("UNKNOWN Role")

	}

	return nil
}
//data map comes from the matchindex or the voteGranted map, where the string is the ID (which we dont care), and the value is the value for the quorum check
func minQuorumValue(dataMap map[string]int) int {
	//here we dont need to check the slice because there isnt even a valid number of votes
	if len(dataMap) < Quorum{
		return 0
	}


	slice:= make([]int,0, len(dataMap))
	for _ , value:= range dataMap{
		slice = append(slice, value)

	}
	slices.Sort(slice)

	idx:= Quorum - 1
	return slice[idx]

}

