<p align="center">
  <img src="data/icons/hicolor/scalable/apps/space.f1nn.chorus.svg" alt="Chorus icon" width="128">
</p>

<h1 align="center">Chorus</h1>

<p align="center">
  View the lyrics for your currently playing music
  <br><br>
  <a href="https://github.com/f1nniboy/chorus/releases"><img src="https://img.shields.io/github/v/release/f1nniboy/chorus?color=blue" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/f1nniboy/chorus?color=blue" alt="License"></a>
  <a href="https://matrix.to/#/#chorus:oss.zone"><img src="https://img.shields.io/matrix/chorus:oss.zone.svg?server_fqdn=matrix.oss.zone&fetchMode=summary&color=blue" alt="Matrix"></a>
  <a href="https://hosted.weblate.org/engage/chorus/"><img src="https://hosted.weblate.org/widget/chorus/svg-badge.svg" alt="Translate"></a>
</p>

Chorus watches whatever media player is running on your system and shows synced or plain lyrics for the current track, no matter which app you're using[*](#recommended-players).

<p align="center">
  <img src="data/screenshots/1.png" width="80%" alt="The app showing synced lyrics for 'Never Gonna Give You Up' by Rick Astley over a blurred album art background">
</p>

## Features

- Follows any MPRIS-compatible player automatically, or you can pick one manually
- Synced lyrics scroll and highlight the current line as the song plays
- Blurred cover art as a live background (*not available for some players when using Flatpak, due to sandbox limitations*)
- Various lyrics providers, [lrcmux](https://github.com/f1nniboy/lrcmux) by default

## Recommended players

Chorus works with any MPRIS-compatible player, but not all players implement MPRIS equally well. Waylyrics maintains a good list of [recommended players](https://github.com/waylyrics/waylyrics/blob/master/README.en.md#recommended-players) that's equally relevant here.

## Installation

### Flatpak

Download `chorus.flatpak` from the [latest release](https://github.com/f1nniboy/chorus/releases/latest), and install it.

### From source

**Requirements**:
- GTK4
- libadwaita


```sh
git clone https://github.com/f1nniboy/chorus
cd chorus
glib-compile-schemas data/
go build -o chorus ./cmd/chorus
GSETTINGS_SCHEMA_DIR=data ./chorus
```

## Translate

Chorus is translated on [Weblate](https://hosted.weblate.org/engage/chorus/). Contributions for new or existing languages are welcome.
