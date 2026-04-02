---
name: Priority 3 data sources implementation
description: All 13 Priority 3 read-only data sources implemented and verified as of 2026-04-02
type: project
---

All 13 T3.x data sources are implemented, tested, documented, and registered.

**Why:** Priority 3 task batch (T3.1–T3.13) from TASKS.md — expanding read-only API coverage.

**How to apply:** These are done; future work starts at Priority 4 (job type expansion) or Priority 5 (CI/CD).

## What was built

- `pkg/datasources/security_roles.go` — T3.1, filter by `role_id`
- `pkg/datasources/security_users.go` — T3.2, filter by `user_id`; uses `SecurityUserDataModel2` to avoid name collision with the resource's `SecurityUserDataModel`
- `pkg/datasources/backup_objects.go` — T3.3, filter by `object_id`; API field `restorePointsCount` mapped to `restore_point_count`
- `pkg/datasources/replicas.go` — T3.4, filter by `replica_id`
- `pkg/datasources/replica_points.go` — T3.5, filter by `replica_point_id`
- `pkg/datasources/proxy_states.go` — T3.6, no filter (like repository_states.go pattern)
- `pkg/datasources/protected_computers.go` — T3.7, filter by `computer_id` via list+scan (no by-ID endpoint in Swagger)
- `pkg/datasources/services.go` — T3.8, no filter
- `pkg/datasources/server_time.go` — T3.9, singleton, ID="server-time"
- `pkg/datasources/server_certificate.go` — T3.10, singleton, ID="server-certificate"
- `pkg/datasources/task_sessions.go` — T3.11, filter by `task_session_id`
- `pkg/datasources/security_analyzer.go` — T3.12, composite singleton; calls both `PathSecurityAnalyzerLastRun` and `PathSecurityAnalyzerBestPractices`
- `pkg/datasources/malware_events.go` — T3.13, no filter

## Endpoints added to endpoints.go

PathSecurityRoles, PathSecurityRoleByID, PathBackupObjects, PathBackupObjectByID, PathReplicas, PathReplicaByID, PathReplicaPoints, PathReplicaPointByID, PathProxyStates, PathServices, PathServerTime, PathServerCertificate, PathTaskSessions, PathTaskSessionByID, PathSecurityAnalyzerBestPractices, PathSecurityAnalyzerLastRun, PathMalwareEvents

## Test file

`pkg/datasources/prio3_datasources_test.go` — covers metadata, schema, configure (nil/invalid), read success/error for all 13, plus security_analyzer lastRun error and bestPractices error paths separately.
