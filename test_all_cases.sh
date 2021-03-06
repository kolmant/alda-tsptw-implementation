#!/bin/bash

for f in testdata/*.txt; do go run main.go "${f##*/}" 50000; done