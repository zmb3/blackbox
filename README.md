# blackbox

A tool for moving secrets from Concourse params files into Vault.

## Installation

Download a [release](https://github.com/zmb3/blackbox/releases) for your platform, or:

```
$ go get -u github.com/zmb3/blackbox
```

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

### Example

```sh
$ cat params.yml
secret1: password1
param1: param1
secret2: password2
username: admin

$ ./blackbox -in params.yml -out params2.yml -path secret/pipeline
move secret1? (n): y
move param1? (n):
move secret2? (n): y
move username? (n):

$ cat params2.yml
param1: param1
username: admin
$ vault list secret/pipeline
Keys
----
secret1
secret2

$ vault read secret/pipeline/secret1
Key             	Value
---             	-----
refresh_interval	768h0m0s
value           	password1
```
