#!bin/sh
GOARCH=amd64
GOOS=linux
go build -ldflags "-s -w" -o build/CrashReportClient main.go