/*
 * Loriot.io app API
 *
 * API to access and configure the Loriot.io app
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

type NewDevice struct {

	// Global ID in IEEE EUI64 address space that uniquely identifies the device
	DevEUI string `json:"devEUI"`

	// Application hexadecimal (uppercase) ID for Loriot
	AppID string `json:"appID"`

	// Name of the asset type to create corresponding asset in Eliona
	AssetTypeName string `json:"assetTypeName"`

	// Configuration id to define the target Loriot.io. If empty all configs are used.
	ConfigID *string `json:"configID,omitempty"`

	// Title for the new device and asset
	Title *string `json:"title,omitempty"`

	// Description for the new device and asset
	Description *string `json:"description,omitempty"`
}

// AssertNewDeviceRequired checks if the required fields are not zero-ed
func AssertNewDeviceRequired(obj NewDevice) error {
	elements := map[string]interface{}{
		"devEUI":        obj.DevEUI,
		"appID":         obj.AppID,
		"assetTypeName": obj.AssetTypeName,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertNewDeviceConstraints checks if the values respects the defined constraints
func AssertNewDeviceConstraints(obj NewDevice) error {
	return nil
}
