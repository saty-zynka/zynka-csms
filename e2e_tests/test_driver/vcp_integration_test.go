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

// VCPClient represents a client for Solidstudio Virtual Charge Point
// This is a proof-of-concept implementation showing how VCP would integrate
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
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
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

// Close closes the WebSocket connection
func (c *VCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close(websocket.StatusGoingAway, "Test complete")
	}
	return nil
}

// TestVCPBasicConnection tests basic WebSocket connection with Basic Auth
// This demonstrates how Solidstudio VCP would connect to zynka-csms
func TestVCPBasicConnection(t *testing.T) {
	// Get configuration from environment or use defaults
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost:9311/ws/cs001"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs001"
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
// This demonstrates how VCP would simulate a charging session
func TestVCPRFIDChargeFlow(t *testing.T) {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "ws://localhost:9311/ws/cs001"
	}

	csID := os.Getenv("CS_ID")
	if csID == "" {
		csID = "cs001"
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
		"idTag":      "DEADBEEF",
		"meterStop":  1000,
		"timestamp":  time.Now().Format(time.RFC3339),
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

// Helper function to hash password like zynka-csms does
func hashPassword(password string) string {
	sha256pw := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(sha256pw[:])
}

