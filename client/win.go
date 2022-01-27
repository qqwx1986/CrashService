package client

import (
	"CrashService/cmn"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func getVersion(appName string) {
	if len(config.Ver) > 0 {
		return
	}
	if platform == "windows" {
		strs := strings.Split(appName, "-")
		if len(strs) != 2 {
			return
		}
		appName = strs[1]
		currentPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		binPath, _ := filepath.Abs(currentPath + "/../../../" + appName + "/Binaries/Win64/")
		logrus.Infof("%s  ==  %s", currentPath, binPath)
		if err := filepath.WalkDir(binPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				logrus.Errorf("%s,%s,%s", binPath, path, err.Error())
				return err
			}
			if d.IsDir() {
				return nil
			}
			logrus.Infof("")
			if d.Name() == fmt.Sprintf("%sClient.exe", appName) || d.Name() == fmt.Sprintf("%sClient-Win64-Shipping.exe", appName) || d.Name() == fmt.Sprintf("%sClient-Win64-Development.exe", appName) {
				config.Ver = cmn.Md5SumFile(path)
			}
			return nil
		}); err != nil {
			logrus.Errorf("walkdir %s", err.Error())
		} else {
			logrus.Errorf("walkdir %s,%s", binPath, config.Ver)
		}
	}
}

func packWinCrashDir(CrashPath string, CrashGUID string) ([]byte, error) {
	packet := &cmn.DumpPacketType{}
	packet.Version = config.Ver
	packet.CrashGuid = CrashGUID
	packet.Files = make([]*cmn.FileContentType, 0)
	filepath.WalkDir(CrashPath, func(path string, d fs.DirEntry, err error) error {
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
