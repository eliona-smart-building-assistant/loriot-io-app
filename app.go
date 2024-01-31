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

package main

import (
	"context"
	"loriot-io/apiserver"
	"loriot-io/apiservices"
	"loriot-io/conf"
	"loriot-io/eliona"
	"loriot-io/loriot"
	"net/http"
	"sync"
	"time"

	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

var once sync.Once

func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		once.Do(func() {
			log.Info("conf", "No configs in DB. Please configure the app in Eliona.")
		})
		return
	}

	for _, config := range configs {
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				_, _ = conf.SetConfigActiveState(context.Background(), config, false)
			}
			continue
		}

		if !conf.IsConfigActive(config) {
			_, _ = conf.SetConfigActiveState(context.Background(), config, true)
			log.Info("conf", "Collecting initialized with Configuration %d:\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Project IDs: %v\n",
				*config.Id,
				*config.Enable,
				config.RefreshInterval,
				*config.RequestTimeout,
				*config.ProjectIDs)
		}

		common.RunOnceWithParam(func(config apiserver.Configuration) {
			log.Info("main", "Collecting %d started.", *config.Id)
			if err := collectResources(&config); err != nil {
				return // Error is handled in the method itself.
			}
			log.Info("main", "Collecting %d finished.", *config.Id)

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, *config.Id)
	}
}

func collectResources(config *apiserver.Configuration) error {
	// Do the magic here
	return nil
}

// listenApi starts the API server and listen for requests
func listenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"),
		frontend.NewEnvironmentHandler(
			utilshttp.NewCORSEnabledHandler(
				apiserver.NewRouter(
					apiserver.NewDevicesAPIController(apiservices.NewDevicesAPIService()),
					apiserver.NewConfigurationAPIController(apiservices.NewConfigurationApiService()),
					apiserver.NewVersionAPIController(apiservices.NewVersionApiService()),
					apiserver.NewCustomizationAPIController(apiservices.NewCustomizationApiService()),
				))))
	log.Fatal("main", "API server: %v", err)
}

func listenForAssetChanges() {
	ctx := context.Background()
	for {
		assetListens, err := eliona.ListenForAssetChanges()
		if err != nil {
			log.Error("eliona", "listening for asset changes: %v", err)
			continue
		}
		log.Debug("eliona", "Started websocket listener")
		for assetListen := range assetListens {
			asset, statusCode := eliona.AssetFromAssetListen(assetListen)
			devEUI := loriot.GetDeviceEUI(asset)
			if devEUI != nil {
				log.Info("eliona", "Asset %v changed: %d", asset.Id, statusCode)

				// todo: filter assets

				configs, err := conf.GetConfigs(ctx)
				if err != nil {
					log.Error("eliona", "Error getting configs: %v", err)
					continue
				}
				for _, config := range configs {
					var device *loriot.Device
					var err error
					if statusCode == 201 {
						// Creation is only possible if the asset is unarchived and the device already exists
						device, err = loriot.UpdateDevice(ctx, config, *devEUI, asset)
					} else if statusCode == 200 {
						device, err = loriot.UpdateDevice(ctx, config, *devEUI, asset)
					} else if statusCode == 204 {
						device, err = loriot.DeleteDevice(ctx, config, *devEUI)
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
					_, err = conf.UpsertDeviceAsset(ctx, config, *device, asset, statusCode)
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
