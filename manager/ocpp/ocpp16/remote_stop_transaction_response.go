// SPDX-License-Identifier: Apache-2.0

package ocpp16

type RemoteStopTransactionResponseJsonStatus string

const RemoteStopTransactionResponseJsonStatusAccepted RemoteStopTransactionResponseJsonStatus = "Accepted"
const RemoteStopTransactionResponseJsonStatusRejected RemoteStopTransactionResponseJsonStatus = "Rejected"

type RemoteStopTransactionResponseJson struct {
	// Status corresponds to the JSON schema field "status".
	Status RemoteStopTransactionResponseJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*RemoteStopTransactionResponseJson) IsResponse() {}

