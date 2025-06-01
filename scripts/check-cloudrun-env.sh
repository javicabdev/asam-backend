#!/bin/bash
# Script to check current environment variables in Cloud Run

SERVICE_NAME="asam-backend"
REGION="europe-west1"

echo "Checking environment variables for $SERVICE_NAME..."
echo "================================================="

# Get the service configuration
gcloud run services describe $SERVICE_NAME \
    --region=$REGION \
    --format="export" | grep -E "^\s*- name:|^\s*value:" | \
    while IFS= read -r line && IFS= read -r value_line; do
        # Extract variable name
        var_name=$(echo "$line" | sed 's/.*name: //')
        # Extract value (masked for sensitive data)
        var_value=$(echo "$value_line" | sed 's/.*value: //')
        
        # Mask sensitive values
        if [[ "$var_name" =~ (PASSWORD|SECRET|KEY) ]]; then
            if [ -z "$var_value" ]; then
                echo "$var_name: <not set>"
            else
                echo "$var_name: <set - masked>"
            fi
        else
            echo "$var_name: $var_value"
        fi
    done

echo "================================================="
echo ""
echo "Missing required variables:"
echo "- ADMIN_USER and ADMIN_PASSWORD are required for production"
echo "- SMTP_USER and SMTP_PASSWORD are optional (service will run without them)"
