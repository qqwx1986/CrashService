package client

import (
	"CrashService/cmn"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

func uploadSymbol(symbolPath string, executePath string, platform string) ([]byte, error) {
	pack := &cmn.SymbolPacketType{}
	pack.Platform = platform
	pack.Files = make([]*cmn.FileContentType, 2)
	pack.Files[0] = cmn.PackFile(symbolPath)
	pack.Files[1] = cmn.PackFile(executePath)
	//用md5作为版本号
	pack.Version = pack.Files[1].Md5
	logrus.Infof("uploadSymbol %s", pack.Version)
	return msgpack.Marshal(pack)
}
