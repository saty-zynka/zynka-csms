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

func TestReserveNowResultHandler(t *testing.T) {
	testCases := []struct {
		name          string
		reservationId int
		status        types.ReserveNowResponseJsonStatus
		expectedAttrs map[string]any
	}{
		{
			name:          "Accepted",
			reservationId: 123,
			status:        types.ReserveNowResponseJsonStatusAccepted,
			expectedAttrs: map[string]any{
				"reserve_now.reservation_id": 123,
				"reserve_now.status":          "Accepted",
			},
		},
		{
			name:          "Rejected",
			reservationId: 456,
			status:        types.ReserveNowResponseJsonStatusRejected,
			expectedAttrs: map[string]any{
				"reserve_now.reservation_id": 456,
				"reserve_now.status":          "Rejected",
			},
		},
		{
			name:          "Faulted",
			reservationId: 789,
			status:        types.ReserveNowResponseJsonStatusFaulted,
			expectedAttrs: map[string]any{
				"reserve_now.reservation_id": 789,
				"reserve_now.status":          "Faulted",
			},
		},
		{
			name:          "Occupied",
			reservationId: 101,
			status:        types.ReserveNowResponseJsonStatusOccupied,
			expectedAttrs: map[string]any{
				"reserve_now.reservation_id": 101,
				"reserve_now.status":          "Occupied",
			},
		},
		{
			name:          "Unavailable",
			reservationId: 202,
			status:        types.ReserveNowResponseJsonStatusUnavailable,
			expectedAttrs: map[string]any{
				"reserve_now.reservation_id": 202,
				"reserve_now.status":          "Unavailable",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := handlers.ReserveNowResultHandler{}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.ReserveNowJson{
					ReservationId: tc.reservationId,
				}
				resp := &types.ReserveNowResponseJson{
					Status: tc.status,
				}

				err := handler.HandleCallResult(ctx, "cs001", req, resp, nil)
				require.NoError(t, err)
			}()

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", tc.expectedAttrs)
		})
	}
}

