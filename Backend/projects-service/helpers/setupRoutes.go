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
	router.Use()

	router.Handle("/projects", auth.EnableBoth(http.HandlerFunc(handler.GetAllProjects))).Methods("GET")
	router.Handle("/projects/{id}", auth.EnableBoth(http.HandlerFunc(handler.GetProjectById))).Methods("GET")
	router.Handle("/projects/create", auth.EnableManager(http.HandlerFunc(handler.CreateProject))).Methods("POST", "OPTIONS")
	router.Handle("/projects/{id}/add-member", auth.EnableManager(http.HandlerFunc(handler.AddMemberToProject))).Methods("POST", "OPTIONS")
	router.Handle("/projects/{id}/remove-member", auth.EnableManager(http.HandlerFunc(handler.RemoveMemberFromProject))).Methods("POST", "OPTIONS")
	router.Handle("/projects/user/{id}", http.HandlerFunc(handler.FindProjectsByUserID)).Methods("GET")
	router.Handle("/projects/manager/{id}", http.HandlerFunc(handler.FindProjectsByManagerID)).Methods("GET")

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetUsersHandler(w, r, userClient)
	}).Methods("GET")

	return router
}
