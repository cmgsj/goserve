# goserve

HTTP file server

## Install

### Go

```bash
go install github.com/cmgsj/goserve@latest
```

### GitHub

<https://github.com/cmgsj/goserve/releases/latest>

```bash
os="darwin"
arch="arm64"
version=""

if [[ -z "${version}" ]]; then
    version="$(curl -sSL "https://api.github.com/repos/cmgsj/goserve/releases/latest" | jq -r '.tag_name' | sed 's/^v//')"
fi

curl -sSLo /tmp/goserve.tar.gz "https://github.com/cmgsj/goserve/releases/download/v${version}/goserve_${version}_${os}_${arch}.tar.gz"

tar xzvf /tmp/goserve.tar.gz -C /tmp

rm -f /tmp/goserve.tar.gz

chmod +x /tmp/goserve

sudo mv /tmp/goserve /usr/local/bin
```

## Demo

```bash
$ goserve /tmp/folder
#
#    __________  ________  ______   _____
#   / __  / __ \/ ___/ _ \/ ___/ | / / _ \
#  / /_/ / /_/ (__  )  __/ /   | |/ /  __/
#  \__, /\____/____/\___/_/    |___/\___/
# /____/
#
#
# Starting HTTP file server
#
# Config:
#   Root:       /tmp/folder
#   Host:       0.0.0.0
#   Port:       80os="darwin"
#   Log Level:  info
#   Log Format: text
#   Log Output: stderr
#
# Routes:
#   GET /                -> Redirect /files
#   GET /files           -> List Files
#   GET /files/{file...} -> List Files
#   GET /health          -> Health Check
#
# Listening at http://localhost:80
#
# Ready to accept connections
#
```
