#!/bin/bash

echo "🧪 Testing Enhanced Actuworry API..."

echo ""
echo "1. Testing Health Check..."
curl -s http://localhost:8080/health | jq '.'

echo ""
echo "2. Testing Available Tables..."
curl -s http://localhost:8080/tables | jq '.'

echo ""
echo "3. Testing Term Life Insurance Calculation..."
curl -s -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d @request.json | jq '.'

echo ""
echo "4. Testing Whole Life Insurance Calculation..."
curl -s -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d @test_whole_life.json | jq '.'

echo ""
echo "5. Testing Batch Calculation..."
curl -s -X POST http://localhost:8080/calculate/batch \
  -H "Content-Type: application/json" \
  -d @test_batch.json | jq '.'

echo ""
echo "✅ All tests completed!"
