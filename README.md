# Chicago COVID Chat Bot

[![Build status](https://github.com/City-Bureau/chicovidchat/workflows/Deploy/badge.svg)](https://github.com/City-Bureau/chicovidchat/actions)

## Setup

You'll need GNU Make, Go and node.js installed as well as credentials for AWS and the Twilio API. Copy `.env.example` to `.env` and fill in the blank values.

To install dependencies, build functions and deploy:

```bash
make install
make build
make deploy
```
