// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ClearCacheResultHandler struct{}

func (h ClearCacheResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	resp := response.(*types.ClearCacheResponseJson)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("clear_cache.status", string(resp.Status)),
	)

	// The handler just logs the result. The actual cache clearing
	// is handled by the Charge Point.

	return nil
}

