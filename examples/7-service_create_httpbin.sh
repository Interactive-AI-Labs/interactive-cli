#!/usr/bin/env bash

interactiveai services create httpbin \
  --project betsson-poc \
  --port 80 \
  --image-type external \
  --image-name httpbin \
  --image-tag latest \
  --endpoint \
  --image-repository kennethreitz
