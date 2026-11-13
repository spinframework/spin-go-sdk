# Requirements
- [**go**](https://go.dev/dl/) - v1.25+
- [**spin**](https://github.com/spinframework/spin) - Latest version

# Usage
The `middleware` component is built against the `spin:up/http-middleware@4.1.0`
world and uses `spinhttp.Next` to forward requests to the `app` component.

Build and run the Spin app:
```sh
spin up --build
```

In another terminal window, you can interact with the Spin app:
```sh
curl -i localhost:3000/hello?name=go
curl -i localhost:3000/blocked
```
