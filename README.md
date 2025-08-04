## oosexclude
Remove outofscope subdomains from https://github.com/rix4uni/scope/blob/main/data/outofscope.txt or a local outofscope.txt file.

## Installation
```
go install github.com/rix4uni/oosexclude@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/oosexclude/releases/download/v0.0.2/oosexclude-linux-amd64-0.0.2.tgz
tar -xvzf oosexclude-linux-amd64-0.0.2.tgz
rm -rf oosexclude-linux-amd64-0.0.2.tgz
mv oosexclude ~/go/bin/oosexclude
```
Or download [binary release](https://github.com/rix4uni/oosexclude/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/oosexclude.git
cd oosexclude; go install
```

## Usage
```
Usage of oosexclude:
  -e, --exclude-list string   Path to exclude list file or URL (default "https://raw.githubusercontent.com/rix4uni/scope/refs/heads/main/data/outofscope.txt")
      --verbose               enable verbose mode
  -v, --version               Print the version of the tool and exit.
```

## Usage Examples
```bash
# Uses the default exclude list URL
cat allsubs.txt | oosexclude

# Specify a custom exclude list file
cat allsubs.txt | oosexclude -e outofscope.txt

# Specify a custom exclude list URL
cat allsubs.txt | oosexclude -e https://example.com/custom_outofscope.txt
```

## Output Examples

Given:
```
allsubs.txt:
community.myfitnesspal.com
community-stage.myfitnesspal.com
img.allin.movilepay.com
dashboard.rapyd.net
argocd.test.uidapi.com
techdev.ibotta.com
exchange.bullish.com
```

With:
```
outofscope.txt:
community*.myfitnesspal.com
*.allin.movilepay.com
*.starsoft.movilepay.com
community.rapyd.net
argocd.*.uidapi.com
*dev.ibotta.com
*.bullish.com
```

Command:
```sh
cat allsubs.txt | oosexclude -e outofscope.txt --verbose
```

Output:
```
IGNORED: community.myfitnesspal.com
IGNORED: community-stage.myfitnesspal.com
IGNORED: img.allin.movilepay.com
NOT IGNORED: dashboard.rapyd.net
IGNORED: argocd.test.uidapi.com
IGNORED: techdev.ibotta.com
IGNORED: exchange.bullish.com
```

Without `-verbose`:
```
dashboard.rapyd.net
```