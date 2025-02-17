<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning (we do at and after v1.0.0).

Usage:

Change log entries are to be added to the Unreleased section
under the appropriate stanza (see below).
Each entry should ideally include the Github issue or PR reference.

The issue numbers will later be link-ified during the
release process so you do not have to worry about including
a link manually, but you can if you wish.

Types of changes (Stanzas):

* __Added__ for new features.
* __Changed__ for changes in existing functionality that did not aim to resolve bugs.
* __Deprecated__ for soon-to-be removed features.
* __Removed__ for now removed features.
* __Fixed__ for any bug fixes that did not threaten user funds or chain continuity.
* __Security__ for any bug fixes that did threaten user funds or chain continuity.

Breaking changes affecting client, API, and state should be mentioned in the release notes.

Ref: https://keepachangelog.com/en/1.0.0/
Ref: https://github.com/osmosis-labs/osmosis/blob/main/CHANGELOG.md
-->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) for all versions `v1.0.0` and beyond (still considered experimental prior to v1.0.0).

## vX.Y.Z - TEMPLATE

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [Unreleased]

## v0.7.0

### Added

* [#634](https://github.com/allora-network/allora-chain/pull/634) Add events for listening coefficients and the computationally cheapest regrets
* [#678](https://github.com/allora-network/allora-chain/pull/678) Add optimistic execution + update CometBFT to v0.38.15
* [#663](https://github.com/allora-network/allora-chain/pull/663) Add whitelists for admins, topic creators, workers, and reputers. Create x/emissions v6 migration and chain upgrade v0.7.0
* [#654](https://github.com/allora-network/allora-chain/pull/654) Reorganize Linter Folder, add linter to check fuzzer state transition probabilities add to 100 percent. (Integrated as part of #653)
* [#685](https://github.com/allora-network/allora-chain/pull/685) Add burner permission to gov module
* [#627](https://github.com/allora-network/allora-chain/pull/627) Add fee market and fee grant module
* [#689](https://github.com/allora-network/allora-chain/pull/689) Add the `CircuitBreakerDecorator` to the `AnteHandler`
* [#697](https://github.com/allora-network/allora-chain/pull/697) Add upgrade guide to contributing.md

### Changed

* [#670](https://github.com/allora-network/allora-chain/pull/670) Adjust topic initial regret initialization 
* [#652](https://github.com/allora-network/allora-chain/pull/652) Reduce code duplication, set local_testnet_upgrade_l1.sh using environment variables with local_testnet_l1.sh instead
* [#650](https://github.com/allora-network/allora-chain/pull/650) Make fuzzer setup run through every state transition once before fuzzing starts
* [#651](https://github.com/allora-network/allora-chain/pull/651) Refactor: Fuzzer rename invariants test to fuzz test
* [#653](https://github.com/allora-network/allora-chain/pull/653) Fuzzer Bugfixes, Allow User to Set Fuzzer Transition Probability Distribution
* [#686](https://github.com/allora-network/allora-chain/pull/686) CLI query commands alignment
* [#691](https://github.com/allora-network/allora-chain/pull/691/files) Fix number of inferers considered in the reward cycle
* [#693](https://github.com/allora-network/allora-chain/pull/693) Migration test updates

### Deprecated

### Removed

* [#680](https://github.com/allora-network/allora-chain/pull/680) Remove custom ProposalHandler

### Fixed

* [#683](https://github.com/allora-network/allora-chain/pull/683) Added new strategy for reputer listening coefficients
* [#668](https://github.com/allora-network/allora-chain/pull/668) Add stake nil amount validation + added tests + fixed other tests
* [#687](https://github.com/allora-network/allora-chain/pull/687) Fix reputer nonce submission boundaries
* [#690](https://github.com/allora-network/allora-chain/pull/690) Make investor token unlock amounts strictly monotonically increasing
* [#695](https://github.com/allora-network/allora-chain/pull/695) Mint module GenesisState, proto, and v3 no-op migration (follow-on PR to 690)
* [#698](https://github.com/allora-network/allora-chain/pull/698) Added docs to CONTRIBUTING.md; Fix `x/mint` version for query autocli; Fix `x/emissions` v6 migration such that `x/gov` actually gets burner permissions; Add props to whitelist-related return protos to ensure 

### Security

* [#682](https://github.com/allora-network/allora-chain/pull/682) Ensure IBC `MsgTransfer` funds are received

## [Released]

## v0.6.0

### Added

* [#624](https://github.com/allora-network/allora-chain/pull/624) Add `nurse` internal healthcheck service based on `pprof`.  See `health/README.md`.
* [#642](https://github.com/allora-network/allora-chain/pull/642) Add release signing keys
* [#648](https://github.com/allora-network/allora-chain/pull/648) Update emissions v5 migration to reset way more maps => complete cleanup of poisonously negative values

### Changed

* [#538](https://github.com/allora-network/allora-chain/pull/538) Refactor Inference Synthesis to use Functions instead of "Builder Pattern"
* [#625](https://github.com/allora-network/allora-chain/pull/625) Determine what percentage of rewards to pay to each topic based on time-based accumulation rather than instantaneously.
* [#684](https://github.com/allora-network/allora-chain/pull/684) Avoid duplication of ABCI methods ordering. 

### Fixed

* [#615](https://github.com/allora-network/allora-chain/pull/615) Improvement to merit-based sortition to enable more cycling through the crowd
* [#615](https://github.com/allora-network/allora-chain/pull/615) Fixed runaway regret calculation where large influx of workers would cause the regret to grow negatively without bound
* [#635](https://github.com/allora-network/allora-chain/pull/635) Fixed occasional nil pointer dereference bug when calling `GetNetworkInferences`. Also, actually filled in loss block height return value.

### Added

* [612](https://github.com/allora-network/allora-chain/pull/612) Load testnet state into local fork via in-place-testnet command 
* [688](https://github.com/allora-network/allora-chain/pull/688) Add the `evidence` module.

### Security

* [#603](https://github.com/allora-network/allora-chain/pull/603) Validate all Values for Correctness Prior to Storing them in the Keeper
* [#620](https://github.com/allora-network/allora-chain/pull/620) Add a static analyzer to detect non-deferred `.Close()` calls, improve migration error handling
* [#622](https://github.com/allora-network/allora-chain/pull/622) Add telemetry metrics on queries/txs

## v0.5.0

### Summary

1. Implement a fix for incomplete migration of topic fields from the v0.4.0 upgrade. 
2. Add additional RPC endpoint to allow for recalculating 
   the inflation rate / target emission rate more often than once per month.
3. Update to use rewardable topics as active topics instead of using them separately.
4. Other miscellaneous minor fixes

### Added

* [#584](https://github.com/allora-network/allora-chain/pull/584) Allow admins to force target emission recalculation on params update and via dedicated mint RPC endpoint
* [#592](https://github.com/allora-network/allora-chain/pull/592) Update to use rewardable topics as active topics and rename `PickChurnableActiveTopics` to `UpdateNoncesOfActiveTopics`
* [#602](https://github.com/allora-network/allora-chain/pull/602) Refactor: Fix query server proto v4 (refactors [#592](https://github.com/allora-network/allora-chain/pull/592))
* [#607](https://github.com/allora-network/allora-chain/pull/607) Add events for building research monitoring suite

### Fixed

* [#582](https://github.com/allora-network/allora-chain/pull/582) Additional topic fields validation
* [#587](https://github.com/allora-network/allora-chain/pull/587) Update Migration Tests to include NaN Initial Regrets Test
* [#590](https://github.com/allora-network/allora-chain/pull/590) Clear NaN in maps by deleting all map values during migration
* [#595](https://github.com/allora-network/allora-chain/pull/592) Emit inference and forecast score events in case of 1 actor
* [#596](https://github.com/allora-network/allora-chain/pull/596) EMA protection against NaN values
* [#598](https://github.com/allora-network/allora-chain/pull/598) Math Operations on NaNs Should Return Errors
* [#604](https://github.com/allora-network/allora-chain/pull/604) Migration: clean additional stores. Ensure record pruning on reward failure. Autocli fix of multi-reputer-stake function.

### Security

* [#588](https://github.com/allora-network/allora-chain/pull/588)  Add String Max Length Module Parameter, Enforce Max String Length on Creation of New Topics
* [#610](https://github.com/allora-network/allora-chain/pull/610) Ensure that BlockHeight of scores is consistently set to be a topic epoch nonce. Also simplify the condition for idempotent payload submission.

## v0.4.0

### Summary

Implements fixes for our [June 2024](https://github.com/sherlock-audit/2024-06-allora-judging) Sherlock.xyz audit, including important fixes for determining which topics are considered active.

### Added

* [#542](https://github.com/allora-network/allora-chain/pull/542) Add scalable management of active topics with associated queries such as `GetActiveTopicsAtBlock` and `GetNextChurningBlockByTopicId`
* [#556](https://github.com/allora-network/allora-chain/pull/556) Scores now take an exponential moving average of the score rather than using the instantaneous score value from this epoch.

### Removed

* [#542](https://github.com/allora-network/allora-chain/pull/542) As part of active topic management, we removed `GetActiveTopics` and other (especially paginated) remnants of an unpartitioned store of active topics.

### Fixed

* [#544](https://github.com/allora-network/allora-chain/pull/544) Added check against zero-rewards after conversion to cosmosInt
* [#547](https://github.com/allora-network/allora-chain/pull/547) Improve error handling on InsertPayload, fixed/added tests err handling
* [#550](https://github.com/allora-network/allora-chain/pull/550) Fix reputer window upper limit
* [#555](https://github.com/allora-network/allora-chain/pull/555) Refactor: Rename TestSuite names
* [#567](https://github.com/allora-network/allora-chain/pull/567) Fix worker nonce window closing as soon as it opens

### Security

* See our recent [June 2024](https://github.com/sherlock-audit/2024-06-allora-judging) security audit for a full description of bugs found during that audit.
* [#554](https://github.com/allora-network/allora-chain/pull/554) Check Signature on Worker or Reputer Payload Matches Inferer/Forecaster/Reputer inside Bundle


## v0.3.0

Refactors to adapt to single transaction insertions from workers and reputers.

### Added

* Changelog added (this), improved [contribution guidelines](./CONTRIBUTING.md), and simplified [PR template](./.github/pull_request_template.md)
* [#443](https://github.com/allora-network/allora-chain/pull/443) Create mechanism to import and export state with `initGenesis` and `exportGenesis`
* [#461](https://github.com/allora-network/allora-chain/pull/461) Create index of topics by block height at which their open worker nonce closes
* [#432](https://github.com/allora-network/allora-chain/pull/432) State sync enabled for faster validator syncing
* [#478](https://github.com/allora-network/allora-chain/pull/478) Create and apply abstraction for augmenting topic fee revenue that is sure to check for topic activation criteria for both funding events, fee payments, and stake additions
* [#482](https://github.com/allora-network/allora-chain/pull/482) Creation of an official Upgrade Flow


### Removed

* [#458](https://github.com/allora-network/allora-chain/pull/458) Removal of Blockless and batch processing; Introduction of online, individual payload processing. This resolves many security, performance, and scalability issues.
* A number of PRs were merged prior to v0.3.0 that improved upon our usage of Blockless, however that has been removed in favor of its removal in #458. Hence, we are not listing those PRs here.
* [#462](https://github.com/allora-network/allora-chain/pull/462) Add individual payload processing
* [#470](https://github.com/allora-network/allora-chain/pull/470) Skim of top performers per topic as they submit payloads ("online skimming")
* [#464](https://github.com/allora-network/allora-chain/pull/464) Remove libp2p peer ids from chain
* [#459](https://github.com/allora-network/allora-chain/pull/459) Revamp nonce management

### Fixed

* [#486](https://github.com/allora-network/allora-chain/pull/486) Correctly set initial emission per unit staked token
* [#460](https://github.com/allora-network/allora-chain/pull/460) Apply a number of bugfixes and high-precision unit tests that ensure our implementation matches our simulation of Allora and therefore the original intentions of the whitepaper and Foundation.
* [#487](https://github.com/allora-network/allora-chain/pull/487) Patch to have the single `allorad` home folder in Docker
* [#437](https://github.com/allora-network/allora-chain/pull/437) Fix issue for excessive effective revenue dripping
* [#441](https://github.com/allora-network/allora-chain/pull/441) Prevent DoS from attempting to withdraw negative stake amounts
* [#472](https://github.com/allora-network/allora-chain/pull/472) Prevent topics to be created with a ground truth lag too big so that the reputation nonce could be dropped when the ground truth is revealed
* [#484](https://github.com/allora-network/allora-chain/pull/484) Fix issue with permissions within Docker container
* [#477](https://github.com/allora-network/allora-chain/pull/477) Fix issue arising from no forecasts being sent in worker payload
* [#436](https://github.com/allora-network/allora-chain/pull/436/files) Fix bug related to excessive usage of topic ground truth lag and misleading error message
* Additional bugfixes and improvements

### Security

* [#454](https://github.com/allora-network/allora-chain/pull/454) Bump CometBFT version
* [#465](https://github.com/allora-network/allora-chain/pull/465) Catch and avert error leading to inappropriate transfer of funds
* [#440](https://github.com/allora-network/allora-chain/pull/440) Remove parallelization to avert nonuniform executions between hardwares

## v0.2.14

* Added many new queries
   * [#430](https://github.com/allora-network/allora-chain/pull/430)
   * [#399](https://github.com/allora-network/allora-chain/pull/399)
   * [#353](https://github.com/allora-network/allora-chain/pull/353)
* Added confidence interval to network inference query
   * [#402](https://github.com/allora-network/allora-chain/pull/402)
* New actors are incorporated more smoothly instead of ignored then included in their next round
   * [#420](https://github.com/allora-network/allora-chain/pull/420)
* Created fuzzer so anyone can help find vulnerabilities (tell them how they can submit)
   * [#407](https://github.com/allora-network/allora-chain/pull/407)
* Parallelized tests for faster CI
   * [#394](https://github.com/allora-network/allora-chain/pull/394)
* Formally adopted gitflow with improved contributor experience
* Simplified nonce semantics
   * [#416](https://github.com/allora-network/allora-chain/pull/416)
* More reliable fetching of previous data
   * [#412](https://github.com/allora-network/allora-chain/pull/412)
   * [#409](https://github.com/allora-network/allora-chain/pull/409)
* Added validations for safer data ingress
   * [#398](https://github.com/allora-network/allora-chain/pull/398)
* Update forecast utility function
    * [#382](https://github.com/allora-network/allora-chain/pull/382)
* Automatically expire stake removals instead of requiring a 2nd tx
   * [#362](https://github.com/allora-network/allora-chain/pull/362)
* Added min bound on epoch length to prevent topics from hogging too many resources
   * [#376](https://github.com/allora-network/allora-chain/pull/376)
* Add restore from snapshot
   * [#352](https://github.com/allora-network/allora-chain/pull/352)
* Version updates to all docker files and deployment scripts
* Removed duplicate computations for increased efficiency
   * [#367](https://github.com/allora-network/allora-chain/pull/367)
* Bugixes, improved error handling, way more coverage, and simplifications including:
   * [#327](https://github.com/allora-network/allora-chain/pull/327)
   * [#318](https://github.com/allora-network/allora-chain/pull/318)
   * [#401](https://github.com/allora-network/allora-chain/pull/401)
    * [#272](https://github.com/allora-network/allora-chain/pull/272)
* More efficient state pruning
   * [#319](https://github.com/allora-network/allora-chain/pull/319)
* Way more validator logging
   * [#305](https://github.com/allora-network/allora-chain/pull/305)
* Removed print statements so more efficient logging
    * [#298](https://github.com/allora-network/allora-chain/pull/298)
* Tuned default global parameters
* Self-delegation disallowed (security patch)
   * [#269](https://github.com/allora-network/allora-chain/pull/269)
* Added stress tests
* Improved [README](./README.md)
* Move expensive logarithm and exponentiation operations to offchain nodes when possible
   * [#258](https://github.com/allora-network/allora-chain/pull/258)


## v0.1.0

*Versions below `v0.2.14` were associated with our "Alpha Testnet" and "Edgenet" deployments, before formal versioning was adopted.*

