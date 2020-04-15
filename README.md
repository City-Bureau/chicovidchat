# Chicago COVID Chat Bot

[![Build status](https://github.com/City-Bureau/chicovidchat/workflows/Deploy/badge.svg)](https://github.com/City-Bureau/chicovidchat/actions)

Find updated, verified information on resources in the Chicago area during the Coronavirus pandemic through SMS. Maintained by [City Bureau](https://www.citybureau.org/).

## Setup

You'll need [GNU Make](https://www.gnu.org/software/make/), [Go](https://golang.org/) and [node.js](https://nodejs.org/en/) installed as well as credentials for AWS, Twilio and Airtable. Your Airtable base will need to have the fields in [`pkg/directory/resource.go`](./pkg/directory/resource.go).

Copy `.env.sample` to `.env` and fill in the values.

To install dependencies, build functions and deploy:

```bash
make install
make build
make deploy
```

## Development

We use `gofmt` for formatting code and `golangci-lint` for linting. Run each of these commands with:

```bash
make format
make lint
```

You can also run all tests with:

```bash
make test
```
