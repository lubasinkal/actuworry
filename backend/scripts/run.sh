#!/bin/bash

# Actuworry - Life Insurance Tool
# Simple startup script

echo "🇧🇼 Starting Actuworry - Life Insurance Tool..."
echo ""
echo "Loading mortality tables and starting server..."
echo ""

# Run the Go backend from the project root
# Use the new cmd/server structure
go run backend/cmd/server/main.go
