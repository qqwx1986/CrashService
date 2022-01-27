set GOARCH=amd64
set GOOS=linux
go build -ldflags "-s -w" -o build/CrashReportClient main.go
exit 0