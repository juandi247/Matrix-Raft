package simulator

import (
	"container/heap"
	"fmt"
	"simba/newraft"
	"simba/sse"
	"strconv"
	"sync"
	"time"
)
 
const (
    StateRunning = iota
    StatePaused
)

type Wctx struct {
    Mu    sync.Mutex
    State int
}

func (w *Wctx) SetState(state int) {
    w.Mu.Lock()
    defer w.Mu.Unlock()
    w.State = state
}

func (w *Wctx) GetState() int {
    w.Mu.Lock()
    defer w.Mu.Unlock()
    return w.State
}

type SimulationRunner struct {
	Time               *SimTime
	Network            *SimNetwork
	FuzzyProbabilities FuzzyConfig
	ShouldPublishEvents bool
	EventChannel chan sse.SseEvent
	PauseContext Wctx  
}


func (s *SimulationRunner) Start() {

	// Config for the simulated Time struct
	s.Time.Tick = 0
	s.Network.IdCounter = 0	
	s.Network.EventChan = s.EventChannel
	s.Network.simClient = SimClient{CachedLeaderId: ""}
	s.Network.ShouldPublishEvent = s.ShouldPublishEvents

	// Config for the simulated Network struct
	s.Network.TimeAdapter = s.Time

	pq:= make(PriorityQueue, 0)
	heap.Init(&pq)
	s.Network.messageQueue = &pq

	s.Network.FuzzyConfig = s.FuzzyProbabilities

	//This is all intiial configuration preivous to the FOR loop that ocntains the running engine steps
	nodeList := initializeNodes(s.FuzzyProbabilities)

	for _ , node := range nodeList{
		PublishEvent(s.ShouldPublishEvents, s.EventChannel, sse.NewNodeStateUpdateEvent(node))
		// s.EventChannel <-sse.NewNodeStateUpdateEvent(node)
	}

	s.PauseContext= Wctx{}


	fmt.Println("Configuration finished. Starting loop")
	for s.Time.Now() <= maxTicks {
	

		switch state:= s.PauseContext.GetState(); state{

		case StatePaused: 
			time.Sleep(300 * time.Millisecond)

		default: 

		//SIMULATION ENGINE
		s.Time.Advance(TickFrequency)
		fmt.Println("Tick: ", s.Time.Now())

		PublishEvent(s.ShouldPublishEvents, s.EventChannel, sse.NewTickAdvanceEvent(int(s.Time.Now())))


		//crashNodes(nodeList, s.FuzzyProbabilities, s.Time.Now())

		updateNodeTimers(nodeList, s.EventChannel, s.ShouldPublishEvents)

		//handleComeBackToLiveNode(nodeList, s.Time.Now())

		handleTimeout(nodeList, s.Network)
		checkInvariants(nodeList)

		//NOTE: handles the  client re1uests, esto podria ir en una funcoin extra
		req:= s.Network.simClient.generateClientRequest(int(s.Network.TimeAdapter.Now()), s.Network.FuzzyConfig.rand)
		if req!=nil{
			s.Network.SendMessage(req)
		}

		//this is ONLY to read the queue and put the messages into the inbox. No logic of delivering messages to any node here.
		if s.Network.messageQueue.Len() > 0 {
			readMessagesToInbox(s.Network, nodeList, s.EventChannel, s.ShouldPublishEvents)
		}

		time.Sleep(200* time.Millisecond)

		}


	}
}

func (s *SimulationRunner) Stop() {
}

/* LAST LOG INDEX del leader o de mi current user, +1 , y claro esto deberia estar es en el propio */
func initializeNextIndex(nodesNumber, id int) map[int]int{
	mapita:= make(map[int]int, nodesNumber-1)

	for i:=1; i<=nodesNumber; i++{
		if i==id{
			continue
		}
		mapita[i]=1
	}

	return mapita
}

func initializeMatchIndex(nodesNumber, id int) map[int]int{
	mapita:= make(map[int]int, nodesNumber-1)

	for i:=1; i<=nodesNumber; i++{
		if i==id{
			continue
		}
		mapita[i]=0
	}

	return mapita
}




func initializeNodes(fuzzyProbabilites FuzzyConfig) []*newraft.Node {
	nodeList:= make([]*newraft.Node, newraft.TotalNodes)

	for i := 1; i <= int(newraft.TotalNodes); i++ {


		heartbeatTimeout:= int(generateHeartbeatTimeout(fuzzyProbabilites.rand))
		//ESTO ES TEMPORALLL!!!
		if i==1{
			heartbeatTimeout=2
		}
		id:= "Node"+strconv.Itoa(i)

		nodeList[i-1] = &newraft.Node{
			Id: id,
			CurrentRole: newraft.FOLLOWER,
			OtherNodesId: buildFriendsIds(newraft.TotalNodes, i),
			CurrentLeader: "",


			//TODO: check this instead of cero would be the normal timoeutj
			HeartbeatTimeout: heartbeatTimeout,
			ElectionTimeout: ElectionTimeout,
			SendAppendEntriesTimeout: SendAppendEntriesFreq,

			CurrTerm: 0,
			VotedFor: "",
			Log: []newraft.Entry{
				{Term: 0, Value: "SKIPPER", Index: 0},
			},

			CommitIndex: 0,
			LastApplied: 0, 

			NextIndex: make(map[string]int),
			MatchIindex: make(map[string]int), 
			VotesGranted: make(map[string]int), 

			//TODO: Simulator fields, que son importantes solo para los ticks de reducir los teimotus nada mas
			
			SimulatorFields: &newraft.SimulatorFields{
				HeartbeatTimeoutCounter: heartbeatTimeout,
				ElectionTimeoutCounter: ElectionTimeout,
				SendAppendEntriesTimeoutCounter: SendAppendEntriesFreq,
				Alive: true,
			},
		}

	}

	return nodeList

}

func crashNodes(nodeList []*newraft.Node, fuzzyProbabilites FuzzyConfig, currentTick int64) {
	for _, node := range nodeList {
		//TODO: uncomment this because its not ACTIVE
		//shouldCrash, comeBackToLiveTick := fuzzyProbabilites.determineCrashingProbabily()
		shouldCrash:=false
		if !shouldCrash {
			continue
		}
		node.SimulatorFields.Alive = false
		//TODO: uncomment thisss
		//node.SimulatorFields.ComeBackToLiveTick = currentTick + comeBackToLiveTick
	}
}
 

func updateNodeTimers(nodeList []*newraft.Node, eventChan chan sse.SseEvent, shouldPublishEvent bool) {
	for _, node := range nodeList {
		switch node.CurrentRole {
		case newraft.FOLLOWER:
			node.SimulatorFields.HeartbeatTimeoutCounter--
			
			PublishEvent(shouldPublishEvent, eventChan, sse.NewHeartbeatTimeoutEvent(node.Id,node.SimulatorFields.HeartbeatTimeoutCounter))
			// eventChan <- sse.NewHeartbeatTimeoutEvent(node.Id, node.SimulatorFields.HeartbeatTimeoutCounter)
			
		case newraft.CANDIDATE:
			node.SimulatorFields.ElectionTimeoutCounter--
		case newraft.LEADER:
			node.SimulatorFields.SendAppendEntriesTimeoutCounter--
		default:
			panic("a node does not have a valid role")
		}

	}
	
}
/*
func handleComeBackToLiveNode(nodeList []*newraft.Node, currentTick int64) {

	for _, node := range nodeList {
		if node.SimulatorFields.ComeBackToLiveTick <= currentTick && !node.SimulatorFields.Alive {
			node.SimulatorFields.Alive = true
			//This is to reestart the values of timeouts, so that the node starts Cleanly from scratch.
			node.SimulatorFields.LeaderHeartbeatCounter = node.LeaderHeartbeat
			node.SimulatorFields.Timeoutcounter = node.Timeout
		}

	}
}*/

func readMessagesToInbox(sn *SimNetwork, nodeList []*newraft.Node, eventChan chan sse.SseEvent, shouldPublishEvent bool) {

	if sn.messageQueue.Len()<=0{
		panic("wtf this hsuold be bigger than cero")
	}

	for i:=0; i<sn.messageQueue.Len(); i++{
		msg:= sn.messageQueue.Peek()
		
		if msg.DeliveryTick > int(sn.TimeAdapter.Now()) {
			return
		}
		//ESTE POP Es lo que me jode no?
		//msg = heap.Pop(sn.messageQueue).(*sse.SimMessage)
		heap.Pop(sn.messageQueue)

/*
		if _, isEntry := msg.Message.(raft.NewEntry); isEntry{
			 //leaderId:= checkLeader(nodeList)
			leaderId := "Node1"
			//if leaderId==0{
			fmt.Println("there is no current LEADER to respond this message")
				continue
			}
			receiverNodeID = leaderId


		}else{
			receiverNodeID= msg.Message.GetReceiver() 
		}
*/ 

		receiverNodeId:= msg.Message.ReceiverId


		if receiverNodeId == ""{
			panic("receiver id was empty")
		}




		//NOTE: CHECK IF THE MESSAGE GOES BACK TO A CLIENT
		if receiverNodeId == ClientId{
			sn.simClient.handleIncomingMessage(*msg)
			return
		}

		
		 for _ , node:=range nodeList{
			if node.Id == receiverNodeId{

				if node.Id == "Node5" || node.Id == "Node1"{
				}
				PublishEvent(shouldPublishEvent, eventChan, sse.NewSimulationMessageDelivered(msg.Id))
				// eventChan <- sse.NewSimulationMessageDelivered(msg.Id)
				responseMessages:=node.HandleEvent(msg.Message)

				//NOTE: checks split brain and that stuff
				checkInvariants(nodeList)

				PublishEvent(shouldPublishEvent, eventChan, sse.NewNodeStateUpdateEvent(node))
				// eventChan <- sse.NewNodeStateUpdateEvent(node)

				sn.SendMessage(responseMessages)
				break
			}
		}
		
	}

}

/*
func checkLeader(nodeList []*newraft.Node) int{
	leader:=0
	maxTerm:=0
	for _ , n :=range nodeList{
		if n.Role == newraft.LEADER && n.CurrentTerm > uint64(maxTerm){
			leader= n.Id
		} 
	}
	return leader
}
*/


/*
ACA ya se habran reducido los tiks por nodo. por lo tanto lo unico seria validar el teimpo no?
*/
func handleTimeout(nodeList []*newraft.Node, sm *SimNetwork) {

	/*TODO: aca tendria que poner esto al principio de la queue de cada NODO, si o si, asi es trigereado el timeout*/
	for _, node := range nodeList {
		if !node.SimulatorFields.Alive {
			continue
		}

		switch node.CurrentRole {

		case newraft.FOLLOWER:
			if node.SimulatorFields.HeartbeatTimeoutCounter <= 0 {
				node.SimulatorFields.HeartbeatTimeoutCounter = node.HeartbeatTimeout
				sm.SendTimeout(newraft.Message{
					Type: newraft.MsgHeartbeatTimeout,
					ReceiverId: node.Id,
					})
			}


		case newraft.CANDIDATE:
			if node.SimulatorFields.ElectionTimeoutCounter <= 0 {
				node.SimulatorFields.ElectionTimeoutCounter = node.ElectionTimeout
				sm.SendTimeout(newraft.Message{
					Type: newraft.MsgElectionTimeout,
					ReceiverId: node.Id,
				})

			}

		case newraft.LEADER:
			if node.SimulatorFields.SendAppendEntriesTimeoutCounter <= 0 {
				node.SimulatorFields.SendAppendEntriesTimeoutCounter = node.SendAppendEntriesTimeout
				sm.SendTimeout(newraft.Message{
					Type: newraft.MsgSendAppendEntriesTimeout,
					ReceiverId: node.Id,
				})
			}
		}
	}
}

func  PublishEvent(shouldPublishEvent bool,eventChan chan sse.SseEvent, event sse.SseEvent){
	if shouldPublishEvent{
		eventChan<- event
	}
}
