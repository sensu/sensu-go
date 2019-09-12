package cmdmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/sensu/sensu-go/util/path"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	bolt "go.etcd.io/bbolt"
)

const (
	dbName = "commands.db"
)

var (
	commandBucketName = []byte("commands")
)

type CommandManager struct {
	assetManager *asset.Manager
	assetGetter  asset.Getter
	db           *bolt.DB
}

type CommandPlugin struct {
	Alias string       `json:"alias"`
	Asset corev2.Asset `json:"asset"`
}

func NewCommandManager() (*CommandManager, error) {
	m := CommandManager{}

	// create an entity for using with command asset filtering
	systemInfo, err := system.Info()
	if err != nil {
		return nil, err
	}
	meta := corev2.NewObjectMeta("sensuctl", "")
	entity := &corev2.Entity{
		EntityClass: "sensuctl",
		ObjectMeta:  meta,
		System:      systemInfo,
	}

	cacheDir := path.UserCacheDir("sensuctl")
	m.db, err = bolt.Open(filepath.Join(cacheDir, dbName), 0600, &bolt.Options{})

	// start the asset manager
	ctx := context.TODO()
	wg := sync.WaitGroup{}
	m.assetManager = asset.NewManager(cacheDir, entity, &wg)
	m.assetGetter, err = m.assetManager.StartAssetManager(ctx)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *CommandManager) InstallCommandFromBonsai(alias, bonsaiAssetName string) error {
	bAsset, err := corev2.NewBonsaiBaseAsset(bonsaiAssetName)
	if err != nil {
		return err
	}

	var version *goversion.Version
	if bAsset.Version != "" {
		version, err = goversion.NewVersion(bAsset.Version)
		if err != nil {
			return err
		}
	}

	bonsaiClient := bonsai.New(bonsai.BonsaiConfig{})
	bonsaiAsset, err := bonsaiClient.FetchAsset(bAsset.Namespace, bAsset.Name)
	if err != nil {
		return err
	}

	if version == nil {
		fmt.Println("no version specified, using latest:", bonsaiAsset.LatestVersion())
		version = bonsaiAsset.LatestVersion()
	} else if !bonsaiAsset.HasVersion(version) {
		availableVersions := bonsaiAsset.ValidVersions()
		sort.Sort(goversion.Collection(availableVersions))
		availableVersionStrs := []string{}
		for _, v := range availableVersions {
			availableVersionStrs = append(availableVersionStrs, v.String())
		}
		return fmt.Errorf("version \"%s\" of asset \"%s/%s\" does not exist\navailable versions: %s",
			version, bAsset.Namespace, bAsset.Name, strings.Join(availableVersionStrs, ", "))
	}

	fmt.Printf("fetching bonsai asset: %s/%s:%s\n", bAsset.Namespace, bAsset.Name, version)

	assetJSON, err := bonsaiClient.FetchAssetVersion(bAsset.Namespace, bAsset.Name, version.String())
	if err != nil {
		return err
	}

	var asset corev2.Asset
	if err := json.Unmarshal([]byte(assetJSON), &asset); err != nil {
		return err
	}

	return m.installCommand(alias, &asset)
}

func (m *CommandManager) InstallCommandFromURL(alias, url, sha512 string) error {
	commandAsset := &corev2.Asset{
		ObjectMeta: corev2.ObjectMeta{
			Name:      alias,
			Namespace: "",
		},
		Builds: []*corev2.AssetBuild{
			{
				URL:     url,
				Sha512:  sha512,
				Filters: []string{},
			},
		},
	}
	return m.installCommand(alias, commandAsset)
}

func (m *CommandManager) installCommand(alias string, commandAsset *corev2.Asset) error {
	err := m.registerCommandPlugin(alias, commandAsset)
	if err != nil {
		return err
	}

	fmt.Println("command was installed successfully")

	return nil
}

func (m *CommandManager) ExecCommand(alias string, args []string) error {
	commandPlugin, err := m.fetchCommandPlugin(alias)
	if err != nil {
		return err
	}

	if commandPlugin == nil {
		return errors.New("the alias specified does not exist")
	}

	ctx := context.TODO()
	runtimeAsset, err := m.assetGetter.Get(ctx, &commandPlugin.Asset)
	if err != nil {
		return err
	}

	if runtimeAsset == nil {
		return errors.New("no asset filters were matched")
	}

	env := environment.MergeEnvironments(os.Environ(), runtimeAsset.Env())

	executor := command.NewExecutor()

	ex := command.ExecutionRequest{
		Env:     env,
		Command: "entrypoint",
		Timeout: 30,
		Name:    commandPlugin.Alias,
	}

	checkExec, err := executor.Execute(context.Background(), ex)
	if err != nil {
		return err
	} else {
		fmt.Printf(checkExec.Output)
	}

	return nil
}

func (m *CommandManager) registerCommandPlugin(alias string, commandAsset *corev2.Asset) error {
	commandPlugin, err := m.fetchCommandPlugin(alias)
	if err != nil {
		return err
	}

	if commandPlugin != nil {
		return errors.New("the alias specified already exists")
	}

	key := []byte(alias)

	localCommandPlugin := CommandPlugin{
		Alias: alias,
		Asset: *commandAsset,
	}

	if err := m.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(commandBucketName)
		if err != nil {
			return err
		}

		// Though we've already attempted to do this, it's possible that a previous
		// call completed installation of the asset while this transaction
		// was blocked on serialization. Re-attempt to get the key in case that is
		// what happened.
		value := bucket.Get(key)
		if value != nil {
			// deserialize command plugin
			if err := json.Unmarshal(value, &localCommandPlugin); err == nil {
				return nil
			}
		}

		// serialize the command plugin and add the command plugin to boltdb
		commandJSON, err := json.Marshal(localCommandPlugin)
		if err != nil {
			panic(err)
		}

		return bucket.Put(key, commandJSON)
	}); err != nil {
		return err
	}

	return nil
}

func (m *CommandManager) fetchCommandPlugin(alias string) (*CommandPlugin, error) {
	key := []byte(alias)
	var localCommandPlugin *CommandPlugin

	// check boltdb for command alias
	if err := m.db.View(func(tx *bolt.Tx) error {
		// If the key exists, the bucket should already exist.
		bucket := tx.Bucket(commandBucketName)
		if bucket == nil {
			return nil
		}

		value := bucket.Get(key)
		if value != nil {
			// deserialize command mapping
			if err := json.Unmarshal(value, &localCommandPlugin); err == nil {
				return nil
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return localCommandPlugin, nil
}
