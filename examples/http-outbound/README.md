# Making outbound HTTP requests from Go Spin components

The Go SDK for building Spin components allows us to granularly allow components
to send HTTP requests to certain hosts. This is configured in `spin.toml`.

Creating and sending HTTP requests from Spin components closely follows the Go
`net/http` API.  See [hello/main.go](./hello/main.go).

Building this as a WebAssembly module can be done using `spin build`:

```shell
$ spin build
Building component http-to-same-app with `componentize-go --world http-trigger --wit-path ../../../wit build`
Working directory: "./http-to-same-app"
Building component hello with `componentize-go --world http-trigger --wit-path ../../../wit build`
Working directory: "./hello"
Finished building all Spin components
```

The component configuration must contain a list of all hosts allowed to send
HTTP requests to, otherwise sending the request results in an error:

```
Cannot send HTTP request: Destination not allowed: <URL>
```

The `hello` component has the following allowed hosts set:

```toml
[component.hello]
source = "hello/main.wasm"
allowed_outbound_hosts = [
    "https://random-data-api.fermyon.app",
    "https://postman-echo.com",
]
```

And the `outbound-http-to-same-app` uses the dedicated `self` keyword to enable making
a request to another component in this same app, via a relative path (in this case, the component
is `hello` at `/hello`):

```toml
[component.outbound-http-to-same-app]
source = "outbound-http-to-same-app/main.wasm"
# Use self to make outbound requests to components in the same Spin application.
allowed_outbound_hosts = ["http://self"]
```

At this point, we can execute the application with the `spin` CLI:

```shell
$ RUST_LOG=spin=trace,wasi_outbound_http=trace spin up
```

The application can now receive requests on `http://localhost:3000/hello`:

```shell
$ curl -i localhost:3000/hello -X POST -d "hello there"
HTTP/1.1 200 OK
content-length: 976
date: Thu, 26 Oct 2023 18:26:17 GMT

{{"timestamp":1698344776965,"fact":"Reindeer grow new antlers every year"}}
...
```

As well as via the `/outbound-http-to-same-app` path to verify outbound http to the `hello` component:

```shell
$ curl -i localhost:3000/outbound-http-to-same-app
HTTP/1.1 200 OK
content-length: 946
date: Thu, 26 Oct 2023 18:26:53 GMT

{{{"timestamp":1698344813408,"fact":"Some hummingbirds weigh less than a penny"}}
...
```

## Notes

- this only implements sending HTTP/1.1 requests
- requests are currently blocking and synchronous
