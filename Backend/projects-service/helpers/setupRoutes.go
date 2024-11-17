package helpers

import (
	"net/http"
	userpb "pb/userpb"
	"projects-service/auth"
	"projects-service/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes(handler *handlers.ProjectHandler, userClient userpb.UserServiceClient) *mux.Router {
	r := mux.NewRouter()

	r.Handle("/projects", auth.EnableBoth(http.HandlerFunc(handler.GetAllProjects))).Methods("GET")
	r.Handle("/project/{id}", auth.EnableBoth(http.HandlerFunc(handler.GetProjectById))).Methods("GET")
	r.Handle("/project/create", auth.EnableManager(http.HandlerFunc(handler.CreateProject))).Methods("POST")
	r.Handle("/project/{id}", auth.EnableManager(http.HandlerFunc(handler.UpdateProject))).Methods("PUT")
	r.HandleFunc("/project/{id}", handler.UpdateProject).Methods("PUT")
	r.HandleFunc("/project/{id}", handler.DeleteProject).Methods("DELETE")

	// Dodavanje rute za gRPC poziv user-service
	r.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUsersHandler(w, r, userClient)
	}).Methods("GET")

	return r
}
