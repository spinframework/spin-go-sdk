# gRPC

This runs the Google 'route guide' sample in Spin. The service code is taken from the Google example, with some Spin prologue material.

## Running the sample

```
spin up --build
```

## Testing with grpcurl

> This sample turns on reflection, so grpcurl does not need the proto files.

Unary call:

```
grpcurl -plaintext -d '{"latitude":413069058,"longitude":-744597778}' localhost:3000 routeguide.RouteGuide/GetFeature
```

Server-streaming call:

```
grpcurl -plaintext -d '{"lo":{"latitude":412000000,"longitude":-800000000},"hi":{"latitude":414000000,"longitude":-700000000}}' localhost:3000 routeguide.RouteGuide/ListFeatures
```

Client-streaming call:

```
grpcurl -plaintext -d '{"latitude":413069058,"longitude":-744597778}{"latitude":413169058,"longitude":-744592778}{"latitude":414008389,"longitude":-743951297}' localhost:3000 routeguide.RouteGuide/RecordRoute
```

Client- and server-streaming:

```
grpcurl -plaintext -d '{"location":{"latitude":413069058,"longitude":-744597778},"message":"hello from america!"}{"location":{"latitude":414008389,"longitude":-743951297},"message":"hello again from america!"}' localhost:3000 routeguide.RouteGuide/RouteChat
```

> The RouteChat feature in this implementation doesn't yet persist messages, so you'll need to hit the endpoint several times within the idle instance timeout, or you'll only see your own messages. (See comment in source code.) The default timeout is 1 second, but you can pass `--idle-instance-timeout 600s` to enjoy a more persistent chat experience!
