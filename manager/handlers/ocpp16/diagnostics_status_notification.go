// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
)

type DiagnosticsStatusNotificationHandler struct{}

func (h DiagnosticsStatusNotificationHandler) HandleCall(ctx context.Context, chargeStationId string, request ocpp.Request) (response ocpp.Response, err error) {
	req := request.(*types.DiagnosticsStatusNotificationJson)

	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("diagnostics_status.status", string(req.Status)))

	return &types.DiagnosticsStatusNotificationResponseJson{}, nil
}

