// SPDX-License-Identifier: Apache-2.0

package ocpp16

type ResetJsonType string

const ResetJsonTypeHard ResetJsonType = "Hard"
const ResetJsonTypeSoft ResetJsonType = "Soft"

type ResetJson struct {
	// Type corresponds to the JSON schema field "type".
	Type ResetJsonType `json:"type" yaml:"type" mapstructure:"type"`
}

func (*ResetJson) IsRequest() {}

