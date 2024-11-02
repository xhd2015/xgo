#!/usr/bin/env bash
set -e
npm install
npm run build

go run ./gen/main.go
