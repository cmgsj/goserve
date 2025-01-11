# goserve

HTTP file server

## Preview

<https://cmgsj.github.io/goserve>

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
    version="$(curl -fsSL "https://api.github.com/repos/cmgsj/goserve/releases/latest" | jq -r '.tag_name' | sed 's/^v//')"
fi

curl -fsSLo /tmp/goserve.tar.gz "https://github.com/cmgsj/goserve/releases/download/v${version}/goserve_${version}_${os}_${arch}.tar.gz"

tar xzvf /tmp/goserve.tar.gz -C /tmp

rm -f /tmp/goserve.tar.gz

chmod +x /tmp/goserve

sudo mv /tmp/goserve /usr/local/bin
```
