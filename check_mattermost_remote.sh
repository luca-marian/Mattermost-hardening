#!/bin/bash

# Configuration
REMOTE_USER="tugas"                                    # your user on the VM
REMOTE_HOST="100.97.5.5"                               # your VM IP or DNS
REMOTE_PORT=22                                         # SSH port, usually 22
MATTERMOST_URL="https://mattermost.beg.05.berylia.org" # your Mattermost URL

REMOTE_COMMANDS=$(
    cat <<'EOF'
echo "🔍 Running Mattermost Health Check on $(hostname)..."

# 1. Linux Services to Test
SERVICES=("mattermost-metrics" "mattermost-satellite" "mattermost-ipdeny")

echo "🔧 Checking Linux services..."
for service in "${SERVICES[@]}"; do
    echo "Checking $service..."
    sudo systemctl is-active --quiet "$service" && echo "✅ $service is running" || echo "❌ $service is NOT running"
done

# 2. Docker Containers
echo "🐳 Checking Docker container health..."

sudo docker ps --format '{{.Names}}' | while read container_name; do
  health=$(sudo docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}no healthcheck{{end}}' "$container_name")

  if [[ "$health" == "healthy" ]]; then
    echo "✅ [$container_name] is healthy."
  elif [[ "$health" == "unhealthy" ]]; then
    echo "❌ [$container_name] is UNHEALTHY!"
    sudo docker logs --tail 10 "$container_name"
  elif [[ "$health" == "no healthcheck" ]]; then
    status=$(sudo docker inspect --format='{{.State.Status}}' "$container_name")
    if [[ "$status" == "running" ]]; then
      echo "⚠️ [$container_name] running, but no healthcheck."
    else
      echo "❌ [$container_name] is NOT running!"
      sudo docker logs --tail 10 "$container_name"
    fi
  else
    echo "❓ [$container_name] unknown health status: $health"
  fi
done

EOF
)

# Now actually connect and run commands
echo "🚀 Connecting to $REMOTE_USER@$REMOTE_HOST..."
ssh -i './open_ssh_key' -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "$REMOTE_COMMANDS"

# === Local Mattermost Web Check ===
echo "🌐 Checking Mattermost Web access at $MATTERMOST_URL..."
http_status=$(curl -k -s -o /dev/null -w "%{http_code}" "$MATTERMOST_URL")
if [[ "$http_status" == "200" || "$http_status" == "302" ]]; then
    echo "✅ Mattermost web server reachable (HTTP $http_status)"
else
    echo "❌ Mattermost web server unreachable (HTTP $http_status)"
fi

echo "✅ Remote Mattermost health check complete."
