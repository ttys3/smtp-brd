#!/bin/bash

golangci-lint run --config ./.golangci.yml

# fixup markdown
markdownlint --fix ./README.md

# fixup go format
gofumpt -w .