// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/handlers"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ResetResultHandler struct {
	CallMaker handlers.CallMaker
}

func (h ResetResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.ResetJson)
	resp := response.(*types.ResetResponseJson)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("reset.type", string(req.Type)),
		attribute.String("reset.status", string(resp.Status)),
	)

	// If reset is accepted, the Charge Point will reboot and send BootNotification
	// We don't need to do anything special here - the BootNotification handler
	// will process the reconnection

	return nil
}

