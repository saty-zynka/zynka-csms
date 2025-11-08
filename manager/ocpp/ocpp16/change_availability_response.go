// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ChangeAvailabilityResponseJsonStatus string

const ChangeAvailabilityResponseJsonStatusAccepted ChangeAvailabilityResponseJsonStatus = "Accepted"
const ChangeAvailabilityResponseJsonStatusRejected ChangeAvailabilityResponseJsonStatus = "Rejected"
const ChangeAvailabilityResponseJsonStatusScheduled ChangeAvailabilityResponseJsonStatus = "Scheduled"

type ChangeAvailabilityResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status ChangeAvailabilityResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*ChangeAvailabilityResponseJson) IsResponse() {}

