#!/bin/bash

# Remove any previously-generated import bindings:
find $(pwd)/imports/ \
  -mindepth 1 \
  -maxdepth 1 \
  -type d \
  ! -name "deps" \
  -exec rm -rf {} +

# Generate bindings for all supported worlds.  We'll only use the imports from
# these bindings, discarding the exports.
componentize-go \
  --ignore-toml-files \
  -w "wasi:http/service@0.3.0-rc-2026-03-15" \
  -w "fermyon:spin/http-trigger@3.0.0" \
  -w "fermyon:spin/redis-trigger" \
  -d wit \
  bindings \
  --format \
  -o imports \
  --pkg-name github.com/spinframework/spin-go-sdk/v3/imports

# For each supported world, generate bindings specific to that world.  We'll use
# only the exports from these bindings, defering to the imports we generated
# above.
for world in \
  "wasi:http/service@0.3.0-rc-2026-03-15" \
  "fermyon:spin/http-trigger@3.0.0" \
  "fermyon:spin/redis-trigger"
do
  rm -rf tmp
  dir=exports/$(echo $world | sed 's+[:/@.-]+_+g')
  rm -rf $dir/wit_exports
  componentize-go \
    --ignore-toml-files \
    -w "$world" \
    -d wit \
    bindings \
    --format \
    -o tmp \
    --export-pkg-name github.com/spinframework/spin-go-sdk/v3/$dir \
    --pkg-name github.com/spinframework/spin-go-sdk/v3/imports
  cp -r tmp/wit_exports $dir/
  rm -rf tmp
done

