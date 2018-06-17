package app

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

//Server is a webserver which
type Server struct {
	webserver *http.Server
	router    *mux.Router
}

//NewServer creates a new abuse mesh server
func NewServer() *Server {
	router := mux.NewRouter()
	webserver := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	server := &Server{
		webserver: webserver,
		router:    router,
	}

	return server
}

func (server *Server) makeRoutes() {
	router := server.router

	router.HandleFunc("/mesh/abuse-report", server.getAbuseReports).Methods("GET")
	router.HandleFunc("/mesh/abuse-report", server.addAbuseReport).Methods("POST")
}

func (server *Server) getAbuseReports(response http.ResponseWriter, request *http.Request) {
	
}

func (server *Server) addAbuseReport(response http.ResponseWriter, request *http.Request) {
	
}
