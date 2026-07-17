# Chorus

> *View the lyrics for your currently playing music.*

Chorus watches whatever media player is running on your system over MPRIS and shows synced or plain lyrics for the current track, no matter which app you're using.

<p align="center">
  <img src="data/screenshots/1.png" width="80%" alt="The app showing synced lyrics for 'Never Gonna Give You Up' by Rick Astley over a blurred album art background">
</p>

## Features

- Follows any MPRIS-compatible player automatically, or pick one manually
- Synced lyrics scroll and highlight the current line as the song plays
- Falls back to plain lyrics when no synced version is available
- Blurred album art as a live background
- Pluggable lyrics providers, [lrcmux](https://github.com/f1nniboy/lrcmux) by default

## Installation

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
