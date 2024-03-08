package plugins

import (
	"bufio"
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
	Path  string
	Log   commonlog.Logger
	Level commonlog.Level
}

func NewLogFIFO(prefix string, log commonlog.Logger, level commonlog.Level) *LogFIFO {
	path := filepath.Join(os.TempDir(), prefix+ksuid.New().String())
	return &LogFIFO{
		Path:  path,
		Log:   commonlog.NewKeyValueLogger(log, "fifo", path),
		Level: level,
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
	self.Log.Debug("creating log FIFO")
	return syscall.Mkfifo(self.Path, 0600)
}

func (self *LogFIFO) start() {
	if file, err := os.Open(self.Path); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			self.Log.Log(self.Level, 0, scanner.Text())

			if err := scanner.Err(); err != nil {
				self.Log.Error(err.Error())
			}
		}
		self.Log.Debug("closing log FIFO")
		if err := os.Remove(self.Path); err != nil {
			self.Log.Error(err.Error())
		}
	} else {
		self.Log.Error(err.Error())
	}
}
