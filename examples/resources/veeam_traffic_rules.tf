resource "veeam_traffic_rules" "main" {
  throttling_enabled = true

  # JSON-encoded array of traffic throttling rule objects.
  # Each rule targets a source/target IP range and sets a bandwidth limit.
  # Pass "[]" to enable the throttling feature with no rules configured.
  throttling_rules = jsonencode([
    {
      name            = "WAN-Throttle-Business-Hours"
      sourceSubnet    = "10.0.0.0/8"
      targetSubnet    = "192.168.0.0/16"
      throttlingValue = 10
      throttlingUnit  = "Mbps"
      isEnabled       = true
    }
  ])
}

output "traffic_rules_id" {
  value = veeam_traffic_rules.main.id
}
