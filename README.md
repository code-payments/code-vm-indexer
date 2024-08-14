# Code VM Indexer

[![Release](https://img.shields.io/github/v/release/code-payments/code-vm-indexer.svg)](https://github.com/code-payments/code-vm-indexer/releases/latest)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/code-payments/code-vm-indexer)](https://pkg.go.dev/github.com/code-payments/code-vm-indexer)
[![Tests](https://github.com/code-payments/code-vm-indexer/actions/workflows/test.yml/badge.svg)](https://github.com/code-payments/code-vm-indexer/actions/workflows/test.yml)
[![GitHub License](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/code-payments/code-vm-indexer/blob/main/LICENSE)

Indexer Service for the [Code VM](https://github.com/code-payments/code-vm). It has two main components:
1. Geyser worker that maintains up-to-date state against instances of the Code VM.
2. RPC service to fetch virtual account state, as well as relevant storage metadata.

The Indexer Service is designed to be run independently, enabling anyone to have full self-custodial access to their funds using the next iteration of the Timelock Explorer.

**Important Note**: Only memory storage is currently supported. Compressed storage support will be added at a later date.

## What is Code?

[Code](https://getcode.com) is a mobile wallet app leveraging self-custodial blockchain technology to provide an instant, global, and private payments experience.

## Quick Start

1. Install Go. See the [official documentation](https://go.dev/doc/install).

2. Download the source code.

```bash
git clone git@github.com:code-payments/code-vm-indexer.git
```

3. Run the test suite:

```bash
make test
```

4. Run the RPC service locally:

```bash
DATA_STORAGE_TYPE=memory make run-rpc
```

5. Run the Geyser worker locally:

```bash
DATA_STORAGE_TYPE=memory GEYSER_WORKER_GRPC_PLUGIN_ENDPOINT=localhost:10000 GEYSER_WORKER_VM_ACCOUNT=<VM_PUBLIC_KEY> SOLANA_RPC_ENDPOINT=http://localhost:8899 make run-geyser
```

## Getting Help

If you have any questions or need help, please reach out to us on [Discord](https://discord.gg/T8Tpj8DBFp) or [Twitter](https://twitter.com/getcode).

## Security and Issue Disclosures

In the interest of protecting the security of our users and their funds, we ask that if you discover any security vulnerabilities please report them using this [Report a Vulnerability](https://github.com/code-wallet/code-program-library/security/advisories/new) link.
