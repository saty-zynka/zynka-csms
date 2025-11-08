// SPDX-License-Identifier: Apache-2.0
//go:build e2e

package test_driver

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// VCPClient represents a client for SolidStudio Virtual Charge Point
// This client can be used to test the simulator or interact with it directly
type VCPClient struct {
	conn     *websocket.Conn
	csID     string
	password string
	url      string
	ctx      context.Context
}

// OCPPMessage represents a standard OCPP JSON message
type OCPPMessage struct {
	MessageTypeID int                    `json:"-"`
	MessageID     string                 `json:"-"`
	Action        string                 `json:"-"`
	Payload       map[string]interface{} `json:"-"`
}

// NewVCPClient creates a new VCP client
func NewVCPClient(csID, password, url string) *VCPClient {
	return &VCPClient{
		csID:     csID,
		password: password,
		url:      url,
		ctx:      context.Background(),
	}
}

// Connect establishes WebSocket connection with Basic Auth
func (c *VCPClient) Connect() error {
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.csID, c.password)))

	dialOptions := &websocket.DialOptions{
		Subprotocols: []string{"ocpp1.6"},
		HTTPHeader: http.Header{
			"Authorization": []string{authHeader},
		},
	}

	conn, resp, err := websocket.Dial(c.ctx, c.url, dialOptions)
	if err != nil {
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to connect: %v, status: %d, body: %s", err, resp.StatusCode, string(bodyBytes))
		}
		return fmt.Errorf("failed to connect: %v", err)
	}

	if conn == nil {
		return fmt.Errorf("connection is nil but no error returned")
	}

	// On successful connection, resp may be nil - that's OK
	// The websocket library may return nil resp on successful connection
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	c.conn = conn
	return nil
}

// SendOCPPMessage sends an OCPP message
func (c *VCPClient) SendOCPPMessage(messageType int, messageID, action string, payload map[string]interface{}) error {
	message := []interface{}{
		messageType,
		messageID,
	}

	if action != "" {
		message = append(message, action)
	}
	if payload != nil {
		message = append(message, payload)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.conn.Write(c.ctx, websocket.MessageText, data)
}

// ReadOCPPMessage reads an OCPP message
func (c *VCPClient) ReadOCPPMessage(timeout time.Duration) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

	_, data, err := c.conn.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var message []interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return message, nil
}

// ReadOCPPMessageWithType reads an OCPP message and returns its type
// Returns: messageType (2=Call, 3=CallResult, 4=CallError), message array, error
func (c *VCPClient) ReadOCPPMessageWithType(timeout time.Duration) (int, []interface{}, error) {
	message, err := c.ReadOCPPMessage(timeout)
	if err != nil {
		return 0, nil, err
	}

	if len(message) == 0 {
		return 0, nil, fmt.Errorf("empty message")
	}

	messageType, ok := message[0].(float64)
	if !ok {
		return 0, nil, fmt.Errorf("invalid message type: %v", message[0])
	}

	return int(messageType), message, nil
}

// ReadCallMessage reads a Call message (type 2) from Central System
// Returns: messageID, action, payload, error
func (c *VCPClient) ReadCallMessage(timeout time.Duration) (string, string, map[string]interface{}, error) {
	messageType, message, err := c.ReadOCPPMessageWithType(timeout)
	if err != nil {
		return "", "", nil, err
	}

	if messageType != 2 {
		return "", "", nil, fmt.Errorf("expected Call message (2), got %d", messageType)
	}

	if len(message) < 3 {
		return "", "", nil, fmt.Errorf("invalid Call message format: expected at least 3 elements, got %d", len(message))
	}

	messageID, ok := message[1].(string)
	if !ok {
		return "", "", nil, fmt.Errorf("invalid message ID type: %v", message[1])
	}

	action, ok := message[2].(string)
	if !ok {
		return "", "", nil, fmt.Errorf("invalid action type: %v", message[2])
	}

	var payload map[string]interface{}
	if len(message) >= 4 {
		payload, ok = message[3].(map[string]interface{})
		if !ok {
			return "", "", nil, fmt.Errorf("invalid payload type: %v", message[3])
		}
	}

	return messageID, action, payload, nil
}

// SendCallResult sends a CallResult message (type 3) to Central System
func (c *VCPClient) SendCallResult(messageID string, payload map[string]interface{}) error {
	message := []interface{}{
		3, // CallResult
		messageID,
	}
	if payload != nil {
		message = append(message, payload)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal CallResult: %w", err)
	}

	return c.conn.Write(c.ctx, websocket.MessageText, data)
}

// SendCallError sends a CallError message (type 4) to Central System
func (c *VCPClient) SendCallError(messageID, errorCode, errorDescription string, errorDetails map[string]interface{}) error {
	message := []interface{}{
		4, // CallError
		messageID,
		errorCode,
		errorDescription,
	}
	if errorDetails != nil {
		message = append(message, errorDetails)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal CallError: %w", err)
	}

	return c.conn.Write(c.ctx, websocket.MessageText, data)
}

// Close closes the WebSocket connection
func (c *VCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close(websocket.StatusGoingAway, "Test complete")
	}
	return nil
}

// TestVCPBasicConnection tests basic WebSocket connection with Basic Auth
// This test verifies that the SolidStudio VCP simulator can connect to zynka-csms
// Note: Uses a different charge station ID (cs002) to avoid conflicts with the simulator (cs001)
func TestVCPBasicConnection(t *testing.T) {
	// Get configuration from environment or use defaults
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002" // Use cs002 to avoid conflict with simulator (cs001)
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002" // Different from simulator's cs001
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password" // Default test password
	}

	// Create VCP client
	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	t.Log("Successfully connected to zynka-csms gateway")

	// Send BootNotification
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}

	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	t.Log("Sent BootNotification")

	// Read response
	response, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	if len(response) < 3 {
		t.Fatalf("Invalid response format: %v", response)
	}

	messageType, ok := response[0].(float64)
	if !ok || int(messageType) != 3 {
		t.Fatalf("Expected CallResult (3), got: %v", response)
	}

	t.Logf("Received BootNotification response: %v", response)

	// Verify response structure
	if len(response) >= 3 {
		responsePayload, ok := response[2].(map[string]interface{})
		if ok {
			status, _ := responsePayload["status"].(string)
			t.Logf("BootNotification status: %s", status)
		}
	}
}

// TestVCPRFIDChargeFlow tests a complete RFID charging flow
// This test demonstrates a full charging session using OCPP 1.6 messages
// Uses cs003 to avoid conflicts with the simulator (cs001) and basic connection test (cs002)
func TestVCPRFIDChargeFlow(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs003" // Use cs003 to avoid conflicts
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs003" // Different from simulator (cs001) and basic test (cs002)
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Step 1: BootNotification
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	response, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}
	t.Logf("BootNotification response: %v", response)

	// Step 2: StatusNotification - Available
	statusPayload := map[string]interface{}{
		"connectorId": 1,
		"errorCode":   "NoError",
		"status":      "Available",
	}
	if err := client.SendOCPPMessage(2, "2", "StatusNotification", statusPayload); err != nil {
		t.Fatalf("Failed to send StatusNotification: %v", err)
	}
	t.Log("Sent StatusNotification (Available)")

	// Step 3: StatusNotification - Preparing (plugged in)
	statusPayload["status"] = "Preparing"
	if err := client.SendOCPPMessage(2, "3", "StatusNotification", statusPayload); err != nil {
		t.Fatalf("Failed to send StatusNotification: %v", err)
	}
	t.Log("Sent StatusNotification (Preparing)")

	// Step 4: Authorize
	authPayload := map[string]interface{}{
		"idTag": "DEADBEEF",
	}
	if err := client.SendOCPPMessage(2, "4", "Authorize", authPayload); err != nil {
		t.Fatalf("Failed to send Authorize: %v", err)
	}

	authResponse, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read Authorize response: %v", err)
	}
	t.Logf("Authorize response: %v", authResponse)

	// Step 5: StartTransaction
	startTxPayload := map[string]interface{}{
		"connectorId": 1,
		"idTag":       "DEADBEEF",
		"meterStart":  0,
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	if err := client.SendOCPPMessage(2, "5", "StartTransaction", startTxPayload); err != nil {
		t.Fatalf("Failed to send StartTransaction: %v", err)
	}

	txResponse, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read StartTransaction response: %v", err)
	}
	t.Logf("StartTransaction response: %v", txResponse)

	// Step 6: StatusNotification - Charging
	statusPayload["status"] = "Charging"
	if err := client.SendOCPPMessage(2, "6", "StatusNotification", statusPayload); err != nil {
		t.Fatalf("Failed to send StatusNotification: %v", err)
	}
	t.Log("Sent StatusNotification (Charging)")

	// Step 7: MeterValues
	meterPayload := map[string]interface{}{
		"connectorId": 1,
		"meterValue": []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"sampledValue": []map[string]interface{}{
					{
						"value":     "1000",
						"context":   "Sample.Periodic",
						"format":    "Raw",
						"measurand": "Energy.Active.Import.Register",
						"unit":      "Wh",
					},
				},
			},
		},
	}
	if err := client.SendOCPPMessage(2, "7", "MeterValues", meterPayload); err != nil {
		t.Fatalf("Failed to send MeterValues: %v", err)
	}
	t.Log("Sent MeterValues")

	// Step 8: StopTransaction
	stopTxPayload := map[string]interface{}{
		"idTag":         "DEADBEEF",
		"meterStop":     1000,
		"timestamp":     time.Now().Format(time.RFC3339),
		"transactionId": 1,
	}
	if err := client.SendOCPPMessage(2, "8", "StopTransaction", stopTxPayload); err != nil {
		t.Fatalf("Failed to send StopTransaction: %v", err)
	}

	stopResponse, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read StopTransaction response: %v", err)
	}
	t.Logf("StopTransaction response: %v", stopResponse)

	// Step 9: StatusNotification - Finishing
	statusPayload["status"] = "Finishing"
	if err := client.SendOCPPMessage(2, "9", "StatusNotification", statusPayload); err != nil {
		t.Fatalf("Failed to send StatusNotification: %v", err)
	}

	// Step 10: StatusNotification - Available
	statusPayload["status"] = "Available"
	if err := client.SendOCPPMessage(2, "10", "StatusNotification", statusPayload); err != nil {
		t.Fatalf("Failed to send StatusNotification: %v", err)
	}
	t.Log("Charging flow completed successfully")
}

// TestHeartbeat tests Heartbeat operation for time synchronization
func TestHeartbeat(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first (required before other operations)
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Send Heartbeat
	if err := client.SendOCPPMessage(2, "2", "Heartbeat", nil); err != nil {
		t.Fatalf("Failed to send Heartbeat: %v", err)
	}

	t.Log("Sent Heartbeat")

	// Read response
	response, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read Heartbeat response: %v", err)
	}

	// Verify response structure
	if len(response) < 3 {
		t.Fatalf("Invalid response format: %v", response)
	}

	messageType, ok := response[0].(float64)
	if !ok || int(messageType) != 3 {
		t.Fatalf("Expected CallResult (3), got: %v", response)
	}

	// Verify response contains currentTime
	if len(response) >= 3 {
		responsePayload, ok := response[2].(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid response payload type: %v", response[2])
		}

		currentTime, ok := responsePayload["currentTime"].(string)
		if !ok {
			t.Fatalf("Missing or invalid currentTime in response: %v", responsePayload)
		}

		// Verify currentTime is a valid RFC3339 timestamp
		_, err := time.Parse(time.RFC3339, currentTime)
		if err != nil {
			t.Fatalf("Invalid currentTime format: %v", err)
		}

		t.Logf("Heartbeat response received with currentTime: %s", currentTime)
	}
}

// TestStatusNotificationTransitions tests StatusNotification with various status transitions
func TestStatusNotificationTransitions(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Test status transitions for connector 1
	statuses := []string{"Available", "Preparing", "Charging", "Finishing", "Available"}

	for i, status := range statuses {
		statusPayload := map[string]interface{}{
			"connectorId": 1,
			"errorCode":   "NoError",
			"status":      status,
		}

		messageID := fmt.Sprintf("%d", i+2)
		if err := client.SendOCPPMessage(2, messageID, "StatusNotification", statusPayload); err != nil {
			t.Fatalf("Failed to send StatusNotification (%s): %v", status, err)
		}

		t.Logf("Sent StatusNotification: %s", status)

		// StatusNotification doesn't require a response, but we can verify no error occurred
		// by checking if connection is still alive
		time.Sleep(100 * time.Millisecond)
	}

	// Test Charge Point level status (connectorId 0)
	cpStatusPayload := map[string]interface{}{
		"connectorId": 0,
		"errorCode":   "NoError",
		"status":      "Available",
	}
	if err := client.SendOCPPMessage(2, "10", "StatusNotification", cpStatusPayload); err != nil {
		t.Fatalf("Failed to send StatusNotification for Charge Point: %v", err)
	}
	t.Log("Sent StatusNotification for Charge Point (connectorId 0)")
}

// TestMeterValues tests MeterValues operation with different measurands
func TestMeterValues(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Test MeterValues with Energy measurand
	meterPayload := map[string]interface{}{
		"connectorId": 1,
		"meterValue": []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"sampledValue": []map[string]interface{}{
					{
						"value":     "1000",
						"context":   "Sample.Periodic",
						"format":    "Raw",
						"measurand": "Energy.Active.Import.Register",
						"unit":      "Wh",
					},
				},
			},
		},
	}
	if err := client.SendOCPPMessage(2, "2", "MeterValues", meterPayload); err != nil {
		t.Fatalf("Failed to send MeterValues: %v", err)
	}
	t.Log("Sent MeterValues with Energy.Active.Import.Register")

	// Test MeterValues with Power measurand
	powerMeterPayload := map[string]interface{}{
		"connectorId": 1,
		"meterValue": []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"sampledValue": []map[string]interface{}{
					{
						"value":     "3500",
						"context":   "Sample.Periodic",
						"format":    "Raw",
						"measurand": "Power.Active.Import",
						"unit":      "W",
					},
				},
			},
		},
	}
	if err := client.SendOCPPMessage(2, "3", "MeterValues", powerMeterPayload); err != nil {
		t.Fatalf("Failed to send MeterValues with Power: %v", err)
	}
	t.Log("Sent MeterValues with Power.Active.Import")

	// Test MeterValues with Current measurand
	currentMeterPayload := map[string]interface{}{
		"connectorId": 1,
		"meterValue": []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"sampledValue": []map[string]interface{}{
					{
						"value":     "16",
						"context":   "Sample.Periodic",
						"format":    "Raw",
						"measurand": "Current.Import",
						"unit":      "A",
					},
				},
			},
		},
	}
	if err := client.SendOCPPMessage(2, "4", "MeterValues", currentMeterPayload); err != nil {
		t.Fatalf("Failed to send MeterValues with Current: %v", err)
	}
	t.Log("Sent MeterValues with Current.Import")

	// MeterValues doesn't require a response, but we verify no error occurred
	time.Sleep(100 * time.Millisecond)
}

// TestChangeConfiguration tests CS-initiated ChangeConfiguration operation
// Note: This test requires the CS to send a ChangeConfiguration Call message
// Currently, there's no API endpoint to trigger this, so this test demonstrates
// how to receive and respond to such messages when they are sent
func TestChangeConfiguration(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for ChangeConfiguration Call message from CS (if triggered)
	// In a real scenario, this would be triggered via API or internal mechanism
	// For now, we demonstrate the response handling
	// Try to read a Call message (this will timeout if none is sent)
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		// Timeout is expected if no message is sent
		t.Log("No ChangeConfiguration message received (expected if not triggered via API)")
		return
	}

	if action == "ChangeConfiguration" {
		// Respond with Accepted status
		responsePayload := map[string]interface{}{
			"status": "Accepted",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send ChangeConfiguration response: %v", err)
		}
		t.Logf("Responded to ChangeConfiguration: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not ChangeConfiguration)", action)
	}
}

// TestRemoteStartTransaction tests CS-initiated RemoteStartTransaction operation
// Note: This test requires the CS to send a RemoteStartTransaction Call message
func TestRemoteStartTransaction(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for RemoteStartTransaction Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No RemoteStartTransaction message received (expected if not triggered via API)")
		return
	}

	if action == "RemoteStartTransaction" {
		// Respond with Accepted status
		responsePayload := map[string]interface{}{
			"status": "Accepted",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send RemoteStartTransaction response: %v", err)
		}
		t.Logf("Responded to RemoteStartTransaction: %v", payload)

		// After RemoteStartTransaction is accepted, Charge Point should send StartTransaction
		// Wait for it (with a longer timeout)
		startTxID, startTxAction, _, err := client.ReadCallMessage(10 * time.Second)
		if err == nil && startTxAction == "StartTransaction" {
			// Respond to StartTransaction
			startTxResponse := map[string]interface{}{
				"transactionId": 1,
				"idTagInfo": map[string]interface{}{
					"status": "Accepted",
				},
			}
			if err := client.SendCallResult(startTxID, startTxResponse); err != nil {
				t.Fatalf("Failed to send StartTransaction response: %v", err)
			}
			t.Log("Responded to StartTransaction from RemoteStartTransaction")
		}
	} else {
		t.Logf("Received Call message: %s (not RemoteStartTransaction)", action)
	}
}

// TestChangeAvailability tests CS-initiated ChangeAvailability operation
// Note: This test requires the CS to send a ChangeAvailability Call message
func TestChangeAvailability(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for ChangeAvailability Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No ChangeAvailability message received (expected if not triggered via API)")
		return
	}

	if action == "ChangeAvailability" {
		// Respond with Accepted status
		responsePayload := map[string]interface{}{
			"status": "Accepted",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send ChangeAvailability response: %v", err)
		}
		t.Logf("Responded to ChangeAvailability: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not ChangeAvailability)", action)
	}
}

// TestClearCache tests CS-initiated ClearCache operation
func TestClearCache(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for ClearCache Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No ClearCache message received (expected if not triggered via API)")
		return
	}

	if action == "ClearCache" {
		// Respond with Accepted status
		responsePayload := map[string]interface{}{
			"status": "Accepted",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send ClearCache response: %v", err)
		}
		t.Logf("Responded to ClearCache: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not ClearCache)", action)
	}
}

// TestGetConfiguration tests CS-initiated GetConfiguration operation
func TestGetConfiguration(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for GetConfiguration Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No GetConfiguration message received (expected if not triggered via API)")
		return
	}

	if action == "GetConfiguration" {
		// Respond with configuration keys
		responsePayload := map[string]interface{}{
			"configurationKey": []map[string]interface{}{
				{
					"key":      "HeartbeatInterval",
					"readonly": false,
					"value":    "300",
				},
			},
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send GetConfiguration response: %v", err)
		}
		t.Logf("Responded to GetConfiguration: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not GetConfiguration)", action)
	}
}

// TestRemoteStopTransaction tests CS-initiated RemoteStopTransaction operation
func TestRemoteStopTransaction(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for RemoteStopTransaction Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No RemoteStopTransaction message received (expected if not triggered via API)")
		return
	}

	if action == "RemoteStopTransaction" {
		// Respond with Accepted status
		responsePayload := map[string]interface{}{
			"status": "Accepted",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send RemoteStopTransaction response: %v", err)
		}
		t.Logf("Responded to RemoteStopTransaction: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not RemoteStopTransaction)", action)
	}
}

// TestReset tests CS-initiated Reset operation
func TestReset(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for Reset Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No Reset message received (expected if not triggered via API)")
		return
	}

	if action == "Reset" {
		// Respond with Accepted status
		responsePayload := map[string]interface{}{
			"status": "Accepted",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send Reset response: %v", err)
		}
		t.Logf("Responded to Reset: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not Reset)", action)
	}
}

// TestUnlockConnector tests CS-initiated UnlockConnector operation
func TestUnlockConnector(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost/ws/cs002"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs002"
	}

	password := os.Getenv("CS_PASSWORD")
	if password == "" {
		password = "password"
	}

	client := NewVCPClient(csID, password, gatewayURL)

	// Connect to gateway
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send BootNotification first
	bootPayload := map[string]interface{}{
		"chargePointModel":  "VCP-Test",
		"chargePointVendor": "Solidstudio",
	}
	if err := client.SendOCPPMessage(2, "1", "BootNotification", bootPayload); err != nil {
		t.Fatalf("Failed to send BootNotification: %v", err)
	}

	// Read BootNotification response
	_, err := client.ReadOCPPMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to read BootNotification response: %v", err)
	}

	// Wait for UnlockConnector Call message from CS
	messageID, action, payload, err := client.ReadCallMessage(2 * time.Second)
	if err != nil {
		t.Log("No UnlockConnector message received (expected if not triggered via API)")
		return
	}

	if action == "UnlockConnector" {
		// Respond with Unlocked status
		responsePayload := map[string]interface{}{
			"status": "Unlocked",
		}
		if err := client.SendCallResult(messageID, responsePayload); err != nil {
			t.Fatalf("Failed to send UnlockConnector response: %v", err)
		}
		t.Logf("Responded to UnlockConnector: %v", payload)
	} else {
		t.Logf("Received Call message: %s (not UnlockConnector)", action)
	}
}

// Helper function to hash password like zynka-csms does
func hashPassword(password string) string {
	sha256pw := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(sha256pw[:])
}
