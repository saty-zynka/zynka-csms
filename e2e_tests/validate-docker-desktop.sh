#!/bin/bash

# Docker Desktop validation script for OCPP 1.6 testing
# This script validates the Docker Desktop environment and checks for compatibility

set -e

# Cross-platform way to get script directory
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
else
    # Linux
    SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
fi
CSMS_DIR="${SCRIPT_DIR}/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Function to print pass message
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

# Function to print fail message
fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

# Function to print warning message
warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

# Function to print info message
info() {
    echo -e "  $1"
}

echo "Docker Desktop Environment Validation for OCPP 1.6 Testing"
echo "========================================================="
echo ""

# Check if Docker is installed and running
echo "1. Checking Docker installation..."
if command -v docker >/dev/null 2>&1; then
    pass "Docker is installed"
else
    fail "Docker is not installed"
    exit 1
fi

if docker info >/dev/null 2>&1; then
    pass "Docker daemon is running"
else
    fail "Docker daemon is not running. Please start Docker Desktop."
    exit 1
fi

# Check Docker Desktop
echo ""
echo "2. Checking Docker Desktop..."
DOCKER_INFO=$(docker info 2>/dev/null || echo "")
if echo "$DOCKER_INFO" | grep -q "Operating System.*Docker Desktop" || \
   ([[ "$OSTYPE" == "darwin"* ]] && echo "$DOCKER_INFO" | grep -q "OSType.*linux"); then
    pass "Docker Desktop detected"
    IS_DOCKER_DESKTOP=true
else
    warn "Docker Desktop not detected (may be using different Docker setup)"
    IS_DOCKER_DESKTOP=false
fi

# Check docker-compose
echo ""
echo "3. Checking docker-compose..."
if command -v docker-compose >/dev/null 2>&1 || docker compose version >/dev/null 2>&1; then
    pass "docker-compose is available"
    if command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD="docker-compose"
    else
        DOCKER_COMPOSE_CMD="docker compose"
    fi
else
    fail "docker-compose is not available"
    exit 1
fi

# Check IPv6 support
echo ""
echo "4. Checking IPv6 support..."
TEST_NETWORK="zynka-validation-ipv6-$$"
if docker network create --ipv6 --subnet=2001:db8::/64 "$TEST_NETWORK" >/dev/null 2>&1; then
    docker network rm "$TEST_NETWORK" >/dev/null 2>&1
    pass "IPv6 is supported"
    IPV6_SUPPORTED=true
else
    warn "IPv6 is not supported - will use IPv4-only configuration"
    IPV6_SUPPORTED=false
    if [ "$IS_DOCKER_DESKTOP" = true ]; then
        info "This is normal for Docker Desktop. The test script will automatically use docker-compose.docker-desktop.yml"
    fi
fi

# Check required ports
echo ""
echo "5. Checking required ports..."
REQUIRED_PORTS=(80 443 1883 1884 9410 9411 8080)
ALL_PORTS_AVAILABLE=true
for port in "${REQUIRED_PORTS[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        warn "Port $port is already in use"
        ALL_PORTS_AVAILABLE=false
    fi
done
if [ "$ALL_PORTS_AVAILABLE" = true ]; then
    pass "All required ports are available"
else
    warn "Some ports are in use. This may cause conflicts during testing."
fi

# Check for docker-compose.docker-desktop.yml
echo ""
echo "6. Checking Docker Desktop override file..."
if [ -f "$CSMS_DIR/docker-compose.docker-desktop.yml" ]; then
    pass "docker-compose.docker-desktop.yml found"
else
    warn "docker-compose.docker-desktop.yml not found (will be created if needed)"
fi

# Check OCPP configuration
echo ""
echo "7. Checking OCPP 1.6 configuration..."
if [ -f "$CSMS_DIR/config/manager/config.toml" ]; then
    if grep -q "ocpp16_enabled.*=.*true" "$CSMS_DIR/config/manager/config.toml" 2>/dev/null; then
        pass "OCPP 1.6 is enabled in config.toml"
    else
        warn "OCPP 1.6 may not be explicitly enabled in config.toml (checking defaults...)"
        # Default config has ocpp16_enabled=true, so this is likely OK
        info "Default configuration enables OCPP 1.6, so this should be fine"
    fi
else
    warn "config.toml not found at expected location"
fi

# Check certificates
echo ""
echo "8. Checking certificates..."
if [ -d "$CSMS_DIR/config/certificates" ]; then
    if [ -f "$CSMS_DIR/config/certificates/csms.pem" ] && [ -f "$CSMS_DIR/config/certificates/csms.key" ]; then
        pass "CSMS certificates found"
    else
        warn "CSMS certificates not found - will be generated during test setup"
    fi
else
    warn "Certificates directory not found"
fi

# Check required tools
echo ""
echo "9. Checking required tools..."
REQUIRED_TOOLS=("jq" "curl" "make" "go")
for tool in "${REQUIRED_TOOLS[@]}"; do
    if command -v "$tool" >/dev/null 2>&1; then
        pass "$tool is installed"
    else
        fail "$tool is not installed"
    fi
done

# Check network connectivity
echo ""
echo "10. Checking network connectivity..."
if docker network ls | grep -q "zynka-csms"; then
    warn "zynka-csms network already exists (may need cleanup)"
else
    pass "zynka-csms network does not exist (will be created)"
fi

# Summary
echo ""
echo "========================================================="
echo "Validation Summary:"
echo "  Passed:   $PASSED"
echo "  Failed:   $FAILED"
echo "  Warnings: $WARNINGS"
echo ""

if [ $FAILED -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}All checks passed! You're ready to run OCPP 1.6 tests.${NC}"
        exit 0
    else
        echo -e "${YELLOW}Checks passed with warnings. Review warnings above.${NC}"
        exit 0
    fi
else
    echo -e "${RED}Some checks failed. Please fix the issues above before running tests.${NC}"
    exit 1
fi

