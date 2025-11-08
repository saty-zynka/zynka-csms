// SPDX-License-Identifier: Apache-2.0

package ocpp16

type UnlockConnectorJson struct {
	// ConnectorId corresponds to the JSON schema field "connectorId".
	ConnectorId int `json:"connectorId" yaml:"connectorId" mapstructure:"connectorId"`
}

func (*UnlockConnectorJson) IsRequest() {}

