# Solidstudio VCP Integration Guide

## Overview

This guide explains how to integrate Solidstudio Virtual Charge Point (VCP) with zynka-csms for E2E testing.

## Prerequisites

1. Docker and Docker Compose installed
2. Go 1.19+ (for test driver)
3. zynka-csms services running (gateway, manager)

## Quick Start

### 1. Start zynka-csms Services

```bash
cd /path/to/zynka-csms
(cd config/certificates && make)
chmod 755 config/certificates/csms.key
export UID=$(id -u)
export GID=$(id -g)
docker-compose up -d
```

Wait for manager to be healthy:
```bash
curl http://localhost:9410/health
```

### 2. Register Charge Station

```bash
# Register with Security Profile 0 (Basic Auth, no TLS)
curl -i http://localhost:9410/api/v0/cs/cs001 \
  -H 'content-type: application/json' \
  -d '{"securityProfile":0}'

# Get the password from response or use default "password"
# Register RFID token
curl -i http://localhost:9410/api/v0/token \
  -H 'content-type: application/json' \
  -d '{
    "countryCode": "GB",
    "partyId": "TWK",
    "type": "RFID",
    "uid": "DEADBEEF",
    "contractId": "GBTWK012345678V",
    "issuer": "Zynka-tech",
    "valid": true,
    "cacheMode": "ALWAYS"
  }'
```

### 3. Run VCP Integration Tests

```bash
cd e2e_tests/test_driver
export GATEWAY_URL="ws://localhost:9311/ws/cs001"
export CS_ID="cs001"
export CS_PASSWORD="password"  # Use the password from registration
go test --tags=e2e -v -run TestVCPBasicConnection
```

## Docker Compose Example for VCP

Create `e2e_tests/docker-compose.vcp.yml`:

```yaml
version: '3.8'

networks:
  default:
    name: zynka-csms
    external: true

services:
  vcp-simulator:
    # Replace with actual Solidstudio VCP image when available
    # image: solidstudio/vcp:latest
    image: node:18-alpine  # Placeholder - replace with VCP image
    command: >
      sh -c "
        apk add --no-cache nodejs npm &&
        npm install -g @solidstudio/vcp &&
        vcp start --cs-id cs001 --url ws://gateway:9311/ws/cs001 --auth basic --username cs001 --password password
      "
    environment:
      - CS_ID=cs001
      - GATEWAY_URL=ws://gateway:9311/ws/cs001
      - AUTH_TYPE=basic
      - CS_USERNAME=cs001
      - CS_PASSWORD=password
    depends_on:
      - gateway
    networks:
      - default
```

**Note:** This is a placeholder configuration. Replace with actual VCP Docker image and commands once available.

## Test Scenarios

### Basic Connection Test

Tests WebSocket connection with Basic Auth:

```bash
go test --tags=e2e -v -run TestVCPBasicConnection
```

### RFID Charging Flow Test

Tests complete RFID charging lifecycle:

```bash
go test --tags=e2e -v -run TestVCPRFIDChargeFlow
```

## Authentication Configuration

### Security Profile 0 (Unsecured + Basic Auth)

```go
client := NewVCPClient("cs001", "password", "ws://localhost:9311/ws/cs001")
client.Connect()
```

### Security Profile 2 (TLS + Basic Auth)

```go
client := NewVCPClient("cs001", "password", "wss://localhost:9311/ws/cs001")
// VCP should handle TLS automatically
client.Connect()
```

### Security Profile 3 (TLS + Client Certificate)

Requires certificate configuration in VCP. See VCP documentation for details.

## Integration Points

### 1. WebSocket Connection

VCP connects to zynka-csms gateway at:
- Endpoint: `/ws/{chargeStationId}`
- Protocol: `ocpp1.6` or `ocpp2.0.1`
- Auth: Basic Auth header

### 2. OCPP Message Format

Standard OCPP JSON format:
```json
[2, "message-id", "Action", {payload}]
```

### 3. Test Flow

1. Connect to gateway with Basic Auth
2. Send BootNotification
3. Send StatusNotification
4. Send Authorize
5. Send StartTransaction
6. Send MeterValues
7. Send StopTransaction
8. Close connection

## Comparison with Previous Approach

| Aspect | Previous (MQTT-based) | SolidStudio VCP |
|--------|---------|-----------------|
| Control Interface | MQTT | Direct API/WebSocket |
| Setup Complexity | High | Low |
| Test Code | MQTT pub/sub | Direct OCPP messages |
| Startup Time | ~30s | ~5s |
| Resource Usage | High | Low |

## Troubleshooting

### Connection Failed

1. Verify gateway is running:
   ```bash
   curl http://localhost:9410/health
   ```

2. Check charge station is registered:
   ```bash
   curl http://localhost:9410/api/v0/cs/cs001
   ```

3. Verify password hash matches:
   ```bash
   # Password "password" should hash to:
   # XohImNooBHFR0OVvjcYpJ3NgPQ1qq73WKhHvch0VQtg=
   ```

### Authentication Failed

1. Check security profile matches connection type
2. Verify username matches charge station ID
3. Verify password hash is correct
4. Check Basic Auth header format

### OCPP Messages Not Received

1. Verify WebSocket connection is established
2. Check subprotocol is set correctly (`ocpp1.6`)
3. Verify message format is correct JSON array
4. Check gateway logs for errors

## Next Steps

1. **Get Solidstudio VCP Docker Image**
   - Download from GitHub releases
   - Or build from source

2. **Update Docker Compose**
   - Replace placeholder image
   - Configure VCP-specific settings

3. **Refactor Test Driver**
   - Replace MQTT-based tests
   - Use VCP client library
   - Maintain same test scenarios

4. **Validate Integration**
   - Run all E2E tests
   - Compare with previous test results
   - Verify consistency

## References

- [Solidstudio VCP GitHub](https://github.com/solidstudiosh/ocpp-virtual-charge-point)
- [OCPP 1.6 Specification](https://www.openchargealliance.org/)
- [zynka-csms Gateway Documentation](../docs/gateway.md)

