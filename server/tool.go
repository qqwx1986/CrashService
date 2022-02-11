package server

import (
	"CrashService/cmn"
	"flag"
)

// 手工解析Dump，补漏

func Tool() {
	cmn.CloseLogFile()

	addr2linePath = flag.String("addr2linePath", "addr2line", "addr2line for android (arm)")
	androidSymbolName = flag.String("androidSymbolName", "MyTestProject-armv7.so", "symbol name for android")
}
