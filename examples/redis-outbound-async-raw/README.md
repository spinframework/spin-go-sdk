# Requirements
- [**go**](https://go.dev/dl/) - v1.25+
- [**spin**](https://github.com/spinframework/spin) - Latest version
- [**docker**](https://docs.docker.com/get-started/get-docker/) - Latest version
- [**componentize-go**](https://github.com/bytecodealliance/componentize-go) - Latest version

# Usage
In one terminal window, you'll run a Redis container:
```sh
docker run -p 6379:6379 redis:8.2
```

In another terminal, you'll run your Spin app:
```sh
spin up --build
```

In yet another terminal, you'll interact with the Spin app:
```sh
curl localhost:3000
```

You should see the following output:
```
mykey value was: myvalue
spin-go-incr value: 1
deleted keys num: 2
```
