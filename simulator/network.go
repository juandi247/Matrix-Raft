package simulator

import (
	"container/heap"
	"fmt"
	"simba/adapters"
	"simba/newraft"
	"simba/sse"
)

type SimNetwork struct {
	messageQueue *PriorityQueue
	FuzzyConfig  FuzzyConfig
	TimeAdapter  adapters.TimeAdapter
	IdCounter int //id counter for the messages to be deliverd
	EventChan chan sse.SseEvent
	ShouldPublishEvent bool
	simClient SimClient
}


func (s *SimNetwork) SendMessage(messages []newraft.Message) {

	events:= []sse.SimMessage{}

	for _, message := range messages {
		var delayTicks int64
		var lost bool
		lost, delayTicks = s.FuzzyConfig.RandomizeNetwork()


		//TODO: there should be a tracker or something for the later UI that indicates that a message was LOST


		s.IdCounter++
		if !lost {
			simMessage := &sse.SimMessage{
				Id: s.IdCounter,
				DeliveryTick: int(s.TimeAdapter.Now() + delayTicks),
				Message:      message,
			}
			heap.Push(s.messageQueue, simMessage)
			//this events is just for the SSE
			events=append(events, *simMessage)
		}
		
	}


	if len(events)!=0{
		PublishEvent(s.ShouldPublishEvent, s.EventChan,sse.NewSimulationMessagesPushed(events))
		// s.EventChan <- sse.NewSimulationMessagesPushed(events)
	}
	// s.printQueueData()
}


func (s *SimNetwork) printQueueData(){
	fmt.Println("-------------- QUEUE DATA on Tick ", s.TimeAdapter.Now(), "-------------------")

	for i, v:= range *s.messageQueue{
 	fmt.Printf(
            "[%d] deliveryTick=%d  senderId=%v receiverId=%v \n",
		i, 
            v.DeliveryTick,
            v.Message.SenderId,
            v.Message.ReceiverId,

        )
		

	}
	fmt.Println("---------------")
}

func (s *SimNetwork) SendTimeout(msg newraft.Message){
	s.IdCounter++
	simMessage:= &sse.SimMessage{
		DeliveryTick: int(s.TimeAdapter.Now()) - 1,
		Message: msg,
		
	}
	s.messageQueue.Push(simMessage)
}




type PriorityQueue []*sse.SimMessage


func (pq *PriorityQueue) Peek() *sse.SimMessage {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0]
}
func (pq PriorityQueue) Len() int { return len(pq) }


func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].DeliveryTick == pq[j].DeliveryTick {
		return pq[i].Index < pq[j].Index
	}
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].DeliveryTick < pq[j].DeliveryTick
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*sse.SimMessage)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *sse.SimMessage, value string, priority int) {
}
