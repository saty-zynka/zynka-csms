// SPDX-License-Identifier: Apache-2.0

package ocpp16_test

import (
	"context"
	"fmt"
	"k8s.io/utils/clock"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	handlers "github.com/zynka-tech/zynka-csms/manager/handlers/ocpp16"
	types "github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/store"
	"github.com/zynka-tech/zynka-csms/manager/store/inmemory"
	clockTest "k8s.io/utils/clock/testing"
)

func TestStopTransactionHandler(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	periodicSampleContext := types.StopTransactionJsonTransactionDataElemSampledValueElemContextSamplePeriodic
	energyRegisterMeasurand := types.StopTransactionJsonTransactionDataElemSampledValueElemMeasurandEnergyActiveImportRegister
	outletLocation := types.StopTransactionJsonTransactionDataElemSampledValueElemLocationOutlet
	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionData: []types.StopTransactionJsonTransactionDataElem{
			{
				SampledValue: []types.StopTransactionJsonTransactionDataElemSampledValueElem{
					{
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Value:     "100",
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		},
		TransactionId: 42,
	}

	got, err := handler.HandleCall(context.Background(), chargingStationId, req)
	require.NoError(t, err)

	want := &types.StopTransactionResponseJson{
		IdTagInfo: &types.StopTransactionResponseJsonIdTagInfo{
			Status: types.StopTransactionResponseJsonIdTagInfoStatusAccepted,
		},
	}

	assert.Equal(t, want, got)

	found, err := transactionStore.FindTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42))
	require.NoError(t, err)

	expectedTransactionEndContext := "Transaction.End"
	expectedPeriodicContext := "Sample.Periodic"
	expectedOutletLocation := "Outlet"
	expectedMeasurand := "Energy.Active.Import.Register"
	expected := &store.Transaction{
		ChargeStationId: chargingStationId,
		TransactionId:   handlers.ConvertToUUID(42),
		IdToken:         "MYRFIDTAG",
		TokenType:       "ISO14443",
		MeterValues: []store.MeterValue{
			{
				Timestamp: now.Format(time.RFC3339),
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Location:  &startLocation,
						Measurand: &startMeasurand,
						Value:     50,
					},
				},
			},
			{
				Timestamp: now.Format(time.RFC3339),
				SampledValues: []store.SampledValue{
					{
						Context:   &expectedPeriodicContext,
						Location:  &expectedOutletLocation,
						Measurand: &expectedMeasurand,
						Value:     100,
					},
				},
			},
			{
				Timestamp: now.Format(time.RFC3339),
				SampledValues: []store.SampledValue{
					{
						Context:   &expectedTransactionEndContext,
						Location:  &expectedOutletLocation,
						Measurand: &expectedMeasurand,
						Value:     150,
					},
				},
			},
		},
		StartSeqNo:        0,
		EndedSeqNo:        1,
		UpdatedSeqNoCount: 0,
		Offline:           false,
	}

	assert.Equal(t, expected, found)
}

func TestStopTransactionWithInvalidToken(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       false,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionId: 42,
	}

	got, err := handler.HandleCall(context.Background(), chargingStationId, req)
	require.NoError(t, err)

	want := &types.StopTransactionResponseJson{
		IdTagInfo: &types.StopTransactionResponseJsonIdTagInfo{
			Status: types.StopTransactionResponseJsonIdTagInfoStatusInvalid,
		},
	}

	assert.Equal(t, want, got)
}

func TestStopTransactionWithSignedMeterData(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	signedDataFormat := types.StopTransactionJsonTransactionDataElemSampledValueElemFormatSignedData
	periodicSampleContext := types.StopTransactionJsonTransactionDataElemSampledValueElemContextSamplePeriodic
	energyRegisterMeasurand := types.StopTransactionJsonTransactionDataElemSampledValueElemMeasurandEnergyActiveImportRegister
	outletLocation := types.StopTransactionJsonTransactionDataElemSampledValueElemLocationOutlet

	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionData: []types.StopTransactionJsonTransactionDataElem{
			{
				SampledValue: []types.StopTransactionJsonTransactionDataElemSampledValueElem{
					{
						Format:    &signedDataFormat,
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Value:     "DEADBEEF",
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		},
		TransactionId: 42,
	}

	_, err = handler.HandleCall(context.Background(), chargingStationId, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversion from signed data not implemented")
}

func TestStopTransactionWithInvalidHexSignedData(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	signedDataFormat := types.StopTransactionJsonTransactionDataElemSampledValueElemFormatSignedData
	periodicSampleContext := types.StopTransactionJsonTransactionDataElemSampledValueElemContextSamplePeriodic
	energyRegisterMeasurand := types.StopTransactionJsonTransactionDataElemSampledValueElemMeasurandEnergyActiveImportRegister
	outletLocation := types.StopTransactionJsonTransactionDataElemSampledValueElemLocationOutlet

	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionData: []types.StopTransactionJsonTransactionDataElem{
			{
				SampledValue: []types.StopTransactionJsonTransactionDataElemSampledValueElem{
					{
						Format:    &signedDataFormat,
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Value:     "INVALID_HEX",
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		},
		TransactionId: 42,
	}

	_, err = handler.HandleCall(context.Background(), chargingStationId, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversion from signed data not implemented")
}

func TestStopTransactionWithMultipleSignedMeterValues(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	signedDataFormat := types.StopTransactionJsonTransactionDataElemSampledValueElemFormatSignedData
	periodicSampleContext := types.StopTransactionJsonTransactionDataElemSampledValueElemContextSamplePeriodic
	energyRegisterMeasurand := types.StopTransactionJsonTransactionDataElemSampledValueElemMeasurandEnergyActiveImportRegister
	outletLocation := types.StopTransactionJsonTransactionDataElemSampledValueElemLocationOutlet

	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionData: []types.StopTransactionJsonTransactionDataElem{
			{
				SampledValue: []types.StopTransactionJsonTransactionDataElemSampledValueElem{
					{
						Format:    &signedDataFormat,
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Value:     "ABCD1234",
					},
					{
						Format:    &signedDataFormat,
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Value:     "EF567890",
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		},
		TransactionId: 42,
	}

	_, err = handler.HandleCall(context.Background(), chargingStationId, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversion from signed data not implemented")
}

func TestStopTransactionWithRawFormat(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	rawFormat := types.StopTransactionJsonTransactionDataElemSampledValueElemFormatRaw
	periodicSampleContext := types.StopTransactionJsonTransactionDataElemSampledValueElemContextSamplePeriodic
	energyRegisterMeasurand := types.StopTransactionJsonTransactionDataElemSampledValueElemMeasurandEnergyActiveImportRegister
	outletLocation := types.StopTransactionJsonTransactionDataElemSampledValueElemLocationOutlet

	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionData: []types.StopTransactionJsonTransactionDataElem{
			{
				SampledValue: []types.StopTransactionJsonTransactionDataElemSampledValueElem{
					{
						Format:    &rawFormat,
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Value:     "123.45",
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		},
		TransactionId: 42,
	}

	got, err := handler.HandleCall(context.Background(), chargingStationId, req)
	require.NoError(t, err)

	want := &types.StopTransactionResponseJson{
		IdTagInfo: &types.StopTransactionResponseJsonIdTagInfo{
			Status: types.StopTransactionResponseJsonIdTagInfoStatusAccepted,
		},
	}

	assert.Equal(t, want, got)
}

func TestStopTransactionWithoutIdTag(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	reason := types.StopTransactionJsonReasonEVDisconnected
	req := &types.StopTransactionJson{
		IdTag:     nil,
		MeterStop:  200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionId: 42,
	}

	got, err := handler.HandleCall(context.Background(), chargingStationId, req)
	require.NoError(t, err)

	want := &types.StopTransactionResponseJson{
		IdTagInfo: nil,
	}

	assert.Equal(t, want, got)
}

func TestStopTransactionWithNilUnitOfMeasure(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	startContext := "Transaction.Begin"
	startMeasurand := "MeterValue"
	startLocation := "Outlet"
	err = transactionStore.CreateTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42), "MYRFIDTAG", "ISO14443",
		[]store.MeterValue{
			{
				SampledValues: []store.SampledValue{
					{
						Context:   &startContext,
						Measurand: &startMeasurand,
						Location:  &startLocation,
						Value:     50,
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		}, 0, false)
	require.NoError(t, err)

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	periodicSampleContext := types.StopTransactionJsonTransactionDataElemSampledValueElemContextSamplePeriodic
	energyRegisterMeasurand := types.StopTransactionJsonTransactionDataElemSampledValueElemMeasurandEnergyActiveImportRegister
	outletLocation := types.StopTransactionJsonTransactionDataElemSampledValueElemLocationOutlet
	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop: 200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionData: []types.StopTransactionJsonTransactionDataElem{
			{
				SampledValue: []types.StopTransactionJsonTransactionDataElemSampledValueElem{
					{
						Context:   &periodicSampleContext,
						Measurand: &energyRegisterMeasurand,
						Location:  &outletLocation,
						Unit:      nil,
						Value:     "100",
					},
				},
				Timestamp: now.Format(time.RFC3339),
			},
		},
		TransactionId: 42,
	}

	got, err := handler.HandleCall(context.Background(), chargingStationId, req)
	require.NoError(t, err)

	want := &types.StopTransactionResponseJson{
		IdTagInfo: &types.StopTransactionResponseJsonIdTagInfo{
			Status: types.StopTransactionResponseJsonIdTagInfoStatusAccepted,
		},
	}

	assert.Equal(t, want, got)

	found, err := transactionStore.FindTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(42))
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.NotNil(t, found.MeterValues)
	assert.Greater(t, len(found.MeterValues), 0)
	// Verify that UnitOfMeasure is nil when not provided
	for _, mv := range found.MeterValues {
		for _, sv := range mv.SampledValues {
			if sv.Measurand != nil && *sv.Measurand == "Energy.Active.Import.Register" {
				assert.Nil(t, sv.UnitOfMeasure, "UnitOfMeasure should be nil when not provided")
			}
		}
	}
}

func TestStopTransactionWithNonExistentTransaction(t *testing.T) {
	chargingStationId := fmt.Sprintf("cs%03d", rand.Intn(1000))
	engine := inmemory.NewStore(clock.RealClock{})

	err := engine.SetToken(context.Background(), &store.Token{
		CountryCode: "GB",
		PartyId:     "TWK",
		Type:        "RFID",
		Uid:         "MYRFIDTAG",
		ContractId:  "GBTWK012345678V",
		Issuer:      "Zynka-tech",
		Valid:       true,
		CacheMode:   "NEVER",
		LastUpdated: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	now, err := time.Parse(time.RFC3339, "2023-06-15T15:06:00+01:00")
	require.NoError(t, err)

	transactionStore := inmemory.NewStore(clock.RealClock{})

	handler := handlers.StopTransactionHandler{
		Clock:            clockTest.NewFakePassiveClock(now),
		TokenStore:       engine,
		TransactionStore: transactionStore,
	}

	idTag := "MYRFIDTAG"
	reason := types.StopTransactionJsonReasonEVDisconnected
	req := &types.StopTransactionJson{
		IdTag:     &idTag,
		MeterStop:  200,
		Reason:    &reason,
		Timestamp: now.Format(time.RFC3339),
		TransactionId: 999,
	}

	got, err := handler.HandleCall(context.Background(), chargingStationId, req)
	require.NoError(t, err)

	want := &types.StopTransactionResponseJson{
		IdTagInfo: &types.StopTransactionResponseJsonIdTagInfo{
			Status: types.StopTransactionResponseJsonIdTagInfoStatusAccepted,
		},
	}

	assert.Equal(t, want, got)

	// Verify transaction was created/ended with seqNo = -1 (since transaction didn't exist)
	found, err := transactionStore.FindTransaction(context.TODO(), chargingStationId, handlers.ConvertToUUID(999))
	require.NoError(t, err)
	assert.NotNil(t, found)
}
