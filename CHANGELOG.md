# Changelog

## [0.4.0](https://github.com/f1nniboy/chorus/compare/v0.3.0...v0.4.0) (2026-07-25)


### Features

* **ci:** generate new metainfo releases section on release ([9b4d43d](https://github.com/f1nniboy/chorus/commit/9b4d43dd2edbc8bad570ce5bee9c24f82fa32ae5))
* **picker:** add fallback icon if there's no cover art ([b22991c](https://github.com/f1nniboy/chorus/commit/b22991c4a391cce5f8673b95e5c03fda51517d13))
* **providers/lrcmux:** only request line sync level ([854d741](https://github.com/f1nniboy/chorus/commit/854d74170e5c3cbf89a3b74d344ecc19df0f25a2))


### Bug fixes

* **background:** use GTK stacks for transition between cover images, ([2597a35](https://github.com/f1nniboy/chorus/commit/2597a35a9554adedce192c79dd91a5ba4de2a279))
* **lyrics:** don't flash first line as current before playback starts ([e9dbabd](https://github.com/f1nniboy/chorus/commit/e9dbabdd311879d2ee9388fe1e09a43074d03064))
* **picker:** use ConnectRowActivated instead of per-row activation func ([6c94752](https://github.com/f1nniboy/chorus/commit/6c947522f161f589149e7d90856d6c3eedee90aa))
* **ui:** clean up ([5081645](https://github.com/f1nniboy/chorus/commit/5081645bc3202bda20923c0f7bf72229e9a886c1))
* **ui:** properly scroll on window maximize/minimize ([eaa00f5](https://github.com/f1nniboy/chorus/commit/eaa00f565ecb6b84f4d278e947c4b6b6b67b1da8))


### Documentation

* **readme:** add translation section ([5081645](https://github.com/f1nniboy/chorus/commit/5081645bc3202bda20923c0f7bf72229e9a886c1))


### Refactors

* **lyrics:** drop redundant blockScroll field ([e9dbabd](https://github.com/f1nniboy/chorus/commit/e9dbabdd311879d2ee9388fe1e09a43074d03064))
* **lyrics:** merge line and widget state into a single type ([e9dbabd](https://github.com/f1nniboy/chorus/commit/e9dbabdd311879d2ee9388fe1e09a43074d03064))
* **mpris:** merge player-tracking channels into ([fb6990b](https://github.com/f1nniboy/chorus/commit/fb6990bb5bbb66d3850f8efa50d063995b455a59))
* remove preferred player, automatically pick first active ([445f63d](https://github.com/f1nniboy/chorus/commit/445f63db1a7e4f31199d8c12e03272dfbe145133))

## [0.3.0](https://github.com/f1nniboy/chorus/compare/v0.2.0...v0.3.0) (2026-07-23)


### Features

* **metainfo:** add more info ([14847ee](https://github.com/f1nniboy/chorus/commit/14847eeb1549e0d44af85e91fabe7ca388eb6585))


### Bug fixes

* **lyrics:** increase fetch timeout ([b88c036](https://github.com/f1nniboy/chorus/commit/b88c036232b3564b656d917fde25a31e5ce3fcbc))
* **meta:** set Version via release-please ([b1ce811](https://github.com/f1nniboy/chorus/commit/b1ce811e838cb7a1871995d074d3be0b20f29c2e))
* **release-please:** fix version marker ([5aa567c](https://github.com/f1nniboy/chorus/commit/5aa567cbbb29936a98f68b64e919c381a3015354))
* **style:** consistent bg opacity ([2c2f473](https://github.com/f1nniboy/chorus/commit/2c2f47384cff11df725297e7137b90fc30fb2612))


### Documentation

* **README:** some changes ([abc7c6e](https://github.com/f1nniboy/chorus/commit/abc7c6ecde2675b9595995e6a7d088964053e78b))


### Refactors

* don't hardcode app name, set proper title ([2c2f473](https://github.com/f1nniboy/chorus/commit/2c2f47384cff11df725297e7137b90fc30fb2612))
* **flatpak:** use vars in manifests, bump commit ([14847ee](https://github.com/f1nniboy/chorus/commit/14847eeb1549e0d44af85e91fabe7ca388eb6585))


### Build

* **flatpak:** polish metadata and manifests ([d37b231](https://github.com/f1nniboy/chorus/commit/d37b2317712bc59117d6c8a4ac4c5520b095a866))

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
