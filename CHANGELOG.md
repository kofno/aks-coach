# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-13

### Added
- `--all-namespaces` flag to query resources across all namespaces
- HPA (Horizontal Pod Autoscaler) columns showing MIN/MAX replicas
- `HPA_TARGET` column displaying CPU current/target metrics
- `--selector` flag for label-based filtering in scope with display in title
- `--output` flag supporting `json` and `table` formats
- `--version` flag showing version information (injected from git tags via ldflags)

### Changed
- Refactored codebase with clean separation: `cli`/`kube`/`compute`/`render` layout

### Improved
- UI now displays pretty tables with intelligent truncation and numeric alignment

[0.1.0]: https://github.com/kofno/aks-coach/releases/tag/v0.1.0
