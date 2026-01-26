# Smart Monitor - API Documentation System

## 🎯 Giải pháp mới: Dynamic Swagger Generation

### ❌ Vấn đề với combined.swagger.json cũ

1. **Thiếu linh động**: Merge từ nhiều proto files → khó customize
2. **Examples nghèo nàn**: Auto-generated → không có context
3. **Phụ thuộc protobuf**: Phải regenerate mỗi lần thay đổi proto
4. **Khó maintain**: Không thể dễ dàng thêm/sửa documentation

### ✅ Giải pháp mới: Script-based Generation

```
┌────────────────────────────────────────────────────────┐
│             Dynamic Swagger Architecture                │
├────────────────────────────────────────────────────────┤
│                                                         │
│  Developer                                              │
│      │                                                  │
│      ├─► Edit generate-swagger.sh                      │
│      │   (Add endpoints, examples, docs)               │
│      │                                                  │
│      ├─► Run ./scripts/generate-swagger.sh             │
│      │   • Generate swagger.json                       │
│      │   • Copy to api-docs.json                       │
│      │   • Validate structure                          │
│      │                                                  │
│      └─► Backend serves:                               │
│          • /v1/swagger.json                            │
│          • /swagger/ (Swagger UI)                      │
│          • /api/docs (alternative)                     │
│                                                         │
│  Benefits:                                              │
│  ✅ Full control over content                          │
│  ✅ Rich examples & descriptions                       │
│  ✅ Easy to update                                     │
│  ✅ Independent of proto changes                       │
│  ✅ Version control friendly                           │
└────────────────────────────────────────────────────────┘
```

## 📁 Files Structure

```
smart-monitor/
├── scripts/
│   └── generate-swagger.sh        # Main generation script
│
├── backend/
│   ├── static/
│   │   ├── swagger.json           # Generated OpenAPI spec
│   │   ├── api-docs.json          # Copy for compatibility
│   │   └── swagger-ui.html        # Custom Swagger UI
│   │
│   └── cmd/server/main.go         # Serves swagger endpoints
│
└── docs/
    ├── SWAGGER_GUIDE.md           # This file
    └── API.md                     # API reference
```

## 🚀 Quick Start

### 1. Generate Swagger

```bash
cd /path/to/smart-monitor
./scripts/generate-swagger.sh
```

Output:
```
🔄 Generating Swagger Documentation...
✅ Base swagger generated: backend/static/swagger.json
✅ Copied to: backend/static/api-docs.json
📊 File size: 20K

🎉 Swagger documentation generated successfully!
📁 Location: backend/static/swagger.json
🌐 Access at: http://localhost:8080/swagger/
```

### 2. Start Backend

```bash
cd backend
go run cmd/server/main.go
```

### 3. Access Swagger UI

Open browser: http://localhost:8080/swagger/

### 4. Test API

Use "Try it out" trong Swagger UI hoặc:

```bash
# Register agent
curl -X POST http://localhost:8080/v1/agent/register \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "server-01",
    "ip_address": "192.168.1.10",
    "agent_version": "1.0.0",
    "metadata": {
      "location": "datacenter-01",
      "environment": "production"
    }
  }'

# Stream metrics (requires token from registration)
curl -X POST http://localhost:8080/v1/stats/stream \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "server-01",
    "agent_id": "agent-a3f5c2d1",
    "access_token": "your-token-here",
    "cpu": 45.2,
    "ram": 68.5,
    "disk": 72.3
  }'
```

## ➕ Adding New Endpoints

### Step 1: Edit generate-swagger.sh

Tìm section `"paths": {` và thêm endpoint mới:

```bash
vim scripts/generate-swagger.sh
```

```json
"/v1/alerts/create": {
  "post": {
    "tags": ["Alerts"],
    "summary": "Create alert rule",
    "description": "Create new alert rule for monitoring metrics",
    "operationId": "createAlert",
    "consumes": ["application/json"],
    "produces": ["application/json"],
    "parameters": [
      {
        "name": "body",
        "in": "body",
        "required": true,
        "schema": {
          "$ref": "#/definitions/CreateAlertRequest"
        }
      }
    ],
    "responses": {
      "200": {
        "description": "Alert created",
        "schema": {
          "$ref": "#/definitions/CreateAlertResponse"
        },
        "examples": {
          "application/json": {
            "success": true,
            "alert_id": "alert-123",
            "message": "Alert rule created successfully"
          }
        }
      },
      "400": {
        "description": "Invalid request",
        "schema": {
          "$ref": "#/definitions/ErrorResponse"
        }
      }
    }
  }
}
```

### Step 2: Add Data Models

Tìm section `"definitions": {` và thêm model:

```json
"CreateAlertRequest": {
  "type": "object",
  "required": ["name", "metric", "threshold"],
  "properties": {
    "name": {
      "type": "string",
      "description": "Alert rule name",
      "example": "High CPU Alert"
    },
    "metric": {
      "type": "string",
      "enum": ["cpu", "ram", "disk"],
      "example": "cpu"
    },
    "threshold": {
      "type": "number",
      "description": "Threshold value (0-100)",
      "example": 80.0
    },
    "duration": {
      "type": "integer",
      "description": "Duration in seconds",
      "example": 300
    }
  }
},
"CreateAlertResponse": {
  "type": "object",
  "properties": {
    "success": {
      "type": "boolean",
      "example": true
    },
    "alert_id": {
      "type": "string",
      "example": "alert-123"
    },
    "message": {
      "type": "string",
      "example": "Alert rule created successfully"
    }
  }
}
```

### Step 3: Regenerate

```bash
./scripts/generate-swagger.sh
```

### Step 4: Verify

```bash
# Check endpoint added
cat backend/static/swagger.json | jq '.paths."/v1/alerts/create"'

# Check model added
cat backend/static/swagger.json | jq '.definitions.CreateAlertRequest'

# Restart backend và test trong Swagger UI
```

## 🔄 Update Existing Endpoints

### Modify Examples

```bash
vim scripts/generate-swagger.sh
```

Tìm endpoint và update examples:

```json
"/v1/agent/register": {
  "post": {
    // ... existing config ...
    "responses": {
      "200": {
        "examples": {
          "application/json": {
            "success": true,
            "message": "Agent registered successfully",
            "agent_id": "agent-new-id-format",
            "access_token": "new-token-format",
            "expires_at": 1737849600,
            "additional_field": "new value"  // New field
          }
        }
      }
    }
  }
}
```

Regenerate:
```bash
./scripts/generate-swagger.sh
```

## 🎨 Customization

### Change API Info

Edit trong `generate-swagger.sh`:

```json
"info": {
  "title": "Your Custom Title",
  "description": "Your custom description\n\nWith markdown support",
  "version": "2.0.0",
  "contact": {
    "name": "Your Team",
    "email": "your-email@example.com"
  }
}
```

### Add New Tags

```json
"tags": [
  {
    "name": "Alerts",
    "description": "Alert management endpoints",
    "externalDocs": {
      "description": "Alert Guide",
      "url": "/docs/ALERTS.md"
    }
  }
]
```

### Modify Security

```json
"securityDefinitions": {
  "ApiKeyAuth": {
    "type": "apiKey",
    "name": "X-API-Key",
    "in": "header"
  }
}
```

## 📊 Comparison

### Old vs New Approach

| Aspect | combined.swagger.json | generate-swagger.sh |
|--------|----------------------|---------------------|
| **Source** | Auto from protobuf | Manual script |
| **Control** | ❌ Limited | ✅ Full control |
| **Examples** | ❌ Basic | ✅ Rich & realistic |
| **Maintenance** | ❌ Hard | ✅ Easy |
| **Updates** | ❌ Regenerate proto | ✅ Edit script |
| **Flexibility** | ❌ Low | ✅ High |
| **Documentation** | ❌ Minimal | ✅ Comprehensive |
| **Version Control** | ❌ Binary diffs | ✅ Text diffs |

### Why Switch?

```
Old Way (Proto → Swagger):
─────────────────────────────────────────
Proto File → protoc → Generated Swagger
   ↓
❌ Limited examples
❌ Generic descriptions  
❌ Hard to customize
❌ Coupling with proto structure


New Way (Script → Swagger):
─────────────────────────────────────────
Script → Generate → Custom Swagger
   ↓
✅ Rich examples
✅ Detailed descriptions
✅ Easy customization
✅ Independent of proto
✅ Full documentation control
```

## 🔍 Validation

### Check JSON Syntax

```bash
cat backend/static/swagger.json | jq '.'
```

### Validate OpenAPI

```bash
npm install -g @apidevtools/swagger-cli
swagger-cli validate backend/static/swagger.json
```

### Lint Swagger

```bash
npm install -g @stoplight/spectral-cli
spectral lint backend/static/swagger.json
```

## 🧪 Testing

### Test Generation

```bash
# Run generation
./scripts/generate-swagger.sh

# Check output
ls -lh backend/static/swagger.json

# Verify content
cat backend/static/swagger.json | jq '.info'
cat backend/static/swagger.json | jq '.paths | keys'
```

### Test Endpoints

```bash
# Start backend
cd backend && go run cmd/server/main.go &

# Wait for startup
sleep 2

# Test swagger.json endpoint
curl http://localhost:8080/v1/swagger.json | jq '.info'

# Test Swagger UI
curl -I http://localhost:8080/swagger/

# Stop backend
pkill -f "go run"
```

### Test in Browser

1. Start backend: `cd backend && go run cmd/server/main.go`
2. Open: http://localhost:8080/swagger/
3. Verify:
   - ✅ All endpoints visible
   - ✅ Examples display correctly
   - ✅ "Try it out" works
   - ✅ Authentication fields present
   - ✅ Models documented

## 📝 Best Practices

### 1. Complete Examples

Always provide realistic, working examples:

```json
"examples": {
  "application/json": {
    "hostname": "prod-server-01",
    "cpu": 45.2,
    "ram": 68.5,
    "timestamp": 1737882600,
    "metadata": {
      "datacenter": "us-east-1",
      "environment": "production"
    }
  }
}
```

### 2. Error Documentation

Document all possible errors:

```json
"responses": {
  "200": { "description": "Success" },
  "400": { 
    "description": "Bad request - Invalid input",
    "examples": {
      "application/json": {
        "error": "Invalid CPU value",
        "code": "VALIDATION_ERROR"
      }
    }
  },
  "401": { "description": "Unauthorized - Invalid token" },
  "404": { "description": "Not found - Host doesn't exist" },
  "500": { "description": "Internal server error" }
}
```

### 3. Field Descriptions

Add helpful descriptions:

```json
"cpu": {
  "type": "number",
  "format": "double",
  "minimum": 0,
  "maximum": 100,
  "description": "CPU usage percentage (0-100). Calculated as average across all cores.",
  "example": 45.2
}
```

### 4. Operation IDs

Use consistent naming:

```json
"operationId": "registerAgent"      // ✅ camelCase
"operationId": "RegisterAgent"      // ❌ PascalCase
"operationId": "register_agent"     // ❌ snake_case
```

### 5. Tags Organization

Group logically:

```json
"tags": [
  "Agent Management",    // Registration, lifecycle
  "Metrics",            // Stats, monitoring data
  "Health",             // Health checks
  "Admin"               // Administrative endpoints
]
```

## 🚀 Advanced Usage

### Multi-Version Support

```bash
# Create versioned generators
cp scripts/generate-swagger.sh scripts/generate-swagger-v1.sh
cp scripts/generate-swagger.sh scripts/generate-swagger-v2.sh

# Generate both
./scripts/generate-swagger-v1.sh  # → swagger-v1.json
./scripts/generate-swagger-v2.sh  # → swagger-v2.json

# Serve both versions
httpMux.HandleFunc("/v1/swagger.json", serveSwaggerV1)
httpMux.HandleFunc("/v2/swagger.json", serveSwaggerV2)
```

### Auto-Generation in CI/CD

```yaml
# .github/workflows/swagger.yml
name: Generate Swagger

on:
  push:
    paths:
      - 'scripts/generate-swagger.sh'
      - 'backend/**/*.go'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Generate Swagger
        run: |
          chmod +x scripts/generate-swagger.sh
          ./scripts/generate-swagger.sh
      
      - name: Validate Swagger
        run: |
          npm install -g @apidevtools/swagger-cli
          swagger-cli validate backend/static/swagger.json
      
      - name: Commit changes
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add backend/static/swagger.json
          git commit -m "Update swagger documentation" || true
          git push
```

### Merge Multiple Sources

```bash
#!/bin/bash
# Advanced: Merge multiple swagger sources

# Generate base
./scripts/generate-swagger.sh

# Merge additional APIs
jq -s '.[0] * .[1]' \
  backend/static/swagger.json \
  additional-api.json \
  > backend/static/merged-swagger.json

mv backend/static/merged-swagger.json backend/static/swagger.json
```

## 📚 Resources

- **OpenAPI 2.0 Spec**: https://swagger.io/specification/v2/
- **Swagger UI**: https://swagger.io/tools/swagger-ui/
- **Swagger Editor**: https://editor.swagger.io/
- **API Best Practices**: https://swagger.io/resources/articles/best-practices-in-api-documentation/

## 🎯 Next Steps

1. ✅ **Generated**: Dynamic swagger documentation system
2. ✅ **Created**: Easy-to-update generation script
3. ✅ **Documented**: Complete guide in SWAGGER_GUIDE.md
4. 🔄 **Next**: Test và customize theo nhu cầu
5. 🔄 **Future**: Add more endpoints as features grow

## 💡 Tips

- **Version Control**: Always commit swagger.json
- **Review Diffs**: Check `git diff` before committing
- **Test Examples**: Verify examples work in "Try it out"
- **Keep Updated**: Regenerate when adding features
- **Document Changes**: Add comments in generation script
- **Use Templates**: Create snippets for common patterns

## 🆘 Troubleshooting

### Issue: Swagger UI not loading

```bash
# Check file exists
ls -lh backend/static/swagger-ui.html

# Check content
cat backend/static/swagger.json | jq '.info'

# Check backend logs
tail -f backend/logs/app.log
```

### Issue: Examples not showing

```bash
# Validate structure
cat backend/static/swagger.json | jq '.paths."/v1/agent/register".post.responses."200".examples'

# Check JSON syntax
jq '.' backend/static/swagger.json
```

### Issue: Endpoint not appearing

```bash
# Check paths section
cat backend/static/swagger.json | jq '.paths | keys'

# Regenerate
./scripts/generate-swagger.sh

# Restart backend
pkill -f "server/main.go" && cd backend && go run cmd/server/main.go &
```

---

**Tổng kết**: Hệ thống swagger mới cung cấp tính linh động cao, dễ dàng cập nhật và maintain. Thay vì phụ thuộc vào protobuf generation, giờ bạn có full control và có thể nhanh chóng thêm/sửa documentation theo nhu cầu!
