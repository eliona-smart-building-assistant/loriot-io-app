//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eliona

import (
	"context"
	"fmt"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"loriot-io/apiserver"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/gorilla/websocket"
)

const (
	RootAssetType = "loriot_io_root"
)

type Asset interface {
	AssetType() string
	Id() string
}

// UpsertAssetWithPutDeviceRequest creates a new or gets an existing Eliona asset. Returns the new or existing asset or error if failed.
func UpsertAssetWithPutDeviceRequest(ctx context.Context, projectID string, putDeviceRequest apiserver.PutDeviceRequest) (*api.Asset, error) {
	return upsertAssetByDeviceId(ctx, api.Asset{
		DeviceIds: []string{
			putDeviceRequest.DevEUI,
		},
		ProjectId:             projectID,
		GlobalAssetIdentifier: fmt.Sprintf("%s %s", putDeviceRequest.Title, putDeviceRequest.DevEUI[len(putDeviceRequest.DevEUI)-4:]),
		Name:                  *api.NewNullableString(&putDeviceRequest.Title),
		Description:           *api.NewNullableString(&putDeviceRequest.Description),
		AssetType:             putDeviceRequest.AssetTypeName,
	})
}

func upsertAssetByDeviceId(ctx context.Context, asset api.Asset) (*api.Asset, error) {
	rootAsset, err := upsertRootAsset(asset.ProjectId)
	if err != nil || rootAsset == nil {
		return rootAsset, err
	}
	asset.ParentLocationalAssetId = rootAsset.Id
	assetReturn, _, err := client.NewClient().AssetsAPI.
		PutAsset(client.AuthenticationContext()).
		IdentifyBy("deviceId").
		Asset(asset).
		Execute()
	return assetReturn, err
}

func upsertRootAsset(projectID string) (*api.Asset, error) {
	assets, _, err := client.NewClient().AssetsAPI.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(RootAssetType).
		Execute()
	if err != nil {
		return nil, err
	}
	if len(assets) > 0 {
		return common.Ptr(assets[0]), nil
	}
	asset, _, err := client.NewClient().AssetsAPI.
		PutAsset(client.AuthenticationContext()).
		Asset(
			api.Asset{
				ProjectId:             projectID,
				GlobalAssetIdentifier: fmt.Sprintf("%s %s", RootAssetType, projectID),
				Name:                  *api.NewNullableString(common.Ptr("Loriot.io")),
				AssetType:             RootAssetType,
			}).
		Execute()
	return asset, err
}

func AssetFromAssetListen(assetListen api.AssetListen) (api.Asset, int32) {
	var statusCode int32
	if assetListen.StatusCode != nil {
		statusCode = *assetListen.StatusCode
	}
	return api.Asset{
		ResourceId:              assetListen.ResourceId,
		Id:                      assetListen.Id,
		DeviceIds:               assetListen.DeviceIds,
		ProjectId:               assetListen.ProjectId,
		GlobalAssetIdentifier:   assetListen.GlobalAssetIdentifier,
		Name:                    assetListen.Name,
		AssetType:               assetListen.AssetType,
		Latitude:                assetListen.Latitude,
		Longitude:               assetListen.Longitude,
		IsTracker:               assetListen.IsTracker,
		Description:             assetListen.Description,
		ParentFunctionalAssetId: assetListen.ParentFunctionalAssetId,
		FunctionalAssetIdPath:   assetListen.FunctionalAssetIdPath,
		ParentLocationalAssetId: assetListen.ParentLocationalAssetId,
		LocationalAssetIdPath:   assetListen.LocationalAssetIdPath,
		Tags:                    assetListen.Tags,
		ChildrenInfo:            assetListen.ChildrenInfo,
		Attachments:             assetListen.Attachments,
	}, statusCode
}

// ListenForAssetChanges returns a channel for listening of asset changes in Eliona
func ListenForAssetChanges() (chan api.AssetListen, error) {
	assets := make(chan api.AssetListen)
	var err error
	go func() {
		err = http.ListenWebSocketWithReconnectAlways(assetListenerWebsocket, time.Duration(0), assets)
	}()
	return assets, err
}

func assetListenerWebsocket() (*websocket.Conn, error) {
	return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/asset-listener?expansions=Asset.deviceIds", "X-API-Key", common.Getenv("API_TOKEN", ""))
}
