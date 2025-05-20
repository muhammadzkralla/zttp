curl -i -X GET "localhost:8080/set/header"
echo "\n"
curl -i -X GET "localhost:8080/get/header" -H "Header1: header1" -H "Header2: header2"
echo "\n"
