# Swagger Documentation Guide

## 📚 Overview

Smart Monitor sử dụng hệ thống Swagger động cho phép dễ dàng cập nhật và mở rộng API documentation.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Swagger Architecture                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐                                        │
│  │  generate-      │  Generate                              │
│  │  swagger.sh     │ ─────────► swagger.json                │
│  └─────────────────┘            (Complete API Spec)         │
│                                                              │
│  ┌─────────────────┐                                        │
│  │  Backend        │  Serve                                 │
│  │  Server         │ ─────────► /v1/swagger.json            │
│  └─────────────────┘            /swagger/ (UI)              │
│                                                              │
│  ┌─────────────────┐                                        │
│  │  Swagger UI     │  Display                               │
│  │  (Browser)      │ ─────────► Interactive Documentation   │
│  └─────────────────┘                                        │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Generation Script

Script `scripts/generate-swagger.sh` tự động tạo swagger documentation:

```bash
cd /path/to/smart-monitor
chmod +x scripts/generate-swagger.sh
./scripts/generate-swagger.sh
```

### Output Files

```
backend/static/
├── swagger.json        # Main OpenAPI specification
├── api-docs.json       # Alternative copy
└── swagger-ui.html     # Custom Swagger UI
```

## 📝 Adding New Endpoints

### 1. Edit generate-swagger.sh

Thêm endpoint vào section `paths`:

```json
"/v1/your-endpoint": {
  "post": {
    "tags": ["Your Tag"],
    "summary": "Your endpoint summary",
    "description": "Detailed description",
    "operationId": "yourOperation",
    "parameters": [...],
    "responses": {
      "200": {
        "description": "Success",
        "schema": {
          "$ref": "#/definitions/YourResponse"
        },
        "examples": {
          "application/json": {
            "key": "value"
          }
        }
      }
    }
  }
}
```

### 2. Add Data Models

Thêm model vào section `definitions`:

```json
"YourResponse": {
  "type": "object",
  "properties": {
    "field1": {
      "type": "string",
      "example": "example value"
    },
    "field2": {
      "type": "integer",
      "example": 123
    }
  }
}
```

### 3. Regenerate Documentation

```bash
./scripts/generate-swagger.sh
```

## 🎨 Customization

### Modify UI Settings

Edit `backend/static/swagger-ui.html`:

```javascript
const ui = SwaggerUIBundle({
    url: "/v1/swagger.json",
    docExpansion: "list",      // "none", "list", "full"
    filter: true,               // Enable search filter
    tryItOutEnabled: true,      // Enable "Try it out"
    defaultModelsExpandDepth: 1,
    // ... more options
});
```

### Change Theme

Add custom CSS in `swagger-ui.html`:

```css
<style>
    .topbar {
        background-color: #your-color;
    }
    .swagger-ui .info .title {
        color: #your-color;
    }
</style>
```

## 📋 Best Practices

### 1. Complete Examples

Luôn cung cấp examples đầy đủ:

```json
"examples": {
  "application/json": {
    "field1": "realistic value",
    "field2": 123,
    "nested": {
      "key": "value"
    }
  }
}
```

### 2. Error Responses

Document tất cả error cases:

```json
"responses": {
  "200": { "description": "Success" },
  "400": { "description": "Bad request" },
  "401": { "description": "Unauthorized" },
  "404": { "description": "Not found" },
  "500": { "description": "Server error" }
}
```

### 3. Security Requirements

Chỉ rõ authentication:

```json
"security": [
  {"ApiKeyAuth": []},
  {"BearerAuth": []}
]
```

### 4. Tags Organization

Group endpoints logically:

```json
"tags": [
  {
    "name": "Agent Management",
    "description": "Agent operations",
    "externalDocs": {
      "description": "Setup guide",
      "url": "/docs/AGENT_SETUP.md"
    }
  }
]
```

## 🔍 Validation

### Test Swagger JSON

```bash
# Check JSON syntax
cat backend/static/swagger.json | jq '.'

# Validate with swagger-cli
npm install -g @apidevtools/swagger-cli
swagger-cli validate backend/static/swagger.json
```

### Test Endpoints

```bash
# Start backend
cd backend
go run cmd/server/main.go

# Access Swagger UI
open http://localhost:8080/swagger/swagger-ui.html

# Download JSON
curl http://localhost:8080/v1/swagger.json
```

## 🚀 Advanced Features

### Auto-Generate from Code

Có thể integrate với code để auto-generate:

```go
// Future: Add swagger annotations to handlers
// @Summary Register agent
// @Description Register new monitoring agent
// @Tags Agent
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Registration data"
// @Success 200 {object} RegisterResponse
// @Router /v1/agent/register [post]
func (h *Handler) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    // Implementation
}
```

### Versioning

Maintain multiple API versions:

```bash
scripts/
├── generate-swagger.sh       # Current version
├── generate-swagger-v1.sh    # Version 1
└── generate-swagger-v2.sh    # Version 2
```

### CI/CD Integration

Add to GitHub Actions:

```yaml
- name: Generate Swagger
  run: |
    chmod +x scripts/generate-swagger.sh
    ./scripts/generate-swagger.sh
    
- name: Validate Swagger
  run: |
    swagger-cli validate backend/static/swagger.json
```

## 📊 Monitoring

### Track Usage

Add analytics to Swagger UI:

```javascript
requestInterceptor: function(request) {
    console.log("API Call:", request.method, request.url);
    // Send to analytics
    return request;
}
```

### API Metrics

Monitor which endpoints are most used:

```go
// Add middleware to track API calls
func swaggerMetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/v1/") {
            log.Printf("Swagger UI accessed: %s", r.URL.Path)
        }
        next.ServeHTTP(w, r)
    })
}
```

## 🔄 Migration from combined.swagger.json

### Why Change?

**Old Approach** (combined.swagger.json):
- ❌ Generated từ protobuf - khó customize
- ❌ Phải regenerate khi thay đổi proto
- ❌ Limited control over examples
- ❌ Merge nhiều services phức tạp

**New Approach** (generate-swagger.sh):
- ✅ Full control over content
- ✅ Easy to add/modify endpoints
- ✅ Rich examples and descriptions
- ✅ Independent of proto changes
- ✅ Flexible merging strategy

### Migration Steps

1. **Backup old file**
   ```bash
   cp pbtypes/combined.swagger.json pbtypes/combined.swagger.json.bak
   ```

2. **Generate new swagger**
   ```bash
   ./scripts/generate-swagger.sh
   ```

3. **Update backend**
   - Point to new swagger.json
   - Remove old combined.swagger.json references

4. **Test thoroughly**
   ```bash
   # Start backend
   cd backend && go run cmd/server/main.go
   
   # Check UI
   open http://localhost:8080/swagger/swagger-ui.html
   
   # Verify all endpoints
   curl http://localhost:8080/v1/swagger.json | jq '.paths | keys'
   ```

## 📚 Resources

### Swagger/OpenAPI Specification
- [OpenAPI 2.0 Spec](https://swagger.io/specification/v2/)
- [Swagger UI Configuration](https://swagger.io/docs/open-source-tools/swagger-ui/usage/configuration/)

### Tools
- [Swagger Editor](https://editor.swagger.io/) - Online editor
- [Swagger CLI](https://www.npmjs.com/package/@apidevtools/swagger-cli) - Validation tool
- [OpenAPI Generator](https://openapi-generator.tech/) - Code generation

### Examples
- [Petstore Example](https://petstore.swagger.io/)
- [Best Practices](https://swagger.io/docs/specification/2-0/basic-structure/)

## 💡 Tips

1. **Version Control**: Commit swagger.json to Git
2. **Review Changes**: Use `git diff` to see API changes
3. **Documentation First**: Update swagger before coding
4. **Test Examples**: Make sure examples work in "Try it out"
5. **Keep Updated**: Regenerate after adding features
6. **Use Templates**: Create templates for common patterns
7. **Automation**: Add pre-commit hooks to validate swagger

## 🎯 Quick Commands

```bash
# Generate swagger
./scripts/generate-swagger.sh

# Validate swagger
swagger-cli validate backend/static/swagger.json

# View JSON pretty
cat backend/static/swagger.json | jq '.'

# Check specific path
cat backend/static/swagger.json | jq '.paths."/v1/agent/register"'

# List all endpoints
cat backend/static/swagger.json | jq '.paths | keys'

# Check definitions
cat backend/static/swagger.json | jq '.definitions | keys'

# Start backend with swagger
cd backend && go run cmd/server/main.go

# Test with curl
curl http://localhost:8080/v1/swagger.json | jq '.info'
```

## 📞 Support

Nếu cần hỗ trợ:
1. Check logs: `tail -f backend/logs/app.log`
2. Validate JSON: `swagger-cli validate backend/static/swagger.json`
3. Review documentation: [docs/API.md](./API.md)
4. Check examples in swagger UI: http://localhost:8080/swagger/
