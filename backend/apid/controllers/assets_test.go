package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHttpApiAssetsGet(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}
	a := &AssetsController{
		Store: store,
	}

	assets := []*types.Asset{
		types.FixtureAsset("one"),
		types.FixtureAsset("two"),
	}
	store.On("GetAssets", "default", "").Return(assets, nil)

	req, _ := http.NewRequest("GET", "/assets", nil)
	res := processRequest(a, req)

	assert.Equal(http.StatusOK, res.Code)

	body := res.Body.Bytes()
	assetList := []*types.Asset{}
	err := json.Unmarshal(body, &assetList)

	assert.NoError(err)
	assert.EqualValues(assets, assetList)
}

func TestHttpApiAssetsGetError(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}
	a := &AssetsController{
		Store: store,
	}

	var nilAssets []*types.Asset
	store.On("GetAssets", "default", "").Return(nilAssets, errors.New("error"))

	req, _ := http.NewRequest("GET", "/assets", nil)
	res := processRequest(a, req)
	body := res.Body.Bytes()

	assert.Equal(http.StatusInternalServerError, res.Code)
	assert.Equal("error\n", string(body))
}

func TestHttpApiAssetGet(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}

	a := &AssetsController{
		Store: store,
	}

	var nilAsset *types.Asset
	store.On("GetAssetByName", "default", "ruby21").Return(nilAsset, nil)
	notFoundReq, _ := http.NewRequest("GET", "/asset/ruby21", nil)
	notFoundRes := processRequest(a, notFoundReq)

	assert.Equal(http.StatusNotFound, notFoundRes.Code)

	asset := types.FixtureAsset("ruby22")
	store.On("GetAssetByName", "default", "ruby22").Return(asset, nil)
	foundReq, _ := http.NewRequest("GET", "/assets/ruby22", nil)
	foundRes := processRequest(a, foundReq)

	assert.Equal(http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()
	asset = &types.Asset{}
	err := json.Unmarshal(body, &asset)

	assert.NoError(err)
	assert.NotNil(asset.Name)
	assert.NotNil(asset.URL)
	assert.NotEqual(asset.Name, "")
	assert.NotEqual(asset.URL, "")
}

func TestHttpApiAssetPut(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}
	a := &AssetsController{
		Store: store,
	}

	asset := types.FixtureAsset("ruby21")
	updatedAssetJSON, _ := json.Marshal(asset)

	store.On("UpdateAsset", mock.AnythingOfType("*types.Asset")).Return(nil).Run(func(args mock.Arguments) {
		receivedAsset := args.Get(0).(*types.Asset)
		assert.NoError(receivedAsset.Validate())
		assert.EqualValues(asset, receivedAsset)
	})
	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/assets/%s", "ruby21"), bytes.NewBuffer(updatedAssetJSON))
	putRes := processRequest(a, putReq)

	assert.Equal(http.StatusOK, putRes.Code)
}
