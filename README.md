# Coordinate Transform

### GRPC Service

- In `coordtransform/coord_transform.proto`, `CoordTransform` is the grpc service that provides the transformation among **WGS84, GCJ02 and BD09** coordinates.
- `Point` is a message representing a localtion with (lat, lon).

### Client

Client calls the grpc service with two optional parameters `host` and `port`

``` go
go run client.go

go run client.go -host=localhost -port=8008
```

### Server

Server runs the grpc service with two optional parameters `host` and `port`

``` go
go run server.go

go run server.go -host=localhost -port=8008
```