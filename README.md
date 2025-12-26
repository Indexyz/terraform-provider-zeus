# Terraform Provider Zeus

Terraform provider for interacting with the Zeus API (see `docs/zeus.md`).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Provider Configuration

```hcl
provider "zeus" {
  endpoint = "http://<host>:<port>"
  token    = "<bearer token>"
}
```

## Supported Resources

- `zeus_pool`
- `zeus_assign`

## Supported Data Sources

- `zeus_pool` (lookup by `id`)
- `zeus_assign` (lookup by `id`)

## Provider Functions

- `provider::zeus::ipv4_ip2long(string)` converts dotted IPv4 to integer.
- `provider::zeus::ipv4_long2ip(int)` converts integer to dotted IPv4.

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate` (requires Terraform installed for formatting examples).

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
