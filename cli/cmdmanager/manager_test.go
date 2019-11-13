package cmdmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	bolt "go.etcd.io/bbolt"

	"github.com/sensu/sensu-go/bonsai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCommandManager() (CommandManager, error) {
	m := CommandManager{}

	cacheDir, err := ioutil.TempDir("", "")
	if err != nil {
		return m, err
	}

	m.db, err = bolt.Open(filepath.Join(cacheDir, dbName), 0600, &bolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return m, err
	}

	return m, nil
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
		m                *CommandManager
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
			m:               &m,
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
			m:               &m,
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
			m:               &m,
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
			m:               &m,
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
			name:            "asset without builds",
			m:               &m,
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
				m.On("FetchAsset", bAsset.namespace, bAsset.name).
					Return(bonsaiAsset, nil)
				m.On("FetchAssetVersion", bAsset.namespace, bAsset.name, bAsset.version).
					Return("{}", nil)
			},
		},
		{
			name:            "invalid asset",
			m:               &m,
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
			name:            "no type annotation",
			m:               &m,
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
			name:            "invalid type annotation",
			m:               &m,
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
			name:            "no provider annotation",
			m:               &m,
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
			name:            "invalid provider annotation",
			m:               &m,
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
			name:            "valid asset with no version specified",
			m:               &m,
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
			m:               &m,
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
			m:               &m,
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
			if tt.m != nil {
				mockBonsaiClient := &MockBonsaiClient{}
				tt.m.bonsaiClient = mockBonsaiClient
				tt.bonsaiClientFunc(mockBonsaiClient)
			}

			err := tt.m.InstallCommandFromBonsai(tt.alias, tt.bonsaiAssetName)

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
		m                     *CommandManager
		alias                 string
		archiveURL            string
		checksum              string
		wantErr               bool
		errMatch              string
		expectedCommandPlugin *CommandPlugin
	}{
		{
			name:    "invalid asset",
			m:       &m,
			alias:   "",
			wantErr: true,
		},
		{
			name:       "valid asset",
			m:          &m,
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
			err := tt.m.InstallCommandFromURL(tt.alias, tt.archiveURL, tt.checksum)
			if err != nil {
				t.Logf("error: %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("CommandManager.InstallCommandFromURL() error = %v, wantErr %v", err, tt.wantErr)
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

func TestCommandManager_installCommand(t *testing.T) {
	type args struct {
		alias        string
		commandAsset *corev2.Asset
	}
	tests := []struct {
		name    string
		m       *CommandManager
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.installCommand(tt.args.alias, tt.args.commandAsset); (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.installCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandManager_ExecCommand(t *testing.T) {
	type args struct {
		ctx        context.Context
		alias      string
		args       []string
		commandEnv []string
	}
	tests := []struct {
		name    string
		m       *CommandManager
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.ExecCommand(tt.args.ctx, tt.args.alias, tt.args.args, tt.args.commandEnv); (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandManager_registerCommandPlugin(t *testing.T) {
	type args struct {
		alias        string
		commandAsset *corev2.Asset
	}
	tests := []struct {
		name    string
		m       *CommandManager
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.registerCommandPlugin(tt.args.alias, tt.args.commandAsset); (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.registerCommandPlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandManager_fetchCommandPlugin(t *testing.T) {
	type args struct {
		alias string
	}
	tests := []struct {
		name    string
		m       *CommandManager
		args    args
		want    *CommandPlugin
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.fetchCommandPlugin(tt.args.alias)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.fetchCommandPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommandManager.fetchCommandPlugin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandManager_FetchCommandPlugins(t *testing.T) {
	tests := []struct {
		name    string
		m       *CommandManager
		want    []*CommandPlugin
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.FetchCommandPlugins()
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.FetchCommandPlugins() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommandManager.FetchCommandPlugins() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandManager_DeleteCommandPlugin(t *testing.T) {
	type args struct {
		alias string
	}
	tests := []struct {
		name    string
		m       *CommandManager
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.DeleteCommandPlugin(tt.args.alias); (err != nil) != tt.wantErr {
				t.Errorf("CommandManager.DeleteCommandPlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
