package main

import (
	"fmt"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

const yml = `
param1: value1
param2: value2
secret1: password1
customKey.secret2: password2
`

type storerFunc func(path string, data map[string]interface{}) error

func (s storerFunc) Store(path string, data map[string]interface{}) error {
	return s(path, data)
}

func TestLoad(t *testing.T) {
	var s sanitizer
	err := s.Load([]byte(yml))
	if err != nil {
		t.Fatal(err)
	}
	if l := len(s.orig); l != 4 {
		t.Errorf("want 4 params, got %d", l)
	}
}

func TestLoadError(t *testing.T) {
	var s sanitizer
	err := s.Load([]byte(`this is not valid yml`))
	if err == nil {
		t.Error("expected error but didn't get one")
	}
}

func TestRun(t *testing.T) {
	var s sanitizer
	err := s.Load([]byte(yml))
	if err != nil {
		t.Fatal(err)
	}

	s.vaultPath = "concourse/myteam/mypipeline"

	// record all vault data to a map
	vaultData := make(map[string]map[string]interface{})
	s.vault = storerFunc(func(path string, data map[string]interface{}) error {
		vaultData[path] = data
		return nil
	})

	// only move params containing 'secret'
	s.shouldMove = func(item yaml.MapItem) bool {
		k := fmt.Sprintf("%v", item.Key)
		return strings.Contains(k, "secret")
	}

	err = s.Run()
	if err != nil {
		t.Fatal(err)
	}

	// verify that only non-sensitive data didn't move
	if l := len(s.out); l != 2 {
		t.Errorf("expected 2 non-sensitive params to stay on disk, got %d", l)
	}
	for i := range s.out {
		k := fmt.Sprintf("%v", s.out[i].Key)
		if strings.HasPrefix(k, "secret") {
			t.Errorf("param %q should have been moved to vault but stayed in yml", k)
		}
	}

	// verify that secrets did move
	if l := len(vaultData); l != 2 {
		t.Errorf("expected 2 secrets to move to vault, got %d", l)
	}

	tt := []struct {
		path  string
		key   string
		value string
	}{
		{"concourse/myteam/mypipeline/secret1", "value", "password1"},
		{"concourse/myteam/mypipeline/secret2", "customKey", "password2"},
	}
	for _, tc := range tt {
		data, ok := vaultData[tc.path]
		if !ok {
			t.Errorf("expected secret in Vault at %s, but got none", tc.path)
			continue
		}
		val := data[tc.key]
		if val != tc.value {
			t.Errorf("expected <%s=%s> at %s, got <%s=%v>",
				tc.key, tc.value, tc.path, tc.key, val)
		}
	}
}
