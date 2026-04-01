## (Raw) WASIp3 HTTP Service Example

This example demonstrates how to do full-duplex streaming and concurrent
outbound requests using the WASIp3 HTTP bindings directly rather than using
"net/http" types (which are not yet supported for WASIp3 use as of this
writing).

### Building and Running

#### Prerequisites

- [componentize-go](https://github.com/bytecodealliance/componentize-go/)
- Curl

```shell
spin build --up
```

While that's running, you can send a request from another shell:

```
curl -i http://127.0.0.1:3000/hello
```

If all goes well, you should see `hello, world!`.

You can also try the other endpoints, e.g. `/echo`, which does full-duplex
streaming:

```
curl -i -H 'content-type: text/plain' --data-binary @- http://127.0.0.1:3000/echo <<EOF
’Twas brillig, and the slithy toves
      Did gyre and gimble in the wabe:
All mimsy were the borogoves,
      And the mome raths outgrabe.
EOF
```

...and `/hash-all`, which concurrently downloads one or more URLs and streams the
SHA-256 hashes of their contents:

```
curl -i \
    -H 'url: https://webassembly.github.io/spec/core/' \
    -H 'url: https://www.w3.org/groups/wg/wasm/' \
    -H 'url: https://bytecodealliance.org/' \
    http://127.0.0.1:3000/hash-all
```

