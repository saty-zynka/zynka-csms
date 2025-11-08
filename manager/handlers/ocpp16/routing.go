// SPDX-License-Identifier: Apache-2.0

package ocpp16

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/zynka-tech/zynka-csms/manager/handlers"
	handlersHasToBe "github.com/zynka-tech/zynka-csms/manager/handlers/has2be"
	handlers201 "github.com/zynka-tech/zynka-csms/manager/handlers/ocpp201"
	"github.com/zynka-tech/zynka-csms/manager/ocpp"
	"github.com/zynka-tech/zynka-csms/manager/ocpp/has2be"
	"github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp16"
	"github.com/zynka-tech/zynka-csms/manager/ocpp/ocpp201"
	"github.com/zynka-tech/zynka-csms/manager/services"
	"github.com/zynka-tech/zynka-csms/manager/store"
	"github.com/zynka-tech/zynka-csms/manager/transport"
	"io/fs"
	"k8s.io/utils/clock"
	"reflect"
	"time"
)

func NewRouter(emitter transport.Emitter,
	clk clock.PassiveClock,
	engine store.Engine,
	certValidationService services.CertificateValidationService,
	chargeStationCertProvider services.ChargeStationCertificateProvider,
	contractCertProvider services.ContractCertificateProvider,
	heartbeatInterval time.Duration,
	schemaFS fs.FS) transport.MessageHandler {

	standardCallMaker := NewCallMaker(emitter)

	return &handlers.Router{
		Emitter:     emitter,
		SchemaFS:    schemaFS,
		OcppVersion: transport.OcppVersion16,
		CallRoutes: map[string]handlers.CallRoute{
			"BootNotification": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.BootNotificationJson) },
				RequestSchema:  "ocpp16/BootNotification.json",
				ResponseSchema: "ocpp16/BootNotificationResponse.json",
				Handler: BootNotificationHandler{
					Clock:               clk,
					RuntimeDetailsStore: engine,
					SettingsStore:       engine,
					HeartbeatInterval:   int(heartbeatInterval.Seconds()),
				},
			},
			"Heartbeat": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.HeartbeatJson) },
				RequestSchema:  "ocpp16/Heartbeat.json",
				ResponseSchema: "ocpp16/HeartbeatResponse.json",
				Handler: HeartbeatHandler{
					Clock: clk,
				},
			},
			"StatusNotification": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.StatusNotificationJson) },
				RequestSchema:  "ocpp16/StatusNotification.json",
				ResponseSchema: "ocpp16/StatusNotificationResponse.json",
				Handler:        handlers.CallHandlerFunc(StatusNotificationHandler),
			},
			"Authorize": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.AuthorizeJson) },
				RequestSchema:  "ocpp16/Authorize.json",
				ResponseSchema: "ocpp16/AuthorizeResponse.json",
				Handler: AuthorizeHandler{
					TokenStore: engine,
				},
			},
			"StartTransaction": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.StartTransactionJson) },
				RequestSchema:  "ocpp16/StartTransaction.json",
				ResponseSchema: "ocpp16/StartTransactionResponse.json",
				Handler: StartTransactionHandler{
					Clock:            clk,
					TokenStore:       engine,
					TransactionStore: engine,
				},
			},
			"StopTransaction": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.StopTransactionJson) },
				RequestSchema:  "ocpp16/StopTransaction.json",
				ResponseSchema: "ocpp16/StopTransactionResponse.json",
				Handler: StopTransactionHandler{
					Clock:            clk,
					TokenStore:       engine,
					TransactionStore: engine,
				},
			},
			"MeterValues": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.MeterValuesJson) },
				RequestSchema:  "ocpp16/MeterValues.json",
				ResponseSchema: "ocpp16/MeterValuesResponse.json",
				Handler: MeterValuesHandler{
					TransactionStore: engine,
				},
			},
			"SecurityEventNotification": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.SecurityEventNotificationJson) },
				RequestSchema:  "ocpp16/SecurityEventNotification.json",
				ResponseSchema: "ocpp16/SecurityEventNotificationResponse.json",
				Handler:        SecurityEventNotificationHandler{},
			},
			"FirmwareStatusNotification": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.FirmwareStatusNotificationJson) },
				RequestSchema:  "ocpp16/FirmwareStatusNotification.json",
				ResponseSchema: "ocpp16/FirmwareStatusNotificationResponse.json",
				Handler:        FirmwareStatusNotificationHandler{},
			},
			"DiagnosticsStatusNotification": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.DiagnosticsStatusNotificationJson) },
				RequestSchema:  "ocpp16/DiagnosticsStatusNotification.json",
				ResponseSchema: "ocpp16/DiagnosticsStatusNotificationResponse.json",
				Handler:        DiagnosticsStatusNotificationHandler{},
			},
			"DataTransfer": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.DataTransferJson) },
				RequestSchema:  "ocpp16/DataTransfer.json",
				ResponseSchema: "ocpp16/DataTransferResponse.json",
				Handler: DataTransferHandler{
					SchemaFS: schemaFS,
					CallRoutes: map[string]map[string]handlers.CallRoute{
						"org.openchargealliance.iso15118pnc": {
							"Authorize": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.AuthorizeRequestJson) },
								RequestSchema:  "ocpp201/AuthorizeRequest.json",
								ResponseSchema: "ocpp201/AuthorizeResponse.json",
								Handler: handlers201.AuthorizeHandler{
									TokenAuthService: &services.OcppTokenAuthService{
										Clock:      clk,
										TokenStore: engine,
									},
									CertificateValidationService: certValidationService,
								},
							},
							"GetCertificateStatus": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.GetCertificateStatusRequestJson) },
								RequestSchema:  "ocpp201/GetCertificateStatusRequest.json",
								ResponseSchema: "ocpp201/GetCertificateStatusResponse.json",
								Handler: handlers201.GetCertificateStatusHandler{
									CertificateValidationService: certValidationService,
								},
							},
							"SignCertificate": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.SignCertificateRequestJson) },
								RequestSchema:  "ocpp201/SignCertificateRequest.json",
								ResponseSchema: "ocpp201/SignCertificateResponse.json",
								Handler: handlers201.SignCertificateHandler{
									ChargeStationCertificateProvider: chargeStationCertProvider,
									Store:                            engine,
								},
							},
							"Get15118EVCertificate": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.Get15118EVCertificateRequestJson) },
								RequestSchema:  "ocpp201/Get15118EVCertificateRequest.json",
								ResponseSchema: "ocpp201/Get15118EVCertificateResponse.json",
								Handler: handlers201.Get15118EvCertificateHandler{
									ContractCertificateProvider: contractCertProvider,
								},
							},
						},
						"iso15118": { // has2be extensions
							"Authorize": {
								NewRequest:     func() ocpp.Request { return new(has2be.AuthorizeRequestJson) },
								RequestSchema:  "has2be/AuthorizeRequest.json",
								ResponseSchema: "has2be/AuthorizeResponse.json",
								Handler: handlersHasToBe.AuthorizeHandler{
									Handler201: handlers201.AuthorizeHandler{
										TokenAuthService: &services.OcppTokenAuthService{
											Clock:      clk,
											TokenStore: engine,
										},
										CertificateValidationService: certValidationService,
									},
								},
							},
							"GetCertificateStatus": {
								NewRequest:     func() ocpp.Request { return new(has2be.GetCertificateStatusRequestJson) },
								RequestSchema:  "has2be/GetCertificateStatusRequest.json",
								ResponseSchema: "has2be/GetCertificateStatusResponse.json",
								Handler: handlersHasToBe.GetCertificateStatusHandler{
									Handler201: handlers201.GetCertificateStatusHandler{
										CertificateValidationService: certValidationService,
									},
								},
							},
							"Get15118EVCertificate": {
								NewRequest:     func() ocpp.Request { return new(has2be.Get15118EVCertificateRequestJson) },
								RequestSchema:  "has2be/Get15118EVCertificateRequest.json",
								ResponseSchema: "has2be/Get15118EVCertificateResponse.json",
								Handler: handlersHasToBe.Get15118EvCertificateHandler{
									Handler201: handlers201.Get15118EvCertificateHandler{
										ContractCertificateProvider: contractCertProvider,
									},
								},
							},
							"SignCertificate": {
								NewRequest:     func() ocpp.Request { return new(has2be.SignCertificateRequestJson) },
								RequestSchema:  "has2be/SignCertificateRequestJson.json",
								ResponseSchema: "has2be/SignCertificateRequestJson.json",
								Handler: handlersHasToBe.SignCertificateHandler{
									Handler201: handlers201.SignCertificateHandler{
										ChargeStationCertificateProvider: chargeStationCertProvider,
										Store:                            engine,
									},
								},
							},
						},
					},
				},
			},
		},
		CallResultRoutes: map[string]handlers.CallResultRoute{
			"DataTransfer": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.DataTransferJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.DataTransferResponseJson) },
				RequestSchema:  "ocpp16/DataTransfer.json",
				ResponseSchema: "ocpp16/DataTransferResponse.json",
				Handler: DataTransferResultHandler{
					SchemaFS: schemaFS,
					CallResultRoutes: map[string]map[string]handlers.CallResultRoute{
						"org.openchargealliance.iso15118pnc": {
							"CertificateSigned": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.CertificateSignedRequestJson) },
								NewResponse:    func() ocpp.Response { return new(ocpp201.CertificateSignedResponseJson) },
								RequestSchema:  "ocpp201/CertificateSignedRequest.json",
								ResponseSchema: "ocpp201/CertificateSignedResponse.json",
								Handler: handlers201.CertificateSignedResultHandler{
									Store: engine,
								},
							},
							"InstallCertificate": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.InstallCertificateRequestJson) },
								NewResponse:    func() ocpp.Response { return new(ocpp201.InstallCertificateResponseJson) },
								RequestSchema:  "ocpp201/InstallCertificateRequest.json",
								ResponseSchema: "ocpp201/InstallCertificateResponse.json",
								Handler: handlers201.InstallCertificateResultHandler{
									Store: engine,
								},
							},
							"TriggerMessage": {
								NewRequest:     func() ocpp.Request { return new(ocpp201.TriggerMessageRequestJson) },
								NewResponse:    func() ocpp.Response { return new(ocpp201.TriggerMessageResponseJson) },
								RequestSchema:  "ocpp201/TriggerMessageRequest.json",
								ResponseSchema: "ocpp201/TriggerMessageResponse.json",
								Handler: handlers201.TriggerMessageResultHandler{
									Store: engine,
								},
							},
						},
						"iso15118": { // has2be extensions
							"CertificateSigned": {
								NewRequest:     func() ocpp.Request { return new(has2be.CertificateSignedRequestJson) },
								NewResponse:    func() ocpp.Response { return new(has2be.CertificateSignedResponseJson) },
								RequestSchema:  "has2be/CertificateSignedRequest.json",
								ResponseSchema: "has2be/CertificateSignedResponse.json",
								Handler:        handlersHasToBe.CertificateSignedResultHandler{},
							},
						},
					},
				},
			},
			"ChangeConfiguration": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.ChangeConfigurationJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.ChangeConfigurationResponseJson) },
				RequestSchema:  "ocpp16/ChangeConfiguration.json",
				ResponseSchema: "ocpp16/ChangeConfigurationResponse.json",
				Handler: ChangeConfigurationResultHandler{
					SettingsStore: engine,
					CallMaker:     standardCallMaker,
				},
			},
			"TriggerMessage": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.TriggerMessageJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.TriggerMessageResponseJson) },
				RequestSchema:  "ocpp16/TriggerMessage.json",
				ResponseSchema: "ocpp16/TriggerMessageResponse.json",
				Handler:        TriggerMessageResultHandler{},
			},
			"RemoteStartTransaction": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.RemoteStartTransactionJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.RemoteStartTransactionResponseJson) },
				RequestSchema:  "ocpp16/RemoteStartTransaction.json",
				ResponseSchema: "ocpp16/RemoteStartTransactionResponse.json",
				Handler:        RemoteStartTransactionResultHandler{},
			},
			"ReserveNow": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.ReserveNowJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.ReserveNowResponseJson) },
				RequestSchema:  "ocpp16/ReserveNow.json",
				ResponseSchema: "ocpp16/ReserveNowResponse.json",
				Handler:        ReserveNowResultHandler{},
			},
			"CancelReservation": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.CancelReservationJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.CancelReservationResponseJson) },
				RequestSchema:  "ocpp16/CancelReservation.json",
				ResponseSchema: "ocpp16/CancelReservationResponse.json",
				Handler:        CancelReservationResultHandler{},
			},
			"ChangeAvailability": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.ChangeAvailabilityJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.ChangeAvailabilityResponseJson) },
				RequestSchema:  "ocpp16/ChangeAvailability.json",
				ResponseSchema: "ocpp16/ChangeAvailabilityResponse.json",
				Handler:        ChangeAvailabilityResultHandler{},
			},
			"ClearCache": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.ClearCacheJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.ClearCacheResponseJson) },
				RequestSchema:  "ocpp16/ClearCache.json",
				ResponseSchema: "ocpp16/ClearCacheResponse.json",
				Handler:        ClearCacheResultHandler{},
			},
			"GetConfiguration": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.GetConfigurationJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.GetConfigurationResponseJson) },
				RequestSchema:  "ocpp16/GetConfiguration.json",
				ResponseSchema: "ocpp16/GetConfigurationResponse.json",
				Handler: GetConfigurationResultHandler{
					SettingsStore: engine,
				},
			},
			"RemoteStopTransaction": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.RemoteStopTransactionJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.RemoteStopTransactionResponseJson) },
				RequestSchema:  "ocpp16/RemoteStopTransaction.json",
				ResponseSchema: "ocpp16/RemoteStopTransactionResponse.json",
				Handler:        RemoteStopTransactionResultHandler{},
			},
			"Reset": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.ResetJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.ResetResponseJson) },
				RequestSchema:  "ocpp16/Reset.json",
				ResponseSchema: "ocpp16/ResetResponse.json",
				Handler: ResetResultHandler{
					CallMaker: standardCallMaker,
				},
			},
			"UnlockConnector": {
				NewRequest:     func() ocpp.Request { return new(ocpp16.UnlockConnectorJson) },
				NewResponse:    func() ocpp.Response { return new(ocpp16.UnlockConnectorResponseJson) },
				RequestSchema:  "ocpp16/UnlockConnector.json",
				ResponseSchema: "ocpp16/UnlockConnectorResponse.json",
				Handler:        UnlockConnectorResultHandler{},
			},
		},
	}
}

func NewCallMaker(e transport.Emitter) *handlers.OcppCallMaker {
	return &handlers.OcppCallMaker{
		Emitter:     e,
		OcppVersion: transport.OcppVersion16,
		Actions: map[reflect.Type]string{
			reflect.TypeOf(&ocpp16.ChangeConfigurationJson{}):    "ChangeConfiguration",
			reflect.TypeOf(&ocpp16.TriggerMessageJson{}):         "TriggerMessage",
			reflect.TypeOf(&ocpp16.RemoteStartTransactionJson{}): "RemoteStartTransaction",
			reflect.TypeOf(&ocpp16.ReserveNowJson{}):             "ReserveNow",
			reflect.TypeOf(&ocpp16.CancelReservationJson{}):      "CancelReservation",
			reflect.TypeOf(&ocpp16.ChangeAvailabilityJson{}):     "ChangeAvailability",
			reflect.TypeOf(&ocpp16.ClearCacheJson{}):              "ClearCache",
			reflect.TypeOf(&ocpp16.GetConfigurationJson{}):       "GetConfiguration",
			reflect.TypeOf(&ocpp16.RemoteStopTransactionJson{}):  "RemoteStopTransaction",
			reflect.TypeOf(&ocpp16.ResetJson{}):                  "Reset",
			reflect.TypeOf(&ocpp16.UnlockConnectorJson{}):        "UnlockConnector",
		},
	}
}

type DataTransferAction struct {
	VendorId  string
	MessageId string
}

type DataTransferCallMaker struct {
	e       transport.Emitter
	actions map[reflect.Type]DataTransferAction
}

func NewDataTransferCallMaker(e transport.Emitter) *DataTransferCallMaker {
	return &DataTransferCallMaker{
		e: e,
		actions: map[reflect.Type]DataTransferAction{
			reflect.TypeOf(&ocpp201.CertificateSignedRequestJson{}): {
				VendorId:  "org.openchargealliance.iso15118pnc",
				MessageId: "CertificateSigned",
			},
			reflect.TypeOf(&ocpp201.InstallCertificateRequestJson{}): {
				VendorId:  "org.openchargealliance.iso15118pnc",
				MessageId: "InstallCertificate",
			},
			reflect.TypeOf(&ocpp201.TriggerMessageRequestJson{}): {
				VendorId:  "org.openchargealliance.iso15118pnc",
				MessageId: "TriggerMessage",
			},
		},
	}
}

func (d DataTransferCallMaker) Send(ctx context.Context, chargeStationId string, request ocpp.Request) error {
	dta, ok := d.actions[reflect.TypeOf(request)]
	if !ok {
		return fmt.Errorf("unknown request type: %T", request)
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}
	requestBytesStr := string(requestBytes)

	dataTransferRequest := ocpp16.DataTransferJson{
		VendorId:  dta.VendorId,
		MessageId: &dta.MessageId,
		Data:      &requestBytesStr,
	}

	dataTransferBytes, err := json.Marshal(dataTransferRequest)
	if err != nil {
		return fmt.Errorf("marshaling data transfer request: %w", err)
	}

	msg := &transport.Message{
		MessageType:    transport.MessageTypeCall,
		MessageId:      uuid.New().String(),
		Action:         "DataTransfer",
		RequestPayload: dataTransferBytes,
	}

	return d.e.Emit(ctx, transport.OcppVersion16, chargeStationId, msg)
}
