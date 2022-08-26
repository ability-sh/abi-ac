#!/bin/sh

go mod tidy
go build -o abi-ac
chmod +x abi-ac
