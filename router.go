package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type Router struct {
	muxRouter *mux.Router
}

type Route struct {
	Path    string
	Method string
	Handler func(http.ResponseWriter, *http.Request)
}

func NewRouter() *Router {
	router := mux.NewRouter()
	return &Router{
		muxRouter: router,
	}
}

func (r Router) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		r.RegisterRoute(route)
		logrus.Infof("Route registered %v", route)
	}
}

func (r Router) RegisterRoute(route Route) {
	r.muxRouter.HandleFunc(route.Path, route.Handler).Methods(route.Method)
}

func (r Router) ServeStatic(path string, withDirListing bool) {
	if withDirListing {
		serveWithDirListing(r.muxRouter, path)
	} else {
		serveWithoutDirListing(r.muxRouter, path)
	}
}

func (r Route) String() string {
	return fmt.Sprintf("%-6v %s", r.Method, r.Path)
}

func serveWithDirListing(router *mux.Router, dir string) {
	serve(router, dir, http.FileServer(http.Dir(dir)))
}

func serveWithoutDirListing(router *mux.Router, dir string) {
	serve(router, dir, withoutDirListing(http.FileServer(http.Dir(dir))))
}

func serve(router *mux.Router, dir string, fileServer http.Handler) *mux.Route {
	return router.PathPrefix("/" + dir + "/").Handler(http.StripPrefix("/"+dir, fileServer))
}

func withoutDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/") {
			http.NotFound(writer, request)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
