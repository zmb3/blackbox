// Command blackbox moves Concourse params to Vault.
// It requires the VAULT_ADDR and VAULT_TOKEN environment variables to be set.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	input := flag.String("in", "", "the input params file")
	output := flag.String("out", "", "the sanitized output params file")
	path := flag.String("path", "", "the base vault path to write to (eg: concourse/myteam/mypipeline)")
	flag.Parse()

	if *input == "" || *output == "" || *path == "" {
		flag.Usage()
		os.Exit(2)
	}

	inBytes, err := ioutil.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read params file: %v\n", err)
		os.Exit(2)
	}

	s := sanitizer{vaultPath: *path}

	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to Vault: %v\n", err)
		os.Exit(2)
	}
	s.vault = &vaultStorer{client: client}

	// read input yml
	err = s.Load(inBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't parse input yml: %v\n", err)
		os.Exit(2)
	}

	// process params
	s.shouldMove = func(item yaml.MapItem) bool {
		fmt.Printf("move %s? (n): ", item.Key)
		var choice string
		fmt.Scanf("%s\n", &choice)
		return strings.HasPrefix(strings.ToLower(choice), "y")
	}

	err = s.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// write output yml
	err = s.Write(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write output yml: %v\n", err)
		os.Exit(2)
	}
}
