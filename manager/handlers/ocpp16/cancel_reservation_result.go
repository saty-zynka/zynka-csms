// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-csms/manager/ocpp"
	types "github.com/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CancelReservationResultHandler struct{}

func (h CancelReservationResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.CancelReservationJson)
	resp := response.(*types.CancelReservationResponseJson)

	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.Int("cancel_reservation.reservation_id", req.ReservationId),
		attribute.String("cancel_reservation.status", string(resp.Status)))

	return nil
}

