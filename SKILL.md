# SKILL — Veeam Terraform Provider

## Purpose
This file contains implementation patterns, code examples, and domain knowledge
for building and maintaining the Veeam V13 Terraform Provider.

Read this before implementing any resource, data source, or client change.

---

## 1. Authentication Pattern (V13 OAuth2)

The Veeam V13 REST API uses OAuth2 with password grant. The token endpoint
is `/api/oauth2/token` and accepts `application/x-www-form-urlencoded`.

### Authenticate
```go
// POST /api/oauth2/token
// Content-Type: application/x-www-form-urlencoded
//
// grant_type=password&username=DOMAIN\admin&password=secret
```

### Refresh Token
```go
// POST /api/oauth2/token
// Content-Type: application/x-www-form-urlencoded
//
// grant_type=refresh_token&refresh_token=<refresh_token>
```

### Token Response (TokenModel)
```json
{
  "access_token": "...",
  "token_type": "bearer",
  "refresh_token": "...",
  "expires_in": 900,
  ".issued": "2024-01-01T00:00:00Z",
  ".expires": "2024-01-01T00:15:00Z"
}
```

**Key rules:**
- Access token lifetime: 15 minutes
- Refresh token lifetime: 14 days (default), single-use
- After refresh, you get a NEW refresh token
- All API requests need `Authorization: Bearer <access_token>`
- All API requests need `x-api-version: 1.3-rev0`

---

## 2. APIClient Interface Pattern

Resources and data sources depend on the `APIClient` interface, not the
concrete `VeeamClient`. This enables testability with mock implementations.

```go
// internal/client/interface.go
type APIClient interface {
    GetJSON(ctx context.Context, endpoint string, result interface{}) error
    PostJSON(ctx context.Context, endpoint string, payload, result interface{}) error
    PutJSON(ctx context.Context, endpoint string, payload, result interface{}) error
    DeleteJSON(ctx context.Context, endpoint string) error
    WaitForTask(ctx context.Context, sessionID string) error
}
```

---

## 3. Terraform Resource Pattern

Every resource follows this structure:

```go
type MyResource struct {
    client client.APIClient  // interface, not concrete
}

// TF model (what Terraform sees)
type MyResourceModel struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
    // ... terraform types only
}

// API model (what the Veeam API sees) — in internal/models/
type MyAPIModel struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    // ... Go native types for JSON serialization
}

func (r *MyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan MyResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }

    // Convert TF model → API model
    apiPayload := MyAPISpec{ Name: plan.Name.ValueString() }

    // Call API
    var apiResult MyAPIModel
    if err := r.client.PostJSON(ctx, "/api/v1/something", apiPayload, &apiResult); err != nil {
        resp.Diagnostics.AddError("Error creating resource", err.Error())
        return
    }

    // Convert API model → TF model
    plan.ID = types.StringValue(apiResult.ID)

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

---

## 4. Async Operation Pattern

Many V13 operations are async. The response is `202 Accepted` with a session ID
in the response body or `Location` header.

```go
func (c *VeeamClient) WaitForTask(ctx context.Context, sessionID string) error {
    for {
        var session SessionModel
        err := c.GetJSON(ctx, "/api/v1/sessions/"+sessionID, &session)
        if err != nil { return err }

        switch session.State {
        case "Stopped":
            if session.Result == "Success" { return nil }
            return fmt.Errorf("task failed: %s", session.Result)
        case "Working":
            // poll again
        default:
            // poll again
        }

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(5 * time.Second):
        }
    }
}
```

---

## 5. Polymorphic Type Pattern

V13 uses `oneOf` with `discriminator` on the `type` field. Example:
- `RepositoryModel` can be `WindowsLocalStorageModel`, `LinuxLocalStorageModel`, etc.
- The `type` field (`"WinLocal"`, `"LinuxLocal"`, etc.) determines the subtype.

**In Go, use embedding + custom JSON unmarshal:**
```go
type RepositoryBase struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Type        string `json:"type"`
}
```

For Terraform schemas, use nested blocks that are type-specific, or a flat
schema with validators that enforce which fields are required per type.

---

## 6. Security Patterns

### Sensitive fields in TF schema
```go
"password": schema.StringAttribute{
    Required:  true,
    Sensitive: true,
    Description: "Password for authentication.",
},
```

### Redact in API models
```go
type CredentialsSpec struct {
    Username string `json:"username"`
    Password string `json:"password"` // sent to API, but never logged
}

// Custom String() to prevent accidental logging
func (c CredentialsSpec) String() string {
    return fmt.Sprintf("CredentialsSpec{Username: %s, Password: [REDACTED]}", c.Username)
}
```

### Environment variable loading
```go
host := data.Host.ValueString()
if host == "" {
    host = os.Getenv("VEEAM_HOST")
}
if host == "" {
    resp.Diagnostics.AddError("Missing Host", "Set host in config or VEEAM_HOST env var")
    return
}
```

---

## 7. Error Handling Pattern

V13 API returns errors in this format:
```json
{
  "errorCode": "InvalidArgument",
  "message": "Repository name already exists",
  "details": "A repository with name 'Default Backup Repository' already exists."
}
```

Parse this in the client and return actionable errors:
```go
type APIError struct {
    ErrorCode string `json:"errorCode"`
    Message   string `json:"message"`
    Details   string `json:"details"`
}

func (e *APIError) Error() string {
    if e.Details != "" {
        return fmt.Sprintf("%s: %s (%s)", e.ErrorCode, e.Message, e.Details)
    }
    return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
}
```

---

## 8. Testing Patterns

### Mock client for resource tests
```go
type MockAPIClient struct {
    mock.Mock
}

func (m *MockAPIClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
    args := m.Called(ctx, endpoint, result)
    if fn, ok := args.Get(0).(func(interface{})); ok {
        fn(result)
    }
    return args.Error(1)
}
```

### httptest for client tests
```go
func TestAuthenticate(t *testing.T) {
    server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/api/oauth2/token", r.URL.Path)
        assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

        body, _ := io.ReadAll(r.Body)
        values, _ := url.ParseQuery(string(body))
        assert.Equal(t, "password", values.Get("grant_type"))

        json.NewEncoder(w).Encode(map[string]interface{}{
            "access_token":  "test-token",
            "refresh_token": "test-refresh",
            "token_type":    "bearer",
            "expires_in":    900,
        })
    }))
    defer server.Close()
    // ... create client with server.URL, test auth
}
```

---

## 9. Veeam V13 API Quick Reference

### Endpoints used by this provider
| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| Credentials | POST /api/v1/credentials | GET /api/v1/credentials/{id} | PUT /api/v1/credentials/{id} | DELETE /api/v1/credentials/{id} |
| Managed Servers | POST /api/v1/backupInfrastructure/managedServers | GET .../{id} | PUT .../{id} | DELETE .../{id} |
| Repositories | POST /api/v1/backupInfrastructure/repositories | GET .../{id} | PUT .../{id} | DELETE .../{id} |
| Proxies | POST /api/v1/backupInfrastructure/proxies | GET .../{id} | PUT .../{id} | DELETE .../{id} |
| Jobs | POST /api/v1/jobs | GET /api/v1/jobs/{id} | PUT /api/v1/jobs/{id} | DELETE /api/v1/jobs/{id} |
| Protection Groups | POST /api/v1/agents/protectionGroups | GET .../{id} | PUT .../{id} | DELETE .../{id} |
| Sessions | — | GET /api/v1/sessions/{id} | — | — |

### Enum values (most common)
- **ECredentialsType:** `Standard`, `Linux`
- **ERepositoryType:** `WinLocal`, `LinuxLocal`, `Smb`, `Nfs`, `AzureBlob`, `AmazonS3`, `S3Compatible`, `GoogleCloud`, `LinuxHardened`, ...
- **EProxyType:** `ViProxy`, `HvProxy`, `FileProxy`
- **EManagedServerType:** `WindowsHost`, `LinuxHost`, `ViHost`, `CloudDirectorHost`, `HvServer`, ...
- **EJobType:** `Backup`, `BackupCopy`, `HyperVBackup`, `VSphereReplica`, `WindowsAgentBackup`, `LinuxAgentBackup`, ...
- **EProtectionGroupType:** `IndividualComputers`, `ADObjects`, `CSVFile`, `PreInstalledAgents`, `CloudMachines`, ...
