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

func TestTriggerMessageResultHandler(t *testing.T) {
	testCases := []struct {
		name          string
		requestedMsg  types.TriggerMessageJsonRequestedMessage
		connectorId   *int
		status        types.TriggerMessageResponseJsonStatus
		expectedAttrs map[string]any
	}{
		{
			name:         "Accepted",
			requestedMsg: types.TriggerMessageJsonRequestedMessageHeartbeat,
			connectorId:  nil,
			status:       types.TriggerMessageResponseJsonStatusAccepted,
			expectedAttrs: map[string]any{
				"trigger.requested_message": "Heartbeat",
				"trigger.status":            "Accepted",
			},
		},
		{
			name:         "Rejected",
			requestedMsg: types.TriggerMessageJsonRequestedMessageStatusNotification,
			connectorId:  nil,
			status:       types.TriggerMessageResponseJsonStatusRejected,
			expectedAttrs: map[string]any{
				"trigger.requested_message": "StatusNotification",
				"trigger.status":            "Rejected",
			},
		},
		{
			name:         "NotImplemented",
			requestedMsg: types.TriggerMessageJsonRequestedMessageMeterValues,
			connectorId:  nil,
			status:       types.TriggerMessageResponseJsonStatusNotImplemented,
			expectedAttrs: map[string]any{
				"trigger.requested_message": "MeterValues",
				"trigger.status":            "NotImplemented",
			},
		},
		{
			name:         "Accepted with ConnectorId",
			requestedMsg: types.TriggerMessageJsonRequestedMessageStatusNotification,
			connectorId:  intPtr(1),
			status:       types.TriggerMessageResponseJsonStatusAccepted,
			expectedAttrs: map[string]any{
				"trigger.requested_message": "StatusNotification",
				"trigger.status":            "Accepted",
				"trigger.connector_id":      1,
			},
		},
		{
			name:         "Rejected with ConnectorId",
			requestedMsg: types.TriggerMessageJsonRequestedMessageMeterValues,
			connectorId:  intPtr(2),
			status:       types.TriggerMessageResponseJsonStatusRejected,
			expectedAttrs: map[string]any{
				"trigger.requested_message": "MeterValues",
				"trigger.status":            "Rejected",
				"trigger.connector_id":      2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := handlers.TriggerMessageResultHandler{}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.TriggerMessageJson{
					RequestedMessage: tc.requestedMsg,
					ConnectorId:      tc.connectorId,
				}
				resp := &types.TriggerMessageResponseJson{
					Status: tc.status,
				}

				err := handler.HandleCallResult(ctx, "cs001", req, resp, nil)
				require.NoError(t, err)
			}()

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", tc.expectedAttrs)
		})
	}
}

