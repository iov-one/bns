# `bnsapi` Blockchain Name Service REST API

`bnsapi` is a proxy all requests to `bnsd`. `bnsapi` is using JSON for payload
serialization and REST for API.

This service is intended to provide very little logic and should be save to be
exposed to the public.


This application follows [12 factor app](https://12factor.net/) principles as
close as possible.

- Logs are written to stdout.
- Configuration is done via environment variables.


# Configuration

To configure `bnsapi` instance use environment variables.

- `HTTP` - the address and the port that the HTTP server listens on
- `TENDERMINT` - the address of the Tendermint API that should be used for data
  queries. For example `https://rpc-private-a-vip-mainnet.iov.one` for the main
  net and http://0.0.0.0:26657 for local instance.
- `HOST_PORT` - HostPort is used for swagger docs configuration

## API

Each listing result set is limited to only a certain amount of entries. Each
result can be paginated using `offset=<key>`. Offset is inclusive.

Each listing result can be filtered using at most one filter at a time.
`offset` is not a filter.

## Docs

To see documentation:
- Run the project `make run`
- Go to `localhost:$PORT/docs`
