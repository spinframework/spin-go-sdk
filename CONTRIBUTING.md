## Generating the WIT bindings

Whenever WIT files are changed, added to, or removed from the `wit` directory, the bindings  in `internal` should be regenerated.

### Prerequisites

- BASH or compatible shell
- [**componentize-go**](https://github.com/bytecodealliance/componentize-go) - Latest version

### Run
```sh
bash regenerate-bindings.sh
```
