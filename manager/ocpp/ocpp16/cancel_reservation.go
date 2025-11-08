// SPDX-License-Identifier: Apache-2.0

package ocpp16

type CancelReservationJson struct {
	// ReservationId corresponds to the JSON schema field "reservationId".
	ReservationId int `json:"reservationId" yaml:"reservationId" mapstructure:"reservationId"`
}

func (*CancelReservationJson) IsRequest() {}

