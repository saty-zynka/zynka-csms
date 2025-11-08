// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ResetResponseJsonStatus string

const ResetResponseJsonStatusAccepted ResetResponseJsonStatus = "Accepted"
const ResetResponseJsonStatusRejected ResetResponseJsonStatus = "Rejected"

type ResetResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status ResetResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*ResetResponseJson) IsResponse() {}

