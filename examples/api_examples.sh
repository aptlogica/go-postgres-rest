#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 GoPostgREST API Examples"
echo "================================="

echo "📊 Health Check"
curl -s "$BASE_URL/health" | jq

echo -e "\n📋 List Tables"
curl -s "$BASE_URL/schema/tables" | jq

echo -e "\n🛠️  Create Table"
curl -X POST "$BASE_URL/ddl/tables" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "employees",
    "columns": [
      {"name": "id", "data_type": "SERIAL", "not_null": true},
      {"name": "name", "data_type": "VARCHAR(255)", "not_null": true},
      {"name": "email", "data_type": "VARCHAR(255)", "unique": true},
      {"name": "department", "data_type": "VARCHAR(100)"},
      {"name": "salary", "data_type": "DECIMAL(10,2)"},
      {"name": "hire_date", "data_type": "DATE", "default_value": "CURRENT_DATE"},
      {"name": "active", "data_type": "BOOLEAN", "default_value": "true"}
    ],
    "primary_key": ["id"],
    "indexes": [
      {"columns": ["email"], "unique": true},
      {"columns": ["department"]}
    ]
  }' | jq

echo -e "\n📝 Insert Records"
curl -X POST "$BASE_URL/employees" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@company.com",
    "department": "Engineering",
    "salary": 75000
  }' | jq

curl -X POST "$BASE_URL/employees" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "email": "jane@company.com",
    "department": "Marketing",
    "salary": 65000
  }' | jq

echo -e "\n📊 Bulk Insert"
curl -X POST "$BASE_URL/bulk/employees/insert" \
  -H "Content-Type: application/json" \
  -d '[
    {"name": "Alice Johnson", "email": "alice@company.com", "department": "Engineering", "salary": 80000},
    {"name": "Bob Williams", "email": "bob@company.com", "department": "Sales", "salary": 60000},
    {"name": "Carol Brown", "email": "carol@company.com", "department": "HR", "salary": 55000}
  ]' | jq

echo -e "\n🔍 Simple Query"
curl -s "$BASE_URL/employees?department=Engineering" | jq

echo -e "\n🔎 Advanced Query with Aggregation"
curl -X POST "$BASE_URL/employees/query" \
  -H "Content-Type: application/json" \
  -d '{
    "aggregates": [
      {"function": "COUNT", "column": "*", "alias": "total_employees"},
      {"function": "AVG", "column": "salary", "alias": "avg_salary"},
      {"function": "MAX", "column": "salary", "alias": "max_salary"}
    ],
    "group_by": ["department"],
    "having": [
      {"column": "COUNT(*)", "operator": ">", "value": 1}
    ],
    "order_by": ["avg_salary DESC"]
  }' | jq

echo -e "\n🔗 Complex Query with Joins"
# First create a departments table
curl -X POST "$BASE_URL/ddl/tables" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "departments",
    "columns": [
      {"name": "id", "data_type": "SERIAL", "not_null": true},
      {"name": "name", "data_type": "VARCHAR(100)", "not_null": true},
      {"name": "budget", "data_type": "DECIMAL(12,2)"}
    ],
    "primary_key": ["id"]
  }'

# Insert department data
curl -X POST "$BASE_URL/bulk/departments/insert" \
  -H "Content-Type: application/json" \
  -d '[
    {"name": "Engineering", "budget": 500000},
    {"name": "Marketing", "budget": 200000},
    {"name": "Sales", "budget": 300000},
    {"name": "HR", "budget": 100000}
  ]'

# Query with join
curl -X POST "$BASE_URL/employees/query" \
  -H "Content-Type: application/json" \
  -d '{
    "select": ["e.name", "e.salary", "d.name as department_name", "d.budget"],
    "joins": [
      {
        "table": "departments",
        "alias": "d", 
        "type": "INNER",
        "on": "e.department = d.name"
      }
    ],
    "complex_filter": {
      "logic": "AND",
      "filters": [
        {"column": "e.active", "operator": "=", "value": true}
      ],
      "groups": [
        {
          "logic": "OR",
          "filters": [
            {"column": "e.salary", "operator": ">=", "value": 70000},
            {"column": "d.budget", "operator": ">", "value": 250000}
          ]
        }
      ]
    },
    "order_by": ["e.salary DESC"]
  }' | jq

echo -e "\n🔍 Full-Text Search"
curl -X POST "$BASE_URL/employees/query" \
  -H "Content-Type: application/json" \
  -d '{
    "full_text": {
      "query": "john engineer",
      "columns": ["name", "department"],
      "type": "websearch"
    }
  }' | jq

echo -e "\n📊 Range Query"
curl -X POST "$BASE_URL/employees/query" \
  -H "Content-Type: application/json" \
  -d '{
    "range": {
      "column": "salary",
      "from": 60000,
      "to": 80000
    }
  }' | jq

echo -e "\n📈 Analytics"
curl -s "$BASE_URL/analytics/database" | jq
curl -s "$BASE_URL/analytics/tables/employees" | jq

echo -e "\n💾 Export to CSV"
curl -s "$BASE_URL/export/employees/csv" > employees.csv
echo "Exported to employees.csv"

echo -e "\n📄 Export Schema"
curl -s "$BASE_URL/export/schema?format=sql" > schema.sql
echo "Schema exported to schema.sql"

echo -e "\n✨ All examples completed!"
