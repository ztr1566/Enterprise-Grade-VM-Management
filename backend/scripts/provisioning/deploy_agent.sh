#!/bin/bash
# deploy_agent.sh - Zero-trust agent deployment script
# Usage: ./deploy_agent.sh <vm_ip> <vm_user> <vm_id> <backend_url>

set -e

# Configuration
VM_IP=$1
VM_USER=$2
VM_ID=$3
BACKEND_URL=$4

if [[ -z "$VM_IP" || -z "$VM_USER" || -z "$VM_ID" || -z "$BACKEND_URL" ]]; then
    echo "Usage: $0 <vm_ip> <vm_user> <vm_id> <backend_url>"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "Building agent binary for Linux/AMD64..."
cd "$REPO_ROOT/backend"
GOOS=linux GOARCH=amd64 go build -o scripts/provisioning/agent ./cmd/agent/

echo "Deploying agent binary to $VM_USER@$VM_IP..."
scp "$REPO_ROOT/backend/scripts/provisioning/agent" "$VM_USER@$VM_IP:/tmp/vm-agent"

echo "Configuring systemd service on remote host..."
ssh "$VM_USER@$VM_IP" bash << EOF
    set -e
    sudo mv /tmp/vm-agent /usr/local/bin/vm-agent
    sudo chmod +x /usr/local/bin/vm-agent

    # Create systemd service file
    cat <<SERVICE | sudo tee /etc/systemd/system/vm-agent.service
[Unit]
Description=Enterprise VM Telemetry Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/vm-agent
WorkingDirectory=/var/lib/vm-agent
Environment=VM_ID=$VM_ID
Environment=BACKEND_URL=$BACKEND_URL
Restart=always
User=root

[Install]
WantedBy=multi-user.target
SERVICE

    # Setup working directory
    sudo mkdir -p /var/lib/vm-agent
    
    # Reload and restart service
    sudo systemctl daemon-reload
    sudo systemctl enable vm-agent
    sudo systemctl restart vm-agent
EOF

echo "------------------------------------------------"
echo "Deployment successful for VM: $VM_ID ($VM_IP)"
echo "The agent will now generate its own private key"
echo "and request an mTLS certificate from the backend."
echo "------------------------------------------------"
