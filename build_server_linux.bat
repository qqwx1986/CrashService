set GOARCH=amd64
set GOOS=linux
go build -ldflags "-s -w" -o build/CrashReceiverServer main.go
exit 0