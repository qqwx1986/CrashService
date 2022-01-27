package deamon

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func Run(argv []string) {
	exePath, _ := filepath.Abs(os.Args[0])
	arr := &os.ProcAttr{
		Dir:   filepath.Dir(exePath),
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, os.Stderr},
		Sys: &syscall.SysProcAttr{},
	}
	process, err := os.StartProcess(os.Args[0], argv, arr)
	if err != nil {
		log.Fatalf("run as deamon %s", err.Error())
	}
	process.Release()
	os.Exit(0)
}
