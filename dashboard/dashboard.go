package dashboard

import (
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
)

func Dashboard(client *clientpkg.Client) error {
	application := NewApplication(client, 3*time.Second)
	err := application.application.Run()
	if application.ticker != nil {
		application.ticker.Stop()
	}
	return err
}
