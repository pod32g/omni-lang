# ADR-001: Compiler Roadmap to Working OmniLang Toolchain

## Status
Accepted

## Context

We now have a parser with error recovery and an initial type checker, but OmniLang still lacks
full semantic analysis, MIR lowering, optimisation passes, and executable backends. To ship a
working compiler and standard library we need a clear incremental roadmap that preserves the
project's small-merge philosophy and keeps `make build`/`make test` green at each stage.

## Decision

We will deliver the remaining milestones in the following phases:

1. **Type Checker Completion**
   - Implement expression type inference and operator compatibility rules.
   - Track variable mutability, struct/enum member access, and control-flow reachability for
     return checking.
   - Expand golden suites with both positive and negative cases so semantics are locked down.

2. **MIR Infrastructure**
   - Elaborate the AST into a typed SSA-based MIR using a builder package.
   - Add verifier and baseline passes (constant folding, dead-code elimination, SCCP, mem2reg).
   - Introduce MIR goldens to guarantee stable textual output.

3. **VM Backend & E2E Harness**
   - Create a bytecode interpreter (`omnir`) consuming MIR to execute programs locally.
   - Wire CLI flags so users can run OmniLang programs via `omnic --backend=vm` followed by
     `omnir` invocations.
   - Seed end-to-end tests for Hello World, arithmetic, and control-flow scenarios.

4. **Cranelift Native Backend**
   - Expand the Rust bridge to lower MIR to Cranelift IR, JIT or emit object code, and link a
     minimal runtime.
   - Integrate caching for native builds and ensure the GitLab pipeline exercises both Linux
     and macOS targets once runners are available.

5. **Runtime & Standard Library**
   - Provide the C shim and Go FFI glue for printing, allocation, and basic OS services.
   - Implement the initial `std` modules (io, math, collections) required by samples and tests.
   - Document the public API surface and add usage goldens.

Throughout these phases we will maintain a strict test-first discipline: each merge must add or
update unit tests, snapshot fixtures, or e2e programs validating the new behaviour. Any partial
functionality will be guarded behind feature flags or placeholder errors to avoid regressing user
experience.

## Consequences

- The roadmap guides prioritisation for future merges and sets expectations for reviewers.
- Contributors can pick discrete tasks (type inference, MIR builder, runtime FFI, etc.) knowing
  how they feed into the working compiler goal.
- Documentation (spec, tour, CLI help) will be updated alongside implementations so the repo
  remains authoritative for new users.
