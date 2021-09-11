`mass-exec`

run scripts (concurrently) in lambda

# What is it

It's a [docker image lambda](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html)
([`./cmd/mass-exec-runtime`](./cmd/mass-exec-runtime)) that calls `os.Exec`.

# Getting started

## Install the CLI

```bash
go install github.com/kantatenbot/mass-exec/cmd/mass-exec
```

## Deploy

```bash
; cd terraform
; mv example.tfvars terraform.tfvars
# set vars in `terraform.tfvars`, then
; terraform init
; terraform apply
```

## Update the Lambda

- M1 Mac users: `./scripts/update.sh`.
- Everyone else: update the image tag in your terraform variables, then `terraform apply`. The
  docker provider handles the image rebuild.
  - I mean, you could use the script too. it's faster, just not iac

## Usage

### Run a script

```bash
; mass-exec run --input="" 'echo hello mass-exec'
run id 1632881136
{"argv":[],"run_id":"1632881136","bucket":"mass-exec-example","key":"runs/1632881136","status":0,"errors":[]}
```

### Run a script and print the output

```bash
; mass-exec run --input="" 'echo hello mass-exec' | mass-exec receive -D
run id 1632881081
hello mass-exec
```

### Run a script with inputs. Inputs are passed to your script as argv:

```bash
# you can use args in your script. note the single quotes
; echo https://google.com | mass-exec run 'curl -q $1 >/dev/null && echo $1 OK' | mass-exec receive -D
run id 1632881114
https://google.com OK
```

### Run a script from a file:

```bash
# example.sh
message=OK
check_url() {
  curl -q $1 >/dev/null && echo $message
}

check_url $@
```

```bash
; echo https://google.com OK | mass-exec run -f example.sh | mass-exec receive -D
https://google.com OK
```

If you add a shebang, you can use other languages:

```python
#!/usr/bin/env python3
import sys

import requests

print(requests.get(sys.argv[1]).headers)
```

# FAQs that I've contrived

## What tools does mass-exec have

BYOT. The Dockerfile is a starting point

## can you show me a better demo than echo and curl

no, but here's a pseudoscript that hopefully illustrates my motivation for making this

```bash
; sc shodan get-host-ports-by-service redis \
  | mass-exec run 'list-redis-keys $1' \
  | mass-exec receive -o data
```

# Troubleshooting

### I use an M1 Mac. I deployed with Terraform, but I get no output from commands run in the lambda. My cloudwatch logs show `command err, fork/exec /usr/bin/bash: exec format error`.

Related: I use an M1 Mac and I get a segfault building the image.

**tl;dr**
Change line 1 of your dockerfile to this

```dockerfile
FROM --platform=aarch64 golang:alpine as builder
```

use `./scripts/update.sh` to update your lambda image.

**explanation**

- [A qemu bug](https://github.com/docker/for-mac/issues/5123) causes some operations in M1 docker to segfault, including go build when statically linking certain libraries
- We can work around this by cross-compiling to `linux/amd64` in a multi-stage, multi-platform build
- But the terraform docker provider doesn't support this
  - it let's you pick a single platform for the whole build, but doesn't understand --platform flags in a multi-stage build atm
  - the lambda module doesn't let you pick your target platform anyway
- So we have to do it outside of terraform
