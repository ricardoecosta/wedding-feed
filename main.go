package main

import (
	"github.com/Sirupsen/logrus"
)

func main() {
	server, err := NewServer("config.json")
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
	server.Start()
	// todo: graceful shutdown
}
