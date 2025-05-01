# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning]. Numbers in parentheses are related issues or pull
requests.

## [Unreleased]

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

[unreleased]: https://github.com/artefactual-sdps/enduro/compare/v0.13.0...HEAD
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
[#1189]: https://github.com/artefactual-sdps/enduro/issues/1189
[#1176]: https://github.com/artefactual-sdps/enduro/issues/1176
[#1168]: https://github.com/artefactual-sdps/enduro/issues/1168
[#1161]: https://github.com/artefactual-sdps/enduro/issues/1161
[#1160]: https://github.com/artefactual-sdps/enduro/issues/1160
[#1156]: https://github.com/artefactual-sdps/enduro/issues/1156
[#1141]: https://github.com/artefactual-sdps/enduro/issues/1141
[#1136]: https://github.com/artefactual-sdps/enduro/issues/1136
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
