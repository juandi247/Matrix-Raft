package server

import (
	"fmt"
	"net/http"
	"simba/sse"
)


type httpServer struct {
	muxer     *http.ServeMux
	port      string
	startFunc func() error
	// RequestTimeout time

}
//TODO: aca agregar el manager con los datos que va a tener, de pronto el mapa de eventos para cada sesion porque se debe pasar a los handlers
func NewHttpServer(port string, isProd bool, certfile, keyfile string) *httpServer {
	eventchannelTmp:= make(chan sse.SseEvent, 100)
	mux := http.NewServeMux()
	// we register it as wildcard
	mux.Handle("/", homePage())
	mux.Handle("/simulation", simulationPage())
	mux.Handle("/startsim", startSimulation(eventchannelTmp) )
	mux.Handle("/events", sseEventsHandler(eventchannelTmp) )
	//TODO: Chcekear esto porque esta suiper mal asi tan directo
	mux.Handle("/js/sim-events.js", serveStaticFiles())

	server := &httpServer{
		muxer: mux,
		port:  port,
	}
	fmt.Println(isProd)

	if isProd {
		server.startFunc = func() error {
			fmt.Println("starting with tls")
			return http.ListenAndServeTLS(port, certfile, keyfile, mux)
		}
	} else {
		server.startFunc = func() error {
			fmt.Println("starting without tls, localhost")
			return http.ListenAndServe(port, mux)
		}
	}
	return server
}

func (s *httpServer) StartServer() error {
	err := s.startFunc()
	if err != nil {
		fmt.Println("ERORRRR", err)
		return fmt.Errorf("Error starting server: %v ", err.Error())
	}
	return nil
}
