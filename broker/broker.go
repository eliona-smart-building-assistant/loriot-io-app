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

package broker

import (
	"context"
	"fmt"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"loriot-io/apiserver"
	"loriot-io/app"
	"loriot-io/eliona"
	"loriot-io/loriot"
	"net/http"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func ListenForAssetChanges() {
	ctx := context.Background()
	for {

		// Listen for asset changes in Eliona
		assetListens, err := eliona.ListenForAssetChanges()
		if err != nil {
			log.Error("eliona", "listening for asset changes: %v", err)
			continue
		}
		log.Debug("eliona", "Started websocket listener")

		for assetListen := range assetListens {
			asset, statusCode := eliona.AssetFromAssetListen(assetListen)

			// Try to get apps information about asset device
			dbAssetDevice, err := app.GetDbDeviceAssetById(asset.Id.Get())
			if err != nil {
				log.Error("eliona", "Error selecting device asset: %v", err)
			}

			// Try to get device EUI. If not defined (e.g. after archiving in frontend) use the app data to find the device EUI
			devEUI := loriot.GetDeviceEUI(asset)
			if devEUI == nil && dbAssetDevice != nil && loriot.IsValidEUI(&dbAssetDevice.DevEui) {
				asset.DeviceIds = []string{
					dbAssetDevice.DevEui,
				}
				devEUI = common.Ptr(dbAssetDevice.DevEui)
			}

			// Perform the action (recreate, delete, update) triggert by Eliona for each config
			if devEUI != nil {
				log.Info("eliona", "Asset %v changed: %d", asset.Id, statusCode)

				configs, err := app.GetConfigs(ctx)
				if err != nil {
					log.Error("eliona", "Error getting configs: %v", err)
					continue
				}
				for _, config := range configs {
					if !app.IsConfigEnabled(config) {
						continue
					}

					// check if project is defined for this asset
					if !sliceContains(app.ProjIds(config), asset.ProjectId) {
						log.Info("asset", "Modified asset with project ID %s doesn't matches project IDs from configuration %d", asset.ProjectId, config.Id)
						continue
					}

					// Perform the action (recreate, delete, update)
					var device *loriot.Device
					var err error

					// Perform creation action.
					if statusCode == http.StatusCreated {
						// at the moment no further action must be performed if an asset is recreated
						app.NotifyUser(config.UserId, &asset.ProjectId, &api.Translation{
							De: api.PtrString(fmt.Sprintf("Loriot App hat Gerät '%s' und Asset '%d' angelegt.", *devEUI, *asset.Id.Get())),
							En: api.PtrString(fmt.Sprintf("Loriot app created device '%s' and asset '%d'.", *devEUI, *asset.Id.Get())),
						})
					}

					// Perform update action.
					if statusCode == http.StatusOK {
						device, err = loriot.UpdateDevice(ctx, config, *devEUI, asset)
						if err == nil {
							app.NotifyUser(config.UserId, &asset.ProjectId, &api.Translation{
								De: api.PtrString(fmt.Sprintf("Loriot App hat Gerät '%s' und Asset '%d' geändert.", *devEUI, *asset.Id.Get())),
								En: api.PtrString(fmt.Sprintf("Loriot app updated device '%s' and asset '%d'.", *devEUI, *asset.Id.Get())),
							})
						}
					}

					// Perform delete action. Perform is always possible.
					if statusCode == http.StatusNoContent {
						device, err = loriot.DeleteDevice(ctx, config, *devEUI)
						if err == nil {
							app.NotifyUser(config.UserId, &asset.ProjectId, &api.Translation{
								De: api.PtrString(fmt.Sprintf("Loriot App hat Gerät '%s' und Asset '%d' gelöscht.", *devEUI, *asset.Id.Get())),
								En: api.PtrString(fmt.Sprintf("Loriot app deleted device '%s' and asset '%d'.", *devEUI, *asset.Id.Get())),
							})
						}
					}
					if err != nil {
						log.Error("loriot", "Error perform operation %d for device %s: %v", statusCode, *devEUI, err)
						continue
					}
					if device == nil {
						log.Warn("loriot", "Device %s for operation %d not found. Changes from Eliona are ignored", *devEUI, statusCode)
						continue
					}
					log.Info("loriot", "Device %s operation %d successfully performed.", *devEUI, statusCode)
					_, err = app.UpsertDeviceAsset(ctx, config, *device, asset, statusCode)
					if err != nil {
						log.Error("app", "Error updating app's device database for operation %d for device %s: %v", statusCode, *devEUI, err)
					}
				}
			}

		}
		log.Warn("Eliona", "Websocket connection broke. Restarting in 5 seconds.")
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
}

func UpsertDevice(ctx context.Context, putDeviceRequest apiserver.PutDeviceRequest) ([]apiserver.DeviceAsset, error) {
	if !loriot.IsValidEUI(&putDeviceRequest.DevEUI) {
		return nil, fmt.Errorf("invalid device EUI: %s", putDeviceRequest.DevEUI)
	}
	var deviceAssets []apiserver.DeviceAsset

	// For all configs update device and asset
	configs, err := app.GetConfigs(ctx)
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configuration found")
	}
	if err != nil {
		return deviceAssets, err
	}
	for _, config := range configs {
		if !app.IsConfigEnabled(config) {
			continue
		}

		if putDeviceRequest.ConfigID != nil && config.Id != nil && int64(*putDeviceRequest.ConfigID) != *config.Id {
			continue
		}

		// Upsert device
		device, err := loriot.UpsertDevice(ctx, config, putDeviceRequest)
		if err != nil {
			return deviceAssets, err
		}
		if device == nil {
			continue
		}

		// For all project IDs upserts the corresponding asset
		for _, projectID := range app.ProjIds(config) {

			asset, err := eliona.UpsertAssetWithPutDeviceRequest(ctx, projectID, putDeviceRequest)
			if err != nil {
				return deviceAssets, err
			}
			if asset == nil {
				continue
			}

			// remember the asset info inside app
			deviceAsset, err := app.UpsertDeviceAsset(ctx, config, *device, *asset, 201)
			if err != nil {
				return deviceAssets, err
			}
			if deviceAsset != nil {
				deviceAssets = append(deviceAssets, *deviceAsset)
			}

		}
	}
	return deviceAssets, nil
}

func sliceContains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
