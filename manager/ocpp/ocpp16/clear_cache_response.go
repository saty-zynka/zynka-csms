// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ClearCacheResponseJsonStatus string

const ClearCacheResponseJsonStatusAccepted ClearCacheResponseJsonStatus = "Accepted"
const ClearCacheResponseJsonStatusRejected ClearCacheResponseJsonStatus = "Rejected"

type ClearCacheResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status ClearCacheResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*ClearCacheResponseJson) IsResponse() {}

