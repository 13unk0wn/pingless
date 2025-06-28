#!/bin/bash

# Configuration
API_BASE_URL="http://127.0.0.1:3000"
USERNAME="13unk0wn"
PASSWORD="RandomPassword"
PFP_FILE="./pfp.jpeg"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting PFP upload process...${NC}"

# Step 1: Verify user and get access token
echo -e "${YELLOW}Step 1: Authenticating user...${NC}"
AUTH_RESPONSE=$(curl -s -X POST "${API_BASE_URL}/api/user/verify_user" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")

# Check if authentication was successful
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to connect to the API server${NC}"
    exit 1
fi

# Extract access token from JSON response
ACCESS_TOKEN=$(echo "$AUTH_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}Error: Failed to get access token from response${NC}"
    echo -e "${YELLOW}Response:${NC} $AUTH_RESPONSE"
    exit 1
fi

echo -e "${GREEN}Successfully obtained access token${NC}"

# Step 2: Upload profile picture using the access token
echo -e "${YELLOW}Step 2: Uploading profile picture...${NC}"

# Check if PFP file exists
if [ ! -f "$PFP_FILE" ]; then
    echo -e "${RED}Error: PFP file not found: $PFP_FILE${NC}"
    exit 1
fi

UPLOAD_RESPONSE=$(curl -s -X POST "${API_BASE_URL}/api/user/upload_pfp" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -F "pfp=@${PFP_FILE}")

# Check if upload was successful
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to upload PFP${NC}"
    exit 1
fi

echo -e "${GREEN}Upload response:${NC} $UPLOAD_RESPONSE"
echo -e "${GREEN}PFP upload completed successfully!${NC}" 
