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

package conf

import (
	"context"
	"fmt"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"loriot-io/apiserver"
	"loriot-io/appdb"
	"loriot-io/loriot"
	http2 "net/http"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func GetDeviceAssets(ctx context.Context) ([]apiserver.DeviceAsset, error) {
	dbAssets, err := appdb.Assets(appdb.AssetWhere.LatestStatusCode.NEQ(null.Int32From(http2.StatusNoContent))).AllG(ctx)
	if err != nil {
		return nil, err
	}
	var deviceAssets []apiserver.DeviceAsset
	for _, dbAsset := range dbAssets {
		deviceAssets = append(deviceAssets, apiserver.DeviceAsset{
			ConfigID:              common.Ptr(dbAsset.ConfigurationID),
			ProjectID:             common.Ptr(dbAsset.ProjectID),
			GlobalAssetIdentifier: dbAsset.GlobalAssetID,
			AppID:                 dbAsset.AppID,
			DevEUI:                dbAsset.DevEui,
			AssetID:               dbAsset.AssetID,
			LatestStatusCode:      dbAsset.LatestStatusCode.Ptr(),
			ModifiedAt:            dbAsset.ModifiedAt.Ptr(),
		})
	}
	return deviceAssets, nil
}

func GetDbDeviceAssetById(assetId *int32) (*appdb.Asset, error) {
	if assetId == nil {
		return nil, nil
	}
	dbDeviceAssets, err := appdb.Assets(
		appdb.AssetWhere.AssetID.EQ(*assetId),
	).AllG(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching asset: %v", err)
	}
	if len(dbDeviceAssets) == 0 {
		return nil, nil
	}
	return dbDeviceAssets[0], nil
}

func UpsertDeviceAsset(ctx context.Context, config apiserver.Configuration, device loriot.Device, asset api.Asset, statusCode int32) (*apiserver.DeviceAsset, error) {
	var dbAsset appdb.Asset
	if asset.Id.Get() == nil {
		return nil, fmt.Errorf("no asset and no id present for %s", asset.AssetType)
	}
	dbAsset.AssetID = *asset.Id.Get()
	dbAsset.GlobalAssetID = asset.GlobalAssetIdentifier
	dbAsset.AppID = device.AppID
	dbAsset.ProjectID = asset.ProjectId
	dbAsset.DevEui = device.DevEUI
	dbAsset.ConfigurationID = null.Int64FromPtr(config.Id).Int64
	dbAsset.LatestStatusCode = null.Int32From(statusCode)
	dbAsset.ModifiedAt = null.TimeFrom(time.Now())
	err := dbAsset.UpsertG(ctx, true, []string{appdb.AssetColumns.AssetID}, boil.Blacklist(appdb.AssetColumns.AssetID), boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("error upserting asset %d device: %w", asset.Id.Get(), err)
	}
	return common.Ptr(apiserver.DeviceAsset{
		ConfigID:              config.Id,
		ProjectID:             common.Ptr(asset.ProjectId),
		GlobalAssetIdentifier: asset.GlobalAssetIdentifier,
		AppID:                 device.AppID,
		DevEUI:                device.DevEUI,
		AssetID:               *asset.Id.Get(),
		LatestStatusCode:      common.Ptr(statusCode),
		ModifiedAt:            common.Ptr(time.Now()),
	}), nil
}
