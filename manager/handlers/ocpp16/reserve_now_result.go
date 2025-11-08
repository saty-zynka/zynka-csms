// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-csms/manager/ocpp"
	types "github.com/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ReserveNowResultHandler struct{}

func (h ReserveNowResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.ReserveNowJson)
	resp := response.(*types.ReserveNowResponseJson)

	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.Int("reserve_now.reservation_id", req.ReservationId),
		attribute.String("reserve_now.status", string(resp.Status)))

	return nil
}

