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

package loriot

import (
	"context"
	"fmt"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"loriot-io/apiserver"
	"regexp"
	"time"

	"github.com/eliona-smart-building-assistant/go-eliona/utils"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

type Meta struct {
	Apps    []App    `json:"apps"`
	Devices []Device `json:"devices"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	PerPage int      `json:"perPage"`
}

type App struct {
	Id             int64     `json:"_id"`
	AppHexID       string    `json:"appHexId"`
	Name           string    `json:"name"`
	OwnerID        int       `json:"ownerid"`
	OrganizationID int       `json:"organizationId"`
	Visibility     string    `json:"visibility"`
	Created        time.Time `json:"created"`
	Devices        int       `json:"devices"`
	DeviceLimit    int       `json:"deviceLimit"`
	Orx            bool      `json:"orx"`
	CanSend        bool      `json:"cansend"`
	CanOTAA        bool      `json:"canotaa"`
	Suspended      bool      `json:"suspended"`
	MasterKey      string    `json:"masterkey"`
	ClientsLimit   int       `json:"clientsLimit"`
	PublishAppSKey bool      `json:"publishAppSKey"`
	AccessRights   []struct {
		Token           string `json:"token"`
		Data            bool   `json:"data"`
		AppServer       bool   `json:"appServer"`
		DevProvisioning bool   `json:"devProvisioning"`
	} `json:"accessRights"`
	Output string `json:"output"`
}

type Device struct {
	AppID             string    `json:"appid"`
	Id                string    `json:"_id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	AppEUI            string    `json:"appeui"`
	OrganizationID    int       `json:"organizationId"`
	Visibility        string    `json:"visibility"`
	DevEUI            string    `json:"deveui"`
	DevAddr           string    `json:"devaddr"`
	SeqNo             int       `json:"seqno"`
	SeqDN             int       `json:"seqdn"`
	AdrCnt            int       `json:"adrCnt"`
	NFCntDwn          int       `json:"NFCntDwn"`
	AFCntDwn          int       `json:"AFCntDwn"`
	Adr               bool      `json:"adr"`
	CreatedAt         time.Time `json:"createdAt"`
	Bat               int       `json:"bat"`
	DevSnr            int       `json:"devSnr"`
	CanSend           bool      `json:"canSend"`
	CanSendFOPTS      bool      `json:"canSendFOPTS"`
	CanSendPayload    bool      `json:"canSendPayload"`
	CanSendADR        bool      `json:"canSendADR"`
	CanRoaming        bool      `json:"canRoaming"`
	AllowDevStatusReq bool      `json:"allowDevStatusReq"`
	Nonce             int       `json:"nonce"`
	Bw                int       `json:"bw"`
	Freq              int       `json:"freq"`
	Gw                string    `json:"gw"`
	LastJoin          time.Time `json:"lastJoin"`
	LastSeen          time.Time `json:"lastSeen"`
	Sf                int       `json:"sf"`
	Rssi              int       `json:"rssi"`
	Snr               float64   `json:"snr"`
	Ant               int       `json:"ant"`
	LastDevStatusReq  time.Time `json:"lastDevStatusReq"`
	LastDevStatusSeen time.Time `json:"lastDevStatusSeen"`
}

func getFromApi[T any](_ context.Context, config apiserver.Configuration, getData func(Meta) []T, urlFormat string, args ...interface{}) ([]T, error) {
	var results []T
	var page = 1
	var perPage = 100

	for {
		url := fmt.Sprintf(urlFormat, args...)
		fullUrl := config.ApiBaseUrl + fmt.Sprintf(url+"?page=%d&perPage=%d", page, perPage)
		request, err := http.NewRequestWithBearer(fullUrl, config.ApiToken)
		if err != nil {
			return nil, fmt.Errorf("error creating request for %s: %w", fullUrl, err)
		}
		meta, err := http.Read[Meta](request, time.Duration(*config.RequestTimeout)*time.Second, true)
		if err != nil {
			return nil, fmt.Errorf("error reading request for %s: %w", fullUrl, err)
		}
		results = append(results, getData(meta)...)
		if page*perPage >= meta.Total {
			break
		}
		page++
	}

	return results, nil
}

func getApps(ctx context.Context, config apiserver.Configuration) ([]App, error) {
	var apps []App
	apps, err := getFromApi[App](ctx, config, func(meta Meta) []App { return meta.Apps }, "/1/nwk/apps")
	return apps, err
}

func getDevices(ctx context.Context, config apiserver.Configuration, appId string) ([]Device, error) {
	var devices []Device
	devices, err := getFromApi[Device](ctx, config, func(meta Meta) []Device { return meta.Devices }, "/1/nwk/app/%s/devices", appId)
	return devices, err
}

func getDevice(ctx context.Context, config apiserver.Configuration, appId string, devEUI string) (*Device, error) {
	request, err := http.NewRequestWithBearer(config.ApiBaseUrl+fmt.Sprintf("/1/nwk/app/%s/device/%s", appId, devEUI), config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", config.ApiBaseUrl, err)
	}
	device, err := http.Read[Device](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("error reading request for %s: %w", config.ApiBaseUrl, err)
	}
	return &device, nil
}

func searchDevice(ctx context.Context, config apiserver.Configuration, devEUI string) (*Device, error) {
	apps, err := getApps(ctx, config)
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		device, err := getDevice(ctx, config, app.AppHexID, devEUI)
		if err != nil {
			return nil, err
		}
		if device != nil {
			return device, nil
		}
	}
	return nil, nil
}

func postDeviceForUpdate(ctx context.Context, config apiserver.Configuration, device Device) error {
	// todo: implement post device call
	return nil
}

func postDeviceForCreate(ctx context.Context, config apiserver.Configuration, type_ string, request apiserver.PostDeviceRequest) (*Device, error) {
	// todo: implement post device call
	return nil, nil
}

func deleteDevice(ctx context.Context, config apiserver.Configuration, device Device) error {
	// todo: implement delete device call
	return nil
}

// UpsertDevice creates or updates a device using the device EUI as primary key.
func UpsertDevice(ctx context.Context, config apiserver.Configuration, type_ string, request apiserver.PostDeviceRequest) (*Device, error) {
	device, err := getDevice(ctx, config, request.AppID, request.DevEUI)
	if err != nil {
		return nil, fmt.Errorf("error getting device for upserting %s: %w", request.DevEUI, err)
	}
	if device == nil {
		device, err = postDeviceForCreate(ctx, config, type_, request)
		if err != nil {
			return device, fmt.Errorf("error creating device %s: %w", request.DevEUI, err)
		}
	} else {
		if request.Title != nil {
			device.Title = *request.Title
		}
		if request.Description != nil {
			device.Description = *request.Description
		}
		err := postDeviceForUpdate(ctx, config, *device)
		if err != nil {
			return device, fmt.Errorf("error updating device %s: %w", request.DevEUI, err)
		}
	}
	return device, nil
}

func UpdateDevice(ctx context.Context, config apiserver.Configuration, devEUI string, asset api.Asset) (*Device, error) {
	device, err := searchDevice(ctx, config, devEUI)
	if err != nil {
		return device, fmt.Errorf("error getting device for updating %s: %w", devEUI, err)
	}
	if device == nil {
		return nil, nil
	}
	if asset.Name.IsSet() {
		device.Title = *asset.Name.Get()
	} else {
		device.Title = devEUI
	}
	if asset.Description.IsSet() {
		device.Description = *asset.Description.Get()
	} else {
		device.Description = ""
	}
	err = postDeviceForUpdate(ctx, config, *device)
	if err != nil {
		return device, fmt.Errorf("error posting device %s: %w", devEUI, err)
	}
	return device, nil
}

func DeleteDevice(ctx context.Context, config apiserver.Configuration, devEUI string) (*Device, error) {
	device, err := searchDevice(ctx, config, devEUI)
	if err != nil {
		return device, fmt.Errorf("error getting device for deletion %s: %w", devEUI, err)
	}
	if device == nil {
		return nil, nil
	}
	err = deleteDevice(ctx, config, *device)
	if err != nil {
		return device, fmt.Errorf("error deleting device %s: %w", devEUI, err)
	}
	return device, nil
}

var euiRegex = regexp.MustCompile(`^[A-Fa-f0-9]{32}$`)

// isValidEUI checks if the given string is a valid EUI-64 identifier.
func isValidEUI(s *string) bool {
	if s == nil {
		return false
	}
	return euiRegex.MatchString(*s)
}

func GetDeviceEUI(asset api.Asset) *string {
	for _, deviceId := range asset.DeviceIds {
		if isValidEUI(&deviceId) {
			return common.Ptr(deviceId)
		}
	}
	return nil
}

type ExampleDevice struct {
	ID   string `eliona:"id" subtype:"info"`
	Name string `eliona:"name,filterable" subtype:"info"`
}

func GetTags(config apiserver.Configuration) ([]ExampleDevice, error) {
	return nil, nil
}

func (tag *ExampleDevice) AdheresToFilter(filter [][]apiserver.FilterRule) (bool, error) {
	f := apiFilterToCommonFilter(filter)
	fp, err := utils.StructToMap(tag)
	if err != nil {
		return false, fmt.Errorf("converting strict to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func apiFilterToCommonFilter(input [][]apiserver.FilterRule) [][]common.FilterRule {
	result := make([][]common.FilterRule, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = make([]common.FilterRule, len(input[i]))
		for j := 0; j < len(input[i]); j++ {
			result[i][j] = common.FilterRule{
				Parameter: input[i][j].Parameter,
				Regex:     input[i][j].Regex,
			}
		}
	}
	return result
}
