package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"simba/simulator"
	"simba/sse"
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

func startSimulation(eventChannel chan sse.SseEvent) http.HandlerFunc{
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
		fuzzyConfig := simulator.FuzzyConfiguration(int64(simConfig.Seed), fuzzyLevel)


		runner := &simulator.SimulationRunner{
			Time:               &simulator.SimTime{},
			Network:            &simulator.SimNetwork{},
			FuzzyProbabilities: fuzzyConfig,
			Port: "8080",
			IsHttps: false,
		}


		//this should be a goruoitne, le pasamos el channel de eventos y listo

	go	 runner.Start(eventChannel)

		w.WriteHeader(http.StatusOK)


	}
}



type SseApiJson struct{
	EventType int
	Payload sse.SseEvent
}
func sseEventsHandler(eventChannel chan sse.SseEvent) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {

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
			case event:= <- eventChannel: 

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
}

