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

### 3. Start SolidStudio VCP Simulator

```bash
cd e2e_tests
docker compose up -d --build
```

The simulator will:
- Build the Docker image from `simulator/` directory
- Connect to gateway at `ws://gateway:9311/ws/cs001`
- Send BootNotification automatically
- Send StatusNotification for connector 1 (Available)
- Start admin API on port 9999

### 4. Run VCP Integration Tests

```bash
cd e2e_tests/test_driver
export GATEWAY_URL="ws://localhost:9311/ws/cs001"
export CS_ID="cs001"
export CS_PASSWORD="password"  # Use the password from registration
go test --tags=e2e -v -run TestVCPBasicConnection
```

Or run all tests:
```bash
go test --tags=e2e -v ./... -count=1
```

## Docker Compose Configuration

The simulator is configured in `e2e_tests/docker-compose.yml`:

```yaml
services:
  solidstudio-vcp:
    build:
      context: ../simulator
      dockerfile: Dockerfile
    environment:
      - WS_URL=ws://gateway:9311/ws
      - CP_ID=cs001
      - PASSWORD=${CS_PASSWORD:-password}
      - ADMIN_PORT=9999
      - OCPP_VERSION=16
    ports:
      - "9999:9999"  # Admin API port
    networks:
      - default
```

**Key Configuration:**
- `WS_URL`: Base WebSocket URL (simulator appends `/CP_ID`)
- `CP_ID`: Charge point identifier (must match CSMS registration)
- `PASSWORD`: Basic Auth password (must match CSMS registration)
- `OCPP_VERSION`: OCPP version (16, 201, or 21)

**Note:** The simulator code is located in the `simulator/` directory and is built into a Docker image during `docker compose up`.

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

The simulator connects to zynka-csms gateway at:
- Base URL: `ws://gateway:9311/ws` (from `WS_URL` env var)
- Full endpoint: `{WS_URL}/{CP_ID}` = `ws://gateway:9311/ws/cs001`
- Protocol: `ocpp1.6` (or `ocpp2.0.1` based on `OCPP_VERSION`)
- Auth: Basic Auth header with `CP_ID:PASSWORD`

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

### Simulator Container Issues

1. **Check simulator logs:**
   ```bash
   docker compose logs solidstudio-vcp
   ```

2. **Verify simulator is running:**
   ```bash
   docker compose ps
   ```

3. **Rebuild simulator image:**
   ```bash
   docker compose build --no-cache solidstudio-vcp
   docker compose up -d
   ```

4. **Check WebSocket connection:**
   - Look for "Connecting..." messages in logs
   - Verify "Sending message ➡️" indicates successful connection
   - Check for connection errors

## Simulator Details

### Building the Simulator

The simulator Docker image is built from the `simulator/` directory:

```bash
cd simulator
docker build -t solidstudio-vcp .
```

Or use docker-compose in `e2e_tests/`:
```bash
cd e2e_tests
docker compose build solidstudio-vcp
```

### Simulator Features

- **OCPP Versions**: Supports 1.6, 2.0.1, and 2.1
- **Auto-connect**: Automatically connects and sends BootNotification on startup
- **Admin API**: HTTP API on port 9999 for sending custom OCPP messages
- **Environment-based**: Configured via environment variables
- **TypeScript/Node.js**: Built with Node.js 18 and TypeScript

### Admin API Usage

The simulator exposes an admin API on port 9999 for sending custom OCPP messages:

```bash
curl -X POST http://localhost:9999/execute \
  -H "Content-Type: application/json" \
  -d '{
    "action": "Authorize",
    "payload": {"idTag": "DEADBEEF"}
  }'
```

## Next Steps

1. **Run E2E Tests**
   - Use `./run-e2e-tests.sh` for automated testing
   - Or follow manual steps above

2. **Customize Simulator**
   - Modify `simulator/index_16.ts` for OCPP 1.6 behavior
   - Modify `simulator/index_201.ts` for OCPP 2.0.1 behavior
   - Add custom scenarios in `simulator/admin/` directory

3. **Extend Tests**
   - Add more test scenarios in `test_driver/vcp_integration_test.go`
   - Use admin API for complex test flows
   - Test different OCPP versions

## References

- [Solidstudio VCP GitHub](https://github.com/solidstudiosh/ocpp-virtual-charge-point)
- [OCPP 1.6 Specification](https://www.openchargealliance.org/)
- [zynka-csms Gateway Documentation](../docs/gateway.md)

