# Changelog

## [1.0.0](https://github.com/jjfantini/humblSKILLS/compare/v0.6.4...v1.0.0) (2026-04-19)


### ⚠ BREAKING CHANGES

* **cli:** the `--adapters-dir` flag has been removed from both `humblskills` and `build-registry`. Any tooling or scripts passing `--adapters-dir=...` will need to drop the flag; the canonical adapters now live alongside the CLI source and are embedded at build time.

### Features

* **cli:** embed adapter catalog in binary ([9ba1920](https://github.com/jjfantini/humblSKILLS/commit/9ba1920de7a3f09cce72cddc17838078b7044898))

## [0.6.4](https://github.com/jjfantini/humblSKILLS/compare/v0.6.3...v0.6.4) (2026-04-18)


### Bug Fixes

* **cli:** align body divider with DETAIL title column ([#26](https://github.com/jjfantini/humblSKILLS/issues/26)) ([879a441](https://github.com/jjfantini/humblSKILLS/commit/879a44129a8cc77884d17c435c113dbdbad39180))

## [0.6.3](https://github.com/jjfantini/humblSKILLS/compare/v0.6.2...v0.6.3) (2026-04-18)


### Bug Fixes

* **cli:** size left pane via SizedItem natural width ([#24](https://github.com/jjfantini/humblSKILLS/issues/24)) ([9fc6fb6](https://github.com/jjfantini/humblSKILLS/commit/9fc6fb60797790a4aee09a8468d4f83e7eb684c4))

## [0.6.2](https://github.com/jjfantini/humblSKILLS/compare/v0.6.1...v0.6.2) (2026-04-18)


### Bug Fixes

* **cli:** tighten left pane + transparent selection for light mode ([#22](https://github.com/jjfantini/humblSKILLS/issues/22)) ([a12ccf1](https://github.com/jjfantini/humblSKILLS/commit/a12ccf14d735f51d4d013c511c82c66e2424ac7f))

## [0.6.1](https://github.com/jjfantini/humblSKILLS/compare/v0.6.0...v0.6.1) (2026-04-18)


### Bug Fixes

* **cli:** align two-pane divider and fill selected row ([#20](https://github.com/jjfantini/humblSKILLS/issues/20)) ([cbdc6cb](https://github.com/jjfantini/humblSKILLS/commit/cbdc6cbb36b2f52ffcb6c04f2a3988676bfc428c))

## [0.6.0](https://github.com/jjfantini/humblSKILLS/compare/v0.5.0...v0.6.0) (2026-04-18)


### Features

* **cli:** unify every command on the Tokyo Night two-pane TUI ([#18](https://github.com/jjfantini/humblSKILLS/issues/18)) ([d2803be](https://github.com/jjfantini/humblSKILLS/commit/d2803befc279a4ab22f1aa04e268db253a27f670))

## [0.5.0](https://github.com/jjfantini/humblSKILLS/compare/v0.4.0...v0.5.0) (2026-04-18)


### Features

* **cli:** unified TUI chrome with bubbletea browse, wizards, and progress ([#16](https://github.com/jjfantini/humblSKILLS/issues/16)) ([9fd5b09](https://github.com/jjfantini/humblSKILLS/commit/9fd5b09cccc773ff1f8e185a530a765a4acf9efe))

## [0.4.0](https://github.com/jjfantini/humblSKILLS/compare/v0.3.0...v0.4.0) (2026-04-18)


### Features

* **cli:** polish search UI and add interactive install picker ([#14](https://github.com/jjfantini/humblSKILLS/issues/14)) ([7c9dfec](https://github.com/jjfantini/humblSKILLS/commit/7c9dfec69c18b5e4d25fc4034d787281edf2b162))

## [0.3.0](https://github.com/jjfantini/humblSKILLS/compare/v0.2.0...v0.3.0) (2026-04-18)


### Features

* **skills:** add use-smart-skill ([6e65535](https://github.com/jjfantini/humblSKILLS/commit/6e65535dec39388e4bcbb38d251596f60e3cb6dc))

## [0.2.0](https://github.com/jjfantini/humblSKILLS/compare/v0.1.1...v0.2.0) (2026-04-17)


### Features

* **cli:** preserve user-owned content across skill updates ([#10](https://github.com/jjfantini/humblSKILLS/issues/10)) ([0c5f845](https://github.com/jjfantini/humblSKILLS/commit/0c5f845097bd65903c682b8a38f0ebb8663b1247))
* minor bump  |  fix: patch  |  feat!/BREAKING CHANGE: major ([a461bd1](https://github.com/jjfantini/humblSKILLS/commit/a461bd1818302269b23be8ca08018362082d3a14))


### Bug Fixes

* **cursor:** default scope to user to match claude-code adapter ([#7](https://github.com/jjfantini/humblSKILLS/issues/7)) ([9dececb](https://github.com/jjfantini/humblSKILLS/commit/9dececb61a703824eb895492c868efafde8791f3))
