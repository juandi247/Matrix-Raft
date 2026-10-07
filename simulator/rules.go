package simulator

import (
	"fmt"
	"simba/newraft"
	"slices"
	"sort"
)

/*
This file will contian all the checks and rules given by RAFT algorithm such as:

No 2 leaders on the same Term
Logs should be the exact same on all the servers from log[0: minimumMatchIndex] (This checks the Commited entries)


wait until the end to check that all the servers contain the same data.


At the end the check will take all the CONFIRMED Entries, and check with the current leader, that will contain that one.
*/


func checkInvariants(nodeList []*newraft.Node){
	checkSplitBrain(nodeList)
	checkCommitedEntries(nodeList)
}

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


//TODO: esto esta mal, deberia ser mucho mejor pero bue
func checkCommitedEntries(nodeList []*newraft.Node){
	//ordenar los nodos de forma que queden en orden ascendente por log size 
	//comparar siempre el menor con el siguiente cada dato del log, si hay algo distinto PARAR
	newList:= nodeList

	sort.Slice(newList, func(i, j int) bool {return newList[i].CommitIndex < newList[j].CommitIndex})



	for i, currNode:=range newList{
		if i==0{
			continue
		}



		prevNode:= newList[i-1]	
		minCommitIndex:= min(currNode.CommitIndex, prevNode.CommitIndex)

	
		if minCommitIndex == 0 {
			continue
		}

		currNodeCommitedLog:= currNode.Log[:minCommitIndex+1]
		prevNodeCommitedLog:= currNode.Log[:minCommitIndex+1]



		equeal:= slices.Equal(currNodeCommitedLog, prevNodeCommitedLog)

		if !equeal{
			fmt.Println("noodos: ", currNode.Id, " y : ", prevNode.Id, "no coindicen")
			panic("logs no matchean sus datos cuando y fueron comiteados, raro")
		}

		
	}


}


