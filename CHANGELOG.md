# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning]. Numbers in parentheses are related issues or pull
requests.

## [Unreleased]

## [0.22.0] - 2026-01-09

### Fixed

- Header breadcrumb overlap ([#1473])

### Changed

- Increase default Archivematica capacity to 20 ([#1405])

### Added

- Initial batch workflow, API endpoints and pages ([#1405])
- Batch UUID filter to list SIPs API endpoint ([#1405])
- New SIP statuses, canceled and validated ([#1405])
- Icons to SIP/AIP/Batch status badges ([#1405])

### Removed

- Capacity configuration for a3m ([#1405])

## [0.21.0] - 2025-12-03

### Added

- Configurable approval for AMSS AIP deletions ([#1401])
- AIP deletion report generation and download ([#1207])

## [0.20.0] - 2025-10-31

### Changed

- Improve WebSocket connections ([#1057])
- Order SIP source objects list by deposit date ([#1373])

### Added

- Retention period to all SIP origin configurations ([#1355])
- File size and deposit date to SIP source list ([#1373])
- Custom home page HTML support to dashboard ([#1383])

## [0.19.0] - 2025-10-01

### Fixed

- Show AIP deletion review buttons based on user authorization ([#1229])

### Changed

- Wording of SIP source upload note ([#1291])
- Auto-expand workflow tasks based on workflow status and count ([#1131])

### Added

- Ability to configure an institution logo ([#1309])
- Initial "Copy SIP" task to ingest workflow ([#1289])
- Spinner to workflow "in progress" badge ([#1289])

## [0.18.0] - 2025-09-04

### Fixed

- Clear SIPs user filter on page load ([#1321])

### Changed

- Change label for SIPs user filter ([#1322])

### Added

- Audit log for certain user actions ([#1282])

## [0.17.0] - 2025-08-08

### Added

- Automatic updates for the storage domain in the dashboard ([#1222])
- SIP ingest from a configurable S3/MinIO bucket source ([#1291])

## [0.16.0] - 2025-07-17

### Fixed

- Increase read and write timeouts for SIP upload ([#1237])

### Changed

- Send websocket events based on user attributes ([#1221])
- Use UUIDs as identifiers for ingest workflows and tasks ([#1209])

## [0.15.0] - 2025-06-27

### Fixed

- Set uploader max filesize from configuration ([#1237])
- Check auth in dashboard before loading SIP workflows ([#1238])
- Don't show AIP workflow auth error ([#1189])
- SIP download from internal bucket in ingest workflow ([#1276])
- AIP download authn/authz checks ([#1265])

### Changed

- Move failed SIP/PIP to internal bucket ([#1202])

### Added

- Support for Azure containers for internal storage ([#1231])
- Record and show the user that uploads a SIP ([#1246])
- Failed SIP/PIP download to dashboard and API ([#1202])
- API endpoint to list ingest users ([#1262])
- Uploaded by filter to the SIPs page ([#1262])

## [0.14.0] - 2025-05-30

### Fixed

- Workflows/tasks status legend ([#1160])
- Return "202 Accepted" from SIP upload endpoint ([#1186])

### Changed

- Use UUIDs as SIP identifiers ([#1209])
- Increase default SIP upload max size to 4 GiB ([#1237])

### Added

- Upload SIPs page to dashboard ([#1186])
- Status and type filter to list AIP workflows endpoint ([#1174])
- Allow users to cancel their deletion requests ([#1174])
- Processing workflow type configuration to watchers ([#1218])
- Return SIP UUID from SIP upload endpoint ([#1210])

### Removed

- Configuration option `stripTopLevelDir` from watchers ([#1218])

## [0.13.0] - 2025-05-01

### Fixed

- Bag validation when running concurrent workflows ([#1123])
- Nondeterministic error with preprocessing child workflow ([#1176])
- Conditionally load AIP workflows based on ABAC ([#1189])

### Changed

- Dashboard pager component for SIP/AIP browse pages ([#1168])
- Update SIP and AIP status values and badges style ([#1160])

### Added

- Failed status for workflows and tasks ([#1161])

### Removed

- Move API endpoints and workflow from ingest domain ([#1117])

## [0.12.0] - 2025-04-02

### Changed

- Rename SIP preservation action/task to workflow/task ([#1117])
- Use ISO 8601 format for datepicker dates ([#1156])

### Added

- Add workflows and tasks to storage service ([#1076])
- Add AIP deletion workflow ([#1076])

## [0.11.0] - 2025-03-13

### Changed

- Reorganize dashboard for ingest and storage ([#1117])
- Move AIP location card to AIP page ([#1117])

### Added

- List AIPs endpoint to storage service API ([#1117])
- AIP list and index pages to dashboard ([#1117])

### Fixed

- AIP download from dashboard ([#1117])
- Pass start and end times to time range filter ([#1102])
- Show an error on an invalid time range ([#1141])

## [0.10.0] - 2025-02-27

### Changed

- Optionally send unzipped bags to Archivematica ([#1136])
- Update packages search box placeholder and label ([#1130])
- Replace package tables by sip and aip tables in database schemas ([#1117])
- Re-design API for ingest and storage ([#1117])

### Added

- Custom date range package filters ([#1102])
- Make preservation task cards expandable ([#1077])

## [0.9.0] - 2025-02-13

### Added

- Created time package list filter ([#1102])

## [0.8.0] - 2025-01-29

### Added

- Search by name to the package list ([#1101])
- Preservation actions help box ([#986])

## [0.7.0] - 2025-01-17

### Changed

- Preservation actions hover/focus effect and expandable behavior ([#986])

## [0.6.0] - 2025-01-09

### Changed

- Preservation tasks view from table to cards ([#1077])

## [0.5.0] - 2024-11-29

### Added

- About information to API and dashboard ([#1062])

### Fixed

- Offcanvas sidebar look and behavior ([#1079])

## [0.4.0] - 2024-11-15

### Added

- Poststorage child workflows ([#886])
- Validate PREMIS XML before Archimatica/a3m processing ([#951])

### Changed

- Use tabs for package status filter ([#989])
- Reduce preservation tasks created from Archivematica jobs ([#950])

### Fixed

- API ABAC for access tokens without attributes ([#1066])

## [0.3.0] - 2024-10-29

### Added

- Status filter to package list page ([#989])
- Render line breaks in preservation tasks notes column ([#1039])

## [0.2.0] - 2024-10-23

### Added

- OIDC scopes configuration ([#1037])
- Option to skip API authentication email verified check ([#1037])
- Failed SIPs/PIPs buckets ([#929])
- ABAC roles mapping configuration ([#1035])
- Total package count and improved pager ([#988])

## [0.1.0] - 2024-10-02

Initial release.

[unreleased]: https://github.com/artefactual-sdps/enduro/compare/v0.22.0...HEAD
[0.22.0]: https://github.com/artefactual-sdps/enduro/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/artefactual-sdps/enduro/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/artefactual-sdps/enduro/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/artefactual-sdps/enduro/compare/v0.18.0...v0.19.0
[0.18.0]: https://github.com/artefactual-sdps/enduro/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/artefactual-sdps/enduro/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/artefactual-sdps/enduro/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/artefactual-sdps/enduro/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/artefactual-sdps/enduro/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/artefactual-sdps/enduro/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/artefactual-sdps/enduro/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/artefactual-sdps/enduro/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/artefactual-sdps/enduro/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/artefactual-sdps/enduro/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/artefactual-sdps/enduro/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/artefactual-sdps/enduro/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/artefactual-sdps/enduro/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/artefactual-sdps/enduro/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/artefactual-sdps/enduro/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/artefactual-sdps/enduro/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/artefactual-sdps/enduro/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/artefactual-sdps/enduro/releases/tag/v0.1.0
[#1473]: https://github.com/artefactual-sdps/enduro/issues/1473
[#1405]: https://github.com/artefactual-sdps/enduro/issues/1405
[#1401]: https://github.com/artefactual-sdps/enduro/issues/1401
[#1383]: https://github.com/artefactual-sdps/enduro/issues/1383
[#1373]: https://github.com/artefactual-sdps/enduro/issues/1373
[#1355]: https://github.com/artefactual-sdps/enduro/issues/1355
[#1322]: https://github.com/artefactual-sdps/enduro/issues/1322
[#1321]: https://github.com/artefactual-sdps/enduro/issues/1321
[#1309]: https://github.com/artefactual-sdps/enduro/issues/1309
[#1291]: https://github.com/artefactual-sdps/enduro/issues/1291
[#1289]: https://github.com/artefactual-sdps/enduro/issues/1289
[#1282]: https://github.com/artefactual-sdps/enduro/issues/1282
[#1276]: https://github.com/artefactual-sdps/enduro/issues/1276
[#1265]: https://github.com/artefactual-sdps/enduro/issues/1265
[#1262]: https://github.com/artefactual-sdps/enduro/issues/1262
[#1246]: https://github.com/artefactual-sdps/enduro/issues/1246
[#1238]: https://github.com/artefactual-sdps/enduro/issues/1238
[#1237]: https://github.com/artefactual-sdps/enduro/issues/1237
[#1231]: https://github.com/artefactual-sdps/enduro/issues/1231
[#1229]: https://github.com/artefactual-sdps/enduro/issues/1229
[#1222]: https://github.com/artefactual-sdps/enduro/issues/1222
[#1221]: https://github.com/artefactual-sdps/enduro/issues/1221
[#1218]: https://github.com/artefactual-sdps/enduro/issues/1218
[#1210]: https://github.com/artefactual-sdps/enduro/issues/1210
[#1209]: https://github.com/artefactual-sdps/enduro/issues/1209
[#1207]: https://github.com/artefactual-sdps/enduro/issues/1207
[#1202]: https://github.com/artefactual-sdps/enduro/issues/1202
[#1189]: https://github.com/artefactual-sdps/enduro/issues/1189
[#1186]: https://github.com/artefactual-sdps/enduro/issues/1186
[#1176]: https://github.com/artefactual-sdps/enduro/issues/1176
[#1174]: https://github.com/artefactual-sdps/enduro/issues/1174
[#1168]: https://github.com/artefactual-sdps/enduro/issues/1168
[#1161]: https://github.com/artefactual-sdps/enduro/issues/1161
[#1160]: https://github.com/artefactual-sdps/enduro/issues/1160
[#1156]: https://github.com/artefactual-sdps/enduro/issues/1156
[#1141]: https://github.com/artefactual-sdps/enduro/issues/1141
[#1136]: https://github.com/artefactual-sdps/enduro/issues/1136
[#1131]: https://github.com/artefactual-sdps/enduro/issues/1131
[#1130]: https://github.com/artefactual-sdps/enduro/issues/1130
[#1123]: https://github.com/artefactual-sdps/enduro/issues/1123
[#1117]: https://github.com/artefactual-sdps/enduro/issues/1117
[#1102]: https://github.com/artefactual-sdps/enduro/issues/1102
[#1101]: https://github.com/artefactual-sdps/enduro/issues/1101
[#1079]: https://github.com/artefactual-sdps/enduro/issues/1079
[#1077]: https://github.com/artefactual-sdps/enduro/issues/1077
[#1076]: https://github.com/artefactual-sdps/enduro/issues/1076
[#1066]: https://github.com/artefactual-sdps/enduro/issues/1066
[#1062]: https://github.com/artefactual-sdps/enduro/issues/1062
[#1057]: https://github.com/artefactual-sdps/enduro/issues/1057
[#1039]: https://github.com/artefactual-sdps/enduro/issues/1039
[#1037]: https://github.com/artefactual-sdps/enduro/issues/1037
[#1035]: https://github.com/artefactual-sdps/enduro/issues/1035
[#989]: https://github.com/artefactual-sdps/enduro/issues/989
[#988]: https://github.com/artefactual-sdps/enduro/issues/988
[#986]: https://github.com/artefactual-sdps/enduro/issues/986
[#951]: https://github.com/artefactual-sdps/enduro/issues/951
[#950]: https://github.com/artefactual-sdps/enduro/issues/950
[#929]: https://github.com/artefactual-sdps/enduro/issues/929
[#886]: https://github.com/artefactual-sdps/enduro/issues/886
[keep a changelog]: https://keepachangelog.com/en/1.1.0
[semantic versioning]: https://semver.org/spec/v2.0.0.html
