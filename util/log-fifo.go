package util

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/tliron/commonlog"
)

//
// LogFIFO
//

type LogFIFO struct {
	Path string
	Log  commonlog.Logger
}

func NewLogFIFO(name string, log commonlog.Logger) *LogFIFO {
	return &LogFIFO{
		Path: filepath.Join(os.TempDir(), name),
		Log:  log,
	}
}

func (self *LogFIFO) Start() error {
	if err := self.create(); err == nil {
		go self.start()
		return nil
	} else {
		return err
	}
}

func (self *LogFIFO) create() error {
	if err := os.Remove(self.Path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	//self.Log.Errorf("creating log file: %s", self.Path)
	return syscall.Mkfifo(self.Path, 0600)
}

func (self *LogFIFO) start() {
	//self.Log.Errorf("reading log file: %s", self.Path)
	if file, err := os.Open(self.Path); err == nil {
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			if line, err := reader.ReadString('\n'); err == nil {
				self.Log.Notice(line)
			} else {
				if err != io.EOF {
					self.Log.Error(err.Error())
				}
				//self.Log.Errorf("stopped reading log file: %s", self.Path)
				break
			}
		}
	} else {
		self.Log.Error(err.Error())
	}
}
