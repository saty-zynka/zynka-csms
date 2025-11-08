// SPDX-License-Identifier: Apache-2.0

package ocpp16

type DiagnosticsStatusNotificationJsonStatus string

const DiagnosticsStatusNotificationJsonStatusIdle DiagnosticsStatusNotificationJsonStatus = "Idle"
const DiagnosticsStatusNotificationJsonStatusUploaded DiagnosticsStatusNotificationJsonStatus = "Uploaded"
const DiagnosticsStatusNotificationJsonStatusUploadFailed DiagnosticsStatusNotificationJsonStatus = "UploadFailed"
const DiagnosticsStatusNotificationJsonStatusUploading DiagnosticsStatusNotificationJsonStatus = "Uploading"

type DiagnosticsStatusNotificationJson struct {
	// Status corresponds to the JSON schema field "status".
	Status DiagnosticsStatusNotificationJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*DiagnosticsStatusNotificationJson) IsRequest() {}

