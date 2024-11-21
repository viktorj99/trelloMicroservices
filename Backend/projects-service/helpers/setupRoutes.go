package helpers

import (
	"net/http"
	userpb "pb/userpb"
	"projects-service/auth"
	"projects-service/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes(handler *handlers.ProjectHandler, userClient userpb.UserServiceClient) *mux.Router {
	router := mux.NewRouter()
	router.Use() // Primeni CORS middleware na sve rute

	router.Handle("/projects", auth.EnableBoth(http.HandlerFunc(handler.GetAllProjects))).Methods("GET")
	router.Handle("/projects/{id}", auth.EnableBoth(http.HandlerFunc(handler.GetProjectById))).Methods("GET")
	router.Handle("/projects/create", auth.EnableManager(http.HandlerFunc(handler.CreateProject))).Methods("POST", "OPTIONS")
	router.Handle("/projects/{id}/update", auth.EnableManager(http.HandlerFunc(handler.UpdateProject))).Methods("PUT", "OPTIONS")
	router.Handle("/projects/{id}/delete", auth.EnableManager(http.HandlerFunc(handler.DeleteProject))).Methods("DELETE", "OPTIONS")

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUsersHandler(w, r, userClient)
	}).Methods("GET")

	return router
}
