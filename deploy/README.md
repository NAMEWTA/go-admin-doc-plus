# Deployment Engineering

`deploy/` is the only owner of runtime deployment definitions for the product. Compose,
container configuration templates and deployment verification live below this directory;
product code and CI workflows consume them through root Taskfile commands.

Deployment definitions must reference credentials through declared secret files or secret
providers. Do not commit runtime credentials, generated secret files or local deployment data.

The legacy deployment entries still located below `go-admin-plus/` and
`go-admin-ui-plus/` are frozen migration inputs. Their exact removal inventory is maintained
in `scripts/go-admin-plus/legacy-governance-t21.txt`; T-21 removes them after all consumers
use the root command plane.
