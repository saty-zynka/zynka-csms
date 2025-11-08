# End-to-end tests

## Description

Runs end-to-end tests against SolidStudio VCP (Virtual Charge Point simulator) along with CSMS. This performs the following tests:
- Plug-in a connector
- Authorise a charge (using RFID or ISO-15118)
- Start a charge
- Stop a charge
- Unplug a connector

## OCPP 1.6 Testing

These tests are configured to run with **OCPP 1.6** by default using SolidStudio VCP.

## Prerequisites

- Docker Desktop (or Docker with docker-compose)
- Go (for running the test driver)
- `jq` (for certificate scripts)
- `curl` (for API calls)
- `make` (for building certificates)

## Docker Desktop Setup

### IPv6 Compatibility

Docker Desktop on macOS and Windows may have IPv6 networking limitations. The test script automatically detects Docker Desktop and uses an IPv4-only configuration when IPv6 is not available.

If you encounter networking issues, you can manually use the Docker Desktop override file:

```shell
docker-compose -f docker-compose.yml -f docker-compose.docker-desktop.yml up -d
```

### Verifying Docker Desktop Environment

Before running tests, you can validate your Docker Desktop setup:

```shell
./validate-docker-desktop.sh
```

This script checks:
- Docker Desktop is running
- IPv6 support status
- Required Docker settings
- Network connectivity

## Steps

### Quick Start

1. Run the following bash script to start the docker containers for SolidStudio VCP and CSMS and execute the end-to-end tests:
```shell
./run-e2e-tests.sh
```

The script will:
- Automatically detect Docker Desktop and configure networking accordingly
- Start CSMS services (gateway, manager, MQTT, Firestore)
- Start SolidStudio VCP charge station simulator
- Run the end-to-end tests
- Clean up containers after tests complete

### Manual Steps

If you prefer to run steps manually:

1. **Start CSMS services:**
```shell
cd ..
(cd config/certificates && make)
chmod 755 config/certificates/csms.key
export UID=$(id -u)
export GID=$(id -g)
docker-compose up -d
```

2. **Wait for manager to be healthy:**
```shell
curl http://localhost:9410/health
```

3. **Register charge station and start SolidStudio VCP:**
```shell
cd e2e_tests
# Register charge station (Security Profile 0 = Basic Auth, no TLS)
curl -i http://localhost:9410/api/v0/cs/cs001 \
  -H 'content-type: application/json' \
  -d '{"securityProfile":0}'

# Start simulator (will use password from environment or default "password")
docker compose up -d --build
```

4. **Run tests:**
```shell
cd test_driver
go test --tags=e2e -v ./... -count=1
```

5. **Cleanup:**
```shell
cd ../..
docker-compose down
cd e2e_tests
docker compose down
```

## Troubleshooting

### Docker Desktop IPv6 Issues

If you see errors related to IPv6 networking:

1. **Check IPv6 support:**
```shell
docker network create --ipv6 --subnet=2001:db8::/64 test-network
docker network rm test-network
```

If this fails, IPv6 is not available and the script will automatically use IPv4-only mode.

2. **Manual override:**
Use the Docker Desktop override file explicitly:
```shell
docker-compose -f docker-compose.yml -f docker-compose.docker-desktop.yml up -d
```

### ARM64 Compatibility Issues

SolidStudio VCP has better cross-platform support than previous simulators. If you encounter any issues:

1. **Ensure Docker Desktop is up to date**
2. **Check container logs:**
   ```shell
   docker compose logs solidstudio-vcp
   ```

If issues persist, check the logs for specific error messages.

### OCPP Version Issues

Ensure OCPP 1.6 is enabled in the manager configuration. Check `config/manager/config.toml`:

```toml
[ocpp]
ocpp16_enabled = true
```

### Certificate Issues

If certificates are missing or invalid:

1. Regenerate certificates:
```shell
cd config/certificates
make
```

2. Copy CSMS certificate if needed (SolidStudio VCP handles this differently):
```shell
cd e2e_tests
# VCP configuration will be handled via environment variables or config files
```

### Port Conflicts

Ensure the following ports are available:
- `1883` - MQTT
- `9410` - Manager API
- `9411` - Manager OCPI
- `8080` - Firestore emulator

### Health Check Failures

If services fail health checks:

1. Check container logs:
```shell
docker-compose logs manager
docker-compose logs gateway
```

2. Verify network connectivity:
```shell
docker network inspect zynka-csms
```

3. Check if services are running:
```shell
docker-compose ps
```





