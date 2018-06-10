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

	router.HandleFunc("/", server.IndexHandler)
	router.HandleFunc("/mesh/config", server.MeshNodeConfigHandler)

	return server
}

//IndexHandler handles the index
func (server *Server) IndexHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprint(response, "Hello world!")
}

func (server *Server) MeshNodeConfigHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprint(response, "Mesh config!")
}
