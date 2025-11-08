// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ChangeAvailabilityJsonType string

const ChangeAvailabilityJsonTypeInoperative ChangeAvailabilityJsonType = "Inoperative"
const ChangeAvailabilityJsonTypeOperative ChangeAvailabilityJsonType = "Operative"

type ChangeAvailabilityJson struct {
	// ConnectorId corresponds to the JSON schema field "connectorId".
	ConnectorId int `json:"connectorId" yaml:"connectorId" mapstructure:"connectorId"`

	// Type corresponds to the JSON schema field "type".
	Type ChangeAvailabilityJsonType `json:"type" yaml:"type" mapstructure:"type"`
}

func (*ChangeAvailabilityJson) IsRequest() {}

