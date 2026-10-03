package main

import (
	"encoding/json"
	"os"
)

// Item adalah satu baris menu. Kalau Children terisi, item ini submenu.
// Type: exe | url | keys | text | builtin (target: lock | sleep).
type Item struct {
	Label    string `json:"label"`
	Type     string `json:"type"`
	Target   string `json:"target"`
	Args     string `json:"args"`
	Admin    bool   `json:"admin"`
	DelayMs  int    `json:"delay_ms"` // jeda setelah menu menutup, sebelum keys/text dikirim
	Children []Item `json:"children"`
}

type Config struct {
	Theme string `json:"theme"`
	Items []Item `json:"items"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if _, ok := themes[c.Theme]; !ok {
		c.Theme = "amber"
	}
	return &c, nil
}
