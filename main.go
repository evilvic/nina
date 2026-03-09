package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	MusicDir string
	Port     string
	User     string
	Pass     string
	DataDir  string
}

func main() {
	cfg := Config{
		MusicDir: envOrDefault("NINA_MUSIC_DIR", "/music"),
		Port:     envOrDefault("NINA_PORT", "4533"),
		User:     envOrDefault("NINA_USER", "admin"),
		Pass:     envOrDefault("NINA_PASS", "admin"),
		DataDir:  envOrDefault("NINA_DATA_DIR", "/data"),
	}

	if _, err := os.Stat(cfg.MusicDir); os.IsNotExist(err) {
		log.Fatalf("music directory does not exist: %s", cfg.MusicDir)
	}

	store, err := NewStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("failed to init store: %v", err)
	}

	log.Printf("nina starting — scanning %s", cfg.MusicDir)

	start := time.Now()
	lib, err := fullScan(cfg.MusicDir)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}
	log.Printf("scan complete: %d artists, %d albums, %d songs in %v",
		len(lib.Artists), len(lib.Albums), len(lib.Songs), time.Since(start).Round(time.Millisecond))

	handler := &API{cfg: cfg, lib: lib, store: store}

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
