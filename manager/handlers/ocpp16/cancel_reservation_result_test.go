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

func TestCancelReservationResultHandler(t *testing.T) {
	testCases := []struct {
		name          string
		reservationId int
		status        types.CancelReservationResponseJsonStatus
		expectedAttrs map[string]any
	}{
		{
			name:          "Accepted",
			reservationId: 123,
			status:        types.CancelReservationResponseJsonStatusAccepted,
			expectedAttrs: map[string]any{
				"cancel_reservation.reservation_id": 123,
				"cancel_reservation.status":         "Accepted",
			},
		},
		{
			name:          "Rejected",
			reservationId: 456,
			status:        types.CancelReservationResponseJsonStatusRejected,
			expectedAttrs: map[string]any{
				"cancel_reservation.reservation_id": 456,
				"cancel_reservation.status":         "Rejected",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := handlers.CancelReservationResultHandler{}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.CancelReservationJson{
					ReservationId: tc.reservationId,
				}
				resp := &types.CancelReservationResponseJson{
					Status: tc.status,
				}

				err := handler.HandleCallResult(ctx, "cs001", req, resp, nil)
				require.NoError(t, err)
			}()

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", tc.expectedAttrs)
		})
	}
}

