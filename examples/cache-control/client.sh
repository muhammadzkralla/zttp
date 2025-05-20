curl -i -X GET "localhost:8080/fresh"
echo "\n"

curl -i -X GET "localhost:8080/fresh" -H 'If-None-Match: version2'
echo "\n"

curl -i -X GET "localhost:8080/fresh" -H 'If-None-Match: version1'
echo "\n"


curl -i -X GET "localhost:8080/no-cache-example"
echo "\n"

curl -i -X GET "localhost:8080/no-cache-example" -H 'If-None-Match: static-version-123' -H 'Cache-Control: no-cache'
echo "\n"

curl -i -X GET "localhost:8080/no-cache-example" -H 'If-None-Match: static-version-123'
echo "\n"


# WARNING: YOU SHOULD ADJUST DATES ACCORDINGLY FOR THIS SECTION
# THIS SCRIPT WAS CREATED AT Tue, 20 May 2025 00:00:00 GMT


# 200
curl -i -X GET "localhost:8080/last-modified"
echo "\n"

# 304
curl -i -X GET "localhost:8080/last-modified" -H 'If-None-Match: version1'
echo "\n"

# Same date, 304, ETag matching override date checking
curl -i -X GET "localhost:8080/last-modified" -H 'If-None-Match: version1' -H "If-Modified-Since: Tue, 20 May 2025 00:00:00 GMT"
echo "\n"

# Before date, 304, ETag matching override date checking
curl -i -X GET "localhost:8080/last-modified" -H 'If-None-Match: version1' -H "If-Modified-Since: Mon, 19 May 2025 00:00:00 GMT"
echo "\n"

# After date, 304, ETag matching override date checking
curl -i -X GET "localhost:8080/last-modified" -H 'If-None-Match: version1' -H "If-Modified-Since: Wed, 21 May 2025 00:00:00 GMT"
echo "\n"

# Same date, 304, date matched
curl -i -X GET "localhost:8080/last-modified" -H "If-Modified-Since: Tue, 20 May 2025 00:00:00 GMT"
echo "\n"

# After date, 304, resouce didn't change after this date
curl -i -X GET "localhost:8080/last-modified" -H "If-Modified-Since: Wed, 21 May 2025 00:00:00 GMT"
echo "\n"

# Before date, 200, resource changed after this date 
curl -i -X GET "localhost:8080/last-modified" -H "If-Modified-Since: Mon, 19 May 2025 00:00:00 GMT"
echo "\n"
