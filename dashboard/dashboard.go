package dashboard

import (
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
)

func Dashboard(client *clientpkg.Client, frequency time.Duration) error {
	application := NewApplication(client, frequency)
	err := application.application.Run()
	if application.ticker != nil {
		application.ticker.Stop()
	}
	return err
}
