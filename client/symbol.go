package client

import (
	"CrashService/cmn"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

func uploadSymbol(symbolPath string, executePath string, platform string, version string) ([]byte, error) {
	pack := &cmn.SymbolPacketType{}
	pack.Platform = platform
	pack.Files = make([]*cmn.FileContentType, 0)
	pack.Files = append(pack.Files, cmn.PackFile(symbolPath))
	if len(executePath) > 0 {
		packFile := cmn.PackFile(executePath)
		pack.Files = append(pack.Files, packFile)
		if len(version) == 0 {
			//如果不传入版本号，用md5作为版本号
			version = packFile.Md5
		}
	}
	pack.Version = version
	logrus.Infof("uploadSymbol %s", pack.Version)
	return msgpack.Marshal(pack)
}
