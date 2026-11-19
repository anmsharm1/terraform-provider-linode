---
page_title: "Linode: Managed Database Provider Development"
description: |-
  Local setup, code structure, and test workflows for managed database provider development.
---

# Managed Database Provider Development

This guide explains how to work on managed database support in the Linode
Terraform provider. It is for developers who know managed databases but are
new to Terraform.

## Terraform in brief

Terraform reads configuration written in HashiCorp Configuration Language
(HCL). The configuration describes the resources that Terraform must manage.
The Linode provider is a plugin. It translates Terraform operations into
Linode API requests.

A **resource** creates and manages an object. A **data source** reads an
object or a catalog without managing its lifecycle. Terraform records managed
resources in a state file. The state links each resource in the configuration
to the real Linode object.

The common workflow is:

1. `terraform plan` shows the changes Terraform intends to make.
2. `terraform apply` performs those changes and updates state.
3. `terraform destroy` deletes the resources managed by that configuration.

Always review a plan before you apply it. Keep state files private. They can
contain credentials and other sensitive values.

## How the code is organized

The provider uses both the older SDKv2 and the newer Terraform Plugin
Framework. New managed database support uses the Plugin Framework.

- `linode/databasemysqlv2/` implements the MySQL cluster resource and data
  source.
- `linode/databasepostgresqlv2/` implements the PostgreSQL cluster resource
  and data source.
- `linode/databases/` implements the cross-engine database listing.
- `linode/databaseengines/` lists database engine versions and types.
- `linode/databaseaccesscontrols/` is an older SDKv2 resource for database
  allow lists.
- `linode/helper/databaseshared/` holds shared engine, maintenance, private
  network, and suspension behavior.
- `linode/framework_provider.go` registers Plugin Framework resources and
  data sources with Terraform.

An engine package keeps its schema, model, resource logic, data source logic,
tests, and HCL test templates together. The schema defines Terraform fields.
The model converts between Terraform state and linodego API types. The
resource implements create, read, update, and delete operations.

Read [AGENTS.md](../../AGENTS.md) before you change the code. It documents
repository patterns and commands.

## Development tools

Install these tools for the work you plan to do:

- **Git** to check out the provider and its test submodule.
- **Go 1.25.8 or later** to build the provider. The provider's `go.mod`
  declares Go 1.25.8.
- **Make** to run the repository commands.
- **golangci-lint v2** to run `make format`, `make lint`, and the formatting
  check used by acceptance tests.
- **Terraform 1.0 or later** only if you want to run a local Terraform
  configuration against your development build.
- **Docker** if you want to run `make docs-check`.

Go downloads the modules listed in `go.mod`. It does not require Terraform to
run the provider's unit tests.

### Use the inflight linodego change

Provider code calls the linodego client. A released client version must
contain the API types and methods used by the provider. To develop against
an inflight linodego change, use a Go workspace. Do not add a local `replace`
directive or a machine-specific `go.work` file to the provider pull request.

Run these commands from the provider repository root. Change
`LINODEGO_REPO` to the path of your linodego checkout:

```sh
PROVIDER_REPO="$PWD"
LINODEGO_REPO="$HOME/Workspace/dx-tools/linodego"
DEV_DIR="$HOME/.cache/linode-provider-dev"

mkdir -p "$DEV_DIR"
go -C "$DEV_DIR" work init "$PROVIDER_REPO" "$LINODEGO_REPO"
export GOWORK="$DEV_DIR/go.work"
```

Create the workspace once. If `go.work` already exists, do not run `go work
init` again. Use the existing workspace and check that it contains both
module paths. The inflight linodego checkout may require a newer Go toolchain
than the provider module. Go can download that toolchain automatically when
`GOTOOLCHAIN=auto` is enabled.

Run provider Go commands from the provider repository with `GOWORK` set. For
example:

```sh
go build ./...
make test-unit
```

Unset `GOWORK` to test against the dependency version selected by the
provider's `go.mod` instead:

```sh
unset GOWORK
```

## Build and test without Linode access

Start with checks that do not call the Linode API:

```sh
go build ./...
make test-unit
```

Run unit tests for one package by setting `PKG_NAME` to its directory under
`linode/`:

```sh
make PKG_NAME=databasemysqlv2 test-unit
make PKG_NAME=databasepostgresqlv2 test-unit
make PKG_NAME=databases test-unit
```

`make build` runs the repository formatter and then builds the provider.
Formatting can change files. Review `git diff` after it runs. The CI workflow
also checks documentation, lint, module tidiness, and unit tests. See the
[Makefile](../../Makefile) and [CI workflow](../../.github/workflows/ci.yml)
for the current commands.

## Run acceptance tests

Acceptance tests send requests to a real Linode account. They can create
billable database instances. Use an account and region approved for testing.
For limited-availability engines, confirm that the account has access before
you start.

Acceptance tests need a Linode API token with access for the operations under
test. In Bash, enter it at a hidden prompt:

```sh
read -r -s -p "Linode API token: " LINODE_TOKEN
export LINODE_TOKEN
printf '\n'
```

Do not commit the token, print it in logs, or include it in a command that
you save in shell history. The Makefile does not load a repository `.env`
file automatically.

The test target initializes the `e2e_scripts` Git submodule when required. It
also runs a script that generates IPv4 and IPv6 test variables. If setup
fails, initialize the submodule explicitly:

```sh
git submodule update --init --recursive
```

Run one test or one database package:

```sh
make PKG_NAME=databasemysqlv2 TEST_CASE=TestAccResourceDatabaseMySQLV2_basic test-int
make TEST_SUITE=databasemysqlv2 test-int
```

The exact test names can change. Check the package's `*_test.go` files before
you use a test name. The default acceptance test target can run many tests and
take a long time. Use a package and test name to limit the first run.

Review the test output and account before and after a live run. The test suite
should remove resources that it creates. If a test is interrupted, identify
and remove only that test's resources. Do not delete resources that existed
before the test. The `make sweep` target can destroy infrastructure; do not
run it unless you understand its target and scope.

## Test a local provider with Terraform

Terraform CLI development overrides tell Terraform to run your local provider
binary instead of a registry download. Build the binary into a stable
directory:

```sh
DEV_PROVIDER_DIR="$HOME/.local/share/terraform-provider-linode-dev"
mkdir -p "$DEV_PROVIDER_DIR"
go build -o "$DEV_PROVIDER_DIR/terraform-provider-linode" .
```

Create a Terraform CLI configuration, for example at
`$HOME/.terraformrc-linode-dev`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/linode/linode" = "/home/USER/.local/share/terraform-provider-linode-dev"
  }
  direct {}
}
```

Replace `/home/USER` with your home directory. The path in the configuration
must be absolute. Set this environment variable in the terminal where you
run Terraform:

```sh
export TF_CLI_CONFIG_FILE="$HOME/.terraformrc-linode-dev"
```

The override affects Terraform commands in that terminal. Rebuild the binary
after each code change. Terraform does not use the development override for
`terraform init`; `init` resolves the provider requirement in the Terraform
configuration. After initialization, normal commands such as `plan` and
`apply` use the local binary. To inspect provider schemas without creating
resources, run `terraform providers schema -json` in a test directory that
declares the `linode/linode` provider.

Keep manual testing in a separate directory. For example, create `main.tf`
with a provider requirement and one MySQL resource:

```hcl
terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "~> 4.0"
    }
  }
}

provider "linode" {}

resource "linode_database_mysql_v2" "dev" {
  label     = "replace-with-a-unique-test-label"
  region    = "replace-with-a-supported-region"
  type      = "replace-with-a-supported-database-plan"
  engine_id = "mysql/8"
}
```

Replace the label, region, plan type, and engine ID with values supported by
the test account. The `linode_database_engines` data source can list available
engine IDs. The `type` is the managed database plan, not a database engine.

From that test directory, run:

```sh
terraform init
terraform validate
terraform plan
```

`init` prepares the working directory and resolves the provider requirement.
`validate` checks the configuration. `plan` shows proposed changes; it does
not create the database. Review the plan, then run `terraform apply` and
confirm the prompt. After apply, run `terraform plan` again and check for
unwanted changes. Run `terraform destroy` when the test is complete. Destroy
also asks for confirmation. Do not remove state as a way to clean up a live
database; Terraform uses state to find the resource that it must delete.

For a Valkey API that requires the beta API version, set the provider's
documented environment variable in the same terminal:

```sh
export LINODE_API_VERSION=v4beta
```

Valkey availability depends on the Linode account. Restore times are a list
of snapshots, not arbitrary point-in-time values. A disabled persistence
setting prevents backups and forks. Verify the current API and test-account
behavior before you run a live restore or suspension test.

Unset the development settings when you finish:

```sh
unset TF_CLI_CONFIG_FILE
unset GOWORK
unset LINODE_API_VERSION
unset LINODE_TOKEN
```

## Submit a change

Keep a pull request focused on one feature or fix. Add unit tests for model
and conversion behavior. Add acceptance tests for changes that call the
Linode API. Update the provider docs when you add or change a resource or
data source.

The CI workflow expects pull request titles in this format:

```text
TPT-1234: Short description
```

Some repository labels exempt a pull request from this rule. Follow the
[contribution guide](../../CONTRIBUTING.md) for the fork and pull process,
tests, commits, and signatures.
