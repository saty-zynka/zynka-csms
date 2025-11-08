// SPDX-License-Identifier: Apache-2.0

package ocpp16_test

import (
	"context"
	"github.com/stretchr/testify/require"
	handlers "github.com/zynka-tech/zynka-csms/manager/handlers/ocpp16"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/testutil"
	"testing"
)

func TestRemoteStartTransactionResultHandler(t *testing.T) {
	testCases := []struct {
		name           string
		idTag          string
		connectorId    *int
		status         types.RemoteStartTransactionResponseJsonStatus
		expectedAttrs  map[string]any
	}{
		{
			name:        "Accepted",
			idTag:       "MYRFIDTAG",
			connectorId: nil,
			status:      types.RemoteStartTransactionResponseJsonStatusAccepted,
			expectedAttrs: map[string]any{
				"remote_start.id_tag": "MYRFIDTAG",
				"remote_start.status": "Accepted",
			},
		},
		{
			name:        "Rejected",
			idTag:       "MYRFIDTAG",
			connectorId: nil,
			status:      types.RemoteStartTransactionResponseJsonStatusRejected,
			expectedAttrs: map[string]any{
				"remote_start.id_tag": "MYRFIDTAG",
				"remote_start.status": "Rejected",
			},
		},
		{
			name:        "Accepted with ConnectorId",
			idTag:       "MYRFIDTAG",
			connectorId: intPtr(1),
			status:      types.RemoteStartTransactionResponseJsonStatusAccepted,
			expectedAttrs: map[string]any{
				"remote_start.id_tag":      "MYRFIDTAG",
				"remote_start.status":      "Accepted",
				"remote_start.connector_id": 1,
			},
		},
		{
			name:        "Rejected with ConnectorId",
			idTag:       "MYRFIDTAG",
			connectorId: intPtr(2),
			status:      types.RemoteStartTransactionResponseJsonStatusRejected,
			expectedAttrs: map[string]any{
				"remote_start.id_tag":      "MYRFIDTAG",
				"remote_start.status":      "Rejected",
				"remote_start.connector_id": 2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := handlers.RemoteStartTransactionResultHandler{}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.RemoteStartTransactionJson{
					IdTag:       tc.idTag,
					ConnectorId: tc.connectorId,
				}
				resp := &types.RemoteStartTransactionResponseJson{
					Status: tc.status,
				}

				err := handler.HandleCallResult(ctx, "cs001", req, resp, nil)
				require.NoError(t, err)
			}()

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", tc.expectedAttrs)
		})
	}
}

func intPtr(i int) *int {
	return &i
}

