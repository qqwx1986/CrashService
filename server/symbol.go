package server

import (
	"CrashService/cmn"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

func downloadSymbol(symbolType *cmn.SymbolPacketType) {
	symbolPath := filepath.Join(symbolBasePath, symbolType.Platform, symbolType.Version)
	cmn.MkdirAllSafe(symbolPath)
	if !cmn.IsDirExists(symbolPath) {
		logrus.Errorf("%s not exists", symbolPath)
		return
	}
	for _, file := range symbolType.Files {
		if err := cmn.UnPackWriteFile(symbolPath, file); err != nil {
			logrus.Errorf("downloadSymbol.UnPackWriteFile %s", err.Error())
		}
	}
}
