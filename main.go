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
	} else if strings.ToLower(os.Args[1]) == "tool" {
		server.Tool()
	} else {
		client.Run()
	}
	println("exit")
}
