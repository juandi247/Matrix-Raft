package simulator

import (
	"fmt"
	"math/rand"
	"simba/newraft"
	"simba/sse"
)


const ClientId = "CLIENT"
var Accounts = []string{"Skipper", "Simba", "John wick", "Tom", "Diego", "Messi", "Sonsota", "Pipolino", "WindiRaptor", "Pedri"}
var TransactionType = []string{"DEBIT", "CREDIT"}



func GenerateRandomEntry(rng *rand.Rand)string {
	randomAccount:= Accounts[ rng.Intn(len(Accounts))]
	transactionType:= TransactionType[ rng.Intn(len(TransactionType))]

	//in between 100 and 2000
	randomMoney:= 100 + rng.Intn(1900)

	return fmt.Sprintf("%s %d $ to acccount: %s", transactionType, randomMoney, randomAccount)
}



type SimClient struct{
	CachedLeaderId string //can be empty if there is no leader known yet
	ClientRequests []sse.SimMessage

	//PendingRequests []sse.SimMessage
}



func (sm *SimClient) generateClientRequest(currentTick int, rng *rand.Rand) []newraft.Message{
	if currentTick % 20 != 0 || currentTick <=70{
		return nil
	} 


	var leaderId string
	if sm.CachedLeaderId == ""{
		leaderId = pickRandomNode(rng)
	}else{
		leaderId = sm.CachedLeaderId
	}


	return []newraft.Message{
		{
			SenderId: ClientId, //TODO: aca neceistaria cundo se reciba, deberia revisar si el recipinete es client, para poder leerlo desde aca
			ReceiverId: leaderId,
			Type: newraft.MsgNewEntry,
			Payload: newraft.FrontEndEntry{
				Magic: "SKIPPER",
				Entry: GenerateRandomEntry(rng),
			},
					},
				}
}


//SHOULD RETURN AN ID of the node 
func pickRandomNode(rng *rand.Rand) string{
	//takes random number between 1 - TotalNodes (including the totalNodes)
	randomId:=  1 + rng.Intn(newraft.TotalNodes) 

	return fmt.Sprintf("Node%d", randomId)
}


/*

forma simple 

si recibo un mensje que dice Lider succesfull, no hago nada.

Si recibo mensaje que diga no soy lider, mira mi lider. Lo ACTUALIZAMOS

si me devulev een un caso l lidr mpty, entonces me jodi supongo xd



complicarla: 
podria guarar el id del mensaje que acabo de enviar o ponerle algun ID o algo y meterlo en una cola de 
delivered. Si me llega una resupest ade ue lalguno de esos delivered, no fue entregado con un id, entonces 
lo intento d nuevo. pero despues
*/
func (sm *SimClient) handleIncomingMessage(msg sse.SimMessage){

	if msg.Message.Type!=newraft.MsgLeaderCheck{
		panic("llego algo invlaido al cliente")

	}
	
	payload, ok:= msg.Message.Payload.(newraft.LeaderCheck)

	if !ok{
		panic("cast failed")
	}


	//set the payload id as the leader when w recieve this.
	//NOTE: we can receive "" empty, or the leader id, if we receive empty, we select a random node again.
	sm.CachedLeaderId = payload.LeaderId

}




