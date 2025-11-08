// SPDX-License-Identifier: Apache-2.0

package ocpp16_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	handlers "github.com/zynka-tech/zynka-csms/manager/handlers/ocpp16"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/testutil"
	"testing"
)

func TestFirmwareStatusNotificationAllStatuses(t *testing.T) {
	testCases := []struct {
		name           string
		status         types.FirmwareStatusNotificationJsonStatus
		expectedStatus string
	}{
		{
			name:           "Downloaded",
			status:         types.FirmwareStatusNotificationJsonStatusDownloaded,
			expectedStatus: "Downloaded",
		},
		{
			name:           "DownloadFailed",
			status:         types.FirmwareStatusNotificationJsonStatusDownloadFailed,
			expectedStatus: "DownloadFailed",
		},
		{
			name:           "Downloading",
			status:         types.FirmwareStatusNotificationJsonStatusDownloading,
			expectedStatus: "Downloading",
		},
		{
			name:           "Idle",
			status:         types.FirmwareStatusNotificationJsonStatusIdle,
			expectedStatus: "Idle",
		},
		{
			name:           "InstallationFailed",
			status:         types.FirmwareStatusNotificationJsonStatusInstallationFailed,
			expectedStatus: "InstallationFailed",
		},
		{
			name:           "Installing",
			status:         types.FirmwareStatusNotificationJsonStatusInstalling,
			expectedStatus: "Installing",
		},
		{
			name:           "Installed",
			status:         types.FirmwareStatusNotificationJsonStatusInstalled,
			expectedStatus: "Installed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := handlers.FirmwareStatusNotificationHandler{}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.FirmwareStatusNotificationJson{
					Status: tc.status,
				}

				resp, err := handler.HandleCall(ctx, "cs001", req)
				require.NoError(t, err)

				assert.Equal(t, &types.FirmwareStatusNotificationResponseJson{}, resp)
			}()

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", map[string]any{
				"firmware_status.status": tc.expectedStatus,
			})
		})
	}
}

