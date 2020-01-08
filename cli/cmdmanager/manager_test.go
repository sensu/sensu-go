package cmdmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockassetgetter"
	"github.com/sensu/sensu-go/testing/mockexecutor"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	bolt "go.etcd.io/bbolt"
)

type setupCommandManagerCallbackFn func(*CommandManager, string)

func setupCommandManager(fns ...setupCommandManagerCallbackFn) (*CommandManager, error) {
	m := CommandManager{}

	cacheDir, err := ioutil.TempDir("", "")
	if err != nil {
		return &m, err
	}

	m.db, err = bolt.Open(filepath.Join(cacheDir, dbName), 0600, &bolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return &m, err
	}

	for _, fn := range fns {
		fn(&m, cacheDir)
	}

	return &m, nil
}

func nextPatchVersion(version string) (string, error) {
	versions := strings.Split(version, ".")
	i, err := strconv.Atoi(versions[2])
	if err != nil {
		return "", err
	}
	versions[2] = strconv.Itoa(i + 1)
	return strings.Join(versions, "."), nil
}

type MockBonsaiClient struct {
	mock.Mock
}

func (m *MockBonsaiClient) FetchAsset(namespace, name string) (*bonsai.Asset, error) {
	args := m.Called(namespace, name)
	return args.Get(0).(*bonsai.Asset), args.Error(1)
}

func (m *MockBonsaiClient) FetchAssetVersion(namespace, name, version string) (string, error) {
	args := m.Called(namespace, name, version)
	return args.Get(0).(string), args.Error(1)
}

func TestCommandManager_InstallCommandFromBonsai(t *testing.T) {
	type bonsaiClientFunc func(*MockBonsaiClient)

	var nilBonsaiAsset *bonsai.Asset

	bAsset := struct {
		name                string
		namespace           string
		version             string
		url                 string
		sha512              string
		fullName            string
		fullNameWithVersion string
	}{
		name:      "bonsaiasset",
		namespace: "sensu",
		version:   "1.0.0",
		url:       "https://fakeurl",
		sha512:    "2842ea31d1b9b68f25a76a3a323f9b480a6e8a499729cbd7d9ff42dd15a233951bfd7b1b14667edad979324476c9f9127ec74662795f37210291d5803d7647db",
	}
	bAsset.fullName = fmt.Sprintf("%s/%s", bAsset.namespace, bAsset.name)
	bAsset.fullNameWithVersion = fmt.Sprintf("%s:%s", bAsset.fullName, bAsset.version)

	m, err := setupCommandManager()
	if err != nil {
		t.Fatal(err)
	}
	defer m.db.Close()

	tests := []struct {
		name             string
		wantErr          bool
		errMatch         string
		alias            string
		bonsaiAssetName  string
		bonsaiClientFunc bonsaiClientFunc
	}{
		{
			name:            "bonsai base asset failure",
			wantErr:         true,
			errMatch:        "asset name must be specified in the format",
			alias:           "testalias",
			bonsaiAssetName: "",
		},
		{
			name:            "bonsai version failure",
			wantErr:         true,
			errMatch:        "Malformed version",
			alias:           "testalias",
			bonsaiAssetName: "sensu/bonsaiasset:badversion",
		},
		{
			name:            "fetch asset failure",
			wantErr:         true,
			errMatch:        "fetch asset failure",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				m.On("FetchAsset", mock.Anything, mock.Anything).
					Return(nilBonsaiAsset, errors.New("fetch asset failure"))
			},
		},
		{
			name:            "non-existent version",
			wantErr:         true,
			errMatch:        fmt.Sprintf("version \"%s\" of asset \"%s\" does not exist", bAsset.version, bAsset.fullName),
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: "0.1.0"},
					},
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
			},
		},
		{
			name:            "fetch asset version failure",
			wantErr:         true,
			errMatch:        "fetch asset version failure",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return("", errors.New("fetch asset version failure"))
			},
		},
		{
			name:            "invalid asset json",
			wantErr:         true,
			errMatch:        "unexpected end of JSON input",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return("", nil)
			},
		},
		{
			name:            "no type annotation",
			wantErr:         true,
			errMatch:        "requested asset does not have a type annotation set",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
					},
					URL:    bAsset.url,
					Sha512: bAsset.sha512,
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "invalid type annotation",
			wantErr:         true,
			errMatch:        "requested asset is not a sensuctl asset",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type": "backend",
						},
					},
					URL:    bAsset.url,
					Sha512: bAsset.sha512,
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "no provider annotation",
			wantErr:         true,
			errMatch:        "requested asset does not have a provider annotation set",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type": "sensuctl",
						},
					},
					URL:    bAsset.url,
					Sha512: bAsset.sha512,
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "invalid provider annotation",
			wantErr:         true,
			errMatch:        "requested asset is not a sensuctl/command asset",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type":     "sensuctl",
							"io.sensu.bonsai.provider": "backend/handler",
						},
					},
					URL:    bAsset.url,
					Sha512: bAsset.sha512,
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "asset without builds",
			wantErr:         true,
			errMatch:        "one or more asset builds are required",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.fullName,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type":     "sensuctl",
							"io.sensu.bonsai.provider": "sensuctl/command",
						},
					},
					URL:    bAsset.url,
					Sha512: bAsset.sha512,
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "invalid asset",
			wantErr:         true,
			errMatch:        "name cannot be empty",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					Builds: []*corev2.AssetBuild{
						{},
					},
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "valid asset with no version specified",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullName,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				newerVersion, err := nextPatchVersion(bAsset.version)
				if err != nil {
					t.Fatal(err)
				}
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
						{Version: newerVersion},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type":     "sensuctl",
							"io.sensu.bonsai.provider": "sensuctl/command",
						},
					},
					Builds: []*corev2.AssetBuild{
						{
							URL:    bAsset.url,
							Sha512: bAsset.sha512,
						},
					},
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, newerVersion).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "valid asset with version specified",
			alias:           "testalias2",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type":     "sensuctl",
							"io.sensu.bonsai.provider": "sensuctl/command",
						},
					},
					Builds: []*corev2.AssetBuild{
						{
							URL:    bAsset.url,
							Sha512: bAsset.sha512,
						},
					},
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
		{
			name:            "alias already exists",
			wantErr:         true,
			errMatch:        "the alias specified already exists",
			alias:           "testalias",
			bonsaiAssetName: bAsset.fullNameWithVersion,
			bonsaiClientFunc: func(m *MockBonsaiClient) {
				bonsaiAsset := &bonsai.Asset{
					Name: bAsset.fullName,
					Versions: []*bonsai.AssetVersionGrouping{
						{Version: bAsset.version},
					},
				}
				asset := corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name:      bAsset.name,
						Namespace: bAsset.namespace,
						Annotations: map[string]string{
							"io.sensu.bonsai.type":     "sensuctl",
							"io.sensu.bonsai.provider": "sensuctl/command",
						},
					},
					Builds: []*corev2.AssetBuild{
						{
							URL:    bAsset.url,
							Sha512: bAsset.sha512,
						},
					},
				}
				assetJSON, err := json.Marshal(asset)
				if err != nil {
					t.Fatal(err)
				}
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return(string(assetJSON), nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bonsaiClientFunc != nil {
				mockBonsaiClient := &MockBonsaiClient{}
				m.bonsaiClient = mockBonsaiClient
				tt.bonsaiClientFunc(mockBonsaiClient)
			}

			err := m.InstallCommandFromBonsai(tt.alias, tt.bonsaiAssetName)

			if (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.InstallCommandFromBonsai() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errMatch != "" {
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMatch)
				} else {
					assert.Contains(t, "", tt.errMatch)
				}
			}
		})
	}
}

func TestCommandManager_InstallCommandFromURL(t *testing.T) {
	m, err := setupCommandManager()
	if err != nil {
		t.Fatal(err)
	}
	defer m.db.Close()

	checksum := "2842ea31d1b9b68f25a76a3a323f9b480a6e8a499729cbd7d9ff42dd15a233951bfd7b1b14667edad979324476c9f9127ec74662795f37210291d5803d7647db"

	tests := []struct {
		name                  string
		alias                 string
		archiveURL            string
		checksum              string
		wantErr               bool
		errMatch              string
		expectedCommandPlugin *CommandPlugin
	}{
		{
			name:     "invalid asset",
			alias:    "",
			wantErr:  true,
			errMatch: "name cannot be empty",
		},
		{
			name:       "valid asset",
			alias:      "testasset",
			checksum:   checksum,
			archiveURL: "https://fake",
			expectedCommandPlugin: &CommandPlugin{
				Alias: "testasset",
				Asset: corev2.Asset{
					Builds: []*corev2.AssetBuild{
						{
							URL:    "https://fake",
							Sha512: checksum,
						},
					},
					ObjectMeta: corev2.ObjectMeta{
						Name:      "testasset",
						Namespace: "sensuctl",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.InstallCommandFromURL(tt.alias, tt.archiveURL, tt.checksum)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CommandManager.InstallCommandFromURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errMatch != "" {
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMatch)
				} else {
					assert.Contains(t, "", tt.errMatch)
				}
			}

			// skip asserting against boltdb if expectedCommandPlugin is nil
			if tt.expectedCommandPlugin == nil {
				return
			}

			var localCommandPlugin CommandPlugin

			if err := m.db.View(func(tx *bolt.Tx) error {
				bucket := tx.Bucket(commandBucketName)
				if bucket == nil {
					return nil
				}

				value := bucket.Get([]byte(tt.alias))
				if value != nil {
					if err := json.Unmarshal(value, &localCommandPlugin); err == nil {
						return nil
					}
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}

			assert.EqualValues(t, *tt.expectedCommandPlugin, localCommandPlugin)
		})
	}
}

func TestCommandManager_ExecCommand(t *testing.T) {
	type assetGetterFunc func(*mockassetgetter.MockAssetGetter)
	type executorFunc func(*mockexecutor.MockExecutor)

	var nilRuntimeAsset *asset.RuntimeAsset

	ctx := context.TODO()
	wg := sync.WaitGroup{}

	entity, err := getEntity()
	if err != nil {
		t.Fatal(err)
	}

	callbackFn := func(m *CommandManager, cacheDir string) {
		m.assetManager = asset.NewManager(cacheDir, entity, &wg)
		m.assetGetter = &mockassetgetter.MockAssetGetter{}
	}

	m, err := setupCommandManager(callbackFn)
	if err != nil {
		t.Fatal(err)
	}
	defer m.db.Close()

	assetPath := "/path/to/asset"
	alias := "testalias"
	checksum := "2842ea31d1b9b68f25a76a3a323f9b480a6e8a499729cbd7d9ff42dd15a233951bfd7b1b14667edad979324476c9f9127ec74662795f37210291d5803d7647db"
	testAsset := corev2.Asset{
		ObjectMeta: corev2.ObjectMeta{
			Name:      "testasset",
			Namespace: "sensuctl",
		},
		Builds: []*corev2.AssetBuild{
			{
				URL:    "https://fake",
				Sha512: checksum,
			},
		},
	}
	if err := m.registerCommandPlugin(alias, &testAsset); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                 string
		alias                string
		args                 []string
		commandEnv           []string
		wantErr              bool
		errMatch             string
		assetGetterFunc      assetGetterFunc
		executorFunc         executorFunc
		executionRequestFunc mockexecutor.RequestFunc
	}{
		{
			name:     "command does not exist",
			alias:    "nonexistentalias",
			wantErr:  true,
			errMatch: "the alias specified does not exist",
		},
		{
			name:     "asset getter failure",
			alias:    alias,
			wantErr:  true,
			errMatch: "asset getter failure",
			assetGetterFunc: func(m *mockassetgetter.MockAssetGetter) {
				m.On("Get", ctx, &testAsset).
					Return(nilRuntimeAsset, errors.New("asset getter failure"))
			},
		},
		{
			name:     "asset getter with no matched filters",
			alias:    alias,
			wantErr:  true,
			errMatch: "no asset filters were matched",
			assetGetterFunc: func(m *mockassetgetter.MockAssetGetter) {
				m.On("Get", ctx, &testAsset).
					Return(nilRuntimeAsset, nil)
			},
		},
		{
			name:     "asset getter with no matched filters",
			alias:    alias,
			wantErr:  true,
			errMatch: "no asset filters were matched",
			assetGetterFunc: func(m *mockassetgetter.MockAssetGetter) {
				m.On("Get", ctx, &testAsset).
					Return(nilRuntimeAsset, nil)
			},
		},
		{
			name:     "executor failure",
			alias:    alias,
			wantErr:  true,
			errMatch: "executor failure",
			assetGetterFunc: func(m *mockassetgetter.MockAssetGetter) {
				runtimeAsset := asset.RuntimeAsset{
					Path:   assetPath,
					SHA512: checksum,
				}
				m.On("Get", ctx, &testAsset).
					Return(&runtimeAsset, nil)
			},
			executorFunc: func(m *mockexecutor.MockExecutor) {
				m.Return(nil, errors.New("executor failure"))
			},
		},
		{
			name:  "executor success without args",
			alias: alias,
			commandEnv: []string{
				"EXAMPLE_SENSU_ENV_VAR=1234",
			},
			assetGetterFunc: func(m *mockassetgetter.MockAssetGetter) {
				runtimeAsset := asset.RuntimeAsset{
					Path:   assetPath,
					SHA512: checksum,
				}
				m.On("Get", ctx, &testAsset).
					Return(&runtimeAsset, nil)
			},
			executorFunc: func(m *mockexecutor.MockExecutor) {
				executionResponse := command.FixtureExecutionResponse(0, "success\n")
				m.Return(executionResponse, nil)
			},
			executionRequestFunc: func(c context.Context, execution command.ExecutionRequest) {
				runtimeAsset := asset.RuntimeAsset{
					Path:   assetPath,
					SHA512: checksum,
				}

				commandEnv := []string{
					"EXAMPLE_SENSU_ENV_VAR=1234",
				}
				env := environment.MergeEnvironments(os.Environ(), commandEnv)
				env = environment.MergeEnvironments(env, runtimeAsset.Env())

				assert.Equal(t, ctx, c)
				assert.Equal(t, alias, execution.Name)
				assert.Equal(t, "entrypoint", execution.Command)
				assert.Equal(t, 30, execution.Timeout)
				assert.Equal(t, env, execution.Env)
			},
		},
		{
			name:  "executor success with args",
			alias: alias,
			args:  []string{"arg1", "arg2"},
			assetGetterFunc: func(m *mockassetgetter.MockAssetGetter) {
				runtimeAsset := asset.RuntimeAsset{
					Path:   assetPath,
					SHA512: checksum,
				}
				m.On("Get", ctx, &testAsset).
					Return(&runtimeAsset, nil)
			},
			executorFunc: func(m *mockexecutor.MockExecutor) {
				executionResponse := command.FixtureExecutionResponse(0, "success\n")
				m.Return(executionResponse, nil)
			},
			executionRequestFunc: func(c context.Context, execution command.ExecutionRequest) {
				assert.Equal(t, "entrypoint arg1 arg2", execution.Command)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.assetGetterFunc != nil {
				mockAssetGetter := mockassetgetter.MockAssetGetter{}
				m.assetGetter = &mockAssetGetter
				tt.assetGetterFunc(&mockAssetGetter)
			}

			if tt.executorFunc != nil {
				mockExecutor := mockexecutor.MockExecutor{}
				if tt.executionRequestFunc != nil {
					mockExecutor.SetRequestFunc(tt.executionRequestFunc)
				}
				m.executor = &mockExecutor
				tt.executorFunc(&mockExecutor)
			}

			err := m.ExecCommand(ctx, tt.alias, tt.args, tt.commandEnv)

			if (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errMatch != "" {
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMatch)
				} else {
					assert.Contains(t, "", tt.errMatch)
				}
			}
		})
	}
}
