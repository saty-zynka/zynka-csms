// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ReserveNowResponseJsonStatus string

const ReserveNowResponseJsonStatusAccepted ReserveNowResponseJsonStatus = "Accepted"
const ReserveNowResponseJsonStatusFaulted ReserveNowResponseJsonStatus = "Faulted"
const ReserveNowResponseJsonStatusOccupied ReserveNowResponseJsonStatus = "Occupied"
const ReserveNowResponseJsonStatusRejected ReserveNowResponseJsonStatus = "Rejected"
const ReserveNowResponseJsonStatusUnavailable ReserveNowResponseJsonStatus = "Unavailable"

type ReserveNowResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status ReserveNowResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*ReserveNowResponseJson) IsResponse() {}

