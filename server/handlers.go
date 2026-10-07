package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		fmt.Println("cayo aca en este server files")
	http.ServeFile(w, r, "templates/js/sim-events.js")
	//TODO: Checkear lo de http.FileServe etc, para los archivos estaticos, ( no habra mucho pero puede ser que ene lfuturo si)
	
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

    		fmt.Println("UID:", uid)

		simRunnerSession, exists:= simulatorSessionManager.SimulatorSessions[uid]

		if !exists{
			//TODO: resopnderle que no hay soimualcion para esa sesion
			fmt.Println("no hay simulacion par ese UID")
			w.WriteHeader(http.StatusNotFound)
			return
		}



		ctx:= context.WithValue(context.Background(), SimRunnerSessionKey ,simRunnerSession )
		newR:= r.WithContext(ctx)

		next.ServeHTTP(w, newR)
	}

}

func configureSimulation(simulatorSessionManager *SimulatorSessionManager) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {


		if r.Method != http.MethodPost{
			return
		}


		data, err:= io.ReadAll(r.Body)

		if err!=nil{
			fmt.Println("Body can not be read",err)
			return
		}


		var simConfig SimulationConfig
		err= json.Unmarshal(data, &simConfig)
		if err!=nil{
			fmt.Println("error unmarshiling", err)
			return
		}


		fmt.Println("LA SEED FUEEEEEEEEE: ", simConfig.Seed)
		fuzzyConfig := simulator.NewFuzzyConfiguration(int64(simConfig.Seed), fuzzyLevel)


		uidGenerated, err:= uuid.NewUUID()

		if err!=nil{
			fmt.Println("Error generating UID", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("ID CREADO FUE: ", uidGenerated)
		simulatorSessionManager.SimulatorSessions[uidGenerated]= &simulator.SimulationRunner{
			Time:               &simulator.SimTime{},
			Network:            &simulator.SimNetwork{},
			FuzzyProbabilities: fuzzyConfig,
			EventChannel: make(chan sse.SseEvent, 100),
			ShouldPublishEvents: true,
		}


		uidStruct:= struct{
			Uid uuid.UUID
		}{
		Uid: uidGenerated,
	}

		data, err = json.Marshal(uidStruct)
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
		simRunner.Start()

		//CONFIG FOR SSEEEE
  		w.Header().Set("Content-Type", "text/event-stream")
    		w.Header().Set("Cache-Control", "no-cache")
    		w.Header().Set("Connection", "keep-alive")


		rc:= http.NewResponseController(w)
		fmt.Println("SSE handler esta activo esperando")
		doneChan:= r.Context().Done()
		
		for{

		select{

			case <-doneChan:
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
					fmt.Println("error formating event")
					return
				}
			rc.Flush()

			}
		}

}



func startSimulation(w http.ResponseWriter, r *http.Request){
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)
		simRunner.Start()
}


func pauseSimulation(w http.ResponseWriter, r *http.Request){
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)
		simRunner.PauseContext.SetState(simulator.StatePaused)
		fmt.Println("Pausing Sim")
		w.WriteHeader(http.StatusOK)
}


func resumeSimulation(w http.ResponseWriter, r *http.Request){
		simRunner:= r.Context().Value(SimRunnerSessionKey).(*simulator.SimulationRunner)
		simRunner.PauseContext.SetState(simulator.StateRunning)
		fmt.Println("Resume sim")
		w.WriteHeader(http.StatusOK)
}


