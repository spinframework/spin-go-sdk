# Overview
The `wasip2` implementation of the Spinframework Go SDK.

## Generating the WIT bindings
Whenever WIT files are changed/added to the `v3/wit` directory, the bindings  in `v3/wit_component` need to be regenerated.

### Prerequisites
- [**componentize-go**](https://github.com/bytecodealliance/componentize-go) - Latest version

### Run
```sh
cd v3

# Delete all non-handwritten code
find $(pwd)/internal/ \
    -mindepth 1 \
    -maxdepth 1 \
    -type d \
    ! -name 'db' \
    ! -name 'export_wasi_http_0_2_0_incoming_handler' \
    -exec rm -rf {} +

componentize-go -w http-trigger -d ./wit bindings --format -o internal --pkg-name github.com/spinframework/spin-go-sdk/v3/internal
```
