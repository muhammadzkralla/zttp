curl -i -X GET "localhost:8080/home"
echo "\n"
curl -i -X GET "localhost:8080/nothome"
echo "\n"

curl -i -X DELETE "localhost:8080/home"
echo "\n"
curl -i -X DELETE "localhost:8080/nothome"
echo "\n"

curl -i -X POST "localhost:8080/home" -d "test"
echo "\n"
curl -i -X POST "localhost:8080/nothome" -d "test"
echo "\n"

curl -i -X PUT "localhost:8080/home" -d "test"
echo "\n"
curl -i -X PUT "localhost:8080/nothome" -d "test"
echo "\n"

curl -i -X PATCH "localhost:8080/home" -d "test"
echo "\n"
curl -i -X PATCH "localhost:8080/nothome" -d "test"
echo "\n"
