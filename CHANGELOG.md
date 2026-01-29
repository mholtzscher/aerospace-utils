# Changelog

## [0.3.2](https://github.com/mholtzscher/aerospace-utils/compare/v0.3.1...v0.3.2) (2026-01-29)


### Features

* **build:** enable CGO and remove amd64 from goarch in goreleaser config ([9dcee05](https://github.com/mholtzscher/aerospace-utils/commit/9dcee0504d0df8975b83b8a62bab3a9f15858241))

## [0.3.1](https://github.com/mholtzscher/aerospace-utils/compare/v0.3.0...v0.3.1) (2026-01-29)


### Features

* **release:** integrate GoReleaser into release-please workflow and remove standalone release workflow ([aa7141e](https://github.com/mholtzscher/aerospace-utils/commit/aa7141e2416505453a022ecf10f76eaf4022ae3c))

## [0.3.0](https://github.com/mholtzscher/aerospace-utils/compare/v0.2.0...v0.3.0) (2026-01-29)


### ⚠ BREAKING CHANGES

* **workspace:** The shift command semantics changed; it is now cumulative and resets when no --by is provided, so existing scripts relying on absolute shifts may produce different gaps

### Features

* add CLI argument parsing for personalized greetings ([413e347](https://github.com/mholtzscher/aerospace-utils/commit/413e347df2ba8af2f3f5e48a6065b3339d2aebed))
* add CLI with config and state management ([0b8fb7d](https://github.com/mholtzscher/aerospace-utils/commit/0b8fb7dc240c5025bb6106e9cc740d5c949dd249))
* add colored output to gaps command ([da5a9e3](https://github.com/mholtzscher/aerospace-utils/commit/da5a9e37da666dbb815e6c43369eabc4c55375e9))
* add colored output with --no-color flag ([b309c1f](https://github.com/mholtzscher/aerospace-utils/commit/b309c1f60bea96033cd8eb103b2c56420878311f))
* add config command to display resolved settings ([5cd4698](https://github.com/mholtzscher/aerospace-utils/commit/5cd469848a2a4490dcec422c20023579d6b76209))
* add dev helper scripts to flake.nix ([280f6c3](https://github.com/mholtzscher/aerospace-utils/commit/280f6c35977aa9a7ca93f2392bde0ad8f919f478))
* add monitor width override and non-macOS support ([1b8d5d3](https://github.com/mholtzscher/aerospace-utils/commit/1b8d5d33943b903879f319ed5b3a6f62a49e6fe7))
* allow negative values for gap adjustment amount ([cf6309d](https://github.com/mholtzscher/aerospace-utils/commit/cf6309d2dc213abef5e95473587b697a29da22de))
* default to 60% workspace width when state is missing ([5108ba1](https://github.com/mholtzscher/aerospace-utils/commit/5108ba1078cf93bf991c9084301e916bcfe368e2))
* pass no_reload flag through to aerospace reload ([d704f7f](https://github.com/mholtzscher/aerospace-utils/commit/d704f7fad24ced42d21a4da7ebe25c55e05f088f))
* **release:** initialize release tooling and CI configurations ([d4299ab](https://github.com/mholtzscher/aerospace-utils/commit/d4299ab944fa499d512967f611c2b1a9d43d76eb))
* **workspace:** add shift support for per-monitor gaps ([d573200](https://github.com/mholtzscher/aerospace-utils/commit/d5732003bf86ec36af4e6e758398c0602bfd5eb0))
* **workspace:** implement cumulative shift for the shift command ([ec87c13](https://github.com/mholtzscher/aerospace-utils/commit/ec87c1311c7d522856e6274e4758d76659e2f4b6))


### Bug Fixes

* correct apple-sdk attribute name in Nix flake ([4235d40](https://github.com/mholtzscher/aerospace-utils/commit/4235d40b469ecfc9d2be0d013dccc43da920714c))
* handle errors from flag operations and output calls ([f08907f](https://github.com/mholtzscher/aerospace-utils/commit/f08907f2a4deb1db31ac7cc4a9e1add5602c6b69))
* replace deprecated CGDisplayIOServicePort with IOKit registry iteration ([5b9c105](https://github.com/mholtzscher/aerospace-utils/commit/5b9c105e8cef39c4f4376ba2f27804b37169ded9))
* validate config exists before updating gaps ([84ce493](https://github.com/mholtzscher/aerospace-utils/commit/84ce493f98a79ee7caa2de823187ebdf1098cbfb))
