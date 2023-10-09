package memory

import (
	"sync"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/commonlog"
)

//
// MemoryBackend
//

type MemoryBackend struct {
	templates   map[string]*backend.Template
	sites       map[string]*backend.Site
	deployments map[string]*Deployment
	plugins     map[backend.PluginID]*backend.Plugin

	log                commonlog.Logger
	modificationWindow int64 // microseconds

	lock sync.Mutex
}

// modificationWindow in seconds
func NewMemoryBackend(modificationWindow int, log commonlog.Logger) *MemoryBackend {
	return &MemoryBackend{
		templates:          make(map[string]*backend.Template),
		sites:              make(map[string]*backend.Site),
		deployments:        make(map[string]*Deployment),
		plugins:            make(map[backend.PluginID]*backend.Plugin),
		log:                log,
		modificationWindow: int64(modificationWindow) * 1_000_000,
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) Connect() error {
	self.log.Notice("connect")
	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) Release() error {
	self.log.Notice("release")
	return nil
}
