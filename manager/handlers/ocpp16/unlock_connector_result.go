// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type UnlockConnectorResultHandler struct{}

func (h UnlockConnectorResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.UnlockConnectorJson)
	resp := response.(*types.UnlockConnectorResponseJson)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("unlock_connector.connector_id", req.ConnectorId),
		attribute.String("unlock_connector.status", string(resp.Status)),
	)

	// The handler just logs the result. The actual connector unlocking
	// is handled by the Charge Point. If there was an active transaction,
	// the Charge Point will stop it before unlocking.

	return nil
}

