package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	ct "github.com/CAIJUNYI/coord/coordtransform"
	"google.golang.org/grpc"
)

var (
	host = flag.String("host", "localhost", "the server address")
	port = flag.String("port", "8008", "the server port")
)

func newPoint(lat, lon float64) *ct.Point {
	return &ct.Point{
		Lat: lat,
		Lon: lon,
	}
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*host+":"+*port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := ct.NewCoordTransformClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pnt := newPoint(37.065, 128.543)
	gcjPnt, err := client.WGS84ToGCJ02(ctx, pnt)
	if err != nil {
		log.Fatalf("failed to transform %v from wgs84 to gcj02", *pnt)
	}
	fmt.Println("gcj02 coordinate:", gcjPnt)

	bdPnt, err := client.WGS84ToBd09(ctx, pnt)
	if err != nil {
		log.Fatalf("failed to transform %v from wgs84 to bd09", *pnt)
	}
	fmt.Println("bd09 coordinate:", bdPnt)
}
