#!/bin/bash

golangci-lint run --config ./.golangci.yml

markdownlint --fix ./README.md