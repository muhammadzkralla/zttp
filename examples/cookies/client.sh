curl -i -X GET "localhost:8080/set"
echo "\n"

curl -i -X GET "localhost:8080/set" -H "Cookie: sessionId=abc123; username=john_doe"
echo "\n"

curl -i -X GET "localhost:8080/clear"
