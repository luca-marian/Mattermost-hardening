#!/bin/bash

# Services
SERVICES=(mattermost-metrics mattermost-satellite mattermost-ipdeny)

echo "🔍 Checking Linux services..."
for service in "${SERVICES[@]}"; do
    echo "Checking $service..."
    systemctl is-active --quiet "$service" && echo "✅ $service is running" || echo "❌ $service is NOT running"
done

# Mattermost Web
MM_URL="https://your-mattermost-url"
echo "🌐 Checking Mattermost Web access at $MM_URL..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$MM_URL")

if [ "$HTTP_STATUS" == "200" ] || [ "$HTTP_STATUS" == "302" ]; then
    echo "✅ Mattermost web server reachable (HTTP $HTTP_STATUS)"
else
    echo "❌ Mattermost web server unreachable (HTTP $HTTP_STATUS)"
fi
