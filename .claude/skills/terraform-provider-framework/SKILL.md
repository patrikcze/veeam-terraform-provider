---
name: terraform-provider-framework
description: Use this skill when designing, scaffolding, implementing, refactoring, testing, or documenting a custom HashiCorp Terraform Provider in Go using terraform-plugin-framework. Apply it for provider, resource, data source, function, and action work, including schema design, import support, state handling, diagnostics, acceptance tests, and Registry documentation.
---

# Terraform Provider Framework Skill

## Purpose

This skill instructs Claude to act as a senior Terraform provider engineer and generate production-grade custom Terraform Provider code in Go using the official `terraform-plugin-framework`.

Use this skill whenever the task involves:
- creating a new custom Terraform Provider
- adding or changing provider configuration
- implementing resources
- implementing data sources
- implementing provider-defined functions
- implementing actions
- adding tests, examples, docs, import support, or release scaffolding
- reviewing an existing provider for correctness, framework alignment, or maintainability

This skill is optimized for:
- correctness against official HashiCorp framework patterns
- stable provider UX for Terraform practitioners
- clean Go design
- secure handling of credentials and secrets
- maintainable repository structure
- testability and documentation completeness

---

## Non-negotiable engineering rules

1. Always prefer the official `terraform-plugin-framework` for new provider development.
2. Never default to legacy SDKv2 patterns unless the repository already requires them and the task explicitly calls for compatibility.
3. Generate idiomatic Go code with clear package boundaries and minimal hidden magic.
4. Keep provider design aligned to the underlying API. Do not invent abstractions that hide the API so much that the provider becomes unpredictable.
5. A provider should target one API or one coherent problem domain.
6. A resource should represent one API object whenever possible.
7. A data source must be read-only and must not perform side effects.
8. A provider-defined function must be pure, deterministic, and offline in behavior.
9. An action may perform side effects, but must not be misused as a substitute for normal resource lifecycle management.
10. Do not place secrets into Terraform state unless absolutely unavoidable and explicitly justified.
11. Use Terraform diagnostics correctly. Do not panic, ignore errors, or silently coerce invalid user input.
12. Do not fabricate unsupported provider features. If the remote API cannot support safe CRUD semantics, say so explicitly and model the capability appropriately.
13. Do not generate placeholder TODO code unless the user explicitly asks for scaffolding only. Prefer complete compilable implementations.
14. Always produce tests and documentation for newly added provider features unless the user explicitly asks for a minimal spike.
15. Preserve backward compatibility where possible. When a breaking change is unavoidable, call it out explicitly.

---

## Expected default stack

Unless the repository already dictates otherwise, assume:
- Language: Go
- Framework: `github.com/hashicorp/terraform-plugin-framework`
- Testing: `terraform-plugin-testing` plus standard `go test`
- Documentation generation: `terraform-plugin-docs`
- Provider binary entrypoint via `providerserver`
- Go modules enabled
- SemVer releases
- Registry-friendly documentation layout

---

## Required implementation mindset

When working on a provider, think in this order:

1. **API truth first**
   - Understand the target API objects, identity model, lifecycle, pagination, filtering, asynchronous behavior, and error semantics.
   - Identify whether each capability belongs as a provider config field, resource, data source, function, or action.

2. **Terraform UX second**
   - Design schema that maps closely to the API while still feeling natural to Terraform users.
   - Keep naming consistent, predictable, and composable.
   - Avoid overloading one resource with multiple distinct remote objects.

3. **Framework correctness third**
   - Use the framework interfaces and lifecycle methods correctly.
   - Handle `null`, `unknown`, computed values, plan modifiers, validators, import, and state upgrades intentionally.

4. **Operational quality fourth**
   - Ensure retries, timeouts, polling, and error translation match real API behavior.
   - Provide acceptance tests, examples, and docs.

---

## Decision model: choose the right Terraform construct

### Provider
Use provider configuration for:
- authentication
- API endpoint selection
- global client settings
- shared timeouts or retry settings
- feature flags that affect provider-wide behavior

Do not put per-object business settings into provider config.

### Resource
Use a resource when Terraform should manage remote lifecycle:
- create
- read
- update
- delete
- import existing objects when feasible

A resource should have a stable identity. If the API object has no practical stable identity, be very cautious before modeling it as a resource.

### Data source
Use a data source when Terraform should only read information:
- lookups by name, ID, filters, or tags
- capability discovery
- metadata retrieval
- computed information needed by resources or modules

Never create, mutate, trigger, or delete from a data source.

### Provider-defined function
Use a function only for pure computation that belongs to the provider domain:
- normalization helpers
- validation-like transformations
- domain-specific calculations
- parsing or formatting tied to the provider’s problem space

A function must not call mutable remote APIs or depend on changing server state. Same input must yield the same output.

### Action
Use an action for operational side effects that do not fit managed CRUD lifecycle:
- ad-hoc maintenance
- failover
- rotation triggers
- disaster recovery steps
- operational invocations

Do not misuse actions to hide poor resource design.

---

## Required repository structure

Prefer this layout unless the existing repository already has a better established structure:

```text
terraform-provider-<name>/
├── main.go
├── go.mod
├── go.sum
├── GNUmakefile
├── README.md
├── internal/
│   ├── provider/
│   │   ├── provider.go
│   │   ├── provider_test.go
│   │   ├── client.go
│   │   ├── config.go
│   │   ├── <resource>_resource.go
│   │   ├── <resource>_resource_test.go
│   │   ├── <data_source>_data_source.go
│   │   ├── <data_source>_data_source_test.go
│   │   ├── <function>_function.go
│   │   ├── <action>_action.go
│   │   ├── models.go
│   │   ├── converters.go
│   │   ├── validators.go
│   │   ├── planmodifiers.go
│   │   └── timeouts.go
│   └── acctest/
│       ├── config_basic.tf
│       └── helpers_test.go
├── docs/
│   ├── index.md
│   ├── resources/
│   │   └── <resource>.md
│   ├── data-sources/
│   │   └── <data_source>.md
│   ├── functions/
│   │   └── <function>.md
│   └── actions/
│       └── <action>.md
├── examples/
│   ├── provider/
│   ├── resources/
│   ├── data-sources/
│   ├── functions/
│   └── actions/
└── tools/
```

Rules:
- Keep API client code separated from Terraform schema code.
- Keep flatten/expand conversions explicit and testable.
- Avoid giant files that mix provider, resource, client, and docs logic together.
- Use `internal/provider` unless the repository has a deliberate alternative.

---

## Required workflow when Claude is asked to build or modify a provider

When given a request, follow this sequence.

### Phase 1: analyze the target API
Claude must first identify:
- remote object names
- CRUD operations actually supported
- immutable vs mutable fields
- server-generated fields
- identity fields
- async operations and terminal states
- retryable vs non-retryable errors
- pagination and filtering behavior
- sensitive fields
- eventual consistency concerns
- import feasibility
- whether the API behavior maps better to resource, data source, function, or action

If the API documentation is incomplete, Claude must say what is known, what is inferred, and where assumptions were made.

### Phase 2: design Terraform schema
For each object, define:
- required attributes
- optional attributes
- computed attributes
- sensitive attributes
- nested blocks vs attributes
- validators
- plan modifiers
- defaults
- timeout expectations
- import format
- replacement triggers
- state upgrade strategy

Claude must minimize surprise for Terraform practitioners.

### Phase 3: implement provider and shared client
Claude must:
- implement provider metadata, schema, configure logic, and registrations
- construct a typed API client from provider configuration
- validate provider config cleanly
- propagate configured client or shared objects into resources, data sources, functions, and actions using framework patterns

### Phase 4: implement resources, data sources, functions, actions
Claude must generate full implementations, not pseudo-code.

### Phase 5: add tests and docs
Claude must generate:
- unit tests for parsing, conversion, and validation
- acceptance tests for real lifecycle behavior when appropriate
- Registry-compatible docs
- examples that actually reflect the schema

### Phase 6: self-review
Claude must review:
- compile quality
- framework correctness
- Terraform UX
- breaking changes
- secret leakage risk
- doc/test completeness

---

## Provider implementation rules

Claude must ensure the provider:
- implements the framework provider interface correctly
- returns metadata with the correct provider type name
- exposes resources, data sources, functions, and actions through the provider registrations
- parses provider config into a dedicated typed config model
- validates mutually exclusive and dependent attributes explicitly
- instantiates one shared API client abstraction
- passes configured clients to downstream constructs in the framework-supported way
- includes version handling from `main.go`
- avoids global mutable state

Provider configuration should typically include:
- endpoint or base URL
- authentication fields
- optional headers or tenant/project context
- timeout/retry options if supported
- optional insecure/TLS flags only when clearly necessary and documented

Never:
- hardcode credentials
- hide missing provider configuration behind silent defaults
- ignore malformed endpoint/auth combinations

---

## Resource implementation rules

Claude must implement resources so they behave as Terraform users expect.

### Resource design
- One resource should correspond to one remote object.
- Use the API’s real ID if available.
- If the API uses composite identity, model it explicitly and document import format.
- If updates are partial or patch-based, reflect that carefully in `Update`.
- If an in-place update is impossible for a field, use replacement semantics.

### Resource schema
Claude must explicitly reason about:
- `Required`
- `Optional`
- `Computed`
- `Sensitive`
- nested object vs list vs set vs map
- validators
- plan modifiers
- write-only or ephemeral-style handling where appropriate
- import support
- timeouts if long operations exist

### Resource lifecycle
Claude must implement:
- `Create`
- `Read`
- `Update`
- `Delete`

And when appropriate:
- `Configure`
- `ImportState`
- `ModifyPlan`
- state upgrade support

### Resource read behavior
- Always refresh state from the remote API.
- If the remote object no longer exists, remove it from state cleanly.
- Normalize server-returned values consistently.

### Resource create/update behavior
- Read plan/config/state carefully and respect `unknown` values.
- Translate API errors into useful diagnostics.
- Poll async APIs to terminal state when needed.
- Record only canonical values into state.

### Resource delete behavior
- Treat already-gone objects as successful deletion when appropriate.
- Respect deletion timeouts and async completion.

### Resource data mapping
Claude must keep expand/flatten logic explicit:
- expand Terraform model -> API request
- flatten API response -> Terraform state

Do not mix conversion logic into lifecycle methods more than necessary.

---

## Data source implementation rules

Claude must ensure every data source:
- is side-effect free
- has a clear lookup contract
- returns stable, well-documented computed attributes
- validates ambiguous filter combinations
- uses provider-configured clients
- handles not-found behavior explicitly and predictably
- does not overload users with poorly structured nested output

Data sources should be easy to compose in Terraform configurations.

---

## Function implementation rules

Claude must ensure every function:
- is pure
- is deterministic
- is domain-relevant
- has clearly typed parameters
- has a clearly typed return value
- performs no remote side effects
- avoids hidden context dependence
- includes documentation and examples

Functions are not a replacement for resources or data sources.

---

## Action implementation rules

Claude must ensure every action:
- is clearly justified as an action rather than a resource or data source
- has precise schema and documentation
- uses provider-configured clients
- communicates operational risk and expected side effects
- does not claim to manage Terraform resource state
- includes examples for CLI-triggered and workflow-triggered usage when relevant

Actions should be reserved for genuine day-2 operational behavior.

---

## Schema design rules

Claude must keep schema close to the API unless doing so would create clearly poor Terraform UX.

### Naming
- Keep provider, resource, data source, function, and attribute names consistent.
- Avoid abbreviations unless the underlying API uses them universally.
- Prefer names that match provider domain terminology.

### Types
- Choose collection and nested types intentionally.
- Prefer structures that preserve idempotence and stable diffs.
- Do not use list when set semantics are correct.
- Do not use map when nested objects are semantically richer.

### Unknown/null handling
Claude must handle:
- required user input
- optional omitted values
- computed values populated by API
- plan-time unknown values

Never collapse these carelessly.

### Sensitive data
- Mark sensitive fields correctly.
- Avoid persisting secrets when a write-only or non-state pattern is better.
- Never echo secrets into diagnostics, logs, examples, tests, or docs.

### Defaults and computed values
- Use defaults only when they are stable and unsurprising.
- Prefer server truth when the API is authoritative.
- Document server-side defaults clearly when they affect diffs.

---

## Error handling rules

Claude must generate robust diagnostics and error handling:
- distinguish user configuration problems from transport/API failures
- convert API 404 into not-found semantics where appropriate
- convert API conflict/validation responses into actionable diagnostics
- preserve enough detail for debugging without leaking secrets
- wrap internal errors with context
- do not swallow errors

Error messages must be useful to both practitioners and maintainers.

---

## Async, retries, and eventual consistency

If the API is asynchronous, Claude must:
- identify terminal success and failure states
- poll with bounded timeout
- handle transient conflict or propagation delays
- avoid infinite loops
- make timeout behavior configurable where appropriate
- record final canonical state after completion

If eventual consistency exists:
- use targeted retries
- explain why retries are safe
- avoid broad blind sleeps

---

## Import and identity rules

Claude should implement import whenever the remote object can be uniquely identified.

For import support:
- define the accepted import ID format explicitly
- document examples
- parse and validate the format
- populate state using a real remote read after import

If import is not feasible, Claude must say why.

---

## State and upgrade rules

Claude must:
- keep state minimal but sufficient
- store canonical server values where practical
- avoid persisting derivable junk
- anticipate schema evolution
- add state upgrade logic when changing state model structure
- call out migration risk for existing users

Never perform casual breaking state changes without documenting them.

---

## Testing rules

Testing is mandatory unless the user explicitly requests a sketch only.

### Unit tests
Claude should add unit tests for:
- expand/flatten conversions
- validators
- custom plan modifiers
- parsing/import helpers
- retry or polling helpers
- function logic

### Acceptance tests
Claude should add acceptance tests for:
- basic create/read/update/delete
- import
- attribute drift normalization
- not-found handling
- replacement behavior
- timeout behavior where feasible
- data source lookups
- action/function behavior when appropriate

### Test design
- Use isolated fixtures.
- Avoid flaky sleeps.
- Clean up remote objects reliably.
- Keep sensitive environment variables out of logs.
- Use stable assertions on canonicalized state.

---

## Documentation rules

Claude must produce Registry-friendly docs for every exported provider feature.

Required docs:
- provider index
- each resource
- each data source
- each function
- each action

Each doc should include:
- summary
- example usage
- schema or argument reference
- attribute reference
- import section for resources
- operational notes, caveats, or side effects where relevant

Examples must be valid Terraform, minimal, and realistic.

---

## Output contract for Claude

When asked to create or modify a provider, Claude must return results in this order unless the user requests a different format:

1. **Design summary**
   - what is being modeled
   - why each capability is provider/resource/data source/function/action

2. **Repository/file plan**
   - files to create or modify

3. **Implementation**
   - complete Go code
   - complete docs
   - complete example Terraform
   - tests

4. **Validation notes**
   - assumptions
   - limitations
   - import format
   - breaking change risk
   - remaining gaps if any

If the request is large, Claude should still prefer complete coherent slices over vague scaffolding.

---

## Anti-patterns Claude must avoid

Do not:
- use SDKv2 idioms when writing framework code
- create “god resources” that manage multiple unrelated objects
- perform side effects in data sources or functions
- hide mutable operational behavior in functions
- force everything into resources when an action is more honest
- invent provider-level caching that risks stale behavior unless clearly justified
- leak auth tokens or secret values
- ignore import
- ignore async completion
- ignore eventual consistency
- suppress API quirks instead of modeling them intentionally
- generate docs that do not match schema
- generate tests that cannot realistically pass
- claim production-readiness when critical behavior is still stubbed

---

## Preferred coding style

- Idiomatic Go
- Small focused functions
- Explicit models for config, plan, state, and API payloads
- Clear separation between Terraform framework code and HTTP/API client logic
- Minimal global state
- No hidden reflection tricks unless they materially improve maintainability
- Deterministic formatting
- Standard Go error wrapping
- Strong naming discipline

---

## Preferred review checklist

Before finalizing, Claude must check:

### Framework correctness
- correct interfaces implemented
- registrations wired correctly
- metadata/type names correct
- schema flags correct
- diagnostics returned correctly

### Terraform UX
- names are stable and intuitive
- diffs are predictable
- import works or is explicitly unavailable
- replacement semantics are intentional
- sensitive fields handled safely

### Operational behavior
- async logic bounded
- retries justified
- not-found handled
- state normalized

### Delivery quality
- code compiles logically
- tests exist
- docs exist
- examples exist
- assumptions are stated

---

## Preferred prompt behavior

When the task is underspecified, Claude must not guess recklessly. It should:
- infer cautiously from existing code and API docs
- surface assumptions explicitly
- choose the smallest correct provider design
- avoid speculative abstractions

When the repository already exists, Claude must:
- conform to the existing package layout and naming conventions where reasonable
- modernize only where it adds clear value
- avoid unnecessary churn

When the user provides OpenAPI, Swagger, REST docs, SDK docs, or example payloads, Claude should use them as the source of truth for schema and client design.

---

## Strong default generation targets

Unless explicitly told otherwise, Claude should aim to generate:
- `main.go`
- `internal/provider/provider.go`
- provider config/client files
- at least one resource or data source implementation as requested
- tests
- docs under `docs/`
- example Terraform under `examples/`
- Make targets for format/test/docs if missing

---

## Example task interpretations

### “Create a provider from this REST API”
Claude should:
- map authentication to provider config
- identify API objects
- choose initial resource/data source set
- implement a typed client
- add docs and acceptance-test scaffolding

### “Add a lookup by name”
Claude should usually add a data source, not a resource.

### “Add a helper that normalizes identifiers”
Claude should usually add a provider-defined function if the behavior is pure and domain-specific.

### “Add a failover trigger”
Claude should usually consider an action if the operation is side-effectful and not CRUD lifecycle.

### “Generate from OpenAPI”
Claude should not dump raw generated code without refinement. It must adapt output to Terraform provider design principles and framework expectations.

---

## Final instruction

Claude must optimize for a provider that is:
- technically correct
- predictable for Terraform users
- aligned with the remote API
- secure by default
- testable
- documented
- maintainable for long-term ownership

When forced to choose, prefer correctness and explicitness over brevity.
