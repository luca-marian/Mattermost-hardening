#!/bin/bash

# Configuration
REMOTE_USER="tugas"                                                             # your user on the VM
REMOTE_HOST="100.97.5.5"                                                        # your VM IP or DNS
REMOTE_PORT=22                                                                  # SSH port, usually 22
MATTERMOST_URL="https://mattermost.beg.05.berylia.org/beg/channels/ipdeny-feed" # your Mattermost URL

# Commands to execute remotely
REMOTE_COMMANDS=$(
    cat <<EOF
echo "🔍 Docker Health Check on \$(hostname)..."

sudo docker ps --format '{{.Names}}' | while read container_name; do
  health=\$(sudo docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}no healthcheck{{end}}' "\$container_name")

  if [[ "\$health" == "healthy" ]]; then
    echo "✅ [\$container_name] is healthy."
  elif [[ "\$health" == "unhealthy" ]]; then
    echo "❌ [\$container_name] is UNHEALTHY!"
    sudo docker logs --tail 10 "\$container_name"
  elif [[ "\$health" == "no healthcheck" ]]; then
    status=\$(sudo docker inspect --format='{{.State.Status}}' "\$container_name")
    if [[ "\$status" == "running" ]]; then
      echo "⚠️  [\$container_name] running, but no healthcheck."
    else
      echo "❌ [\$container_name] is NOT running!"
      sudo docker logs --tail 10 "\$container_name"
    fi
  else
    echo "❓ [\$container_name] health status unknown: \$health"
  fi
done

# Check Mattermost Web
HTTP_STATUS=\$(curl -k -s -o /dev/null -w "%{http_code}" "$MATTERMOST_URL")
if [[ "\$HTTP_STATUS" == "200" ]] || [[ "\$HTTP_STATUS" == "302" ]]; then
  echo "✅ Mattermost web server reachable (HTTP \$HTTP_STATUS)"
else
  echo "❌ Mattermost web server unreachable (HTTP \$HTTP_STATUS)"
fi
EOF
)

# Now actually connect and run commands
echo "🚀 Connecting to $REMOTE_USER@$REMOTE_HOST..."
ssh -i './open_ssh_key' -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "$REMOTE_COMMANDS"

echo "✅ Remote Mattermost health check complete."
