#!/bin/bash

echo "Testing EasilyPanel5 Download API"

# Test status endpoint
echo "1. Testing status endpoint..."
wget -q -O - http://localhost:8080/api/status
echo -e "\n"

# Test Java detection
echo "2. Testing Java detection..."
wget -q -O - http://localhost:8080/api/java/detect
echo -e "\n"

# Test cores list
echo "3. Testing cores list..."
wget -q -O - http://localhost:8080/api/cores/list | head -20
echo -e "\n"

# Test Paper versions
echo "4. Testing Paper versions..."
wget -q -O - "http://localhost:8080/api/cores/versions?type=Paper" | head -20
echo -e "\n"

echo "All tests completed!"
