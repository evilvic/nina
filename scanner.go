package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dhowden/tag"
)

var audioExtensions = map[string]string{
	".mp3":  "audio/mpeg",
	".flac": "audio/flac",
	".m4a":  "audio/mp4",
	".aac":  "audio/mp4",
	".ogg":  "audio/ogg",
	".opus": "audio/ogg",
	".wav":  "audio/wav",
	".wma":  "audio/x-ms-wma",
}

func isAudioFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := audioExtensions[ext]
	return ok
}

func contentTypeForFile(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ct, ok := audioExtensions[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func makeID(prefix, input string) string {
	h := md5.Sum([]byte(input))
	return fmt.Sprintf("%s-%x", prefix, h[:8])
}

func fullScan(musicDir string) (*Library, error) {
	lib := NewLibrary()
	if err := scanLibrary(musicDir, lib); err != nil {
		return nil, err
	}
	lib.SortArtists()
	probeDurations(lib)
	return lib, nil
}

func scanLibrary(musicDir string, lib *Library) error {
	count := 0
	return filepath.Walk(musicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("walk error: %s: %v", path, err)
			return nil
		}
		if info.IsDir() || !isAudioFile(info.Name()) {
			return nil
		}

		relPath, _ := filepath.Rel(musicDir, path)

		artistName, albumName, title, track, disc, year := readTags(path, relPath)

		suffix := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

		artistID := makeID("ar", strings.ToLower(artistName))
		albumID := makeID("al", strings.ToLower(artistName+"/"+albumName))
		songID := makeID("sg", relPath)

		artist, ok := lib.Artists[artistID]
		if !ok {
			artist = &Artist{ID: artistID, Name: artistName}
			lib.Artists[artistID] = artist
		}

		album, ok := lib.Albums[albumID]
		if !ok {
			album = &Album{
				ID:         albumID,
				Name:       albumName,
				ArtistID:   artistID,
				ArtistName: artistName,
				Year:       year,
			}
			lib.Albums[albumID] = album
			artist.AlbumIDs = append(artist.AlbumIDs, albumID)

			dir := filepath.Dir(path)
			for _, coverName := range []string{"cover.jpg", "cover.png", "folder.jpg", "folder.png", "front.jpg", "front.png", "Cover.jpg", "Folder.jpg"} {
				coverPath := filepath.Join(dir, coverName)
				if _, err := os.Stat(coverPath); err == nil {
					album.CoverPath = coverPath
					break
				}
			}
		}

		if album.Year == 0 && year > 0 {
			album.Year = year
		}

		song := &Song{
			ID:          songID,
			Title:       title,
			AlbumID:     albumID,
			AlbumName:   albumName,
			ArtistID:    artistID,
			ArtistName:  artistName,
			Track:       track,
			Disc:        disc,
			Year:        year,
			Size:        info.Size(),
			Path:        path,
			Suffix:      suffix,
			ContentType: contentTypeForFile(info.Name()),
		}
		lib.Songs[songID] = song
		album.SongIDs = append(album.SongIDs, songID)

		count++
		if count%500 == 0 {
			log.Printf("scanned %d files...", count)
		}

		return nil
	})
}

func probeDurations(lib *Library) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Printf("ffprobe not found, skipping duration probing")
		return
	}

	total := len(lib.Songs)
	log.Printf("probing durations for %d songs...", total)

	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	var probed int64

	for _, song := range lib.Songs {
		wg.Add(1)
		sem <- struct{}{}
		go func(s *Song) {
			defer wg.Done()
			defer func() { <-sem }()
			s.Duration = probeDuration(s.Path)
			n := atomic.AddInt64(&probed, 1)
			if n%200 == 0 {
				log.Printf("probed %d/%d durations...", n, total)
			}
		}(song)
	}
	wg.Wait()

	for _, album := range lib.Albums {
		album.Duration = 0
		for _, songID := range album.SongIDs {
			if s, ok := lib.Songs[songID]; ok {
				album.Duration += s.Duration
			}
		}
	}

	log.Printf("duration probing complete")
}

func probeDuration(path string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		path)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return int(math.Round(dur))
}

func readTags(path, relPath string) (artist, album, title string, track, disc, year int) {
	parts := strings.Split(relPath, string(filepath.Separator))
	switch len(parts) {
	case 1:
		title = strings.TrimSuffix(parts[0], filepath.Ext(parts[0]))
		artist = "Unknown Artist"
		album = "Unknown Album"
	case 2:
		artist = parts[0]
		title = strings.TrimSuffix(parts[1], filepath.Ext(parts[1]))
		album = "Unknown Album"
	default:
		artist = parts[0]
		album = parts[1]
		title = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return
	}

	if m.Artist() != "" {
		artist = m.Artist()
	}
	if m.Album() != "" {
		album = m.Album()
	}
	if m.Title() != "" {
		title = m.Title()
	}
	if t, _ := m.Track(); t > 0 {
		track = t
	}
	if d, _ := m.Disc(); d > 0 {
		disc = d
	}
	if m.Year() > 0 {
		year = m.Year()
	}
	return
}
