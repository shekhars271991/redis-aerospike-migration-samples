#!/bin/bash

# AWS Deployment Script for Load Test Client
# This script sets up a high-performance load testing environment on AWS

set -e

echo "=== Load Test Client AWS Deployment ==="

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo "This script should not be run as root for security reasons"
   exit 1
fi

# System Information
echo "System Information:"
echo "  OS: $(uname -a)"
echo "  CPU Cores: $(nproc)"
echo "  Memory: $(free -h | grep '^Mem:' | awk '{print $2}')"
echo "  Network Interfaces: $(ip route | grep default | awk '{print $5}')"
echo ""

# Update system packages
echo "Updating system packages..."
if command -v yum &> /dev/null; then
    sudo yum update -y
    sudo yum install -y wget git htop iftop
elif command -v apt &> /dev/null; then
    sudo apt update
    sudo apt upgrade -y
    sudo apt install -y wget git htop iftop
fi

# Install Go 1.23 if not present
if ! command -v go &> /dev/null || [[ $(go version | grep -o 'go1\.[0-9]*') < "go1.23" ]]; then
    echo "Installing Go 1.23..."
    cd /tmp
    wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    
    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    source ~/.bashrc
    
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
fi

echo "Go version: $(go version)"

# System tuning for high performance
echo "Applying system tuning for high performance..."

# Increase file descriptor limits
echo "* soft nofile 1000000" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 1000000" | sudo tee -a /etc/security/limits.conf
echo "root soft nofile 1000000" | sudo tee -a /etc/security/limits.conf
echo "root hard nofile 1000000" | sudo tee -a /etc/security/limits.conf

# Network tuning
sudo sysctl -w net.core.somaxconn=65535
sudo sysctl -w net.ipv4.tcp_max_syn_backlog=65535
sudo sysctl -w net.core.netdev_max_backlog=65535
sudo sysctl -w net.ipv4.tcp_keepalive_time=600
sudo sysctl -w net.ipv4.tcp_keepalive_intvl=60
sudo sysctl -w net.ipv4.tcp_keepalive_probes=3
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
sudo sysctl -w net.ipv4.tcp_rmem="4096 65536 134217728"
sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 134217728"

# Make sysctl changes permanent
echo "net.core.somaxconn = 65535" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" | sudo tee -a /etc/sysctl.conf
echo "net.core.netdev_max_backlog = 65535" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_keepalive_time = 600" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_keepalive_intvl = 60" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_keepalive_probes = 3" | sudo tee -a /etc/sysctl.conf
echo "net.core.rmem_max = 134217728" | sudo tee -a /etc/sysctl.conf
echo "net.core.wmem_max = 134217728" | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_rmem = 4096 65536 134217728' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_wmem = 4096 65536 134217728' | sudo tee -a /etc/sysctl.conf

# Apply current session limits
ulimit -n 1000000

# Set Go runtime optimization
export GOMAXPROCS=$(nproc)
export GOGC=100

echo "GOMAXPROCS=$GOMAXPROCS" >> ~/.bashrc
echo "GOGC=100" >> ~/.bashrc

# Clone and build load test client
echo "Setting up load test client..."
cd $HOME

if [ -d "loadtest-client" ]; then
    echo "Updating existing loadtest-client..."
    cd loadtest-client
    git pull
else
    echo "Cloning loadtest-client repository..."
    # Replace with actual repository URL
    echo "Please clone the loadtest-client repository manually:"
    echo "git clone <repository-url> loadtest-client"
    echo "Then run this script again."
    exit 1
fi

# Build the client
echo "Building load test client..."
make deps
make build-prod

# Verify build
if [ ! -f "bin/loadtest" ]; then
    echo "Build failed - loadtest binary not found"
    exit 1
fi

echo "Load test client built successfully"

# Create results directory
mkdir -p results

# Display system status
echo ""
echo "=== System Status ==="
echo "File descriptor limit: $(ulimit -n)"
echo "Go version: $(go version)"
echo "CPU cores: $(nproc)"
echo "Available memory: $(free -h | grep '^Mem:' | awk '{print $7}')"
echo "Network settings applied: $(sysctl net.core.somaxconn)"

# Display usage instructions
echo ""
echo "=== Deployment Complete ==="
echo "Load test client is ready for high-performance testing!"
echo ""
echo "Quick start commands:"
echo "  cd ~/loadtest-client"
echo "  ./bin/loadtest --help"
echo "  ./bin/loadtest --config configs/default.yaml"
echo "  ./bin/loadtest --config configs/high-load.yaml"
echo ""
echo "For maximum performance (1M+ QPS):"
echo "  ./bin/loadtest --config configs/stress-test.yaml"
echo ""
echo "Monitor performance with:"
echo "  htop    # CPU and memory"
echo "  iftop   # Network usage"
echo "  ss -s   # Connection statistics"
echo ""
echo "IMPORTANT: Test gradually - start with default.yaml, then high-load.yaml"
echo "Only use stress-test.yaml on powerful instances (c5.24xlarge or larger)"

# Create a quick test script
cat > ~/quick-test.sh << 'EOF'
#!/bin/bash
cd ~/loadtest-client
echo "Running quick performance test..."
./bin/loadtest --url http://localhost:8080 --duration 30s --qps 1000 --workers 50
EOF

chmod +x ~/quick-test.sh

echo ""
echo "Created ~/quick-test.sh for quick testing"
echo "Run it with: ~/quick-test.sh"
