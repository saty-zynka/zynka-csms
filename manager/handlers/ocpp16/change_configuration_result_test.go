// SPDX-License-Identifier: Apache-2.0

package ocpp16_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	handlers16 "github.com/zynka-tech/zynka-csms/manager/handlers/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/store"
	"github.com/zynka-tech/zynka-csms/manager/store/inmemory"
	"github.com/zynka-tech/zynka-csms/manager/testutil"
	"k8s.io/utils/clock"
	"testing"
)

type mockCallMaker struct {
	calls []struct {
		chargeStationId string
		request         ocpp.Request
	}
}

func (m *mockCallMaker) Send(ctx context.Context, chargeStationId string, request ocpp.Request) error {
	m.calls = append(m.calls, struct {
		chargeStationId string
		request         ocpp.Request
	}{chargeStationId: chargeStationId, request: request})
	return nil
}

func TestChangeConfigurationResultHandler(t *testing.T) {
	testCases := []struct {
		name                string
		key                 string
		value               string
		status              types.ChangeConfigurationResponseJsonStatus
		existingSettings    map[string]*store.ChargeStationSetting
		expectedAttrs       map[string]any
		expectBootTrigger   bool
		expectError         bool
	}{
		{
			name:   "Accepted",
			key:     "HeartbeatInterval",
			value:   "300",
			status:  types.ChangeConfigurationResponseJsonStatusAccepted,
			existingSettings: nil,
			expectedAttrs: map[string]any{
				"setting.key":    "HeartbeatInterval",
				"setting.value":   "300",
				"setting.status":  "Accepted",
			},
			expectBootTrigger: false,
			expectError:      false,
		},
		{
			name:   "Rejected",
			key:     "InvalidKey",
			value:   "value",
			status:  types.ChangeConfigurationResponseJsonStatusRejected,
			existingSettings: nil,
			expectedAttrs: map[string]any{
				"setting.key":    "InvalidKey",
				"setting.value":   "value",
				"setting.status":  "Rejected",
			},
			expectBootTrigger: false,
			expectError:       false,
		},
		{
			name:   "NotSupported",
			key:     "UnsupportedKey",
			value:   "value",
			status:  types.ChangeConfigurationResponseJsonStatusNotSupported,
			existingSettings: nil,
			expectedAttrs: map[string]any{
				"setting.key":    "UnsupportedKey",
				"setting.value":   "value",
				"setting.status":  "NotSupported",
			},
			expectBootTrigger: false,
			expectError:       false,
		},
		{
			name:   "RebootRequired with all settings done",
			key:     "MeterValueSampleInterval",
			value:   "60",
			status:  types.ChangeConfigurationResponseJsonStatusRebootRequired,
			existingSettings: map[string]*store.ChargeStationSetting{
				"OtherSetting": {
					Value:  "value",
					Status: store.ChargeStationSettingStatusAccepted,
				},
			},
			expectedAttrs: map[string]any{
				"setting.key":    "MeterValueSampleInterval",
				"setting.value":   "60",
				"setting.status":  "RebootRequired",
			},
			expectBootTrigger: true,
			expectError:       false,
		},
		{
			name:   "RebootRequired with pending settings",
			key:     "MeterValueSampleInterval",
			value:   "60",
			status:  types.ChangeConfigurationResponseJsonStatusRebootRequired,
			existingSettings: map[string]*store.ChargeStationSetting{
				"OtherSetting": {
					Value:  "value",
					Status: store.ChargeStationSettingStatusPending,
				},
			},
			expectedAttrs: map[string]any{
				"setting.key":    "MeterValueSampleInterval",
				"setting.value":   "60",
				"setting.status":  "RebootRequired",
			},
			expectBootTrigger: false,
			expectError:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			settingsStore := inmemory.NewStore(clock.RealClock{})
			mockCallMaker := &mockCallMaker{}

			if tc.existingSettings != nil {
				err := settingsStore.UpdateChargeStationSettings(context.Background(), "cs001", &store.ChargeStationSettings{
					ChargeStationId: "cs001",
					Settings:        tc.existingSettings,
				})
				require.NoError(t, err)
			}

			handler := handlers16.ChangeConfigurationResultHandler{
				SettingsStore: settingsStore,
				CallMaker:     mockCallMaker,
			}

			tracer, exporter := testutil.GetTracer()

			ctx := context.Background()

			var handlerErr error
			func() {
				ctx, span := tracer.Start(ctx, "test")
				defer span.End()

				req := &types.ChangeConfigurationJson{
					Key:   tc.key,
					Value: tc.value,
				}
				resp := &types.ChangeConfigurationResponseJson{
					Status: tc.status,
				}

				handlerErr = handler.HandleCallResult(ctx, "cs001", req, resp, nil)
			}()

			if tc.expectError {
				assert.Error(t, handlerErr)
			} else {
				assert.NoError(t, handlerErr)
			}

			testutil.AssertSpan(t, &exporter.GetSpans()[0], "test", tc.expectedAttrs)

			// Verify settings were updated
			settings, err := settingsStore.LookupChargeStationSettings(context.Background(), "cs001")
			require.NoError(t, err)
			assert.NotNil(t, settings)
			assert.NotNil(t, settings.Settings[tc.key])
			assert.Equal(t, tc.value, settings.Settings[tc.key].Value)
			assert.Equal(t, store.ChargeStationSettingStatus(tc.status), settings.Settings[tc.key].Status)

			// Verify boot notification trigger
			if tc.expectBootTrigger {
				assert.Len(t, mockCallMaker.calls, 1)
				assert.Equal(t, "cs001", mockCallMaker.calls[0].chargeStationId)
				triggerMsg, ok := mockCallMaker.calls[0].request.(*types.TriggerMessageJson)
				require.True(t, ok)
				assert.Equal(t, types.TriggerMessageJsonRequestedMessageBootNotification, triggerMsg.RequestedMessage)
			} else {
				assert.Len(t, mockCallMaker.calls, 0)
			}
		})
	}
}

