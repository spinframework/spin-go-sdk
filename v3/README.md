WIP

spin-go-sdk with wasip2 support

Notes:

The current version of tooling used for this work:

- wit-bindgen-go `wit-bindgen-go version (devel) (0e3b31e354b31b4f2d5e7d5163e29fb2ffd0b052)`
- wasm-tools `wasm-tools 1.227.1`
- tinygo `tinygo version 0.37.0 darwin/arm64 (using go version go1.24.1 and LLVM version 19.1.2)`
- spin `spin 3.2.0-pre0 (3d07b0cb 2025-03-14)`
- go `go version go1.24.1 darwin/arm64`
- binaryen tools `binaryen-version_123`


Regeneratin bindings:

- install tooling as specified above
- make sure they are on PATH and picking up the versions as specified above
- cd `<root>/v3`
- Run: `wit-bindgen-go generate -w http-trigger -p github.com/spinframework/spin-go-sdk/v3/internal --out internal ./wit`

Testing:

- cd `<root>/v3/examples/http`
- Run `spin build`
- Run `spin up`
- In a separate terminal, run: `curl http://127.0.0.1:3000/hello`
