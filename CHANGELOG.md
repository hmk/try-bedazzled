# Changelog

## [0.3.1](https://github.com/hmk/try-bedazzled/compare/v0.3.0...v0.3.1) (2026-05-12)


### Documentation

* **readme:** lead with serious framing, drop "rainbow" tagline ([#24](https://github.com/hmk/try-bedazzled/issues/24)) ([ad8e7c1](https://github.com/hmk/try-bedazzled/commit/ad8e7c1230244ae8ba727ccff94b1147e637e42c))

## [0.3.0](https://github.com/hmk/try-bedazzled/compare/v0.2.0...v0.3.0) (2026-05-11)


### Features

* **install:** publish a homebrew cask to hmk/homebrew-tap on each release ([#22](https://github.com/hmk/try-bedazzled/issues/22)) ([c17aead](https://github.com/hmk/try-bedazzled/commit/c17aead06b3b0110e1950116ce2f2d6c75d6d056))

## [0.2.0](https://github.com/hmk/try-bedazzled/compare/v0.1.3...v0.2.0) (2026-05-11)


### Features

* add install script and improve readme ([#20](https://github.com/hmk/try-bedazzled/issues/20)) ([ee3af3e](https://github.com/hmk/try-bedazzled/commit/ee3af3e8b72018e112afad968812960757a21c4c))

## [0.1.3](https://github.com/hmk/try-bedazzled/compare/v0.1.2...v0.1.3) (2026-05-11)


### Bug Fixes

* **release:** read notary issuer_id and key_id from vars, not secrets ([#16](https://github.com/hmk/try-bedazzled/issues/16)) ([f2efd48](https://github.com/hmk/try-bedazzled/commit/f2efd48010fb368a6f35a52b31770aae1b5e63c8))

## [0.1.2](https://github.com/hmk/try-bedazzled/compare/v0.1.1...v0.1.2) (2026-05-11)


### Bug Fixes

* **release:** notarize.macos.ids must reference build id, not archive id ([#14](https://github.com/hmk/try-bedazzled/issues/14)) ([0d9b45e](https://github.com/hmk/try-bedazzled/commit/0d9b45e58f80e47fd7d1afc14e4d2fbe6528578c))

## [0.1.1](https://github.com/hmk/try-bedazzled/compare/v0.1.0...v0.1.1) (2026-05-11)


### Bug Fixes

* trigger 0.1.1 to validate signed release pipeline end-to-end ([8beb264](https://github.com/hmk/try-bedazzled/commit/8beb264d146bbcfa5216b48688100ccfcb8216d0))

## 0.1.0 (2026-05-11)


### Features

* adaptive fullscreen sizing, space-as-dash queries, wider-rendering icons ([1f85de4](https://github.com/hmk/try-bedazzled/commit/1f85de4afc447ade847811098d4cb5916c2ad6e8))
* add benchmark harness comparing Go, C, and Ruby ([62377b2](https://github.com/hmk/try-bedazzled/commit/62377b2299837b0126ff2f467e39ec0742cab728))
* add Bubble Tea TUI selector with fuzzy filtering ([d5ba591](https://github.com/hmk/try-bedazzled/commit/d5ba5917a704db072738a6b9149cb0d131d30d4b))
* add GoReleaser, CI workflows, and README ([b3c0eac](https://github.com/hmk/try-bedazzled/commit/b3c0eaccce343fd9d5784d583b118f9ecf368d0b))
* add interactive theme picker with live preview ([bad2913](https://github.com/hmk/try-bedazzled/commit/bad29134c41ce2164bfb37b635ddc14aa8bb6ffd))
* add theming system with four built-in themes ([f716f23](https://github.com/hmk/try-bedazzled/commit/f716f2362186e3593dfe889f4871660f25748411))
* bedazzle TUI with icons, columns, and visual polish ([d7ae28e](https://github.com/hmk/try-bedazzled/commit/d7ae28ec614df38afb2670e194bc9750dd19ab67))
* bubbles table picker + file tree preview panel ([98c4d70](https://github.com/hmk/try-bedazzled/commit/98c4d70cf302c1b251078753cc43785a01543020))
* expand Config with prefs and custom icons support ([698ed31](https://github.com/hmk/try-bedazzled/commit/698ed313206f3dea9542294beec7ba876d413263))
* **readme:** lean into the bedazzling ([60559a2](https://github.com/hmk/try-bedazzled/commit/60559a2b4666b9a4d9ccfd0357daf96ad06d366e))
* rename default to bedazzled with rainbow bars + right-aligned score ([a8dda9a](https://github.com/hmk/try-bedazzled/commit/a8dda9a4b4c0e8b73c1aab6b406930e082a87f49))
* selected-row BG highlight, rainbow gradient, bedazzled→catppuccin palette ([6beeeed](https://github.com/hmk/try-bedazzled/commit/6beeeedbbe45ebf70626ffb1a433d3e428da6c59))
* settings menu (ctrl-,), search bar top/bottom rules, display mode ([424fb85](https://github.com/hmk/try-bedazzled/commit/424fb85bc748f693d493cee876e2761e2fcd8532))
* showcase rainbow theme in second half of demo ([#9](https://github.com/hmk/try-bedazzled/issues/9)) ([b5478bd](https://github.com/hmk/try-bedazzled/commit/b5478bde2acefd4e96f81f63adfb9e08aaedeb72))
* true scrolling, column reorder, custom icons, ctrl-p preview toggle ([ebed1c2](https://github.com/hmk/try-bedazzled/commit/ebed1c2bd03191181f56fef17ce677c7e61b9f54))
* wire up Cobra CLI with all commands ([1b705c4](https://github.com/hmk/try-bedazzled/commit/1b705c4909f94ed5e80133014024adcab5c308b4))


### Bug Fixes

* address lint issues in source files ([968950d](https://github.com/hmk/try-bedazzled/commit/968950dc996e891d26bdb1b31bcf91037ae5834b))
* change settings key from ctrl-, to ctrl-g ([6d2ace8](https://github.com/hmk/try-bedazzled/commit/6d2ace8867d056211f326ecfbababecdb193d591))
* enable colors in theme picker via shared TTY renderer ([a44ae28](https://github.com/hmk/try-bedazzled/commit/a44ae283fdb8fdebd9eb0e819a2832bba980d8f3))
* keep selection on-screen when preview toggles, trim preview row, lint ([35a5226](https://github.com/hmk/try-bedazzled/commit/35a5226af491c476e30134753816086b74333313))
* read theme from config file when resolving ([45e4137](https://github.com/hmk/try-bedazzled/commit/45e4137f201ee27b0e35f81598a828dc08f6ef0c))
* render last visible frame in key injection mode ([4eb54bc](https://github.com/hmk/try-bedazzled/commit/4eb54bcb4d0af8bc7ad2412641a1192b84fb415e))
* route theme, help, and version through exec mode ([dfd11c1](https://github.com/hmk/try-bedazzled/commit/dfd11c164771995d5b836742c40e6b21e6bb2deb))
* show_emojis setting now actually hides icons ([7a445b4](https://github.com/hmk/try-bedazzled/commit/7a445b472d10d0b73ad52fff0c8be5395d82aa23))
* use /dev/tty for Lip Gloss color detection ([a605a98](https://github.com/hmk/try-bedazzled/commit/a605a98ab9a0470817a2175fd90da8cd2bdcf193))


### Refactors

* replace huh settings form with native Bubble Tea selector ([bbcff59](https://github.com/hmk/try-bedazzled/commit/bbcff59b954fe6f91c58570868ce0f225dc7ac2b))


### Documentation

* rewrite README with full feature docs and differentiators ([3b3b372](https://github.com/hmk/try-bedazzled/commit/3b3b372f70096f4377cb03f157c92de9a3c585d7))
