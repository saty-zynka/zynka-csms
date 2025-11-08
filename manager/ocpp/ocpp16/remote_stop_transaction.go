// SPDX-License-Identifier: Apache-2.0

package ocpp16

type RemoteStopTransactionJson struct {
	// TransactionId corresponds to the JSON schema field "transactionId".
	TransactionId int `json:"transactionId" yaml:"transactionId" mapstructure:"transactionId"`
}

func (*RemoteStopTransactionJson) IsRequest() {}

