package asset

import (
	"errors"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/bonsai"
	cliClient "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Match a tabular output header
var tabularHeaderPattern = ".*\n( |─)+\n"

func TestOutdatedCommand(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	config := cli.Config.(*cliClient.MockConfig)
	config.On("Format").Return("none")

	assets := []corev2.Asset{}
	client := cli.Client.(*cliClient.MockClient)
	client.On("List", mock.Anything, &assets, mock.Anything, mock.Anything).Return(nil)

	cmd := OutdatedCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	// Match a tabular output header with nothing else under it
	assert.Regexp(tabularHeaderPattern+"$", out)
	assert.Nil(err)
}

type mockedBonsaiClient struct {
	mock.Mock
}

func (c *mockedBonsaiClient) FetchAsset(namespace, name string) (*bonsai.Asset, error) {
	args := c.Called(namespace, name)
	return args.Get(0).(*bonsai.Asset), args.Error(1)
}

func (c *mockedBonsaiClient) FetchAssetVersion(namespace, name, version string) (string, error) {
	args := c.Called(namespace, name)
	return args.String(0), args.Error(1)
}

func Test_outdatedAssets(t *testing.T) {
	type clientFunc func(*mockedBonsaiClient)

	bonsaiAsset := corev2.Asset{
		ObjectMeta: corev2.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				bonsai.URLAnnotation:       "http://127.0.0.1",
				bonsai.VersionAnnotation:   "0.1.0",
				bonsai.NamespaceAnnotation: "sensu",
				bonsai.NameAnnotation:      "testasset",
			},
		},
	}

	tests := []struct {
		name       string
		assets     []corev2.Asset
		clientFunc clientFunc
		want       []bonsai.OutdatedAsset
		wantErr    bool
	}{
		{
			name: "asset without bonsai API URL annotation is ignored",
			assets: []corev2.Asset{
				*corev2.FixtureAsset("foo"),
			},
			want: []bonsai.OutdatedAsset{},
		},
		{
			name: "asset without bonsai version returns an error",
			assets: []corev2.Asset{
				corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
						Annotations: map[string]string{
							bonsai.URLAnnotation: "http://127.0.0.1",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "asset without bonsai namespace returns an error",
			assets: []corev2.Asset{
				corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
						Annotations: map[string]string{
							bonsai.URLAnnotation:     "http://127.0.0.1",
							bonsai.VersionAnnotation: "0.1.0",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "asset without bonsai name returns an error",
			assets: []corev2.Asset{
				corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
						Annotations: map[string]string{
							bonsai.URLAnnotation:       "http://127.0.0.1",
							bonsai.VersionAnnotation:   "0.1.0",
							bonsai.NamespaceAnnotation: "sensu",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "asset without bonsai name returns an error",
			assets: []corev2.Asset{
				corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
						Annotations: map[string]string{
							bonsai.URLAnnotation:       "http://127.0.0.1",
							bonsai.VersionAnnotation:   "0.1.0",
							bonsai.NamespaceAnnotation: "sensu",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "asset with an invalid version returns an error",
			assets: []corev2.Asset{
				corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
						Annotations: map[string]string{
							bonsai.URLAnnotation:       "http://127.0.0.1",
							bonsai.VersionAnnotation:   "invalid",
							bonsai.NamespaceAnnotation: "sensu",
							bonsai.NameAnnotation:      "testasset",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "bonsai client error",
			assets: []corev2.Asset{bonsaiAsset},
			clientFunc: func(c *mockedBonsaiClient) {
				c.On("FetchAsset", "sensu", "testasset").
					Return(&bonsai.Asset{}, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name:   "invalid asset version in fetched asset",
			assets: []corev2.Asset{bonsaiAsset},
			clientFunc: func(c *mockedBonsaiClient) {
				c.On("FetchAsset", "sensu", "testasset").
					Return(
						&bonsai.Asset{Versions: []*bonsai.AssetVersionGrouping{&bonsai.AssetVersionGrouping{Version: "invalid"}}},
						nil,
					)
			},
			wantErr: true,
		},
		{
			name:   "up-to-date assset is not marked as outdated",
			assets: []corev2.Asset{bonsaiAsset},
			clientFunc: func(c *mockedBonsaiClient) {
				c.On("FetchAsset", "sensu", "testasset").
					Return(
						&bonsai.Asset{Versions: []*bonsai.AssetVersionGrouping{&bonsai.AssetVersionGrouping{Version: "0.1.0"}}},
						nil,
					)
			},
			want: []bonsai.OutdatedAsset{},
		},
		{
			name:   "older asset is marked as outdated",
			assets: []corev2.Asset{bonsaiAsset},
			clientFunc: func(c *mockedBonsaiClient) {
				c.On("FetchAsset", "sensu", "testasset").
					Return(
						&bonsai.Asset{Versions: []*bonsai.AssetVersionGrouping{&bonsai.AssetVersionGrouping{Version: "0.2.0"}}},
						nil,
					)
			},
			want: []bonsai.OutdatedAsset{
				bonsai.OutdatedAsset{
					BonsaiName:      "testasset",
					BonsaiNamespace: "sensu",
					AssetName:       "foo",
					CurrentVersion:  "0.1.0",
					LatestVersion:   "0.2.0",
				},
			},
		},
		{
			name:   "new prefixed version in Bonsai is still considered",
			assets: []corev2.Asset{bonsaiAsset},
			clientFunc: func(c *mockedBonsaiClient) {
				c.On("FetchAsset", "sensu", "testasset").
					Return(
						&bonsai.Asset{Versions: []*bonsai.AssetVersionGrouping{&bonsai.AssetVersionGrouping{Version: "v0.2.0"}}},
						nil,
					)
			},
			want: []bonsai.OutdatedAsset{
				bonsai.OutdatedAsset{
					BonsaiName:      "testasset",
					BonsaiNamespace: "sensu",
					AssetName:       "foo",
					CurrentVersion:  "0.1.0",
					LatestVersion:   "v0.2.0",
				},
			},
		},
		{
			name: "local prefixed assets are still considered outdated with new Bonsai version that's not prefixed",
			assets: []corev2.Asset{
				corev2.Asset{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
						Annotations: map[string]string{
							bonsai.URLAnnotation:       "http://127.0.0.1",
							bonsai.VersionAnnotation:   "v0.1.0",
							bonsai.NamespaceAnnotation: "sensu",
							bonsai.NameAnnotation:      "testasset",
						},
					},
				},
			},
			clientFunc: func(c *mockedBonsaiClient) {
				c.On("FetchAsset", "sensu", "testasset").
					Return(
						&bonsai.Asset{Versions: []*bonsai.AssetVersionGrouping{&bonsai.AssetVersionGrouping{Version: "0.2.0"}}},
						nil,
					)
			},
			want: []bonsai.OutdatedAsset{
				bonsai.OutdatedAsset{
					BonsaiName:      "testasset",
					BonsaiNamespace: "sensu",
					AssetName:       "foo",
					CurrentVersion:  "v0.1.0",
					LatestVersion:   "0.2.0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockedBonsaiClient{}
			if tt.clientFunc != nil {
				tt.clientFunc(client)
			}

			got, err := outdatedAssets(tt.assets, client)
			if (err != nil) != tt.wantErr {
				t.Errorf("outdatedAssets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("outdatedAssets() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
