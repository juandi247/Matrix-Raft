package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"simba/simulator"
	"simba/sse"

	"github.com/google/uuid"
)
const fuzzyLevel simulator.FuzzyLevel = simulator.LOW



func homePage() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	}
}

func simulationPage() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/simulation.html")
	}
}

func serveStaticFiles() http.HandlerFunc{
return func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/js/sim-events.js")
	//TODO: Checkear lo de http.FileServe etc, para los archivos estaticos, ( no habra mucho pero puede ser que ene lfuturo si)
	
	}
}

func checkMap(sm *SimulatorSessionManager) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "We have %d simulations running", len(sm.simulatorSessions))
	}
}


type SimulationConfig struct{
	Seed int `json:"seed"`

}

const SimRunnerSessionKey = "simRunnerSession"
func simulationUidCheckMiddleware(next http.HandlerFunc,simulatorSessionManager *SimulatorSessionManager ) http.HandlerFunc{

	return func(w http.ResponseWriter, r *http.Request) {
		readUid := r.URL.Query().Get("uid")
		// if r.Method != http.MethodPost{
		// 	fmt.Println("error en el methodos")
		// 	return
		// }


		uid, err:= uuid.Parse(readUid)
		if err!=nil{
			fmt.Println("error parsing uid: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}


		simRunnerSession, exists:= simulatorSessionManager.simulatorSessions[uid]

		if !exists{
			w.WriteHeader(http.StatusNotFound)
			return
		}



		newCtx:= (context.WithValue(r.Context(), SimRunnerSessionKey ,simRunnerSession ))

		next.ServeHTTP(w, r.WithContext(newCtx))
	}

}

func configureSimulation(simulatorSessionManager *SimulatorSessionManager) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {


		if r.Method != http.MethodPost{
			return
		}




		var simConfig SimulationConfig
		err:= json.NewDecoder(r.Body).Decode(&simConfig)
		if err!=nil{
			fmt.Println("error unmarshiling", err)
			return
		}


		fuzzyConfig := simulator.NewFuzzyConfiguration(int64(simConfig.Seed), fuzzyLevel)


		uidGenerated, err:= uuid.NewUUID()

		if err!=nil{
			fmt.Println("Error generating UID", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		simulatorSessionManager.simulatorSessions[uidGenerated]= &simulator.SimulationRunner{
			Uid: uidGenerated,
			Time:               &simulator.SimTime{},
			Network:            &simulator.SimNetwork{},
			FuzzyProbabilities: fuzzyConfig,
			EventChannel: make(chan sse.SseEvent, 100),
			ShouldPublishEvents: true,
			
			//TODO: DEFAULT FOR NOW this 2
			SleepingTime: 300,
			MaxTicks: 1000,
		}


		uidStruct:= struct{
			Uid uuid.UUID
		}{
		Uid: uidGenerated,
	}

		data, err := json.Marshal(uidStruct)
		if err!=nil{
			fmt.Println("error marshaling struct", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
		


	}
}



type SseApiJson struct{
	EventType int
	Payload sse.SseEvent
}

func sseEventsHandler(w http.ResponseWriter, r *http.Request) {
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)

		//CONFIG FOR SSEEEE
  		w.Header().Set("Content-Type", "text/event-stream")
    		w.Header().Set("Cache-Control", "no-cache")
    		w.Header().Set("Connection", "keep-alive")


		rc:= http.NewResponseController(w)
		doneChan:= r.Context().Done()
		
		for{

		select{

			case <-doneChan:
				//this would clean it up autoamtcially, it will end, and the own simulator 
				simRunner.PauseContext.SetState(simulator.StateStopDefinetly)
				return 
			case event:= <- simRunner.EventChannel: 

				encodedData, err:= json.Marshal(SseApiJson{
					EventType: int(event.GetEventType()),
					Payload: event,
				})

				if err!=nil{
					fmt.Println("error marshaling json", err)
					return
				}
				

				_, err = fmt.Fprintf(w, "data: %s \n\n", string(encodedData))
				if err!=nil{
					fmt.Println("error formating event", err)
					return
				}
			rc.Flush()
			}
		}

}



func startSimulation(simSessionManager *SimulatorSessionManager) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)

		//cleans up automatically when its stopped or it finishes the ticks 
		go func() {
			defer simSessionManager.cleanupSession(simRunner.Uid)

			simRunner.Start()
		}()
		w.WriteHeader(http.StatusOK)
	}
}


func pauseSimulation(w http.ResponseWriter, r *http.Request){
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)
		simRunner.PauseContext.SetState(simulator.StatePaused)
		w.WriteHeader(http.StatusOK)
}


func resumeSimulation(w http.ResponseWriter, r *http.Request){
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)
		simRunner.PauseContext.SetState(simulator.StateRunning)
		w.WriteHeader(http.StatusOK)
}


func changeSimulationSleep(w http.ResponseWriter, r *http.Request){
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)

	var changeApi struct{
		SleepMode simulator.SleepMode

	}

	err:= json.NewDecoder(r.Body).Decode(&changeApi)

	if err!=nil{
		fmt.Println("error decoding json")
	}
	//TODO: change this, esto esta como feo xd
		simRunner.SleepingTime = changeApi.SleepMode
		w.WriteHeader(http.StatusOK)
}


