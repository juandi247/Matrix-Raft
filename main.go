package main

import (
	"fmt"
	// "simba/adapters"
	"simba/server"

	// "simba/reality"
	"simba/simulator"
)

const matrixMode bool = true
const SEED = 12345
// By DEFAULT LOW but this should ve changed for the simulations, and for runtime too(?)
const fuzzyLevel simulator.FuzzyLevel = simulator.LOW

func main() {
	// var runner adapters.Runner

	if matrixMode {
		
		httpServer:=server.NewHttpServer(":8090", false, "", "")


		go func(){
			err:= httpServer.StartServer()
			if err!=nil{
				fmt.Println("error: ", err)
				return
			}
		}()

		/*
		fuzzyConfig := simulator.FuzzyConfiguration(SEED, fuzzyLevel)


		runner = &simulator.SimulationRunner{
			Time:               &simulator.SimTime{},
			Network:            &simulator.SimNetwork{},
			FuzzyProbabilities: fuzzyConfig,
			Port: "8080",
			IsHttps: false,
		}
*/
		/*
crear la f uzzyConfiguration con la seed, y el level (po rahora el level siempre sera igual)


iniciar el similatorRunner

*/
	} else {
		// transportAdapter:= &reality.RealNetwork{}
		// timeAdapter := &reality.PhysicalTime{}

		// runner= some runner

	}
	fmt.Println("starting program")
//	runner.Start()
	select{}
}





func coso(i, j int) int{

return i+j
}
