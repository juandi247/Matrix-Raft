package simulator

import (
	"fmt"
	"math/rand"
	"simba/newraft"
	"simba/sse"
)


const numberOfRequests = 1000
const maxEntryLength = 10
const ClientId = "CLIENT"
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")



func GenerateRandomString(rng *rand.Rand)string {
	arr:= make([]rune, len(letters))

	for i:=range arr{
 		number:= rng.Intn(len(letters))
		arr[i] = letters[number]
	}

return string(arr)	
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
				Entry: GenerateRandomString(rng),
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
	fmt.Println("LLEGO ALGO AL CLIENTE: ")

	if msg.Message.Type!=newraft.MsgLeaderCheck{
		panic("llego algo invlaido al cliente")

	}
	
	payload, ok:= msg.Message.Payload.(newraft.LeaderCheck)

	if !ok{
		panic("cast failed")
	}



	if payload.LeaderId==""{
		fmt.Println("llego empty del random al clinte")
		return
	}
	//set the payload id as the leader when w recieve this.
	sm.CachedLeaderId = payload.LeaderId
	fmt.Println("el leaderID supustamente ahora seria: ", payload.LeaderId)

}




/* 
idear el requst generator 


pero seria mejor generarlas on the FLy??? mas complicado, pero mas sencillo para volver a enviarlas de nuevo

1. Debo tener una variable ques eria una especie de "cache" de quien es el lider ahora. 
2. Cuando envio una request,  debo enviarsela al currentLeader de mi Cliente simulado. 
3.  Si su currentLeader es null, entonces escojo uno rnadom, si ese random me devuelve como no estoyuy listo deberia intentar mas tarde, en unos +20 ticks o algo similar. Seria una queue de mensages pendientes para enviar al lider






opciones: 
- Tener todo listo con nsu delivery tick, lo cual me complicaria porque luego, a que leader seria entregao entonces? y si no fue entregado debo ponerlo de nuevo en la cola no? 


	[]sse.SimMessage, messages que van a entregarse ya con su tick y todo.  Ya estan en la queue y todo, pero entonces como eliho enviarselo a X nodo? porque si no se quien es el lidr, debo posponer esos. 



- Ir pasando por tick e ir enviando cada X ticks un menaje, pero eso condicoina los ticks, y si los timeouts no salen bien etnonces mal. 

, 

- Opcion pouede ser valida: 
	- Cada X ticks enviar un mensaje - Siempre al lider Id, si es null, enviarlo a uno random con la seed y esperar su respuesta. 

	- Si responde un leaderId - ACTUALIZAMOS el cache e intentmos volver a enviar el mensaje ese. 
	- Si enviamos y nos dice que no peude aceptar nada - lo ponemos de nuevo par vovler a enviar en X ticks


*/


/* TODO:  HOY

- Request generat5or funcionando 

- Pausar la simulacion (haciendola una gorouitine, en el for loop, tenerle un channel para "pausarla")

- Mandar en la UI para ver todo el contenido del mensaje, haciendo click en la network

*/
