// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/zynka-csms/manager/ocpp"
	types "github.com/zynka-csms/manager/ocpp/ocpp16"
)

type FirmwareStatusNotificationHandler struct{}

func (h FirmwareStatusNotificationHandler) HandleCall(ctx context.Context, chargeStationId string, request ocpp.Request) (response ocpp.Response, err error) {
	req := request.(*types.FirmwareStatusNotificationJson)

	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("firmware_status.status", string(req.Status)))

	return &types.FirmwareStatusNotificationResponseJson{}, nil
}

