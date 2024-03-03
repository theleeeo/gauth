#!/bin/sh

# Create the keys directory
mkdir -p keys

# Generate the private key
openssl genpkey -algorithm Ed25519 --out keys/thor.pem

# Generate the public key
openssl pkey -in keys/thor.pem -pubout -out keys/thor.pub.pem