package server

import (
	"CrashService/cmn"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type androidCrashChanType struct {
	path     string
	dump     string
	platform string
	version  string
}

func unPackAndroidCrash(packet *cmn.AndroidPacketType) {
	path := filepath.Join(crashBasePath, packet.Platform, packet.Version, time.Now().Format("2006-01-02"), packet.CrashGuid)
	cmn.MkdirAllSafe(path)
	if len(packet.Log) > 0 {
		ioutil.WriteFile(filepath.Join(path, "Log.txt"), []byte(packet.Log), os.ModePerm)
	}
	ioutil.WriteFile(filepath.Join(path, "Dump.txt"), []byte(packet.Dump), os.ModePerm)
	deviceInfo, _ := json.MarshalIndent(packet.DeviceInfo, "", "\t")
	ioutil.WriteFile(filepath.Join(path, "Device.json"), deviceInfo, os.ModePerm)
	go func() {
		androidCrashChan <- &androidCrashChanType{
			path:     path,
			dump:     packet.Dump,
			platform: packet.Platform,
			version:  packet.Version,
		}
	}()
}
func anlyiseAndroidCrash(crash *androidCrashChanType) {
	dumpLines := strings.Split(crash.dump, "\n")
	addrs := make([]string, 0)
	for _, line := range dumpLines {
		if len(line) == 0 {
			continue
		}
		first := strings.Split(line, "(")[1]
		addrs = append(addrs, strings.Split(first, ")")[0])
	}
	symbolPath := filepath.Join(symbolBasePath, crash.platform, crash.version, *androidSymbolName)

	argv := append(addrs, "-e", symbolPath, "-f", "-C", "-s")
	stdout, _ := os.Create(filepath.Join(crash.path, "Callstack.txt"))
	defer stdout.Close()

	cmd := exec.Command(*addr2linePath, argv...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = logrus.StandardLogger().Out
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Command err %s", err.Error())
		return
	}
	stackLines := strings.Split(strings.ReplaceAll(string(out.Bytes()), "\r\n", "\n"), "\n")
	for i, j := 0, 0; i < len(stackLines); i += 2 {
		if len(stackLines[i]) == 0 {
			break
		}
		stdout.WriteString(fmt.Sprintf("[%d](%s)%s ==> %s \n", j, addrs[j], stackLines[i], stackLines[i+1]))
		j++
	}
}
