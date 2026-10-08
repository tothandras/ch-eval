## MODIFIED Requirements

### Requirement: Pricing from named profiles
Every priced value SHALL be derived from a pricing profile checked into `pricing/`, and the full profile SHALL be embedded in the result JSON. Rates and service shape SHALL live in profile data, never in code. A profile SHALL declare `kind`: `cloud` (compute units × rate, the existing semantics) or `self-managed` (line items for node roles × count × hourly rate, EBS GB-month plus provisioned IOPS and throughput, S3 GB-month plus per-request rates, egress per GB, and an ops line of hourly rate × hours per month). Profiles without `kind` SHALL be read as `cloud`. Reports SHALL show both billed-shape and cpu-linear costs.

#### Scenario: zero-rate profile decouples accounting from pricing
- **WHEN** a run uses the `local-zero` profile
- **THEN** all USD values are 0 and all resource-accounting fields are still fully populated

#### Scenario: self-managed window cost
- **WHEN** a measure window is priced with a `self-managed` profile
- **THEN** the billed-shape window cost equals the sum of node-role hourly costs over the window, split across insert, merge, query, and idle by CPU share exactly as for cloud profiles, with storage, request, egress, and ops lines reported separately
