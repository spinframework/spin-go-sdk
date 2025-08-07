# Requirements
- Latest version of [TinyGo](https://tinygo.org/getting-started/)
- Latest version of [Docker](https://docs.docker.com/get-started/get-docker/)

# Usage

In one terminal window, run:
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