package router

import (
	"github.com/gorilla/mux"
	"net/http"
)

func Initialize() *mux.Router {
	router := mux.NewRouter()

	//router.Use(mux.CORSMethodMiddleware(router))
	//router.Use(authentication)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	router.HandleFunc("/messages", getMessages).Methods("GET")
	router.HandleFunc("/messages", createMessage).Methods("POST")
	router.HandleFunc("/messages/{id}", updateMessage).Methods("PUT")
	router.HandleFunc("/messages/{id}", getMessage).Methods("GET")
	router.HandleFunc("/messages/{id}", deleteMessage).Methods("DELETE")

	return router
}

func authentication(next http.Handler) http.Handler  {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if (isAuthorized(request)) {
			next.ServeHTTP(writer, request)
		}

		http.Error(writer, "Forbidden", http.StatusForbidden)
	})
}

func isAuthorized(request *http.Request) bool {
	return request.Header.Get("X-Auth-Token") == "FooBar"
}
