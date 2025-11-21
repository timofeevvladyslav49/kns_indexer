# KNS Indexer - Keeta Name Service Indexer

**KNS Indexer** is a fully decentralized, open-source indexer for **Keeta Name Service (KNS)** - an ENS-like naming
system built on the Keeta blockchain.

Unlike traditional centralized indexers, KNS Indexer **does not custody or intermediate any data**. All name
registrations, transfers, and metadata updates happen directly on-chain. The indexer only listens to events and stores
them in a queryable database, meaning **anyone can run their own indexer instance** and get exactly the same data as
the "official" one.

## Key Features

- **100% on-chain operations** - no off-chain dependencies or trusted parties
- Token-based usernames - enables true atomic swaps and composability with DeFi
- Decentralized resolution - names can resolve not only to blockchain addresses but also to IPFS/IPNS content hashes
- Fully reproducible - run your own node and query API without relying on third parties
- Easy deployment via Docker Compose

### Current Tech Stack

- **Indexer**: Python (planning migration to Go for better performance)
- **Database**: PostgreSQL

## Why Token-Based Names Matter

Every KNS name is a token living on the Keeta blockchain. This design unlocks:

- Instant trustless trading via atomic swaps
- Seamless integration with wallets and DeFi applications

## Roadmap

| Status | Feature                  | Description                                                                |
|--------|--------------------------|----------------------------------------------------------------------------|
| ⏳      | Primary name             | Set and resolve primary name per address                                   |
| ⏳      | IPFS content resolution  | Enable decentralized websites                                              |
| ⏳      | Public REST API          | Allow dApps and wallets to resolve names and search without running a node |
| ⏳      | Comprehensive test suite | Unit, integration, and e2e tests                                           |
| ⏳      | Full migration to Go     | Significant performance boost and lower resource usage                     |

## Quick Start

```shell
git clone https://github.com/yourusername/kns_indexer.git
cd kns_indexer
cp .env.example .env
docker compose up
```

## Run Your Own - Be Truly Decentralized

There is no "official" indexer. You are the infrastructure.

Any instance you run will give you and your users the same data as everyone else - because everything is verified
on-chain.

## Contributing

Contributions are very welcome! Feel free to open issues, submit PRs, or suggest new features.