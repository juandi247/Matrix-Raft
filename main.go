package main

import (
	"flag"
	"fmt"
	"simba/server"
	"simba/simulator"
)

const matrixMode bool = true
const fuzzyLevel simulator.FuzzyLevel = simulator.LOW




func main() {




	mode:=	flag.String("mode","V", "Usage: Pass as value:\n V (for visualization mode) \n M (matrix mode) \n R (real implementation)\n Default value is V")

	//NOTE: Only for Visualizatoin MODE OR REAL MODE
		port:=	flag.Int("port", 8080, "Usage: Pass a port such as 8090. Port is by default 8080")
		isHttps:=flag.Bool("https", false, "Pass boolean to activate https. If true, use the flags of -certfile & -keyfile to pass the path for both")
		certFilePath:=flag.String("certfile", "", "Pass the path for the certfile (only when -http flag is true)")
		keyFilePath:=flag.String("keyfile", "", "Pass the path for the keyfile (only when -https flag is true)")



	//NOTE: Only for Simulator MODE
		seed:=	flag.Int64("seed", 0, "Default seed is 1234")
		maxTicks:=flag.Int64("maxticks", 100_000, "Max ticks for the simulation to run")
		//TODO: config for the fuzzy probabitlies
		//fuzzyLevel:=flag.String("fuzzylevel", "LOW", "")




	flag.Parse()

	switch *mode{

	case "V":
		err:=	StartVisualizationMode(*port, *isHttps, *certFilePath, *keyFilePath)
		if err!=nil{
			fmt.Println("Error found: ", err)
			return
		}
	case "M":
		err:=StartMatrixMode(*seed, *maxTicks)
		if err!=nil{
			fmt.Println("Error found: ", err)
			return
		}
		
	case "R":


	default: 
		fmt.Println("Mode not recognized, please provide V (visualization mode), M (matix mode) or R (for real mode)")
		return
	}
	select{}
}




func StartVisualizationMode(port int, isHttps bool, certFilePath, keyFilePath string) error{
	fmt.Println("visualizaton mode")
	fmt.Println("https: ", isHttps)

		if isHttps && (certFilePath == "" || keyFilePath ==""){
			return fmt.Errorf("certfile or keyfile not provided for https")
		}


		httpServer:=server.NewHttpServer(fmt.Sprintf(":%d",port), isHttps, certFilePath, keyFilePath)

		go func(){
			err:= httpServer.StartServer()
			if err!=nil{
				fmt.Println("error starting server: ", err)
				return
			}
		}()

		return nil

}




func StartMatrixMode(seed int64, maxTicks int64) error{

	if seed==0{
		return fmt.Errorf("Seed not provided, please provide the -seed flag with a random int64")
	}

	fmt.Println("Matrix mode")

	fuzzyConfig := simulator.NewFuzzyConfiguration(seed, fuzzyLevel)




	sim:= &simulator.SimulationRunner{
			Time:               &simulator.SimTime{},
			Network:            &simulator.SimNetwork{},
			FuzzyProbabilities: fuzzyConfig,
			EventChannel: nil,
			ShouldPublishEvents: false,
		}

	sim.Start()

	return nil

}
