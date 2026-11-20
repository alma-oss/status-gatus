# Fork Delta

This repository is a fork of [TwiN/gatus](https://github.com/TwiN/gatus).
Changes listed here are carried on top of upstream and must be reviewed during
upstream synchronization.

## Downstream changes

### External-only configuration

- Local change: `fix(config): Allow external-only configuration`
- Behavior: An `external-endpoints` entry satisfies the configuration requirement;
  native endpoints and suites are optional.
- Removal: Drop after an upstream release includes it.

### External endpoint condition results

- Upstream PR: [TwiN/gatus#1742](https://github.com/TwiN/gatus/pull/1742)
- Upstream commit: `869abe88`
- Local commit: `4cf36952` (`feat(api): Support pushing condition results to external endpoints`)
- Removal: Drop after the PR is merged into upstream

### Web base-path support

- Upstream issue: [TwiN/gatus#88](https://github.com/TwiN/gatus/issues/88)
- Upstream PR: [TwiN/gatus#1726](https://github.com/TwiN/gatus/pull/1726)
- Imported commits:
  - `72a71b47` as `c5223d8e`
  - `74fe98fa` as `4a072684`
  - `eece8143` as `d0ec1726`
- Removal: Drop after the PR is merged into upstream

### Config reload file modification time

- Upstream PR: [TwiN/gatus#1668](https://github.com/TwiN/gatus/pull/1668)
- Imported commits:
  - `fd334248` as `31dbf24d`
  - `6deab334` as `71fd4f30`
  - `5447c6f7` as `66df2fb8`
  - `e02d0766` as `4b47be27`
- This prevents repeated configuration reloads when a
  mounted ConfigMap timestamp is ahead of the container clock.
- Removal: Drop after the PR is merged into upstream

### Config reload evaluation counters

- Upstream PR: [TwiN/gatus#1791](https://github.com/TwiN/gatus/pull/1791)
- Upstream commit: `1814ba6e` as `3995b626`
- This restores persisted consecutive success and failure
  counters after a configuration reload.
- Removal: Drop after the PR is merged into upstream

### Large result pages are not write-through cached

- Local change: `fix(storage): Do not cache results pages larger than 100`
- Behavior: `getEndpointStatusByKey` skips the write-through cache when
  `pageSize > maximumCachedResultsPageSize` (100, upstream's hard `pageSize`
  cap). The fork allows `pageSize` up to `maximum-number-of-results` for the
  long history; without this, one such request stays cached forever (every
  insert refreshes it and resets its TTL) and every later insert re-reads the
  whole history (~300 ms and ~230 MB per push at 5 000 results × 8 conditions).
- Test: `storage/store/sql/sql_cache_test.go`.
- Removal: Keep while the fork allows `pageSize` above upstream's cap.

### Header buttons rendered

- Local change: `fix(web): Render ui.buttons in the index template`
- Behavior: `{{- range .UI.Buttons}}` sat under a `//` line comment and trimmed the newline that
  ended it; `html/template` strips JavaScript comments, so the generated `buttons.push(...)`
  calls were elided and `ui.buttons` never reached the header. The comment and the trim are
  gone (`web/app/public/index.html` and the built `web/static/index.html`).
- Test: `config/ui/ui_test.go` (`TestConfig_ValidateAndSetDefaults_RendersButtons`).
- Removal: Drop after an upstream release includes an equivalent; propose upstream.

## Synchronization

Before merging upstream changes, compare this list with upstream release notes
and merged pull requests. Remove entries only after the corresponding downstream
commits have been replaced by upstream equivalents.

### Bounded counter restore at startup

- Local change: `fix(main): Restore evaluation counters from at most the alert threshold`
- Behavior: `restoreNumberOfEvaluationsInARow` skips endpoints without enabled alerts and
  otherwise loads only `max(failure-threshold, success-threshold)` results instead of
  `maximum-number-of-results`. With the 90-day history (259 200 results per endpoint) the
  previous behaviour loaded the whole results and conditions history of every endpoint at
  every start, which kept the internal instance unready for 10+ minutes (2026-09-29).
- Related upstream: [TwiN/gatus#1791](https://github.com/TwiN/gatus/pull/1791) (the imported
  counter restore). Propose upstream; drop after an upstream release includes an equivalent.

### Derived `health` on status responses

- Local change: `feat(api): Add derived health to endpoint statuses`
- Behavior: `GET /api/v1/endpoints/statuses` and `GET /api/v1/endpoints/{key}/statuses` set
  `health` on every result and on the endpoint (health of the latest result):
  `HEALTHY` (success, all conditions passed), `HEALTHY_WITH_WARNINGS` (success, at least one
  failed condition — how the status stack encodes Warning, see open-questions D12) and
  `UNHEALTHY` (failure). Derived at response time, nothing is persisted.
- Removal: keep; candidate for an upstream proposal once the condition-results push
  (TwiN/gatus#1742) is merged, since the warning state depends on it.

## Release

The image of this fork is published to the GitHub Container Registry as `ghcr.io/alma-oss/status-gatus`:

- a tag `X.Y.Z` publishes the version `X.Y.Z` (workflow "Publish"),
- every push to a `feature/**` branch publishes the pre-release `0.0.0-<branch>.<run number>`
  (workflow "Pre-release"), e.g. `0.0.0-initial-implementation.7`.
