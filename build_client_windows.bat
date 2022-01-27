set GOARCH=amd64
set GOOS=windows
go build -ldflags "-s -w" -o build/CrashReportClient.exe main.go
exit 0