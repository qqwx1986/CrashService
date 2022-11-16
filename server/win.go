package server

import (
	"CrashService/cmn"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func unPackData(compressData []byte) ([]byte, error) {
	r, _err := zlib.NewReader(bytes.NewBuffer(compressData))
	if _err != nil {
		return nil, _err
	}
	defer r.Close()
	buff := make([]byte, 0)
	totalRead := 0
	for {
		tmp := make([]byte, 32768)
		nRead, err := r.Read(tmp)
		if nRead == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("UnPackWriteFile Read %s", err.Error())
		}
		buff = append(buff, tmp[:nRead]...)
		totalRead += nRead
	}
	return buff, nil
}

type compressedHeader struct {
	ver              string
	directoryName    string
	fileName         string
	uncompressedSize int32
	fileCount        int32
}

func (th *compressedHeader) toString() string {
	return fmt.Sprintf("%s,%s,%s,%d,%d", th.ver, th.directoryName, th.fileName, th.uncompressedSize, th.fileCount)
}
func (th *compressedHeader) read(r io.Reader) (err error) {
	ver := make([]byte, 3, 3)
	if _, err = r.Read(ver); err != nil {
		return
	}
	var size int32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		return
	}
	directoryName := make([]byte, size, size)
	if _, err = r.Read(directoryName); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		return
	}
	fileName := make([]byte, size, size)
	if _, err = r.Read(fileName); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &th.uncompressedSize); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &th.fileCount); err != nil {
		return
	}
	th.ver = string(ver)
	th.directoryName = getString(directoryName)
	th.fileName = getString(fileName)
	return
}

func getString(data []byte) string {
	i := 0
	for ; i < len(data); i++ {
		if data[i] == 0 {
			break
		}
	}
	return string(data[:i])
}

type compressedCrashFile struct {
	currentFileIndex int32
	filename         string
	fileData         []byte
}

func (th *compressedCrashFile) read(r io.Reader) (err error) {
	if err = binary.Read(r, binary.LittleEndian, &th.currentFileIndex); err != nil {
		return
	}
	var size int32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		return
	}
	fileName := make([]byte, size, size)
	if _, err = r.Read(fileName); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		return
	}
	th.fileData = make([]byte, size, size)
	if _, err = r.Read(th.fileData); err != nil {
		return
	}
	th.filename = getString(fileName)
	return
}
func (th *compressedCrashFile) saveFile(dir string) error {
	dir = filepath.Join(dir, th.filename)
	if err := os.WriteFile(dir, th.fileData, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func unPackUECrash(compressData []byte, directory string) (string, error) {
	unCompressData, err := unPackData(compressData)
	if err != nil {
		return "", err
	}
	r := bytes.NewReader(unCompressData)
	var header compressedHeader
	if err = header.read(r); err != nil {
		return "", err
	}
	directory = filepath.Join(directory, header.directoryName)
	os.MkdirAll(directory, os.ModePerm)
	for {
		var file compressedCrashFile
		if err = file.read(r); err != nil {
			break
		}
		file.saveFile(directory)
	}
	return directory, nil
}

type crashInfo struct {
	LoginId        string
	EpicAccountId  string
	SystemId       string
	AppId          string
	AppVersion     string
	AppEnvironment string
	UploadType     string
}

func getCrashInfo(request *http.Request, directory string) error {
	var info crashInfo
	queries := request.URL.Query()
	userId := queries.Get("UserID")
	userIds := strings.Split(userId, "|")
	info.LoginId = userIds[0]
	info.EpicAccountId = userIds[1]
	info.SystemId = userIds[2]
	info.AppId = queries.Get("AppId")
	info.AppVersion = queries.Get("AppVersion")
	info.AppEnvironment = queries.Get("AppEnvironment")
	info.UploadType = queries.Get("UploadType")
	data, _ := json.MarshalIndent(info, "", "\t")
	os.WriteFile(filepath.Join(directory, "Info.json"), data, os.ModePerm)
	return nil
}
