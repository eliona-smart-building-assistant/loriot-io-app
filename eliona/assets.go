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
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"loriot-io/apiserver"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/gorilla/websocket"
)

type Asset interface {
	AssetType() string
	Id() string
}

//
// Todo: Define anything for eliona like writing assets or heap data
//

func notifyUser(userId string, projectId string, assetsCreated int) error {
	receipt, _, err := client.NewClient().CommunicationAPI.
		PostNotification(client.AuthenticationContext()).
		Notification(
			api.Notification{
				User:      userId,
				ProjectId: *api.NewNullableString(&projectId),
				Message: *api.NewNullableTranslation(&api.Translation{
					De: api.PtrString(fmt.Sprintf("Template App hat %d neue Assets angelegt. Diese sind nun im Asset-Management verfügbar.", assetsCreated)),
					En: api.PtrString(fmt.Sprintf("Template app added %v new assets. They are now available in Asset Management.", assetsCreated)),
				}),
			}).
		Execute()
	log.Debug("eliona", "posted notification about CAC: %v", receipt)
	if err != nil {
		return fmt.Errorf("posting CAC notification: %v", err)
	}
	return nil
}

// UpsertAsset creates a new or gets an existing Eliona asset. Returns the new or existing asset or error if failed.
func UpsertAsset(ctx context.Context, projectID string, postDeviceRequest apiserver.PostDeviceRequest) (api.Asset, error) {

	// todo: implement asset insert or update to Eliona

	return api.Asset{}, nil
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
