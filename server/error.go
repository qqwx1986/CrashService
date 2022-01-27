package server

import (
	"CrashService/cmn"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func receiveError(errPacket *cmn.ErrorPacketType) {
	path := filepath.Join(errorBasePath, errPacket.DeviceInfo.Platform, time.Now().Format("2006-01-02"))
	cmn.MkdirAllSafe(path)
	errBuffer := []byte(errPacket.Error)
	if json.Valid(errBuffer) {
		if err := json.Unmarshal(errBuffer, &errPacket.FormatErr); err == nil {
			errPacket.Error = ""
			errPacket.FormatErr.CallStacks = strings.Split(errPacket.FormatErr.CallStack, "\n\t")
			if len(errPacket.FormatErr.CallStacks) > 0 {
				errPacket.FormatErr.CallStack = ""
			}
		}
	}
	errPacket.RecvTime = time.Now().Format(time.RFC3339)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")
	if err := enc.Encode(errPacket); err != nil {
	}
	ioutil.WriteFile(filepath.Join(path, errPacket.ErrorGuid+".json"), buf.Bytes(), os.ModePerm)
}
