#!/bin/bash

echo "🚀 Testing Enhanced Actuworry Features..."

SERVER="http://localhost:8080"

echo ""
echo "1. 🏥 Testing Health Check..."
curl -s "$SERVER/health" | jq '.'

echo ""
echo "2. 📊 Testing Immediate Annuity Calculation..."
curl -s -X POST "$SERVER/calculate" \
  -H "Content-Type: application/json" \
  -d '{
    "age": 65,
    "sum_assured": 12000,
    "interest_rate": 0.04,
    "table_name": "female",
    "product_type": "immediate_annuity"
  }' | jq '.'

echo ""
echo "3. ⏰ Testing Deferred Annuity Calculation..."
curl -s -X POST "$SERVER/calculate" \
  -H "Content-Type: application/json" \
  -d '{
    "age": 45,
    "sum_assured": 18000,
    "interest_rate": 0.05,
    "table_name": "male",
    "product_type": "deferred_annuity",
    "deferral_period": 20
  }' | jq '.'

echo ""
echo "4. 🚬 Testing Smoker Underwriting..."
curl -s -X POST "$SERVER/calculate" \
  -H "Content-Type: application/json" \
  -d '{
    "age": 40,
    "term": 15,
    "sum_assured": 200000,
    "interest_rate": 0.05,
    "table_name": "male",
    "product_type": "term_life",
    "smoker_status": "smoker",
    "health_rating": "standard"
  }' | jq '.'

echo ""
echo "5. ⭐ Testing Preferred Risk Underwriting..."
curl -s -X POST "$SERVER/calculate" \
  -H "Content-Type: application/json" \
  -d '{
    "age": 35,
    "term": 20,
    "sum_assured": 150000,
    "interest_rate": 0.05,
    "table_name": "female",
    "product_type": "whole_life",
    "smoker_status": "non_smoker",
    "health_rating": "preferred"
  }' | jq '.'

echo ""
echo "6. 📈 Testing Sensitivity Analysis..."
curl -s -X POST "$SERVER/calculate/sensitivity" \
  -H "Content-Type: application/json" \
  -d '{
    "base_policy": {
      "age": 35,
      "term": 10,
      "sum_assured": 100000,
      "interest_rate": 0.05,
      "table_name": "male",
      "product_type": "term_life"
    },
    "interest_rates": [0.03, 0.04, 0.05, 0.06, 0.07],
    "ages": [30, 35, 40, 45, 50]
  }' | jq '.analysis | keys'

echo ""
echo "7. 🎯 Testing Portfolio Analysis..."
curl -s -X POST "$SERVER/analyze/portfolio" \
  -H "Content-Type: application/json" \
  -d '{
    "policies": [
      {
        "age": 35,
        "term": 10,
        "sum_assured": 100000,
        "interest_rate": 0.05,
        "table_name": "male",
        "product_type": "term_life",
        "smoker_status": "non_smoker"
      },
      {
        "age": 42,
        "term": 20,
        "sum_assured": 200000,
        "interest_rate": 0.05,
        "table_name": "female",
        "product_type": "whole_life",
        "health_rating": "preferred"
      },
      {
        "age": 28,
        "term": 15,
        "sum_assured": 75000,
        "interest_rate": 0.045,
        "table_name": "male",
        "product_type": "term_life",
        "smoker_status": "smoker"
      }
    ]
  }' | jq '.'

echo ""
echo "8. 🔄 Testing Batch Calculation with Mixed Products..."
curl -s -X POST "$SERVER/calculate/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "policies": [
      {
        "age": 30,
        "term": 20,
        "sum_assured": 100000,
        "interest_rate": 0.05,
        "table_name": "male",
        "product_type": "term_life"
      },
      {
        "age": 40,
        "term": 15,
        "sum_assured": 150000,
        "interest_rate": 0.05,
        "table_name": "female",
        "product_type": "whole_life"
      },
      {
        "age": 65,
        "sum_assured": 12000,
        "interest_rate": 0.04,
        "table_name": "male",
        "product_type": "immediate_annuity"
      }
    ]
  }' | jq '.summary'

echo ""
echo "✅ All enhanced feature tests completed!"
echo ""
echo "🎉 Your Actuworry system now supports:"
echo "   • Term & Whole Life Insurance"
echo "   • Immediate & Deferred Annuities"
echo "   • Underwriting Factors (Smoker status, Health ratings)"
echo "   • Sensitivity Analysis"
echo "   • Portfolio Analytics"
echo "   • Business Intelligence Metrics"
