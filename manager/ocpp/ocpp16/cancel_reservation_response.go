// SPDX-License-Identifier: Apache-2.0

package ocpp16

type CancelReservationResponseJsonStatus string

const CancelReservationResponseJsonStatusAccepted CancelReservationResponseJsonStatus = "Accepted"
const CancelReservationResponseJsonStatusRejected CancelReservationResponseJsonStatus = "Rejected"

type CancelReservationResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status CancelReservationResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*CancelReservationResponseJson) IsResponse() {}

