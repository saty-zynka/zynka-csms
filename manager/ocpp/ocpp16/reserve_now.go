// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ReserveNowJson struct {
	// ConnectorId corresponds to the JSON schema field "connectorId".
	ConnectorId int `json:"connectorId" yaml:"connectorId" mapstructure:"connectorId"`

	// ExpiryDate corresponds to the JSON schema field "expiryDate".
	ExpiryDate string `json:"expiryDate" yaml:"expiryDate" mapstructure:"expiryDate"`

	// IdTag corresponds to the JSON schema field "idTag".
	IdTag string `json:"idTag" yaml:"idTag" mapstructure:"idTag"`

	// ParentIdTag corresponds to the JSON schema field "parentIdTag".
	ParentIdTag *string `json:"parentIdTag,omitempty" yaml:"parentIdTag,omitempty" mapstructure:"parentIdTag,omitempty"`

	// ReservationId corresponds to the JSON schema field "reservationId".
	ReservationId int `json:"reservationId" yaml:"reservationId" mapstructure:"reservationId"`
}

func (*ReserveNowJson) IsRequest() {}

