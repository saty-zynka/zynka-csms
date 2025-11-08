// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/store"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type GetConfigurationResultHandler struct {
	SettingsStore store.ChargeStationSettingsStore
}

func (h GetConfigurationResultHandler) HandleCallResult(ctx context.Context, chargeStationId string, request ocpp.Request, response ocpp.Response, state any) error {
	req := request.(*types.GetConfigurationJson)
	resp := response.(*types.GetConfigurationResponseJson)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("get_configuration.keys_requested", len(req.Key)),
		attribute.Int("get_configuration.keys_returned", len(resp.ConfigurationKey)),
		attribute.Int("get_configuration.unknown_keys", len(resp.UnknownKey)),
	)

	// Store the configuration keys returned by the Charge Point
	// This can be used to track Charge Point capabilities
	if len(resp.ConfigurationKey) > 0 {
		settings := &store.ChargeStationSettings{
			ChargeStationId: chargeStationId,
			Settings:        make(map[string]*store.ChargeStationSetting),
		}

		for _, key := range resp.ConfigurationKey {
			value := ""
			if key.Value != nil {
				value = *key.Value
			}
			settings.Settings[key.Key] = &store.ChargeStationSetting{
				Value:  value,
				Status: store.ChargeStationSettingStatusAccepted,
			}
		}

		// Update settings store with returned configuration
		// Note: This merges with existing settings
		_ = h.SettingsStore.UpdateChargeStationSettings(ctx, chargeStationId, settings)
	}

	return nil
}

