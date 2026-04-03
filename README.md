## The Go SDK for Spin

This is an SDK for developing [Spin](https://github.com/spinframework/spin) applications using the Go programming language.

> Note: This SDK temporarily relies on an unreleased version of [componentize-go](https://github.com/bytecodealliance/componentize-go). For the time being, please install the [canary build](https://github.com/bytecodealliance/componentize-go/releases/tag/canary).

## Example

```go
package main

import (
	"fmt"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Spin!")
	})
}

func main() {}
```

See the [examples](./examples) directory for more examples.
