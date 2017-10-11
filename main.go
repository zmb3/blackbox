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

func ensureEnv(varname string) {
	if ev := os.Getenv(varname); ev == "" {
		fatalf("required environment variable %q is missing", varname)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(2)
}

func main() {
	ensureEnv("VAULT_TOKEN")
	ensureEnv("VAULT_ADDR")

	input := flag.String("in", "", "the input params file")
	output := flag.String("out", "", "the sanitized output params file")
	path := flag.String("path", "", "the base Vault path to write to (eg: concourse/myteam/mypipeline)")
	all := flag.Bool("all", false, "move all params to Vault (don't prompt for each)")
	verbose := flag.Bool("verbose", false, "display verbose output")
	flag.Parse()

	if *input == "" || *path == "" || (!*all && *output == "") {
		flag.Usage()
		os.Exit(2)
	}

	inBytes, err := ioutil.ReadFile(*input)
	if err != nil {
		fatalf("could not read params file: %v\n", err)
	}

	client, err := vault.NewClient(nil)
	if err != nil {
		fatalf("could not connect to Vault: %v\n", err)
	}
	s := sanitizer{
		vaultPath: *path,
		vault:     &vaultStorer{client: client},
		shouldMove: func(item yaml.MapItem) bool {
			if *all {
				return true
			}
			fmt.Printf("move %s? (n): ", item.Key)
			var choice string
			fmt.Scanf("%s\n", &choice)
			return strings.HasPrefix(strings.ToLower(choice), "y")
		},
	}

	if *verbose {
		s.verbose = os.Stdout
	}

	// read input yml
	err = s.Load(inBytes)
	if err != nil {
		fatalf("could not parse input yml: %v\n", err)
	}

	// process params
	err = s.Run()
	if err != nil {
		fatalf("%v\n", err)
	}

	// write output yml
	if *output != "" {
		err = s.Write(*output)
		if err != nil {
			fatalf("could not write output yml: %v\n", err)
		}
	}
}
