package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Load(dir, packID, lang string) (*QuizPack, error) {
	filename := fmt.Sprintf("%s_%s.json", packID, lang)
	path := filepath.Join(dir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("pack %s/%s not found: %w", packID, lang, err)
	}

	var p QuizPack
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("malformed pack %s: %w", filename, err)
	}

	if err := validate(&p); err != nil {
		return nil, fmt.Errorf("invalid pack %s: %w", filename, err)
	}

	return &p, nil
}

func validate(p *QuizPack) error {
	if len(p.Categories) == 0 {
		return fmt.Errorf("pack has no categories")
	}
	for i, cat := range p.Categories {
		if cat.Name == "" {
			return fmt.Errorf("category %d has no name", i)
		}
		if len(cat.Questions) == 0 {
			return fmt.Errorf("category %q has no questions", cat.Name)
		}
		for j, q := range cat.Questions {
			if q.Text == "" {
				return fmt.Errorf("category %q question %d has no text", cat.Name, j)
			}
			if q.Value <= 0 {
				return fmt.Errorf("category %q question %d has invalid value %d", cat.Name, j, q.Value)
			}
		}
	}
	return nil
}

// ListAvailable returns unique pack IDs found in dir along with their available languages.
func ListAvailable(dir string) ([]PackInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type entry struct {
		langs []string
	}
	seen := make(map[string]*entry)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".json")
		idx := strings.LastIndex(base, "_")
		if idx <= 0 {
			continue
		}
		id := base[:idx]
		lang := base[idx+1:]
		if _, ok := seen[id]; !ok {
			seen[id] = &entry{}
		}
		seen[id].langs = append(seen[id].langs, lang)
	}

	var result []PackInfo
	for id, e := range seen {
		result = append(result, PackInfo{ID: id, Langs: e.langs})
	}
	return result, nil
}

type PackInfo struct {
	ID    string   `json:"id"`
	Langs []string `json:"langs"`
}
