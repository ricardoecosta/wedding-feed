#!/usr/bin/env bash
GOOS=linux GOARCH=amd64 go build -o wedding-feed *.go || exit 0
chmod +x wedding-feed
rev=$(git rev-parse HEAD) || exit 0
zip -r build-$rev.zip wedding-feed config.json Procfile static/**
