package main

import (
	"CrashService/client"
	"CrashService/server"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	for i, arg := range os.Args {
		logrus.Infof("%d %s", i, arg)
	}
	if len(os.Args) < 2 {
		return
	}
	if strings.ToLower(os.Args[1]) == "server" {
		server.Run()
	} else if strings.ToLower(os.Args[1]) == "tool" {
		server.Tool()
	} else {
		client.Run()
	}
	println("exit")
}
