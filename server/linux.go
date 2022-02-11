package server

import (
	"CrashService/cmn"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type linuxCrashChanType struct {
	path string
}

func unPackLinuxCrash(packet *cmn.DumpPacketType, platform string) {
	path := filepath.Join(crashBasePath, platform, packet.Version, time.Now().Format("2006-01-02"), packet.CrashGuid)
	cmn.MkdirAllSafe(path)
	for _, File := range packet.Files {
		if err := cmn.UnPackWriteFile(path, File); err != nil {
			logrus.Errorf("UnPackWriteFile %s", err.Error())
		}
	}
	go func() {
		linuxCrashChan <- &linuxCrashChanType{
			path: path,
		}
	}()
}
func anlyiseLinuxCrash(crash *linuxCrashChanType) {

}
