#!/bin/bash

# Configuration
API_BASE_URL="http://127.0.0.1:3000"
NGINX_URL="http://localhost"
USERNAME="13unk0wn"
PASSWORD="RandomPassword"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Getting profile picture URL for user: ${USERNAME}${NC}"

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

# Step 2: Get profile picture URL using the access token
echo -e "${YELLOW}Step 2: Retrieving profile picture URL...${NC}"

# Method 1: Get specific PFP image by type
PFP_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/api/user/image?username=${USERNAME}&type=pfp" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

# Check if the request was successful
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to get profile picture URL${NC}"
    exit 1
fi

# Extract URL from JSON response
PFP_URL=$(echo "$PFP_RESPONSE" | grep -o '"url":"[^"]*"' | cut -d'"' -f4)

if [ -z "$PFP_URL" ]; then
    echo -e "${YELLOW}No profile picture found for user: ${USERNAME}${NC}"
    echo -e "${YELLOW}Response:${NC} $PFP_RESPONSE"
    
    # Method 2: Get all user images as fallback
    echo -e "${YELLOW}Trying to get all user images...${NC}"
    ALL_IMAGES_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/api/user/images?username=${USERNAME}" \
      -H "Authorization: Bearer ${ACCESS_TOKEN}")
    
    if [ $? -eq 0 ]; then
        echo -e "${BLUE}All user images:${NC} $ALL_IMAGES_RESPONSE"
    fi
else
    # Construct full URLs for both backend and nginx
    BACKEND_URL="${API_BASE_URL}${PFP_URL}"
    NGINX_FULL_URL="${NGINX_URL}${PFP_URL}"
    
    echo -e "${GREEN}Profile picture found!${NC}"
    echo -e "${BLUE}Relative URL:${NC} $PFP_URL"
    echo -e "${BLUE}Backend URL (API):${NC} $BACKEND_URL"
    echo -e "${BLUE}Nginx URL (View Image):${NC} $NGINX_FULL_URL"
    
    # Display additional image info
    echo -e "${YELLOW}Image details:${NC}"
    echo "$PFP_RESPONSE" | jq '.' 2>/dev/null || echo "$PFP_RESPONSE"
    
    echo -e "${GREEN}To view the image in your browser, open:${NC}"
    echo -e "${BLUE}$NGINX_FULL_URL${NC}"
fi

echo -e "${GREEN}Profile picture URL retrieval completed!${NC}" 
