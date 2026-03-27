# Spin component in Go using the Redis trigger

```shell
$ spin build --up
```

```shell
$ redis-cli
127.0.0.1:6379> PUBLISH messages test-message
(integer) 1
```
