package simulator

import "simba/newraft"

/*
This file will contian all the checks and rules given by RAFT algorithm such as:

No 2 leaders on the same Term
Logs should be the exact same on all the servers from log[0: minimumMatchIndex] (This checks the Commited entries)


wait until the end to check that all the servers contain the same data.


At the end the check will take all the CONFIRMED Entries, and check with the current leader, that will contain that one.
*/

func checkSplitBrain(nodeList []*newraft.Node){
	var biggestTerm int

	for _ , node:= range nodeList{
		if node.CurrentRole == newraft.LEADER {
			if biggestTerm == node.CurrTerm{
				panic("There are 2 leaders on the same term")
			}
		biggestTerm = max(biggestTerm, node.CurrTerm)

		}

	}

}


//aca chekear las comited entries, del ultimo leader, con las del resto de logs. tomando cada log de cada server follower
//tomamos ese valor y las entradas hasta ahi, deben ser igual que las del lider, el lider edberia tener el commitIndex siempre mayor o igual al del log dec ada follower
func checkCommitedEntries(nodeList []*newraft.Node){
	

}
