# Valkey Terraform support: local testing

This guide is for Managed Database developers. Use it to test the local provider
code against the live Linode API.

## What changed

The provider now supports:

* `linode_database_valkey_v2`: create, update, import, and delete a Valkey database.
* `linode_database_valkey_v2` data source: read an existing database.
* `linode_database_valkey_config`: read the engine configuration catalog.
* Valkey entries in database and engine listings, and Valkey access controls.

The implementation follows the MySQL and PostgreSQL provider patterns.
Valkey uses discrete RDB snapshots for forks, not a continuous restore range.

Three Terraform terms are useful:

* **Resource**: an object that Terraform manages through the API.
* **Data source**: an API read that does not create an object.
* **State**: Terraform's record of an object and its ID. Import adds an existing
  object to state; it does not create another database.

## 1. Prepare the checkout

Use Bash. Open the provider repository on the branch that contains this change.
Install Git, Make, Go 1.26 or a compatible later version, and Terraform CLI.
Make sure Go and Terraform are on `PATH`.

```bash
cd /path/to/terraform-provider-linode
go version
terraform version
```

The Go test commands below fetch any missing module dependencies.

Use a development account with Valkey enabled, database write access, and quota
for test databases. VPC tests also need VPC and subnet write access.
The test setup reads the image, region, engine, and plan catalogs.
The access-controls package also reads the MySQL and PostgreSQL engine catalogs.

**Live tests create billable resources.** Fork and resize tests create additional
nodes or databases. Provisioning can take 30 minutes per database.

## 2. Run local unit tests

These tests do not need a token and do not call the live API.

```bash
go test -tags=unit -count=1 \
  ./linode/databasevalkeyv2 \
  ./linode/databasevalkeyconfig \
  ./linode/helper/databaseshared \
  ./linode/databases
```

## 3. Set the token

Enter the development account token without putting it in shell history.
Do not paste tokens into source files, test output, or this guide.
The test runner does not load `.env` automatically.

```bash
read -r -s -p "Linode development token: " LINODE_TOKEN
printf '\n'
export LINODE_TOKEN
export LINODE_API_VERSION=v4beta
export TF_ACC=1
export TF_ACC_TERRAFORM_PATH="$(command -v terraform)"
: "${LINODE_TOKEN:?Set a development account token first}"
```

By default, the client uses the public Linode API. Set `LINODE_URL` only when
your team requires a different API endpoint. Do not enable API request logging:
it can expose credentials.

## 4. Run the live tests

Run this command from the repository root, in the same shell as step 3.

```bash
go test -tags=integration -v -count=1 -p=1 -parallel=1 -timeout=240m \
  -ldflags="-X=github.com/linode/terraform-provider-linode/v4/version.ProviderVersion=acc" \
  -run='^(TestAccResourceDatabaseValkeyV2_.*|TestAccDataSourceDatabaseValkeyV2_.*|TestAccDataSourceDatabaseValkeyConfig_.*|TestAccResourceDatabaseAccessControls_Valkey)$' \
  ./linode/databasevalkeyv2 \
  ./linode/databasevalkeyconfig \
  ./linode/databaseaccesscontrols
```

The tests start the provider from this checkout and run Terraform automatically.
You do not need `terraform init`, a provider installation, or a development override.
This focused command avoids the unrelated formatting and firewall-IP setup used
by `make test-int`.

Coverage includes create/read/delete, import verification, external deletion,
allow lists, maintenance settings, resize, engine version changes, forks,
suspend/resume, engine configuration, VPC changes, listings, and access controls.
Tests choose regions from the API catalog and use plan `g7-dedicated-4-2`.
Region capability does not guarantee capacity. If the API reports that it cannot
allocate resources, confirm regional capacity with the Managed Database team.
The tests do not provide a region environment override.

Import tests verify `engine_id` but exclude the computed `version` from exact
state comparisons. The API can report `8.1` at one read and `8.1.11` at another.
The provider keeps the API-reported version; it does not remove the patch number.

## 5. Check results and cleanup

* Each selected test must report `PASS`. Check every `SKIP`: unavailable Valkey,
  VPC regions, resize plans, or upgrade versions can prevent coverage.
* Tests normally delete their resources, including after a test failure.
  An interrupted run can leave resources behind.
* After the run, check Cloud Manager for test databases, VPCs, and subnets with
  labels that start with `tf_test`. Delete only resources from your run.
  Do not run a broad account sweeper in a shared account.
* Keep Terraform state and test logs private. State can contain database passwords.

```bash
unset LINODE_TOKEN TF_ACC TF_ACC_TERRAFORM_PATH
```

## Validation results (2026-11-02)

Unit tests, acceptance-test compilation, and Markdown lint passed.
Live tests passed for the configuration catalog, access controls including import,
the basic resource, allow-list and maintenance updates, and resource/data-source VPCs.
The affected-test rerun took about 40 minutes.

The full suite is not yet verified. The initial run hit regional capacity errors
for several database creates. The engine-upgrade test was skipped because the
catalog did not offer two versions within one major release.
An additional basic-test run found a version difference during import.
The import checks now exclude `version`, as the PostgreSQL tests do.
This adjustment still needs a live rerun.
