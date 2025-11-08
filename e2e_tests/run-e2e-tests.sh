#!/bin/bash

# Get the directory where the script is located
# Cross-platform way to get script directory
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
else
    # Linux
    SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
fi

# Get the directory where the CSMS is located
DEFAULT_CSMS_DIR="${SCRIPT_DIR}"/..
CSMS_DIR="${1:-$DEFAULT_CSMS_DIR}"

# Define paths relative to the script's location
VCP_DIR="$CSMS_DIR/e2e_tests"
TEST_DIR="$CSMS_DIR/e2e_tests/test_driver"

# Function to detect Docker Desktop
detect_docker_desktop() {
    if [[ "$OSTYPE" == "darwin"* ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        # Check if Docker Desktop is running
        DOCKER_INFO=$(docker info 2>/dev/null || echo "")
        if echo "$DOCKER_INFO" | grep -q "Operating System.*Docker Desktop" || \
           (echo "$DOCKER_INFO" | grep -q "OSType.*linux" && [[ "$OSTYPE" == "darwin"* ]]); then
            return 0
        fi
    fi
    return 1
}

# Function to check IPv6 support
check_ipv6_support() {
    # Try to create a test network with IPv6
    TEST_NETWORK="zynka-test-ipv6-$$"
    if docker network create --ipv6 --subnet=2001:db8::/64 "$TEST_NETWORK" 2>/dev/null; then
        docker network rm "$TEST_NETWORK" >/dev/null 2>&1
        return 0
    else
        return 1
    fi
}

# Function to start Docker Compose
start_docker_compose_for_zynka_csms() {
    cd "$CSMS_DIR"
    (cd config/certificates && make)
    chmod 755 $CSMS_DIR/config/certificates/csms.key
    
    # Get user/group IDs for docker-compose
    DOCKER_UID=$(id -u)
    DOCKER_GID=$(id -g)
    
    # Determine which compose files to use
    COMPOSE_FILES="-f docker-compose.yml"
    if detect_docker_desktop && ! check_ipv6_support; then
        echo "Docker Desktop detected without IPv6 support. Using IPv4-only configuration."
        if [ -f "$CSMS_DIR/docker-compose.docker-desktop.yml" ]; then
            COMPOSE_FILES="$COMPOSE_FILES -f docker-compose.docker-desktop.yml"
        else
            echo "Warning: docker-compose.docker-desktop.yml not found. IPv6 may cause issues."
        fi
    fi
    
    # Use env to set UID/GID without readonly variable issues
    env UID=$DOCKER_UID GID=$DOCKER_GID docker-compose $COMPOSE_FILES up -d
    if [ $? -eq 0 ]; then
        echo "Docker Compose started successfully"
    else
        echo "Failed to start Docker Compose"
        stop_docker_compose_for_zynka_csms
        exit 1
    fi
}

# Function to register charge station
register_charge_station() {
    CS_ID="${CS_ID:-cs001}"
    SECURITY_PROFILE="${SECURITY_PROFILE:-0}"
    PASSWORD="${CS_PASSWORD:-password}"
    
    # Calculate password hash (SHA-256 + base64)
    PASSWORD_HASH=$(echo -n "$PASSWORD" | shasum -a 256 | awk '{print $1}' | xxd -r -p | base64)
    
    echo "Registering charge station $CS_ID..."
    curl -s -i http://localhost:9410/api/v0/cs/$CS_ID \
        -H 'content-type: application/json' \
        -d "{\"securityProfile\":$SECURITY_PROFILE,\"base64SHA256Password\":\"$PASSWORD_HASH\"}" > /dev/null
    
    if [ $? -eq 0 ]; then
        echo "Charge station $CS_ID registered successfully"
        echo "Using password: $PASSWORD"
        export CS_PASSWORD=$PASSWORD
    else
        echo "Warning: Failed to register charge station (may already be registered)"
    fi
    
    # Also register cs002 for tests (to avoid conflicts with simulator using cs001)
    echo "Registering charge station cs002 for tests..."
    curl -s -i http://localhost:9410/api/v0/cs/cs002 \
        -H 'content-type: application/json' \
        -d "{\"securityProfile\":$SECURITY_PROFILE,\"base64SHA256Password\":\"$PASSWORD_HASH\"}" > /dev/null
    if [ $? -eq 0 ]; then
        echo "Charge station cs002 registered successfully"
    fi
}

# Function to start Docker Compose for SolidStudio VCP
start_docker_compose_for_vcp() {
        cd "$VCP_DIR"
        
        # Register charge station before starting simulator
        register_charge_station
        
        echo "Starting SolidStudio VCP simulator..."
        docker compose up -d --build
        if [ $? -ne 0 ]; then
            echo "Failed to start Docker Compose for SolidStudio VCP"
            stop_docker_compose_for_vcp
            exit 1
        fi

        echo "Waiting for SolidStudio VCP to initialize and connect..."
        sleep 10
        
        # Check if simulator container is running
        if ! docker compose ps | grep -q "solidstudio-vcp.*Up"; then
            echo "Warning: SolidStudio VCP container may not be running properly"
            docker compose logs solidstudio-vcp
        fi
}

# Function to stop Docker Compose for SolidStudio VCP
stop_docker_compose_for_vcp() {
    cd "$VCP_DIR" && docker compose down
}

stop_docker_compose_for_zynka_csms() {
    cd "$CSMS_DIR"
    # Use same compose files as start
    COMPOSE_FILES="-f docker-compose.yml"
    if detect_docker_desktop && ! check_ipv6_support && [ -f "$CSMS_DIR/docker-compose.docker-desktop.yml" ]; then
        COMPOSE_FILES="$COMPOSE_FILES -f docker-compose.docker-desktop.yml"
    fi
    DOCKER_UID=$(id -u)
    DOCKER_GID=$(id -g)
    env UID=$DOCKER_UID GID=$DOCKER_GID docker-compose $COMPOSE_FILES down
}

# Function to check health endpoint
check_health_endpoint() {
    HEALTH_ENDPOINT="http://localhost:9410/health"
    echo "$(date +"%Y-%m-%d %H:%M:%S"):Waiting for the health endpoint to become available..."
    while true; do
        STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_ENDPOINT)
        if [ $STATUS_CODE -eq 200 ]; then
            echo "$(date +"%Y-%m-%d %H:%M:%S"):Health endpoint is available (HTTP 200)"
            break
        else
            echo "$(date +"%Y-%m-%d %H:%M:%S"):Health endpoint is not yet available (HTTP $STATUS_CODE)"
            sleep 5
        fi
    done
}

# Function to run tests
run_tests() {
    echo "Running test command..."
    cd "$TEST_DIR"
    go test --tags=e2e -v ./... -count=1
    TEST_RESULT=$?
    cd "$CSMS_DIR"
    if [ $TEST_RESULT -eq 0 ]; then
        echo "Tests completed successfully"
    else
        echo "Tests failed"
    fi

    stop_docker_compose_for_vcp
    stop_docker_compose_for_zynka_csms
}

# Main script execution
start_docker_compose_for_zynka_csms
check_health_endpoint
start_docker_compose_for_vcp
run_tests
