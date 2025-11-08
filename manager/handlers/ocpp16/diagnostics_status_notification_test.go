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

func TestDiagnosticsStatusNotificationAllStatuses(t *testing.T) {
	testCases := []struct {
		name           string
		status         types.DiagnosticsStatusNotificationJsonStatus
		expectedStatus string
	}{
		{
			name:           "Idle",
			status:         types.DiagnosticsStatusNotificationJsonStatusIdle,
			expectedStatus: "Idle",
		},
		{
			name:           "Uploaded",
			status:         types.DiagnosticsStatusNotificationJsonStatusUploaded,
			expectedStatus: "Uploaded",
		},
		{
			name:           "UploadFailed",
			status:         types.DiagnosticsStatusNotificationJsonStatusUploadFailed,
			expectedStatus: "UploadFailed",
		},
		{
			name:           "Uploading",
			status:         types.DiagnosticsStatusNotificationJsonStatusUploading,
			expectedStatus: "Uploading",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := handlers.DiagnosticsStatusNotificationHandler{}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.DiagnosticsStatusNotificationJson{
					Status: tc.status,
				}

				resp, err := handler.HandleCall(ctx, "cs001", req)
				require.NoError(t, err)

				assert.Equal(t, &types.DiagnosticsStatusNotificationResponseJson{}, resp)
			}()

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", map[string]any{
				"diagnostics_status.status": tc.expectedStatus,
			})
		})
	}
}

