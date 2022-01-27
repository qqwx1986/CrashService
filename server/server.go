package server

import (
	"CrashService/cmn"
	"CrashService/deamon"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

// Server 相关参数
var httpPort *int
var crashBasePath string
var symbolBasePath string
var errorBasePath string
var basePath *string
var dumpAnlyisePath string
var dumpAnlyiseUrl *string
var daemon *bool

//channel 数量
var chanNum *int

type minidumpChanType struct {
	path     string
	ver      string
	platform string
}

var minidumpChan chan *minidumpChanType
var exitSignal chan os.Signal

var exitChan []chan struct{}
var waitChan *sync.WaitGroup

func Run() {
	httpPort = flag.Int("httpPort", 13333, "port for http listen")
	basePath = flag.String("basePath", "./", "work base path")
	dumpAnlyiseUrl = flag.String("dumpAnlyiseUrl", "https://raw.githubusercontent.com/qqwx1986/DumpAnlyise/main/Release/DumpAnlyise.exe", "DumpAnlyise.exe download url")
	chanNum = flag.Int("chanNum", 1, "channel num for dumpAnlyise")
	daemon = flag.Bool("daemon", false, "run as daemon (only windows)")
	flag.CommandLine.Parse(os.Args[2:])
	if *daemon {
		args := make([]string, 0)
		for i := 0; i < len(os.Args); i++ {
			if strings.Contains(os.Args[i], "Daemon") {
				continue
			}
			args = append(args, os.Args[i])
		}
		deamon.Run(args)
	}
	var err error
	*basePath, err = filepath.Abs(*basePath)
	cmn.TrimPath(basePath)
	crashBasePath = filepath.Join(*basePath, "crash")
	symbolBasePath = filepath.Join(*basePath, "symbol")
	errorBasePath = filepath.Join(*basePath, "error")
	dumpAnlyisePath = filepath.Join(*basePath, "tool")
	os.MkdirAll(crashBasePath, os.ModePerm)
	os.MkdirAll(symbolBasePath, os.ModePerm)
	os.MkdirAll(errorBasePath, os.ModePerm)
	os.MkdirAll(dumpAnlyisePath, os.ModePerm)

	dumpAnlyisePath = filepath.Join(dumpAnlyisePath, "DumpAnlyise.exe")
	if !cmn.IsFileExists(dumpAnlyisePath) {
		if err := downloadFileFromHttpUrl(dumpAnlyisePath, *dumpAnlyiseUrl); err != nil {
			logrus.Errorf("Downlaod from %s %s", *dumpAnlyiseUrl, err.Error())
		}
	}
	if *chanNum < 1 {
		*chanNum = 1
	}
	receiveBody := func(body io.ReadCloser) []byte {
		var buff = make([]byte, 0)
		for true {
			var tmp = make([]byte, 2048)
			recvLen, _ := body.Read(tmp)
			if recvLen > 0 {
				buff = append(buff, tmp[:recvLen]...)
			} else {
				break
			}
		}
		return buff
	}
	http.Handle("/error/", http.StripPrefix("/error/", http.FileServer(http.Dir(errorBasePath))))
	http.Handle("/crash/", http.StripPrefix("/crash/", http.FileServer(http.Dir(crashBasePath))))
	http.Handle("/symbol/", http.StripPrefix("/symbol/", http.FileServer(http.Dir(symbolBasePath))))

	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, "OK")
	})
	http.HandleFunc(cmn.HttpReceiverErrorPath, func(writer http.ResponseWriter, request *http.Request) {
		if strings.ToUpper(request.Method) != "POST" {
			writer.WriteHeader(405)
			logrus.Errorf("Http recv %s,%s,need POST", request.URL.Path, request.Method)
			return
		}
		var buff = receiveBody(request.Body)
		packet := &cmn.ErrorPacketType{}
		if err := json.Unmarshal(buff, packet); err != nil {
			logrus.Errorf("json.Unmarshal %s,%s", string(buff), err.Error())
			return
		}
		logrus.Infof("Recv error %s", string(buff))
		receiveError(packet)
		io.WriteString(writer, "OK")
	})
	http.HandleFunc(cmn.HttpUploadSymbolPath, func(writer http.ResponseWriter, request *http.Request) {
		if strings.ToUpper(request.Method) != "POST" {
			writer.WriteHeader(405)
			logrus.Errorf("Http recv %s,%s,need POST", request.URL.Path, request.Method)
			return
		}
		var buff = receiveBody(request.Body)

		// 包含 linux,android 的 symbol 和 windows的pdb
		//platform := strings.ToLower(request.Header.Get("Platform"))
		packet := &cmn.SymbolPacketType{}
		err := msgpack.Unmarshal(buff, packet)
		if err != nil {
			logrus.Errorf("msgpack.Unmarshal %d,%d,%s", request.ContentLength, len(buff), err.Error())
			return
		}
		logrus.Infof("Recv symbol %s,%d", packet.Version, len(packet.Files))
		downloadSymbol(packet)
		io.WriteString(writer, "OK")
	})
	http.HandleFunc(cmn.HttpReceiverCrashPath, func(writer http.ResponseWriter, request *http.Request) {
		if strings.ToUpper(request.Method) != "POST" {
			writer.WriteHeader(405)
			logrus.Errorf("Http recv %s,%s,need POST", request.URL.Path, request.Method)
			return
		}
		var buff = receiveBody(request.Body)
		platform := strings.ToLower(request.Header.Get("Platform"))
		if platform == "windows" {
			packet := &cmn.DumpPacketType{}
			err := msgpack.Unmarshal(buff, packet)
			if err != nil {
				logrus.Errorf("msgpack.Unmarshal %d,%d,%s", request.ContentLength, len(buff), err.Error())
				return
			}
			logrus.Infof("Recv %s,%s", packet.Version, packet.CrashGuid)
			unPackMinidump(packet, platform)
		} else if platform == "android" {

		} else if platform == "linux" {

		} else if platform == "ios" {

		} else if platform == "mac" {

		}

		io.WriteString(writer, "OK")
	})
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		logrus.Fatalf("listen err %s", err.Error())
	}
	go func() {
		logrus.Printf("Listen Http :%d", *httpPort)
		if err := http.Serve(listener, nil); err != nil {
			logrus.Fatalf("http server err %s", err.Error())
		}
	}()
	minidumpChan = make(chan *minidumpChanType, 1024)
	exitSignal = make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	waitChan = &sync.WaitGroup{}
	exitChan = make([]chan struct{}, *chanNum)
	for i := 0; i < *chanNum; i++ {
		go processChan(i)
	}
	logrus.Infof("Recv signal %s", (<-exitSignal).String())
	for _, c := range exitChan {
		c <- struct{}{}
	}
	waitChan.Wait()
	logrus.Infof("Server end(%d)", len(minidumpChan))
}

func processChan(idx int) {
	exitChan[idx] = make(chan struct{}, 1)
	waitChan.Add(1)
	defer waitChan.Done()
	logrus.Infof("processChan start %d", idx)
	defer logrus.Infof("processChan end %d", idx)
	for {
		select {
		case <-exitChan[idx]:
			logrus.Infof("")
			return
		case dump := <-minidumpChan:
			anlyiseMiniDump(dump)
		}
	}
}
