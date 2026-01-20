# Requirements
- [**go**](https://go.dev/dl/) - v1.25+
- [**spin**](https://github.com/spinframework/spin) - Latest version
- [**docker**](https://docs.docker.com/get-started/get-docker/) - Latest version
- [**componentize-go**](https://github.com/asteurer/componentize-go) - Latest version

# Usage
In one terminal window, run:
```sh
# Note that the `-d` flag is intentionally omitted
docker compose up
```

In another terminal, you'll run your Spin app:
```sh
spin up --build
```

In yet another terminal, you'll interact with the Spin app:
```sh
curl localhost:3000/publish
```

You will see logs appear in the `docker compose` window that look something like this:
```sh
$ docker compose up
...
broker      | 1754324646: New connection from 172.18.0.1:36970 on port 1883.
broker      | 1754324646: New client connected from 172.18.0.1:36970 as client001 (p2, c1, k30, u'user').
subscriber  | telemetry Eureka!
broker      | 1754324646: Client client001 closed its connection.
```
