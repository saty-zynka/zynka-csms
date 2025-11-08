// SPDX-License-Identifier: Apache-2.0

package ocpp16

type UnlockConnectorResponseJsonStatus string

const UnlockConnectorResponseJsonStatusUnlocked UnlockConnectorResponseJsonStatus = "Unlocked"
const UnlockConnectorResponseJsonStatusUnlockFailed UnlockConnectorResponseJsonStatus = "UnlockFailed"
const UnlockConnectorResponseJsonStatusNotSupported UnlockConnectorResponseJsonStatus = "NotSupported"

type UnlockConnectorResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status UnlockConnectorResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*UnlockConnectorResponseJson) IsResponse() {}

