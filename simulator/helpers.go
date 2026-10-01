package simulator

import (
	"fmt"
	"math/rand"
	"strconv"
)


func buildFriendsIds(numberOfFriends int, nodeIndex int) []string{
	rta:= []string{}
	for i:=1; i<=numberOfFriends; i++{
		if i!=nodeIndex{
		rta = append(rta, "Node"+strconv.Itoa(i))
		}

	}
	return rta
}



func generateHeartbeatTimeout(rng *rand.Rand) uint32 {
	t:= uint32(MinHeartBeatTimeout + rng.Intn(MaxHeartBeatTimeout-MinHeartBeatTimeout+1))
	fmt.Println("timeout generado es: ", t) 
	return t
}




