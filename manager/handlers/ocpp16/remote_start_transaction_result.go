// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type RemoteStartTransactionResultHandler struct{}

func (h RemoteStartTransactionResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.RemoteStartTransactionJson)
	resp := response.(*types.RemoteStartTransactionResponseJson)

	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("remote_start.id_tag", req.IdTag),
		attribute.String("remote_start.status", string(resp.Status)))

	if req.ConnectorId != nil {
		span.SetAttributes(attribute.Int("remote_start.connector_id", *req.ConnectorId))
	}

	return nil
}

