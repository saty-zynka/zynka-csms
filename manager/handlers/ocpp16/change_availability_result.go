// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ChangeAvailabilityResultHandler struct{}

func (h ChangeAvailabilityResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.ChangeAvailabilityJson)
	resp := response.(*types.ChangeAvailabilityResponseJson)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("change_availability.connector_id", req.ConnectorId),
		attribute.String("change_availability.type", string(req.Type)),
		attribute.String("change_availability.status", string(resp.Status)),
	)

	// The handler just logs the result. The actual availability change
	// is handled by the Charge Point, which will send a StatusNotification
	// when the change takes effect.

	return nil
}

