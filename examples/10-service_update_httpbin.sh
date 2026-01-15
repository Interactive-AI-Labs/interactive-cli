#!/usr/bin/env bash
iai services update httpbin \
  --project demo-1 \
  --port 80 \
  --image-type external \
  --image-name httpbin \
  --image-tag latest \
  --endpoint \
  --image-repository kennethreitz
