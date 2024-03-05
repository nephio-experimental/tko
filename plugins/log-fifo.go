package plugins

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/segmentio/ksuid"
	"github.com/tliron/commonlog"
)

// TODO: move to commonlog

//
// LogFIFO
//

type LogFIFO struct {
	Path string
	Log  commonlog.Logger
}

func NewLogFIFO(prefix string, log commonlog.Logger) *LogFIFO {
	return &LogFIFO{
		Path: filepath.Join(os.TempDir(), prefix+ksuid.New().String()),
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
	self.Log.Infof("creating log FIFO: %s", self.Path)
	return syscall.Mkfifo(self.Path, 0600)
}

func (self *LogFIFO) start() {
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
				self.Log.Infof("stopped reading from log FIFO: %s", self.Path)
				break
			}
		}
	} else {
		self.Log.Error(err.Error())
	}
}
