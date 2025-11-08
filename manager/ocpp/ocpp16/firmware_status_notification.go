// SPDX-License-Identifier: Apache-2.0

package ocpp16

type FirmwareStatusNotificationJsonStatus string

const FirmwareStatusNotificationJsonStatusDownloaded FirmwareStatusNotificationJsonStatus = "Downloaded"
const FirmwareStatusNotificationJsonStatusDownloadFailed FirmwareStatusNotificationJsonStatus = "DownloadFailed"
const FirmwareStatusNotificationJsonStatusDownloading FirmwareStatusNotificationJsonStatus = "Downloading"
const FirmwareStatusNotificationJsonStatusIdle FirmwareStatusNotificationJsonStatus = "Idle"
const FirmwareStatusNotificationJsonStatusInstallationFailed FirmwareStatusNotificationJsonStatus = "InstallationFailed"
const FirmwareStatusNotificationJsonStatusInstalling FirmwareStatusNotificationJsonStatus = "Installing"
const FirmwareStatusNotificationJsonStatusInstalled FirmwareStatusNotificationJsonStatus = "Installed"

type FirmwareStatusNotificationJson struct {
	// Status corresponds to the JSON schema field "status".
	Status FirmwareStatusNotificationJsonStatus `json:"status" yaml:"status" mapstructure:"status"`
}

func (*FirmwareStatusNotificationJson) IsRequest() {}

