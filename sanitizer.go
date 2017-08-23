package main

import (
	"fmt"
	"io"
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
	verbose    io.Writer
}

// Load initializes the sanitizer with the specified yaml data.
func (s *sanitizer) Load(yml []byte) error {
	s.out = nil
	return yaml.Unmarshal(yml, &s.orig)
}

// Run iterates through each input param, either storing the data
// in Vault, or recording it as non-sensitive.
func (s *sanitizer) Run() error {
	for _, item := range s.orig {
		if !s.shouldMove(item) {
			s.out = append(s.out, item)
			continue
		}

		k := fmt.Sprintf("%v", item.Key)

		// write to "value" property unless param uses ((property.varname)) syntax
		vaultKey := "value"
		if i := strings.Index(k, "."); i != -1 {
			vaultKey = k[:i]
			k = k[i+1:]
		}

		p := path.Join(s.vaultPath, k)
		err := s.vault.Store(p, map[string]interface{}{
			vaultKey: item.Value,
		})
		if err != nil {
			return fmt.Errorf("could not write to Vault at %s: %v", p, err)
		}
		if s.verbose != nil {
			fmt.Fprintf(s.verbose, "wrote %s\n", p)
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
