# OpenTofu Compatibility Analysis

This document analyses what would be required to publish and support this provider
for [OpenTofu](https://opentofu.org), the open-source Terraform fork maintained by
the Linux Foundation.

**TL;DR: Zero code changes are needed. The provider binary already works with OpenTofu.**

---

## Technical Compatibility

### Plugin Protocol

Both Terraform and OpenTofu use the same **gRPC Plugin Protocol v6**.
This provider is built on `terraform-plugin-framework`, which communicates exclusively
over that protocol. The provider binary is therefore already compatible with OpenTofu
without modification.

### Dependency Licensing

All provider SDK dependencies are under the **Mozilla Public License 2.0 (MPL-2.0)**,
the same licence used by OpenTofu. There are no BSL-1.1 or proprietary dependencies
in the `go.mod` that would require a licensing review for OpenTofu distribution.

| Package | Licence |
|---|---|
| `hashicorp/terraform-plugin-framework` | MPL-2.0 |
| `hashicorp/terraform-plugin-go` | MPL-2.0 |
| `hashicorp/terraform-plugin-log` | MPL-2.0 |
| `hashicorp/terraform-plugin-testing` | MPL-2.0 |

### Verified: No Code Changes Required

| Concern | Verdict |
|---|---|
| Provider schema API | Identical — same framework |
| CRUD lifecycle hooks | Identical |
| Import support | Identical |
| Acceptance test helpers | Identical (`terraform-plugin-testing` works with OpenTofu) |
| Sensitive field handling | Identical |

---

## What Would Need to Happen

### 1. Publish to the OpenTofu Registry (registry.opentofu.org)

The OpenTofu Registry operates similarly to the Terraform Registry:

1. Create an account at `registry.opentofu.org` under the `patrikcze` namespace.
2. Register a GPG public key in the Registry UI (can reuse the same key as Terraform Registry).
3. Push a signed `v*` tag — the same GoReleaser artefacts produced today are accepted.

The registry source address for users would be:

```hcl
terraform {
  required_providers {
    veeam = {
      source  = "opentofu/registry.opentofu.org/patrikcze/veeam"
      # or simply:
      source  = "registry.opentofu.org/patrikcze/veeam"
      version = "~> 0.1"
    }
  }
}
```

### 2. Dual-Registry CI/CD (Optional)

The current GoReleaser workflow can publish to both registries simultaneously by
adding a second `release` publish step in `.github/workflows/release.yml`. No changes
to `.goreleaser.yml` are needed — the same artefacts satisfy both registries.

### 3. Documentation Updates

- `README.md`: add an OpenTofu installation block alongside the Terraform one.
- `.goreleaser.yml` release footer: add OpenTofu installation instructions.
- No changes to provider source code or examples are required.

---

## Effort Estimate

| Task | Effort | Notes |
|---|---|---|
| Provider code changes | **None** | Already compatible |
| Registry account setup | ~1–2 hours | One-time, administrative |
| GPG key registration | ~30 minutes | Can reuse Terraform Registry key |
| CI/CD dual-publish step | ~2 hours | Additive to existing workflow |
| Documentation | ~1 hour | README + goreleaser footer |
| **Total** | **~4–6 hours** | Mostly one-time setup |

---

## Testing Against OpenTofu

To verify the provider works locally with OpenTofu before any registry publishing:

```bash
# Install the provider locally
make install

# In your Terraform/OpenTofu configuration, use the local mirror path:
terraform {
  required_providers {
    veeam = {
      source  = "registry.terraform.io/patrikcze/veeam"
      version = "dev"
    }
  }
}

# Run with OpenTofu instead of Terraform:
tofu init
tofu plan
tofu apply
```

Because the binary and plugin protocol are identical, local testing with `tofu` requires
no changes to `make install` or the provider itself.

---

## Summary

Transitioning to OpenTofu support is **low-effort and low-risk**. The entire work
is account/registry administration plus small documentation additions. No code, no
schema changes, no test changes. The existing CI/CD pipeline produces artefacts
that satisfy both registries out of the box.
