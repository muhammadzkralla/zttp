curl -i -X GET "localhost:8080/accepts" -H "Accept: text/html, application/json; q=0.8, text/plain; q=0.5; charset=\"utf-8\""
echo "\n"

curl -i -X GET "localhost:8080/" -H "Accept-Charset: utf-8, iso-8859-1;q=0.2" -H "Accept-Encoding: gzip, compress;q=0.2" -H "Accept-Language: en;q=0.8, nl, ru"
echo "\n"

