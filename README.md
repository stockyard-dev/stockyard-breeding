# Stockyard Breeding

**Self-hosted pedigree tracking and breeding records**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/breeding/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9803:9803 -v breeding_data:/data ghcr.io/stockyard-dev/stockyard-breeding
```

Open `http://localhost:9803` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9803` | HTTP port |
| `DATA_DIR` | `./breeding-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
