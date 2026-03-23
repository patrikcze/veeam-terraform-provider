# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Fixed
- `veeam_backup_job`: preserve state stability for agent job `storage` and `schedule` optional/computed attributes after apply; avoid inconsistent-result errors when optional blocks are omitted.
- `veeam_backup_job`: preserve configured `storage.proxy_auto_select` for agent jobs when API responses do not return proxy selection fields.
- `veeam_repository`: normalize `use_fast_cloning_on_xfs_volumes` to a known value for non-Linux repository types to avoid unknown-after-apply errors.

### Validation
- Real VBR apply/destroy workflow validated for combined resources: `backup_job` (`LinuxAgentBackup`), `repository` (`WinLocal`), `proxy`, `managed_server` (`LinuxHost`), `protection_group`, `credential`, and `cloud_credential` (`AzureStorage`).
- Confirmed expected runtime behavior in VBR UI for:
  - Job storage repository selection via Terraform-managed repository ID.
  - Daily schedule settings and retry controls.
