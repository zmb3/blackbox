package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type storer interface {
	Store(path string, data map[string]interface{}) error
}

type sanitizer struct {
	orig       yaml.MapSlice
	out        yaml.MapSlice
	vault      storer
	vaultPath  string
	shouldMove func(item yaml.MapItem) bool
}

// Load initializes the sanitizer with the specified yaml data.
func (s *sanitizer) Load(yml []byte) error {
	return yaml.Unmarshal(yml, &s.orig)
}

// Run processes each input param.
func (s *sanitizer) Run() error {
	for _, item := range s.orig {
		if s.shouldMove(item) {
			k := fmt.Sprintf("%v", item.Key)

			// write to "value" property unless param uses ((property.varname)) syntax
			key := "value"
			if i := strings.Index(k, "."); i != -1 {
				key = k[:i]
				k = k[i+1:]
			}

			p := path.Join(s.vaultPath, k)
			err := s.vault.Store(p, map[string]interface{}{
				key: item.Value,
			})
			if err != nil {
				return fmt.Errorf("could not write to Vault at %s: %v", p, err)
			}
		} else {
			s.out = append(s.out, item)
		}
	}
	return nil
}

// Write dumps the data that was not stored to Vault back to disk.
func (s *sanitizer) Write(path string) error {
	b, err := yaml.Marshal(s.out)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}
