package simulator

import (
	"math/rand"
	"simba/newraft"
)


const numberOfRequests = 1000
const maxEntryLength = 10
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")


func GenerateRequests(rng *rand.Rand) []SimMessage {

	arr:= make([]SimMessage, numberOfRequests)

	for i:=range 1000{
	arr[i] = SimMessage{
			id: i,
			DeliveryTick: 10 + rng.Intn(maxTicks-10 +1) ,
			Message: newraft.Message{
				Type: newraft.MsgNewEntry,
				Payload: newraft.FrontEndEntry{
					Magic: "SKIPPER",
					Entry: GenerateRandomString(rng),
					},
				},
		}

	}
	return arr

}


func GenerateRandomString(rng *rand.Rand)string {
	arr:= make([]rune, len(letters))

	for i:=range arr{
 		number:= rng.Intn(len(letters))
		arr[i] = letters[number]
	}

return string(arr)	
}
