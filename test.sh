#!/bin/bash

echo "Testing Fake Hue Lights"
echo "======================="

echo "Getting all lights:"
curl -s http://localhost:8080/api/testuser/lights | python3 -m json.tool

echo -e "\n\nTurning on light 1 (white):"
curl -X PUT -H "Content-Type: application/json" -d '{"on":true}' http://localhost:8080/api/testuser/lights/1/state

echo -e "\n\nTurning on light 2 (red):"
curl -X PUT -H "Content-Type: application/json" -d '{"on":true, "hue":0, "sat":254, "bri":200}' http://localhost:8080/api/testuser/lights/2/state

echo -e "\n\nTurning on light 3 (blue):"
curl -X PUT -H "Content-Type: application/json" -d '{"on":true, "hue":46920, "sat":254, "bri":150}' http://localhost:8080/api/testuser/lights/3/state

echo -e "\n\nChanging light 1 to green:"
curl -X PUT -H "Content-Type: application/json" -d '{"hue":21845, "sat":254, "bri":180}' http://localhost:8080/api/testuser/lights/1/state

echo -e "\n\nDimming light 2:"
curl -X PUT -H "Content-Type: application/json" -d '{"bri":50}' http://localhost:8080/api/testuser/lights/2/state

echo -e "\n\nTurning off light 3:"
curl -X PUT -H "Content-Type: application/json" -d '{"on":false}' http://localhost:8080/api/testuser/lights/3/state

echo -e "\n\nCheck the web interface at: http://localhost:9080"
echo "You should see the color changes in real-time!"
