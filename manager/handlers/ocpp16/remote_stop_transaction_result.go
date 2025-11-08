// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type RemoteStopTransactionResultHandler struct{}

func (h RemoteStopTransactionResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.RemoteStopTransactionJson)
	resp := response.(*types.RemoteStopTransactionResponseJson)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("remote_stop.transaction_id", req.TransactionId),
		attribute.String("remote_stop.status", string(resp.Status)),
	)

	// The handler just logs the result. The actual transaction stopping
	// is handled by the Charge Point, which will send a StopTransaction.req
	// when the transaction is stopped.

	return nil
}

