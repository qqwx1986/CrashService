package cmn

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileContentType struct {
	FileName  string
	Content   []byte
	Len       int
	Md5       string //原始md5
	EncodeMd5 string //压缩后的md5
}
type DumpPacketType struct {
	Version   string
	CrashGuid string
	Files     []*FileContentType
}

type AndroidPacketType struct {
	Version   string
	CrashGuid string
}

type SymbolPacketType struct {
	Platform string
	Version  string
	Files    []*FileContentType
}
type DeviceInfoType struct {
	Platform  string
	DeviceId  string
	Cpu       string
	Gpu       string
	Memory    int
	Os        string
	OsSub     string
	OsVersion string
}
type FormatErrorType struct {
	CallStack string
	Reason    string
	Name      string
	CallStacks []string
}
type ErrorPacketType struct {
	Error      string
	ErrorGuid  string
	Version    string
	Time       string
	RecvTime   string
	DeviceInfo DeviceInfoType
	FormatErr  FormatErrorType
}

func Md5SumFile(fileName string) string {
	if file, err := ioutil.ReadFile(fileName); err == nil {
		sum := md5.Sum(file)
		return hex.EncodeToString(sum[:])
	}
	return ""
}

var HttpReceiverCrashPath = "/receiverCrash"
var HttpReceiverErrorPath = "/receiverError"
var HttpUploadSymbolPath = "/uploadSymbol"

type fileOutput struct {
	file     *os.File
	fileName string
	close    bool
}

func (fileOutput) getFileName() string {
	now := time.Now()
	return fmt.Sprintf("./log/%s.log", now.Format("2006-01-02"))
}
func (output *fileOutput) init() {
	if output.close {
		return
	}
	newFileName := output.getFileName()
	if output.fileName != newFileName {
		if output.file != nil {
			output.file.Close()
		}
		MkdirAllSafe("./log")
		output.file, _ = os.OpenFile(newFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		output.fileName = newFileName
	}
}
func (output *fileOutput) closeFile() {
	output.close = true
	if output.file != nil {
		output.file.Close()
		output.file = nil
	}
}
func (output *fileOutput) Write(p []byte) (n int, err error) {
	output.init()
	os.Stdout.Write(p)
	if output.file != nil {
		return output.file.Write(p)
	}
	return 0, nil
}

func IsDirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
func IsFileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
func TrimPath(path *string) {
	*path = strings.Replace(*path, `\\`, "/", -1)
	*path = strings.Replace(*path, `\`, "/", -1)
}

var fileLog *fileOutput

func CloseLogFile() {
	fileLog.closeFile()
}

func PackFile(filePath string) (fileType *FileContentType) {
	logrus.Infof("PackFile %s", filePath)
	fileType = &FileContentType{}
	defer logrus.Infof("PackFile finished %s,%d,%s,%s", filePath, fileType.Len, fileType.Md5, fileType.EncodeMd5)
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("PackFile %s", err.Error())
		return nil
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	n, err := w.Write(file)
	if n != len(file) || err != nil {
		logrus.Errorf("%d,%d", n, len(file))
	}
	w.Close()
	md5Sum := md5.Sum(file)
	encodeMd5 := md5.Sum(b.Bytes())
	fileType.FileName = filepath.Base(filePath)
	fileType.Content = b.Bytes()
	fileType.Len = len(file)
	fileType.Md5 = hex.EncodeToString(md5Sum[:])
	fileType.EncodeMd5 = hex.EncodeToString(encodeMd5[:])
	return
}
func UnPackWriteFile(path string, fileType *FileContentType) error {
	fileName := path + "/" + fileType.FileName
	logrus.Infof("UnPackWriteFile %s", fileName)

	encodeMd5 := md5.Sum(fileType.Content)
	encodeMd5Str := hex.EncodeToString(encodeMd5[:])
	if encodeMd5Str != fileType.EncodeMd5 {
		return fmt.Errorf("mismatch md5 %s:%s,%d", encodeMd5Str, fileType.EncodeMd5, len(fileType.Content))
	}
	r, _err := zlib.NewReader(bytes.NewBuffer(fileType.Content))
	if _err != nil {
		return _err
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
			logrus.Errorf("UnPackWriteFile Read %s", err.Error())
		}
		buff = append(buff, tmp[:nRead]...)
		totalRead += nRead
	}
	md5Sum := md5.Sum(buff)
	md5Str := hex.EncodeToString(md5Sum[:])
	if md5Str != fileType.Md5 {
		return fmt.Errorf("mismatch unzip md5 %s(%d):%s(%d),%d", md5Str, len(buff), fileType.Md5, fileType.Len, fileType.Len-totalRead)
	}
	logrus.Infof("UnPackWriteFile finished %s,%s,%s,%d,%d", fileType.FileName, md5Str, encodeMd5Str, totalRead, fileType.Len)
	return ioutil.WriteFile(fileName, buff, os.ModePerm)
}

var mkDirLock sync.Mutex

func MkdirAllSafe(path string) {
	mkDirLock.Lock()
	defer mkDirLock.Unlock()
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		logrus.Errorf("MkdirAllSafe %s", err.Error())
	}
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	fileLog = &fileOutput{}
	logrus.SetOutput(fileLog)
}
