package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ScryfallSet struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ScryfallResponse struct {
	Data []ScryfallSet `json:"data"`
}

var (
	setMap  map[string]string
	setOnce sync.Once
)

// GetSetName returns the full set name for a given set code (e.g. "MH2" -> "Modern Horizons 2").
// It fetches from Scryfall if not already loaded.
func GetSetName(code string) (string, error) {
	setOnce.Do(func() {
		setMap = make(map[string]string)
		if err := loadSets(); err != nil {
			fmt.Printf("Warning: failed to load sets: %v\n", err)
		}
	})

	name, ok := setMap[code]
	if !ok {
		// Try refreshing if not found? Or just return empty.
		// For now, simple lookup.
		return "", fmt.Errorf("set code %s not found", code)
	}
	return name, nil
}

func loadSets() error {
	cachePath := filepath.Join(os.TempDir(), "scryfall_sets.json")
	
	// Check cache age
	info, err := os.Stat(cachePath)
	if err == nil && time.Since(info.ModTime()) < 24*time.Hour {
		if err := loadFromCache(cachePath); err == nil {
			return nil
		}
		// If load fails, fall through to fetch
	}

	if err := fetchAndCache(cachePath); err != nil {
		return err
	}
	return nil
}

func loadFromCache(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var sets []ScryfallSet
	// The cache file stores a raw array of ScryfallSet, not the wrapped response
	if err := json.NewDecoder(f).Decode(&sets); err != nil {
		return err
	}

	if setMap == nil {
		setMap = make(map[string]string)
	}
	for _, s := range sets {
		setMap[s.Code] = s.Name
	}
	return nil
}

func fetchAndCache(path string) error {
	resp, err := http.Get("https://api.scryfall.com/sets")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var scryResp struct {
		Data []ScryfallSet `json:"data"`
	}
	if err := json.Unmarshal(body, &scryResp); err != nil {
		return fmt.Errorf("failed to unmarshal scryfall response: %w", err)
	}

	if setMap == nil {
		setMap = make(map[string]string)
	}
	
	var sets []ScryfallSet
	for _, s := range scryResp.Data {
		setMap[s.Code] = s.Name
		sets = append(sets, s)
	}

	// Save cache as simple array
	f, err := os.Create(path)
	if err != nil {
		return nil // Cache write failure is not fatal
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(sets)
}
