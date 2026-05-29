 # Contributing

## Setup

Requires Docker and Go 1.22+.

Clone and run:

make run

This builds the image and starts the server on port 8080.

## Adding a language

1. Install the runtime in the Dockerfile
2. Add a block to config/languages.yaml
3. Run make test
4. Open a PR

No Go code change needed.

## Commit prefixes

Use feat:, fix:, or docs: only.

## Running tests

make test