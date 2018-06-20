package controllers

import "github.com/gorilla/mux"

type Controller interface {
	RegisterRoutes(*mux.Router)
}
