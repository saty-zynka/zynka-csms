// SPDX-License-Identifier: Apache-2.0

package ocpp16

type GetConfigurationResponseJsonConfigurationKey struct {
	// Key corresponds to the JSON schema field "key".
	Key string `json:"key" yaml:"key" mapstructure:"key"`

	// Readonly corresponds to the JSON schema field "readonly".
	Readonly bool `json:"readonly" yaml:"readonly" mapstructure:"readonly"`

	// Value corresponds to the JSON schema field "value".
	Value *string `json:"value,omitempty" yaml:"value,omitempty" mapstructure:"value,omitempty"`
}

type GetConfigurationResponseJson struct {
	// ConfigurationKey corresponds to the JSON schema field "configurationKey".
	ConfigurationKey []GetConfigurationResponseJsonConfigurationKey `json:"configurationKey,omitempty" yaml:"configurationKey,omitempty" mapstructure:"configurationKey,omitempty"`

	// UnknownKey corresponds to the JSON schema field "unknownKey".
	UnknownKey []string `json:"unknownKey,omitempty" yaml:"unknownKey,omitempty" mapstructure:"unknownKey,omitempty"`
}

func (*GetConfigurationResponseJson) IsResponse() {}

