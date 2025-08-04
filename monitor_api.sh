#!/bin/bash

# API monitoring script
# Continuously calls the API every second

API_URL="http://localhost:8080/data"
INTERVAL=1  # 1 second

echo "Starting API monitoring..."
echo "URL: $API_URL"
echo "Interval: ${INTERVAL}s"
echo "Press Ctrl+C to stop"
echo "----------------------------------------"

# Counter for requests
counter=1

while true; do
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Request #$counter"
    
    # Make the API call
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}\nTIME:%{time_total}s" "$API_URL")
    
    # Extract the response body and status
    body=$(echo "$response" | head -n -2)
    status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    time=$(echo "$response" | grep "TIME:" | cut -d: -f2)
    
    # Display results
    if [ "$status" = "200" ]; then
        echo "✅ Status: $status | Time: ${time}s"
        echo "Response length: ${#body} characters"
        echo "First 100 chars: ${body:0:100}..."
    else
        echo "❌ Status: $status | Time: ${time}s"
        echo "Error: $body"
    fi
    
    echo "----------------------------------------"
    
    # Increment counter
    ((counter++))
    
    # Wait for the specified interval
    sleep $INTERVAL
done 