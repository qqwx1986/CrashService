package client

import (
	"CrashService/cmn"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
	"io/fs"
	"path/filepath"
)

func packLinuxCrashDir(crashPath string) ([]byte, error) {
	packet := &cmn.DumpPacketType{}
	packet.Version = config.Ver
	packet.CrashGuid = filepath.Base(crashPath)
	packet.Files = make([]*cmn.FileContentType, 0)
	filepath.WalkDir(crashPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logrus.Errorf("packLinuxCrashDir %s,%s,%s", crashPath, path, err.Error())
			return nil
		}
		if d.IsDir() {
			return nil
		}
		fileType := cmn.PackFile(path)
		if fileType != nil {
			packet.Files = append(packet.Files, fileType)
		}
		fmt.Printf("FileName %s,%s,%s,%d,%d\n", d.Name(), fileType.Md5, fileType.EncodeMd5, fileType.Len, len(fileType.Content))
		return nil
	})
	return msgpack.Marshal(packet)
}
