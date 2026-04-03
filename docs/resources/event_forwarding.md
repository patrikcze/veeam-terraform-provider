---
page_title: "veeam_event_forwarding Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages the Veeam event forwarding configuration (SNMP traps and syslog).
---

# veeam_event_forwarding (Resource)

Manages the Veeam event forwarding singleton (`/api/v1/generalOptions/eventForwarding`). Controls where Veeam sends SNMP traps and syslog messages for backup events.

This is a singleton resource — only one instance may exist per provider configuration. Deleting the resource removes it from Terraform state only; the server configuration is not reset.

## Example Usage

```hcl
resource "veeam_event_forwarding" "main" {
  snmp_enabled   = true
  snmp_host      = "snmp-receiver.example.com"
  snmp_port      = 162
  snmp_community = "veeam-public"

  syslog_enabled  = true
  syslog_host     = "syslog.example.com"
  syslog_port     = 514
  syslog_protocol = "UDP"
}
```

## Schema

### Optional

- `snmp_enabled` (Boolean) Whether SNMP trap forwarding is enabled.
- `snmp_host` (String) Hostname or IP address of the SNMP trap receiver.
- `snmp_port` (Number) UDP port of the SNMP trap receiver (default: 162).
- `snmp_community` (String) SNMP community string for trap authentication.
- `syslog_enabled` (Boolean) Whether syslog event forwarding is enabled.
- `syslog_host` (String) Hostname or IP address of the syslog server.
- `syslog_port` (Number) Port of the syslog server.
- `syslog_protocol` (String) Transport protocol for syslog messages. Allowed values: `UDP`, `TCP`.

### Read-Only

- `id` (String) Always `"event-forwarding"`. Fixed singleton identifier.

## Import

This resource uses a fixed singleton ID:

```bash
terraform import veeam_event_forwarding.main event-forwarding
```
