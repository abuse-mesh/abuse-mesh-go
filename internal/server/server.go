package server

import (
	"fmt"
	"go/build"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/coreos/bbolt"
	"github.com/olekukonko/tablewriter"

	controllers "github.com/abuse-mesh/abuse-mesh-go/internal/controllers"
	"github.com/abuse-mesh/abuse-mesh-go/internal/report"
	"github.com/gorilla/mux"
)

//Server is a webserver which
type Server struct {
	Webserver    *http.Server
	Router       *mux.Router
	Controllers  []controllers.Controller
	DatabasePath string
	Database     *bolt.DB
	DebugMode    bool
}

//NewServer creates a new abuse mesh server
func NewServer(debugMode bool) *Server {
	router := mux.NewRouter()
	webserver := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	server := &Server{
		Webserver:    webserver,
		Router:       router,
		Controllers:  []controllers.Controller{},
		DatabasePath: "/tmp/abusemesh.db", //TODO change to more permanent default valuedebugRouter
		DebugMode:    debugMode,
	}

	return server
}

func (server *Server) ListenAndServe() error {

	server.initServer()

	log.Println("Starting Abuse Mesh server on " + server.Webserver.Addr)

	return server.Webserver.ListenAndServe()
}

func (server *Server) initServer() {
	if db, err := bolt.Open(server.DatabasePath, 0600, nil); db != nil {
		server.Database = db
	} else {
		log.Printf("Unable to open database at '%s'", server.DatabasePath)
		log.Fatal(err)
	}

	defaultControllers := server.defaultControllers()

	//Add the default controllers to the list of controllers already in the server
	server.Controllers = append(server.Controllers, defaultControllers...)

	//We only have a v1 at the moment so we can prefix all the routes with the v1 protocol prefix
	v1Router := server.Router.PathPrefix("/mesh/v1").Subrouter()
	for _, controller := range server.Controllers {
		controller.RegisterRoutes(v1Router)
	}

	if server.DebugMode {
		debugRouter(server.Router)
	}
}

func (server *Server) defaultControllers() []controllers.Controller {

	reportController := &controllers.ReportController{
		AbuseReportStorage: report.NewAbuseReportBoltStorage(server.Database),
	}

	return []controllers.Controller{
		reportController,
	}
}

func debugRouter(router *mux.Router) {

	fmt.Println("\nRoutes in router: ")

	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Path", "Methods", "Handler"})

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		handler := route.GetHandler()

		function := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())

		file, line := function.FileLine(function.Entry())

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		gopath += "/src/"

		relativePath := file[len(gopath):len(file)]

		table.Append([]string{
			path,
			strings.Join(methods, ", "),
			relativePath + ":" + strconv.Itoa(line),
		})

		return nil
	})

	table.Render()
}
