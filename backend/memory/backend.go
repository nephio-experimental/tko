package memory

import (
	contextpkg "context"
	"sync"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
)

const Name = "memory"

var _ backend.Backend = new(MemoryBackend)

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
func (self *MemoryBackend) Connect(context contextpkg.Context) error {
	self.log.Notice("connect")
	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) Release(context contextpkg.Context) error {
	self.log.Notice("release")
	return nil
}

// ([fmt.Stringer] interface)
// ([backend.Backend] interface)
func (self *MemoryBackend) String() string {
	return "Memory"
}
