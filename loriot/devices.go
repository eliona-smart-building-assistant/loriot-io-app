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
	http2 "net/http"
	"regexp"
	"time"

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

type DeviceForUpdate struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type DeviceForCreate struct {
	DevEUI      string `json:"deveui,omitempty"`
	AppEUI      string `json:"appeui,omitempty"`
	JoinEUI     string `json:"joineui,omitempty"`
	AppKey      string `json:"appkey,omitempty"`
	NwkKey      string `json:"nwkkey,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	DevClass    string `json:"devclass,omitempty"`
	DevAddr     string `json:"devaddr,omitempty"`
	SeqNo       string `json:"seqno,omitempty"`
	SeqDN       string `json:"seqdn,omitempty"`
	NwkSKey     string `json:"nwkskey,omitempty"`
	NetID       string `json:"netid,omitempty"`
	AppSKey     string `json:"AppSKey,omitempty"`
	NFCntDwn    string `json:"NFCntDwn,omitempty"`
	AFCntDwn    string `json:"AFCntDwn,omitempty"`
	FNwkSIntKey string `json:"FNwkSIntKey,omitempty"`
	SNwkSIntKey string `json:"SNwkSIntKey,omitempty"`
	NwkSEncKey  string `json:"NwkSEncKey,omitempty"`
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
		meta, statusCode, err := http.ReadWithStatusCode[Meta](request, time.Duration(*config.RequestTimeout)*time.Second, true)
		if err != nil || statusCode != http2.StatusOK {
			return nil, fmt.Errorf("error reading request for %s: %d %w", fullUrl, statusCode, err)
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
	for idx, _ := range devices {
		devices[idx].AppID = appId
	}
	return devices, err
}

func getDevice(ctx context.Context, config apiserver.Configuration, appId string, devEUI string) (*Device, error) {
	request, err := http.NewRequestWithBearer(config.ApiBaseUrl+fmt.Sprintf("/1/nwk/app/%s/device/%s", appId, devEUI), config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("error creating get request for %s: %w", config.ApiBaseUrl, err)
	}
	device, statusCode, err := http.ReadWithStatusCode[Device](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if statusCode == http2.StatusNotFound {
		return nil, nil
	}
	if err != nil || statusCode != http2.StatusOK {
		return nil, fmt.Errorf("error reading get request for %s: %d %w", config.ApiBaseUrl, statusCode, err)
	}
	device.AppID = appId
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
	fullUrl := config.ApiBaseUrl + fmt.Sprintf("/1/nwk/app/%s/device/%s", device.AppID, device.DevEUI)
	deviceForUpdate := DeviceForUpdate{
		Title:       device.Title,
		Description: device.Description,
	}
	request, err := http.NewPostRequestWithBearer(fullUrl, deviceForUpdate, config.ApiToken)
	if err != nil {
		return fmt.Errorf("error creating post request for %s: %w", fullUrl, err)
	}
	_, statusCode, err := http.ReadWithStatusCode[any](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil || statusCode != http2.StatusOK {
		return fmt.Errorf("error reading post request for %s: %d %w", fullUrl, statusCode, err)
	}
	return nil
}

func postDeviceForCreate(ctx context.Context, config apiserver.Configuration, putDeviceRequest apiserver.PutDeviceRequest) (*Device, error) {
	fullUrl := config.ApiBaseUrl + fmt.Sprintf("/1/nwk/app/%s/devices", putDeviceRequest.AppID)
	deviceForCreate := DeviceForCreate{
		DevEUI:      putDeviceRequest.DevEUI,
		AppEUI:      putDeviceRequest.AppEUI,
		JoinEUI:     putDeviceRequest.JoinEUI,
		AppKey:      putDeviceRequest.AppKey,
		NwkKey:      putDeviceRequest.NwkKey,
		Title:       putDeviceRequest.Title,
		Description: putDeviceRequest.Description,
		DevClass:    putDeviceRequest.DevClass,
		DevAddr:     putDeviceRequest.DevAddr,
		SeqNo:       putDeviceRequest.SeqNo,
		SeqDN:       putDeviceRequest.SeqDN,
		NwkSKey:     putDeviceRequest.NwkSKey,
		NetID:       putDeviceRequest.NetID,
		AppSKey:     putDeviceRequest.AppSKey,
		NFCntDwn:    putDeviceRequest.NfCntDwn,
		AFCntDwn:    putDeviceRequest.AfCntDwn,
		FNwkSIntKey: putDeviceRequest.FNwkSIntKey,
		SNwkSIntKey: putDeviceRequest.SNwkSIntKey,
		NwkSEncKey:  putDeviceRequest.NwkSEncKey,
	}
	request, err := http.NewPostRequestWithBearer(fullUrl, deviceForCreate, config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("error creating post request for %s: %w", fullUrl, err)
	}
	device, statusCode, err := http.ReadWithStatusCode[Device](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil || statusCode != http2.StatusOK {
		return &device, fmt.Errorf("error reading post request for %s: %d %w", fullUrl, statusCode, err)
	}
	return &device, nil
}

func deleteDevice(ctx context.Context, config apiserver.Configuration, device Device) error {
	fullUrl := config.ApiBaseUrl + fmt.Sprintf("/1/nwk/app/%s/device/%s", device.AppID, device.DevEUI)
	request, err := http.NewDeleteRequestWithBearer(fullUrl, config.ApiToken)
	if err != nil {
		return fmt.Errorf("error creating delete request for %s: %w", fullUrl, err)
	}
	_, statusCode, err := http.ReadWithStatusCode[any](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil || statusCode != http2.StatusOK {
		return fmt.Errorf("error reading delete request for %s: %d %w", fullUrl, statusCode, err)
	}
	return nil
}

// UpsertDevice creates or updates a device using the device EUI as primary key.
func UpsertDevice(ctx context.Context, config apiserver.Configuration, request apiserver.PutDeviceRequest) (*Device, error) {
	device, err := getDevice(ctx, config, request.AppID, request.DevEUI)
	if err != nil {
		return nil, fmt.Errorf("error getting device for upserting %s: %w", request.DevEUI, err)
	}
	if device == nil {
		device, err = postDeviceForCreate(ctx, config, request)
		if err != nil {
			return device, fmt.Errorf("error creating device %s: %w", request.DevEUI, err)
		}
		return device, nil
	} else {
		if len(request.Title) > 0 {
			device.Title = request.Title
		}
		if len(request.Description) > 0 {
			device.Description = request.Description
		}
		err := postDeviceForUpdate(ctx, config, *device)
		if err != nil {
			return device, fmt.Errorf("error updating device %s: %w", request.DevEUI, err)
		}
		return device, nil
	}
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

var euiRegex = regexp.MustCompile(`^[A-Fa-f0-9]{16}$|^[A-Fa-f0-9]{32}$|^[A-Fa-f0-9]{64}$`)

// IsValidEUI checks if the given string is a valid EUI-64 identifier.
func IsValidEUI(s *string) bool {
	if s == nil {
		return false
	}
	return euiRegex.MatchString(*s)
}

func GetDeviceEUI(asset api.Asset) *string {
	for _, deviceId := range asset.DeviceIds {
		if IsValidEUI(&deviceId) {
			return common.Ptr(deviceId)
		}
	}
	return nil
}

type ExampleDevice struct {
	ID   string `eliona:"id" subtype:"info"`
	Name string `eliona:"name,filterable" subtype:"info"`
}
