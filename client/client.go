package client

import (
	"CrashService/cmn"
	"bytes"
	"encoding/json"
	"flag"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//windows平台
var appName *string
var crashGUID *string
var debugSymbols *string
var crashPath string

//linux平台
var crc *string
var absLog *string

var platform = runtime.GOOS

type clientConfig struct {
	Url string
	Ver string
}

var config *clientConfig

func Run() {
	var pack []byte
	var err error
	var httpPath string
	config = &clientConfig{}
	cmn.CloseLogFile()
	if os.Args[1] == "uploadSymbol" {
		var symbolPath = flag.String("symbolPath", "", "")
		var executePath = flag.String("executePath", "", "")
		var _platform = flag.String("platform", "windows", "windows/linux/android")
		var url = flag.String("url", "http://localhost:13333", "")
		if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
			logrus.Fatalf("Parse %s", err.Error())
		}
		cmn.TrimPath(symbolPath)
		cmn.TrimPath(executePath)
		if !cmn.IsFileExists(*symbolPath) {
			logrus.Fatalf("symbolPath not exists")
		}
		if !cmn.IsFileExists(*executePath) {
			logrus.Fatalf("executePath not exists")
		}
		pack, err = uploadSymbol(*symbolPath, *executePath, strings.ToLower(*_platform))
		if err != nil {
			logrus.Fatalf("uploadSymbol %s", err.Error())
		}
		httpPath = cmn.HttpUploadSymbolPath
		config.Url = *url
		platform = *_platform
	} else {
		executerPath,_ := filepath.Abs(os.Args[0])
		executerPath = filepath.Join(filepath.Dir(executerPath),"CrashReportClient.json")
		if file, err := ioutil.ReadFile(executerPath); err == nil {
			if err = json.Unmarshal(file, config); err != nil {
				logrus.Fatalf("json.Unmarshal %s", err.Error())
			}
		} else {
			logrus.Fatalf("ioutil.ReadFile %s",err.Error())
		}
		if platform == "windows" {
			appName = flag.String("AppName", "AppName", "")
			crashGUID = flag.String("CrashGUID", "CrashGUID", "")
			debugSymbols = flag.String("DebugSymbols", "DebugSymbols", "")
			if len(os.Args) < 2 {
				logrus.Fatalf("Wrong Args")
			}
			beginArg := 2
			if strings.Contains(os.Args[1], "CrashReportClient") {
				beginArg = 3
			}
			if err := flag.CommandLine.Parse(os.Args[beginArg:]); err != nil {
				logrus.Fatalf("Parse %s", err.Error())
			}
			getVersion(*appName)
			crashPath = os.Args[beginArg-1]
			pack, err = packWinCrashDir(crashPath, *crashGUID)
			if err != nil {
				logrus.Fatalf("packWinCrashDir %s", err.Error())
			}
			httpPath = cmn.HttpReceiverCrashPath
		} else if platform == "linux" {

		}
	}
	reader := bytes.NewBuffer(pack)
	if rsp, err := httpPost(config.Url+httpPath, "application/octet-stream", reader); err != nil {
		logrus.Printf("post %s", err.Error())
	} else {
		var buff = make([]byte, 400)
		n, _ := rsp.Body.Read(buff)
		rsp.Body.Close()
		logrus.Printf("rsp %s", string(buff[:n]))
	}
}

func httpPost(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	var c = &http.Client{}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Platform", platform)
	return c.Do(req)
}
