# AGENTS.md

Instructions for building and testing Zynka CSMS after feature implementation.

## Prerequisites

- Go 1.20+ (Gateway requires Go 1.20, Manager requires Go 1.21)
- Docker and docker-compose
- `jq` (for certificate scripts)
- `make` (for building certificates)
- `curl` (for API calls and health checks)

## Build Instructions

### Build Gateway Component

```bash
cd gateway
go build ./...
```

### Build Manager Component

```bash
cd manager
go build ./...
```

### Build Docker Images

Build gateway Docker image:
```bash
cd gateway
docker build . --file Dockerfile --tag gateway:latest --build-arg TARGETARCH=amd64
```

Build manager Docker image:
```bash
cd manager
docker build . --file Dockerfile --tag manager:latest --build-arg TARGETARCH=amd64
```

### Generate Certificates

Before running the system, generate required certificates:
```bash
cd config/certificates
make
chmod 755 csms.key
```

## Unit Testing

### Run Unit Tests with Coverage

**Gateway:**
```bash
cd gateway
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**Manager:**
```bash
cd manager
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run Integration Tests

Integration tests require Docker to be running. Set up environment variables first:

```bash
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
export DOCKER_HOST=$(docker context inspect -f '{{ .Endpoints.docker.Host }}')
```

Then run integration tests:

**Manager (integration tests):**
```bash
cd manager
go test ./store/... --tags=integration
```

### Security Scanning (Optional)

Run Gosec security scanner (matches CI workflow):

**Gateway:**
```bash
cd gateway
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec -exclude-generated ./...
```

**Manager:**
```bash
cd manager
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec -exclude-generated ./...
```

## E2E Testing

### Quick E2E Test Run

Run the automated E2E test script (handles setup, execution, and cleanup):
```bash
./e2e_tests/run-e2e-tests.sh
```

### Manual E2E Test Steps

1. **Start CSMS services:**
```bash
(cd config/certificates && make)
chmod 755 config/certificates/csms.key
export UID=$(id -u)
export GID=$(id -g)
docker-compose up -d
```

2. **Wait for manager to be healthy:**
```bash
curl http://localhost:9410/health
```

3. **Register charge station and start simulator:**
```bash
cd e2e_tests
curl -i http://localhost:9410/api/v0/cs/cs001 \
  -H 'content-type: application/json' \
  -d '{"securityProfile":0}'
docker compose up -d --build
```

4. **Run E2E tests:**
```bash
cd test_driver
go test --tags=e2e -v ./... -count=1
```

5. **Cleanup:**
```bash
cd ../..
docker-compose down
cd e2e_tests
docker compose down
```

## Post-Feature Implementation Workflow

After implementing a feature, follow these steps:

1. **Build the affected component(s):**
   - If changes are in `gateway/`, run `cd gateway && go build ./...`
   - If changes are in `manager/`, run `cd manager && go build ./...`

2. **Run unit tests with coverage:**
   - Run `go test ./... -coverprofile=coverage.out` in the affected component directory
   - Verify all tests pass and coverage is acceptable

3. **Run integration tests (if applicable):**
   - If changes affect storage or external integrations, run integration tests with `--tags=integration`
   - Ensure Docker daemon is running before executing

4. **Run E2E tests:**
   - Execute `./e2e_tests/run-e2e-tests.sh` to validate end-to-end functionality
   - Ensure all E2E tests pass before submitting changes

5. **Verify Docker builds:**
   - Build Docker images to ensure Dockerfile changes (if any) work correctly
   - Test that containers start successfully

6. **Check code quality:**
   - Run security scanner: `gosec -exclude-generated ./...`
   - Ensure no critical or high severity vulnerabilities

## Testing Notes

- Unit tests should pass before committing changes
- Integration tests require Docker and may take longer to execute
- E2E tests require the full stack (CSMS + simulator) and may take several minutes
- For Docker Desktop on macOS/Windows, IPv6 may not be available; the E2E script handles this automatically
- Certificate generation is required before first run or if certificates are missing

## Component-Specific Notes

- **Gateway**: Handles WebSocket connections from charge stations, forwards messages via MQTT
- **Manager**: Processes OCPP messages, manages charge stations and tokens, provides REST API
- Both components are built as separate Go modules with their own `go.mod` files
- CI workflows run tests and build Docker images automatically on push

