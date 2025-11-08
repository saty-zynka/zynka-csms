// SPDX-License-Identifier: Apache-2.0

package ocpp16

type GetConfigurationJson struct {
	// Key corresponds to the JSON schema field "key".
	Key []string `json:"key,omitempty" yaml:"key,omitempty" mapstructure:"key,omitempty"`
}

func (*GetConfigurationJson) IsRequest() {}

