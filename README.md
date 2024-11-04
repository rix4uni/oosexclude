## oosexclude
Remove outofscope subdomains from a updated or a local file.

## Installation
```
go install github.com/rix4uni/oosexclude@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/oosexclude/releases/download/v0.0.1/oosexclude-linux-amd64-0.0.1.tgz
tar -xvzf oosexclude-linux-amd64-0.0.1.tgz
rm -rf oosexclude-linux-amd64-0.0.1.tgz
mv oosexclude ~/go/bin/oosexclude
```
Or download [binary release](https://github.com/rix4uni/oosexclude/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/oosexclude.git
cd oosexclude; go install
```

## Usage Examples
```
Usage of oosexclude:
  -exclude-list string
        Path to exclude list file or URL (default "https://raw.githubusercontent.com/rix4uni/scope/refs/heads/main/outofscope.txt")
```

## Usage Examples
```bash
# Uses the default exclude list URL
cat urls.txt | oosexclude

# Specify a custom exclude list file
cat urls.txt | oosexclude -exclude-list outofscope.txt

# Specify a custom exclude list URL
cat urls.txt | oosexclude -exclude-list https://example.com/custom_outofscope.txt
```