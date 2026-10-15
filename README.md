# nina

A lightweight, single-binary music server that implements the [Subsonic API](http://www.subsonic.org/pages/api.jsp). Point it at a directory of music files and connect with any Subsonic-compatible client.

~2,000 lines of Go. ~9 MB binary. No SQLite. No web UI. Just music.

## Features

- Scans and serves your local music library over the Subsonic API
- Reads metadata from audio file tags (artist, album artist, album, title, track, disc, year)
- Groups albums by **album artist**, so featured tracks ("X & Y") stay in X's album: the album artist tag, else "Various Artists" for compilations (flag or a `Compilations` folder), else the top-level artist folder
- Duration probing via ffprobe (concurrent, 8 workers)
- Transcoding via ffmpeg (MP3, Opus, AAC, FLAC) when clients request a lower bitrate
- Supports MP3, FLAC, M4A, AAC, OGG, Opus, WAV, and WMA
- Cover art discovery (`cover.jpg`, `folder.jpg`, `front.jpg`, etc.)
- Search across artists, albums, and songs
- Album lists (alphabetical, by artist, random, starred)
- Playlists with full CRUD (persisted to JSON)
- Star/unstar artists, albums, and songs (persisted to JSON)
- Play count tracking via scrobble
- Rescan library without restart (`startScan`)
- XML and JSON response formats
- Authentication via token (MD5 + salt) or plaintext/hex-encoded password
- Library indexed in memory, state persisted to a single JSON file
- Single static binary, easy to containerize

## Usage

```
go build -o nina .
NINA_MUSIC_DIR=/path/to/music ./nina
```

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `NINA_MUSIC_DIR` | `/music` | Path to your music library |
| `NINA_PORT` | `4533` | Port to listen on |
| `NINA_USER` | `admin` | Username for authentication |
| `NINA_PASS` | `admin` | Password for authentication |
| `NINA_DATA_DIR` | `/data` | Directory for persistent state (`nina.json`) |

### Container

```bash
# Apple Containers
container build --tag nina:latest .
container run -d --name nina -p 0.0.0.0:4533:4533 \
  -v /path/to/music:/music \
  -v /path/to/data:/data \
  nina:latest

# Docker / Podman
podman build -f Containerfile -t nina .
podman run -p 4533:4533 -v /path/to/music:/music:ro -v nina-data:/data nina
```

### Optional dependencies

- **ffprobe** - enables song duration detection during scan
- **ffmpeg** - enables on-the-fly transcoding when clients request a lower bitrate

Both are included in the container image.

## Subsonic API coverage

### Fully implemented

`ping`, `getLicense`, `getMusicFolders`, `getIndexes`, `getArtists`, `getArtist`, `getAlbum`, `getSong`, `getMusicDirectory`, `stream`, `download`, `getCoverArt`, `search2`, `search3`, `getAlbumList`, `getAlbumList2`, `getRandomSongs`, `getUser`, `getPlaylists`, `getPlaylist`, `createPlaylist`, `updatePlaylist`, `deletePlaylist`, `star`, `unstar`, `getStarred`, `getStarred2`, `scrobble`, `startScan`, `getScanStatus`

### Stubs (return OK, no persistence)

`setRating`

### Not implemented

Lyrics, podcasts, bookmarks, internet radio, chat, shares, user management, multi-user

## nina vs Navidrome

[Navidrome](https://www.navidrome.org/) is the most popular self-hosted Subsonic-compatible server (~20k GitHub stars, 215 contributors, GPL-3.0). nina aims to cover the same core use case in a fraction of the code.

### At a glance

|  | nina | Navidrome |
|---|---|---|
| Codebase | ~2,000 lines of Go | ~120k+ lines (Go 75%, JS 19%, Rust 3%) |
| Binary | ~9 MB | ~50 MB (embedded React web UI) |
| Database | Single JSON file | SQLite |
| State persisted | Playlists, stars, play counts | Everything (ratings, bookmarks, play queues, lyrics, shares, etc.) |
| Web UI | None — bring your own client | Full Material UI / React frontend with themes, 34+ languages |
| Users | Single | Multi-user with per-user playlists, play counts, and preferences |

### Features side by side

| Feature | nina | Navidrome |
|---|---|---|
| Browse (artists, albums, songs) | Yes | Yes |
| Search | Yes | Yes |
| Streaming | Yes | Yes |
| Transcoding | Yes (ffmpeg, optional) — MP3, Opus, AAC, FLAC | Yes (ffmpeg) — per-user/per-player settings, Opus supported |
| Cover art | File-based discovery | File-based + embedded + fetched from external sources |
| Playlists | CRUD, persisted to JSON | CRUD, M3U import, smart/dynamic playlists |
| Stars / favorites | Yes, persisted | Yes, persisted |
| Ratings (5-star) | Stub (no-op) | Yes |
| Scrobbling | Play count tracking (local) | Last.fm, ListenBrainz, Maloja |
| Smart playlists | No | Yes |
| Lyrics | No | Yes |
| Bookmarks / play queue | No | Yes |
| Sharing (public links) | No | Yes |
| Jukebox mode | No | Yes |
| Internet radio | No | Yes |
| Auto library monitoring | Rescan on demand (`startScan`) | Auto-detects changes, imports new files |
| External metadata | No | Last.fm, Spotify, Deezer artist/album info |
| Folder-based browsing | Yes (`getMusicDirectory`) | Simulated from tags (no real folder browsing) |

### Operational

| | nina | Navidrome |
|---|---|---|
| Startup | Scan to memory → serve | Scan → index to SQLite → serve |
| Memory model | Entire library in RAM (~4 MB per 10k songs) | SQLite on disk, lower memory for huge libraries |
| Persistence | `nina.json` (~KB) | SQLite database (~MB) |
| Runtime dependencies | ffmpeg/ffprobe (optional, for transcoding + durations) | ffmpeg (optional, for transcoding) |
| Platforms | Anywhere Go compiles + container | macOS, Linux, Windows, Raspberry Pi, Docker |
| Configuration | 5 env vars | YAML/TOML config file with 50+ options |

### When to use which

**Choose nina if** you want the simplest possible Subsonic server — no SQLite, no web UI, minimal config, runs anywhere with a single binary and a music folder. Good for single-user setups where you already have a preferred Subsonic client.

**Choose Navidrome if** you need multi-user, a web UI, Last.fm/ListenBrainz scrobbling, smart playlists, lyrics, 5-star ratings, bookmarks, jukebox mode, or external metadata. It's a more complete solution at the cost of more moving parts.
