package main

import (
	"net/http"
	"fmt"
	"log"
	"github.com/ricardoecosta/weddingfeed/router"
	"github.com/ricardoecosta/weddingfeed/config"
)

func main() {
	config := config.Load("config.json")
	router := router.Initialize()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router))
}
