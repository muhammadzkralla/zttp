openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

curl -i -k -X GET "https://localhost:8080/tls"
