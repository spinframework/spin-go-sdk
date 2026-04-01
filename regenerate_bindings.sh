#!/bin/bash

# Remove any previously-generated import bindings:
rm -r imports

# Generate bindings for all supported worlds.  We'll only use the imports from
# these bindings, discarding the exports.
componentize-go \
  --ignore-toml-files \
  -w "spin:up/http-trigger@4.0.0" \
  -w "spin:up/redis-trigger@4.0.0" \
  -w "wasi:http/service@0.3.0-rc-2026-03-15" \
  -w "fermyon:spin/http-trigger@3.0.0" \
  -w "fermyon:spin/redis-trigger" \
  -d wit \
  bindings \
  --format \
  -o imports \
  --pkg-name github.com/spinframework/spin-go-sdk/v3/imports \
  --include-versions

rm -r imports/wit_exports

# For each supported world, generate bindings specific to that world.  We'll use
# only the exports from these bindings, defering to the imports we generated
# above.
for world in \
  "spin:up/http-trigger@4.0.0" \
  "spin:up/redis-trigger@4.0.0" \
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
    --pkg-name github.com/spinframework/spin-go-sdk/v3/imports \
    --include-versions
  cp -r tmp/wit_exports $dir/
  rm -rf tmp
done

