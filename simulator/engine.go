package simulator

import (
	"container/heap"
	"fmt"
	"simba/newraft"
	"simba/sse"
	"strconv"
	"time"
)
 
type SimulationRunner struct {
	Time               *SimTime
	Network            *SimNetwork
	FuzzyProbabilities FuzzyConfig
	Port               string
	IsHttps            bool
	LeaderId 	int
}

func (s *SimulationRunner) Start(eventChan chan sse.SseEvent) {

	// Config for the simulated Time struct
	s.Time.Tick = 0
	s.Network.IdCounter = 0	
	s.Network.EventChan = eventChan
	s.Network.simClient = SimClient{CachedLeaderId: ""}

	// Config for the simulated Network struct
	s.Network.TimeAdapter = s.Time

	pq:= make(PriorityQueue, 0)
	heap.Init(&pq)
	s.Network.messageQueue = &pq

	s.Network.FuzzyConfig = s.FuzzyProbabilities

	//This is all intiial configuration preivous to the FOR loop that ocntains the running engine steps
	nodeList := initializeNodes(s.FuzzyProbabilities)

	for _ , node := range nodeList{
		eventChan <-sse.NewNodeStateUpdateEvent(node)
	}

	//requests:= GenerateRequests(s.FuzzyProbabilities.rand)

	fmt.Println("Configuration finished. Starting loop")
	// Engine Loop of execution
	for s.Time.Now() <= maxTicks {

		// advance 1 tick
		s.Time.Advance(TickFrequency)
		eventChan <- sse.NewTickAdvanceEvent(int(s.Time.Now()))


		fmt.Printf("Starting tick:  %v \n", s.Time.Now())
		//crashNodes(nodeList, s.FuzzyProbabilities, s.Time.Now())

		updateNodeTimers(nodeList, eventChan)

		//handleComeBackToLiveNode(nodeList, s.Time.Now())

		handleTimeout(nodeList, s.Network)

		//NOTE: handles the  client re1uests, esto podria ir en una funcoin extra
		req:= s.Network.simClient.generateClientRequest(int(s.Network.TimeAdapter.Now()), s.Network.FuzzyConfig.rand)
		if req!=nil{
			fmt.Println("ENVIANDO nueva ENTRY")
		s.Network.SendMessage(req)
		}

		//this is ONLY to read the queue and put the messages into the inbox. No logic of delivering messages to any node here.
		if s.Network.messageQueue.Len() > 0 {
			readMessagesToInbox(s.Network, nodeList, eventChan)
		}

		time.Sleep(200* time.Millisecond)

		
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
				newraft.Entry{Term: 0, Value: "SKIPPER", Index: 0},
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
		fmt.Println("crashed node: ", node.Id)
		//TODO: uncomment thisss
		//node.SimulatorFields.ComeBackToLiveTick = currentTick + comeBackToLiveTick
	}
}
 

func updateNodeTimers(nodeList []*newraft.Node, eventChan chan sse.SseEvent) {
	for _, node := range nodeList {
		switch node.CurrentRole {
		case newraft.FOLLOWER:
			node.SimulatorFields.HeartbeatTimeoutCounter--
			eventChan <- sse.NewHeartbeatTimeoutEvent(node.Id, node.SimulatorFields.HeartbeatTimeoutCounter)
			
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

func readMessagesToInbox(sn *SimNetwork, nodeList []*newraft.Node, eventChan chan sse.SseEvent) {

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
				fmt.Printf("%v VA a hacer HANLDE EVENT de mensaje desde: %v con deliveryTick: %v\n", node.Id, msg.Message.SenderId, msg.DeliveryTick)
				}
				eventChan <- sse.NewSimulationMessageDelivered(msg.Id)
				responseMessages:=node.HandleEvent(msg.Message)
				eventChan <- sse.NewNodeStateUpdateEvent(node)

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
				fmt.Printf("we reached a timeout follower id: %v, this should trigger a election \n", node.Id)
				node.SimulatorFields.HeartbeatTimeoutCounter = node.HeartbeatTimeout
				sm.SendTimeout(newraft.Message{
					Type: newraft.MsgHeartbeatTimeout,
					ReceiverId: node.Id,
					})
			}


		case newraft.CANDIDATE:
			if node.SimulatorFields.ElectionTimeoutCounter <= 0 {
				fmt.Println("we reached a timeout candidate")
				node.SimulatorFields.ElectionTimeoutCounter = node.ElectionTimeout
				sm.SendTimeout(newraft.Message{
					Type: newraft.MsgElectionTimeout,
					ReceiverId: node.Id,
				})

			}

		case newraft.LEADER:
			if node.SimulatorFields.SendAppendEntriesTimeoutCounter <= 0 {
				fmt.Println("we reached a timeout leader")
				node.SimulatorFields.SendAppendEntriesTimeoutCounter = node.SendAppendEntriesTimeout
				sm.SendTimeout(newraft.Message{
					Type: newraft.MsgSendAppendEntriesTimeout,
					ReceiverId: node.Id,
				})
			}
		}
	}
}
