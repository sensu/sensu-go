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
	store.On("GetAssets", mock.Anything).Return(assets, nil)

	req := newRequest("GET", "/assets", nil)
	res := processRequest(a, req)

	assert.Equal(http.StatusOK, res.Code)

	body := res.Body.Bytes()
	assetList := []*types.Asset{}
	err := json.Unmarshal(body, &assetList)

	assert.NoError(err)
	assert.EqualValues(assets, assetList)

	unauthReq := newRequest("GET", "/assets", nil)
	unauthReq = requestWithNoAccess(unauthReq)
	unauthRes := processRequest(a, unauthReq)
	assert.Equal(http.StatusUnauthorized, unauthRes.Code)
}

func TestHttpApiAssetsGetError(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}
	a := &AssetsController{
		Store: store,
	}

	var nilAssets []*types.Asset
	store.On("GetAssets", mock.Anything).Return(nilAssets, errors.New("error"))

	req := newRequest("GET", "/assets", nil)
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
	store.On("GetAssetByName", mock.Anything, "ruby21").Return(nilAsset, nil)
	notFoundReq := newRequest("GET", "/asset/ruby21", nil)
	notFoundRes := processRequest(a, notFoundReq)

	assert.Equal(http.StatusNotFound, notFoundRes.Code)

	asset := types.FixtureAsset("ruby22")
	store.On("GetAssetByName", mock.Anything, "ruby22").Return(asset, nil)
	foundReq := newRequest("GET", "/assets/ruby22", nil)
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

func TestHttpApiAssetGetUnauthorized(t *testing.T) {
	assert := assert.New(t)
	controller := AssetsController{}

	req := newRequest("GET", "/assets/ruby23", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(http.StatusUnauthorized, res.Code)
}

func TestHttpApiAssetPut(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}
	a := &AssetsController{
		Store: store,
	}

	asset := types.FixtureAsset("ruby21")
	updatedAssetJSON, _ := json.Marshal(asset)

	store.On("UpdateAsset", mock.Anything, mock.AnythingOfType("*types.Asset")).Return(nil).Run(func(args mock.Arguments) {
		receivedAsset := args.Get(1).(*types.Asset)
		assert.NoError(receivedAsset.Validate())
		assert.EqualValues(asset, receivedAsset)
	})
	putReq := newRequest("PUT", fmt.Sprintf("/assets/%s", "ruby21"), bytes.NewBuffer(updatedAssetJSON))
	putRes := processRequest(a, putReq)

	assert.Equal(http.StatusOK, putRes.Code)
}
