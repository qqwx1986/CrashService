package main

import (
	"CrashService/client"
	"CrashService/server"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	if strings.ToLower(os.Args[1]) == "server" {
		server.Run()
	} else {
		client.Run()
	}
	println("exit")
}
