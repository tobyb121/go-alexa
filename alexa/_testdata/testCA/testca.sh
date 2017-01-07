#!/bin/bash

echo "Generating Root CA..."
openssl req -new -newkey rsa:2048 -nodes \
 -x509 -days 36500 -subj "/CN=Test root CA" \
 -keyout root.key.pem -out root.crt.pem

echo "Generating Intermediate CA..."
openssl req -new -newkey rsa:2048 -nodes \
 -subj "/CN=Test Intermediate CA" \
 -keyout intermediate.key.pem -out intermediate.csr.pem

openssl x509 -req -in intermediate.csr.pem  \
 -extfile ext.cnf -extensions intermediate_ca \
 -set_serial 0001 -out intermediate.crt.pem \
 -CA root.crt.pem -CAkey root.key.pem -days 36500

echo "Generating Signing CA..."
openssl req -new -newkey rsa:2048 -nodes \
-subj "/CN=Test Signing CA" -keyout signing.key.pem \
-out signing.csr.pem

openssl x509 -req -in signing.csr.pem -extfile ext.cnf \
-extensions intermediate_ca -set_serial 0001 \
-out signing.crt.pem -CA intermediate.crt.pem \
-CAkey intermediate.key.pem -days 36500

echo "Generating Server Cert..."
openssl req -new -newkey rsa:2048 -nodes \
 -subj "/CN=Test Server" -keyout server.key.pem -out server.csr.pem

openssl x509 -req -in server.csr.pem -extfile ext.cnf \
 -extensions server_crt -set_serial 0001 -days 36500 \
 -out server.crt.pem -CA signing.crt.pem -CAkey signing.key.pem

cat server.crt.pem signing.crt.pem intermediate.crt.pem root.crt.pem > valid-chain.crt.pem
cat server.crt.pem signing.crt.pem intermediate.crt.pem > missing-root.crt.pem
cat server.crt.pem root.crt.pem > missing-intermediate.crt.pem

echo "Generating Expired Server Cert..."
openssl req -new -key server.key.pem \
 -subj "/CN=Expired Test Server" -out expired-server.csr.pem

openssl x509 -req -in expired-server.csr.pem -extfile ext.cnf \
 -extensions server_crt -set_serial 0001 -days -1 \
 -out expired-server.crt.pem -CA signing.crt.pem -CAkey signing.key.pem

cat expired-server.crt.pem signing.crt.pem intermediate.crt.pem root.crt.pem > expired-chain.crt.pem


echo "Generating Expired Intermediate CA..."
openssl req -new -newkey rsa:2048 -nodes \
 -subj "/CN=Test Intermediate CA" \
 -keyout expired-intermediate.key.pem -out expired-intermediate.csr.pem

openssl x509 -req -in expired-intermediate.csr.pem  \
 -extfile ext.cnf -extensions intermediate_ca \
 -set_serial 0001 -out expired-intermediate.crt.pem \
 -CA root.crt.pem -CAkey root.key.pem -days -1

echo "Generating Expired Chain Server Cert..."
openssl req -new -key server.key.pem \
 -subj "/CN=Expired Chain Test Server" -out expired-chain-server.csr.pem

openssl x509 -req -in expired-chain-server.csr.pem -extfile ext.cnf \
 -extensions server_crt -set_serial 0001 -days 36500 \
 -out expired-chain-server.crt.pem -CA expired-intermediate.crt.pem -CAkey expired-intermediate.key.pem

cat expired-chain-server.crt.pem expired-intermediate.crt.pem root.crt.pem > expired-intermediate-chain.crt.pem

rm *.csr.pem