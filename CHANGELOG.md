# Changelog

## [0.2.0](https://github.com/f1nniboy/chorus/compare/v0.1.0...v0.2.0) (2026-07-22)


### Features

* add Flatpak manifest, desktop file, metainfo, and icons ([5f07b25](https://github.com/f1nniboy/chorus/commit/5f07b25ce6715c6019f0cd856a759dffd0c25a49))
* add Flatpak packaging and release automation ([5f07b25](https://github.com/f1nniboy/chorus/commit/5f07b25ce6715c6019f0cd856a759dffd0c25a49))
* add localization ([c64c0c8](https://github.com/f1nniboy/chorus/commit/c64c0c8f8a8d847c34704a587a52b96c11d28a04))
* **i18n:** add German translation ([c64c0c8](https://github.com/f1nniboy/chorus/commit/c64c0c8f8a8d847c34704a587a52b96c11d28a04))


### Bug fixes

* **lyrics:** block touch-drag scrolling during synced playback ([085e2d7](https://github.com/f1nniboy/chorus/commit/085e2d7551d12752dc10c35f8a64a9394587e92d))
* **lyrics:** escape status description as markup ([d3a9bc4](https://github.com/f1nniboy/chorus/commit/d3a9bc4f66920b798b984dbf46cd4fce7d2d422c))
* **lyrics:** fix scroll centering when lines wrap ([b603baf](https://github.com/f1nniboy/chorus/commit/b603baf72f61ec642b6e264d736c07c02bef60b8))
* **lyrics:** remove dead spacer fields and fix stuck empty-result state ([454f4c7](https://github.com/f1nniboy/chorus/commit/454f4c7ea3c560ef72d7e973569dc54c9765d817))
* **lyrics:** scroll to top when position is before the first line ([9df63f7](https://github.com/f1nniboy/chorus/commit/9df63f785651bbd58973bbe6a42fa76b4d2c2074))
* **lyrics:** stop scroll position jumping to 0% on content rebuild ([9df63f7](https://github.com/f1nniboy/chorus/commit/9df63f785651bbd58973bbe6a42fa76b4d2c2074))
* **providers:** apply bool config fields to provider instances ([3e25b42](https://github.com/f1nniboy/chorus/commit/3e25b42d3240e80b38b8860f2694b00a24b6966e))
* **settings:** refresh cache size on dialog open instead of at startup ([5f07b25](https://github.com/f1nniboy/chorus/commit/5f07b25ce6715c6019f0cd856a759dffd0c25a49))


### Documentation

* **README:** add ([5ab522e](https://github.com/f1nniboy/chorus/commit/5ab522e07c1506b8d14f48d17922a49abc4e79a5))
* **readme:** note Flatpak cover-art sandbox limitation ([5f07b25](https://github.com/f1nniboy/chorus/commit/5f07b25ce6715c6019f0cd856a759dffd0c25a49))


### Refactors

* fix golangci-lint findings ([c64c0c8](https://github.com/f1nniboy/chorus/commit/c64c0c8f8a8d847c34704a587a52b96c11d28a04))
* **lrcmux:** normalize base URL once in Init instead of per-fetch ([1b366a7](https://github.com/f1nniboy/chorus/commit/1b366a78c5cdbe6dfb700ccecc6cf385d9b8a196))
* **lyrics:** drop custom lineWidget interface for *gtk.Widget ([9df63f7](https://github.com/f1nniboy/chorus/commit/9df63f785651bbd58973bbe6a42fa76b4d2c2074))
* **lyrics:** move provider rebuild off the main thread ([81ecdf0](https://github.com/f1nniboy/chorus/commit/81ecdf0321861c88a63dccf3f5a48b92ab41680c))
* **providers:** avoid instantiating a live provider for settings UI ([de36705](https://github.com/f1nniboy/chorus/commit/de36705886e88d13f9275e1d7093898de733f9e0))
