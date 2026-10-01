package simulator

import (
	"fmt"
	"simba/adapters"
	"simba/newraft"
)

type SimNetwork struct {
	messageQueue *PriorityQueue
	FuzzyConfig  FuzzyConfig
	TimeAdapter  adapters.TimeAdapter
}

type SimMessage struct {
	index int
	id int
	DeliveryTick int
	Message      newraft.Message
}



func (s *SimNetwork) SendMessage(messages []newraft.Message) {
	fmt.Printf("we are going to put in the queue %v messages \n\n", len(messages))
	for _, message := range messages {
		var delayTicks int64
		var lost bool
		lost, delayTicks = s.FuzzyConfig.RandomizeNetwork()


		fmt.Println("ReceiverID: ", message.ReceiverId, "deliveryTick: ", s.TimeAdapter.Now()+delayTicks)
		//TODO: there should be a tracker or something for the later UI that indicates that a message was LOST
		if !lost {
			simMessage := &SimMessage{
				DeliveryTick: int(s.TimeAdapter.Now() + delayTicks),
				Message:      message,
			}
			s.messageQueue.Push(simMessage)
		}
		
	}

	s.printQueueData()
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
	simMessage:= &SimMessage{
		DeliveryTick: int(s.TimeAdapter.Now()) - 1,
		Message: msg,
	}
	s.messageQueue.Push(simMessage)
}




type PriorityQueue []*SimMessage


func (pq *PriorityQueue) Peek() *SimMessage {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0]
}
func (pq PriorityQueue) Len() int { return len(pq) }


func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].DeliveryTick == pq[j].DeliveryTick {
		return pq[i].index < pq[j].index
	}
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].DeliveryTick < pq[j].DeliveryTick
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*SimMessage)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *SimMessage, value string, priority int) {
}
