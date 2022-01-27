package deamon

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func Run(argv []string) {
	var sysproc = &syscall.SysProcAttr{
		CreationFlags: 0x00000008,
	}
	arr := &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, os.Stderr},
		Sys:   sysproc,
	}
	exePath, _ := filepath.Abs(os.Args[0])
	process, err := os.StartProcess(exePath, argv, arr)
	if err != nil {
		log.Fatalf("run as deamon %s", err.Error())
	}
	process.Release()
	os.Exit(0)
}
