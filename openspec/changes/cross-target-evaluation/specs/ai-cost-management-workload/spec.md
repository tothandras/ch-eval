## ADDED Requirements

### Requirement: Internal workload isolation
AI Cost Management material SHALL live only under `internal/ai-cost-management/`. It SHALL use only synthetic data, and it SHALL be listed as a publication blocker in `PROGRESS.md` until it is removed or scrubbed. Items the brief marks out of scope SHALL NOT be re-raised.

#### Scenario: Publication check
- **WHEN** the repository is prepared for publication
- **THEN** `internal/` is removed or reviewed, and `PROGRESS.md` records the outcome

### Requirement: Evidence-labelled deliverable
The deliverable SHALL follow the brief's output format. Every finding SHALL be labelled verified, inferred, or assumed. Every verified local claim SHALL cite a script and a results file. Local timings SHALL NOT be presented as platform performance.

#### Scenario: Traceable claim
- **WHEN** a reviewer reads a verified finding
- **THEN** the evidence index names the script that reproduces it and the results file that records it
