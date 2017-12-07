package migration

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/sensu/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// environments performs a migration on the environments in etcd, required by a
// breaking change introduced in https://github.com/sensu/sensu-go/pull/574,
// which effectively prevent users to update their environments because the new
// organization attribute is required.
func environments(storeURL string) {
	logger.Info("running environments migration")

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{storeURL},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		logger.Fatal(err)
	}

	envsResponse, err := client.Get(context.Background(), "/sensu.io/environments", clientv3.WithPrefix())
	if err != nil {
		logger.Fatal(err)
	}

	for _, kv := range envsResponse.Kvs {
		envBytes := kv.Value
		env := &types.Environment{}
		if err := json.Unmarshal(envBytes, env); err != nil {
			logger.WithError(err).Info("error unmarshaling environment: ")
			continue
		}

		if env.Organization == "" {
			pathParts := strings.Split(string(kv.Key), "/")
			if len(pathParts) != 5 {
				logger.Info("cannot parse environment key for migration: ", kv.Key)
				continue
			}

			env.Organization = pathParts[3]
			envBytes, _ := json.Marshal(env)
			_, err := client.Put(context.Background(), string(kv.Key), string(envBytes))
			if err != nil {
				logger.WithError(err).Info("error updating environment in store: ")
			}
		}
	}

	logger.Info("migration of environment completed")
}
