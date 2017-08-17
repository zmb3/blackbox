# blackbox

A tool for moving secrets from Concourse params files into Vault.

## Installation

`$ go get -u github.com/zmb3/blackbox`

## Usage

First, set the `VAULT_ADDR` and `VAULT_TOKEN` environment variables.

Run blackbox with:

```
$ blackbox -in params.yml -out sanitized.yml -path concourse/myteam/mypipeline
```

For each parameter, the tool will ask you whether or not you would like to move it
to Vault.  To accept the default value (no), simply press enter.  To move the param
to Vault, enter `y` and press enter.

When the tool completes, it will write a new YML file containing only the non-sensitive
Values that were not moved to vault.
