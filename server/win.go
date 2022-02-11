package server

import (
	"CrashService/cmn"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

type minidumpChanType struct {
	path     string
	ver      string
	platform string
}

func unPackMinidump(packet *cmn.DumpPacketType, platform string) {
	path := filepath.Join(crashBasePath, platform, packet.Version, time.Now().Format("2006-01-02"), packet.CrashGuid)
	cmn.MkdirAllSafe(path)
	for _, File := range packet.Files {
		if err := cmn.UnPackWriteFile(path, File); err != nil {
			logrus.Errorf("UnPackWriteFile %s", err.Error())
		}
	}
	go func() {
		minidumpChan <- &minidumpChanType{
			path:     path,
			ver:      packet.Version,
			platform: platform,
		}
	}()
}

func anlyiseMiniDump(miniDump *minidumpChanType) {
	argv := make([]string, 0)
	miniDumpFile := filepath.Join(miniDump.path, "UE4Minidump.dmp")
	cmn.TrimPath(&miniDumpFile)
	logrus.Infof("anlyiseMiniDump %s", miniDumpFile)
	argv = append(argv, fmt.Sprintf("-MiniDump=%s", miniDumpFile))
	path := filepath.Join(symbolBasePath, miniDump.platform, miniDump.ver)
	cmn.TrimPath(&path)
	argv = append(argv, fmt.Sprintf("-DebugSymbols=;%s", path))
	attr := &os.ProcAttr{
		Files: make([]*os.File, 3),
	}
	stdout, _ := os.Create(filepath.Join(miniDump.path, "Callstack.txt"))
	defer stdout.Close()
	attr.Files[1] = stdout
	attr.Files[2] = stdout
	process, err := os.StartProcess(dumpAnlyisePath, argv, attr)
	if err != nil {
		logrus.Errorf("StartProcess %s", err.Error())
		return
	}
	process.Wait()
}
