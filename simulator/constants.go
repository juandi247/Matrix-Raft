package simulator


const TickFrequency = 1

const SendAppendEntriesFreq = 40
const ElectionTimeout = 100
const MinHeartBeatTimeout = 60
const MaxHeartBeatTimeout = 150

/*
Assertions for ranges inside the constants, to have it in compile time. 
*/
func init() {
	if SendAppendEntriesFreq >= MinHeartBeatTimeout || SendAppendEntriesFreq >= MaxHeartBeatTimeout {
		panic("the leader heart beat MUST be smaller than both minfollower and maxfollower timeouts")
	}

	if MinHeartBeatTimeout >= MaxHeartBeatTimeout {
		panic("the minfollowre must be smaller than the max")
	}

	if minCrashNodeDowntime >= maxCrashNodeDowntime {
		panic("the minCrashNodeDowntime must be smaller than the max")
	}

	if minLatencyDelay >= maxCrashNodeDowntime {
		panic("the minCrashNodeDowntime must be smaller than the max")
	}
}

