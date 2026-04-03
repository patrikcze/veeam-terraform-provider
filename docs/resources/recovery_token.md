---
page_title: "veeam_recovery_token Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages an agent recovery token in Veeam Backup & Replication.
---

# veeam_recovery_token (Resource)

Manages an agent recovery token in Veeam Backup & Replication (`/api/v1/agents/recoveryTokens`). Recovery tokens are issued to Veeam agents to authorize recovery operations without requiring a full agent re-enrollment.

The `token_value` attribute is set only at creation time and is never returned by subsequent API reads. It is preserved in Terraform state — importing this resource will result in an empty `token_value`.

## Example Usage

```hcl
resource "veeam_managed_server" "agent_host" {
  name           = "agent01.example.com"
  type           = "WindowsHost"
  credentials_id = var.windows_credential_id
}

resource "veeam_recovery_token" "agent01" {
  name              = "agent01-recovery"
  description       = "Recovery token for agent01"
  managed_server_id = veeam_managed_server.agent_host.id
}

output "recovery_token_value" {
  value     = veeam_recovery_token.agent01.token_value
  sensitive = true
}
```

## Schema

### Required

- `name` (String) Display name of the recovery token.
- `managed_server_id` (String) UUID of the managed server this token is issued for. Changing this forces a destroy and recreate.

### Optional

- `description` (String) Optional description of the recovery token.

### Read-Only

- `id` (String) Recovery token identifier (assigned by the server).
- `token_value` (String, Sensitive) The actual recovery token string. Only available immediately after creation — not returned by subsequent API reads. Preserved in Terraform state.

## Import

```bash
terraform import veeam_recovery_token.main <uuid>
```

The UUID can be retrieved via `GET /api/v1/agents/recoveryTokens`.

**Important:** After import, `token_value` will be empty because the API does not return the token value after creation. The token itself is not regenerated; only the Terraform state entry is created.
