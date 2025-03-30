#!/bin/bash

# Define variables to edit
REGION="ap-south-1"                     # Your AWS region
ALB_SECURITY_GROUP_ID="sg-"       # Replace with your ALB's Security Group ID
ALB_PORT=80                             # Port for the ALB listener (default: 80)

# Fetch all API Gateway IP ranges
echo "Fetching API Gateway IP ranges globally..."
API_GATEWAY_IPS=$(curl -s https://ip-ranges.amazonaws.com/ip-ranges.json | \
  jq -r ".prefixes[] | select(.service==\"API_GATEWAY\" and .region==\"${REGION}\") | .ip_prefix")

echo -e "API Gateway IP ranges for ${REGION}:\n${API_GATEWAY_IPS}"

if [ -z "$API_GATEWAY_IPS" ]; then
    echo "Error: Unable to fetch API Gateway IP ranges. Please check your internet connection."
    exit 1
fi

# Remove existing rules for the ALB Security Group for the specified port
echo "Removing existing inbound rules for ALB Security Group: ${ALB_SECURITY_GROUP_ID}..."
EXISTING_RULES=$(aws ec2 describe-security-groups --group-ids ${ALB_SECURITY_GROUP_ID} \
--query "SecurityGroups[0].IpPermissions[?FromPort==\`${ALB_PORT}\`].IpRanges[].CidrIp" --output text)

if [ -n "$EXISTING_RULES" ]; then
    for CIDR in ${EXISTING_RULES}; do
        aws ec2 revoke-security-group-ingress --group-id ${ALB_SECURITY_GROUP_ID} \
        --protocol tcp --port ${ALB_PORT} --cidr ${CIDR}
        echo "Removed rule allowing ${CIDR} on port ${ALB_PORT}"
    done
else
    echo "No existing rules found for port ${ALB_PORT}."
fi

# Add API Gateway IP ranges to the ALB Security Group
echo "Adding API Gateway IP ranges to ALB Security Group..."
for CIDR in ${API_GATEWAY_IPS}; do
    aws ec2 authorize-security-group-ingress --group-id ${ALB_SECURITY_GROUP_ID} \
    --protocol tcp --port ${ALB_PORT} --cidr ${CIDR}
    echo "Added rule allowing ${CIDR} on port ${ALB_PORT}"
done

# Verify the updated Security Group rules
echo "Verifying updated rules for ALB Security Group: ${ALB_SECURITY_GROUP_ID}..."
aws ec2 describe-security-groups --group-ids ${ALB_SECURITY_GROUP_ID} --query "SecurityGroups[0].IpPermissions"
echo "Security Group updated successfully."
