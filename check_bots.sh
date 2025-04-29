#!/bin/bash

# Configuration
CONTAINER_NAME="mattermost_app_1" # Replace with your container name
MATTERMOST_URL="https://mattermost.beg.##.berylia.org"
ACCESS_TOKEN_DIR="./bot_tokens" # Directory with saved bot tokens

SEND_TEST_MESSAGE=true # Set to false if you want only token tests

# ---- Script starts ----

echo "🔍 Finding bots inside container $CONTAINER_NAME..."

# Extract bot list inside the container
docker exec "$CONTAINER_NAME" bash -c "cd /mattermost && ./bin/mattermost user list --all" | grep Bot | awk '{print $2}' >bots_list.txt

while read botname; do
    echo "----------------------------------------"
    echo "🤖 Bot: $botname"

    # Check roles inside container
    ROLES=$(docker exec "$CONTAINER_NAME" bash -c "cd /mattermost && ./bin/mattermost user get $botname | grep Roles")
    echo "$ROLES"

    if [[ "$ROLES" == *"system_admin"* ]]; then
        echo "⚠️  WARNING: Bot has system_admin role!"
    else
        echo "✅ OK: Bot does NOT have system_admin role."
    fi

    # Check token
    TOKEN_FILE="$ACCESS_TOKEN_DIR/${botname}.token"
    if [ ! -f "$TOKEN_FILE" ]; then
        echo "❌ No token found for $botname at $TOKEN_FILE. Skipping token test."
        continue
    fi
    TOKEN=$(cat "$TOKEN_FILE")

    echo "🔗 Testing access token..."
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$MATTERMOST_URL/api/v4/users/me")

    if [ "$HTTP_STATUS" == "200" ]; then
        echo "✅ Token is valid."
    else
        echo "❌ Token test failed with HTTP $HTTP_STATUS."
        continue
    fi

    # Optionally, send test post
    if [ "$SEND_TEST_MESSAGE" = true ]; then
        echo "📢 Sending test message..."
        read -p "Enter channel_id where $botname is a member (or leave empty to skip): " channel_id

        if [ ! -z "$channel_id" ]; then
            RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H 'Content-Type: application/json' \
                -d "{\"channel_id\":\"$channel_id\", \"message\":\"Test message from $botname after hardening.\"}" \
                -H "Authorization: Bearer $TOKEN" \
                "$MATTERMOST_URL/api/v4/posts")

            if [ "$RESPONSE" == "201" ]; then
                echo "✅ Successfully posted message."
            else
                echo "❌ Failed to post message. HTTP $RESPONSE."
            fi
        else
            echo "ℹ️ Skipped test post for $botname."
        fi
    fi

done <bots_list.txt

echo "✅ All bots processed."
