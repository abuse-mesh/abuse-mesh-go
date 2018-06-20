package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/abuse-mesh/abuse-mesh-go/internal/report"
	"github.com/gorilla/mux"
)

type ReportController struct {
	AbuseReportStorage report.AbuseReportStorage
}

func (controller *ReportController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/abuse-report", controller.getAbuseAllReports).Methods("GET")
	router.HandleFunc("/abuse-report/{id}", controller.getAbuseReport).Methods("GET")

	router.HandleFunc("/abuse-report", controller.addAbuseReport).Methods("POST")
}

func (controller *ReportController) getAbuseAllReports(response http.ResponseWriter, request *http.Request) {
	//Get a report
	reports := controller.AbuseReportStorage.GetAll()

	//Else return json
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(reports)
}

func (controller *ReportController) getAbuseReport(response http.ResponseWriter, request *http.Request) {
	//Parse the vars from the request using the router
	vars := mux.Vars(request)

	//Get a report
	report := controller.AbuseReportStorage.GetOneByID([]byte(vars["id"]))

	//If not found return a 404
	if report == nil {
		http.NotFound(response, request)
		return
	}

	//Else return json
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(report)
}

func (controller *ReportController) addAbuseReport(response http.ResponseWriter, request *http.Request) {
	controller.AbuseReportStorage.Insert(&report.AbuseReport{
		ID:                  []byte("ABC"),
		AbuseType:           report.CopyrightInfringement,
		SuspectResourceType: report.DomainName,
		SuspectResourceID:   "serverius.net",
	})
}
