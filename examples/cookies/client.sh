curl -i -X GET "localhost:8080/"
echo "\n"

curl -i -X GET "localhost:8080/" -H "Cookie: sessionId=abc123; username=john_doe"
