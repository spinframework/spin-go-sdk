## The Go SDK for Spin

This is an SDK for developing [Spin](https://github.com/spinframework/spin) applications using the Go programming language.

> Note: This SDK temporarily relies on [a fork](https://github.com/dicej/componentize-go) of [componentize-go](https://github.com/bytecodealliance/componentize-go) until [this PR](https://github.com/bytecodealliance/componentize-go/pull/35) has been accepted.  For the time being, please install [this build](https://github.com/dicej/componentize-go/releases/tag/canary).

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
