## oosexclude
Filter subdomains using exclude (`--egrep`) or include (`--grep`) pattern lists, with support for glob wildcards and full regular expressions.

## Installation
```
go install github.com/rix4uni/oosexclude@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/oosexclude/releases/download/v0.0.4/oosexclude-linux-amd64-0.0.4.tgz
tar -xvzf oosexclude-linux-amd64-0.0.4.tgz
rm -rf oosexclude-linux-amd64-0.0.4.tgz
mv oosexclude ~/go/bin/oosexclude
```
Or download [binary release](https://github.com/rix4uni/oosexclude/releases) for your platform.

## Compile from source
```
git clone --depth 1 https://github.com/rix4uni/oosexclude.git
cd oosexclude; go install
```

## Usage
```yaml
Usage of oosexclude:
      --egrep string   Path to exclude list file or URL (default "https://raw.githubusercontent.com/rix4uni/scope/refs/heads/main/data/outofscope.txt")
      --grep string    Path to include list file or URL
      --ignore-case    Match patterns case-insensitively
      --stats          Print filtering stats to stderr after processing
      --version        Print the version of the tool and exit.
```

## Usage Examples
```yaml
# Uses the default exclude list URL
cat allsubs.txt | oosexclude

# Specify a custom exclude list file
cat allsubs.txt | oosexclude --egrep match-list.txt

# Specify a custom include list file
cat allsubs.txt | oosexclude --grep match-list.txt

# Single inline pattern
cat allsubs.txt | oosexclude --egrep "v[1-9].hack.com"
cat allsubs.txt | oosexclude --grep "v[1-9].hack.com"

# Multiple inline patterns (comma-separated)
cat allsubs.txt | oosexclude --egrep "v[1-9].hack.com, argocd.*.uidapi.com, *dev.ibotta.com"
cat allsubs.txt | oosexclude --grep "v[1-9].hack.com, argocd.*.uidapi.com, community.rapyd.net"

# Case-insensitive matching
cat allsubs.txt | oosexclude --egrep match-list.txt --ignore-case
cat allsubs.txt | oosexclude --grep match-list.txt --ignore-case

# Show filtering stats
cat allsubs.txt | oosexclude --egrep match-list.txt --stats
```

## Output Examples

Given:
```yaml
allsubs.txt:
_acme-challenge.hack.com
api.dev-us.hack.com
api.hack.com
auth-v2.hack.com
_autodiscover.hack.com
beta-login.hack.com
cdn-assets.hack.com
client-portal.hack.com
db-admin.internal.hack.com
dev-api.hack.com
edge-eu-west.hack.com
files.backup.hack.com
grafana.monitoring.hack.com
img.cdn.hack.com
jenkins-ci.hack.com
k8s-master01.hack.com
mail01.hack.com
mobile-app.hack.com
mta-sts.hack.com
node-1.cluster.hack.com
pre-prod.hack.com
s3-upload.hack.com
secure.payments.hack.com
shop.api-v2.hack.com
smtp-relay.hack.com
staging.api.hack.com
static-v1.hack.com
test123.hack.com
uat.portal.hack.com
vpn-gateway.hack.com
auth-v2.hack.com
shop.api-v2.hack.com
static-v1.hack.com
community.myfitnesspal.com
community-stage.myfitnesspal.com
img.allin.movilepay.com
dashboard.rapyd.net
argocd.test.uidapi.com
techdev.ibotta.com
exchange.bullish.com
```

With:
```yaml
match-list.txt:
v[1-9].hack.com
community*.myfitnesspal.com
*.allin.movilepay.com
*.starsoft.movilepay.com
community.rapyd.net
argocd.*.uidapi.com
*dev.ibotta.com
*.bullish.com
```

Command (`--egrep` removes lines matching any pattern):
```yaml
cat allsubs.txt | oosexclude --egrep match-list.txt
```

Output:
```yaml
_acme-challenge.hack.com
api.dev-us.hack.com
api.hack.com
_autodiscover.hack.com
beta-login.hack.com
cdn-assets.hack.com
client-portal.hack.com
db-admin.internal.hack.com
dev-api.hack.com
edge-eu-west.hack.com
files.backup.hack.com
grafana.monitoring.hack.com
img.cdn.hack.com
jenkins-ci.hack.com
k8s-master01.hack.com
mail01.hack.com
mobile-app.hack.com
mta-sts.hack.com
node-1.cluster.hack.com
pre-prod.hack.com
s3-upload.hack.com
secure.payments.hack.com
smtp-relay.hack.com
staging.api.hack.com
test123.hack.com
uat.portal.hack.com
vpn-gateway.hack.com
dashboard.rapyd.net
```

Command (`--grep` keeps only lines matching any pattern, matched part highlighted in terminal):
```yaml
cat allsubs.txt | oosexclude --grep match-list.txt
```

Output:
```yaml
auth-v2.hack.com
shop.api-v2.hack.com
static-v1.hack.com
auth-v2.hack.com
shop.api-v2.hack.com
static-v1.hack.com
community.myfitnesspal.com
community-stage.myfitnesspal.com
img.allin.movilepay.com
argocd.test.uidapi.com
techdev.ibotta.com
exchange.bullish.com
```

Command (`--stats` prints a summary to stderr):
```yaml
cat allsubs.txt | oosexclude --egrep match-list.txt --stats
```

Stderr output:
```yaml
[stats] input: 40  kept: 28  removed: 12
```
