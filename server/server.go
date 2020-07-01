package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"

	ct "github.com/junyicc/coord/coordtransform"
	"google.golang.org/grpc"
)

const (
	xPi = 3.14159265358979324 * 3000.0 / 180.0
	pi  = 3.1415926535897932384626 // π
	a   = 6378245.0                // 长半轴
	ee  = 0.00669342162296594323   // 偏心率平方
)

var (
	host = flag.String("host", "localhost", "the server address")
	port = flag.String("port", "8008", "the server port")
)

func inChina(p *ct.Point) bool {
	return (p.Lon > 73.66 && p.Lon < 135.05 && p.Lat > 3.86 && p.Lat < 53.55)
}

func transformlat(lon, lat float64) float64 {
	ret := -100.0 + 2.0*lon + 3.0*lat + 0.2*lat*lat + 0.1*lon*lat + 0.2*math.Sqrt(math.Abs(lon))
	ret += (20.0*math.Sin(6.0*lon*pi) + 20.0*
		math.Sin(2.0*lon*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*pi) + 40.0*
		math.Sin(lat/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*pi) + 320*
		math.Sin(lat*pi/30.0)) * 2.0 / 3.0
	return ret
}

func transformlon(lon, lat float64) float64 {
	ret := 300.0 + lon + 2.0*lat + 0.1*lon*lon + 0.1*lon*lat + 0.1*math.Sqrt(math.Abs(lon))
	ret += (20.0*math.Sin(6.0*lon*pi) + 20.0*
		math.Sin(2.0*lon*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lon*pi) + 40.0*
		math.Sin(lon/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lon/12.0*pi) + 300.0*
		math.Sin(lon/30.0*pi)) * 2.0 / 3.0
	return ret
}

type coordTransformServer struct {
}

func NewCoordTransformServer() *coordTransformServer {
	return &coordTransformServer{}
}

func (s *coordTransformServer) GCJ02ToWGS84(ctx context.Context, p *ct.Point) (*ct.Point, error) {
	if !inChina(p) {
		return nil, fmt.Errorf("(%f, %f) out of China", p.Lat, p.Lon)
	}
	dlat := transformlat(p.Lon-105.0, p.Lat-35.0)
	dlng := transformlon(p.Lon-105.0, p.Lat-35.0)
	radlat := p.Lat / 180.0 * pi
	magic := math.Sin(radlat)
	magic = 1 - ee*magic*magic
	sqrtmagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((a * (1 - ee)) / (magic * sqrtmagic) * pi)
	dlng = (dlng * 180.0) / (a / sqrtmagic * math.Cos(radlat) * pi)
	mglat := p.Lat + dlat
	mglng := p.Lon + dlng
	return &ct.Point{
		Lon: p.Lon*2 - mglng,
		Lat: p.Lat*2 - mglat,
	}, nil
}
func (s *coordTransformServer) WGS84ToGCJ02(ctx context.Context, p *ct.Point) (*ct.Point, error) {
	if !inChina(p) {
		return nil, fmt.Errorf("(%f, %f) out of China", p.Lat, p.Lon)
	}
	dlat := transformlat(p.Lon-105.0, p.Lat-35.0)
	dlng := transformlon(p.Lon-105.0, p.Lat-35.0)
	radlat := p.Lat / 180.0 * pi
	magic := math.Sin(radlat)
	magic = 1 - ee*magic*magic
	sqrtmagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((a * (1 - ee)) / (magic * sqrtmagic) * pi)
	dlng = (dlng * 180.0) / (a / sqrtmagic * math.Cos(radlat) * pi)
	mglat := p.Lat + dlat
	mglng := p.Lon + dlng
	return &ct.Point{
		Lon: mglng,
		Lat: mglat,
	}, nil
}
func (s *coordTransformServer) Bd09ToGCJ02(ctx context.Context, p *ct.Point) (*ct.Point, error) {
	x := p.Lon - 0.0065
	y := p.Lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPi)
	gcjLon := z * math.Cos(theta)
	gcjLat := z * math.Sin(theta)
	return &ct.Point{
		Lat: gcjLat,
		Lon: gcjLon,
	}, nil
}
func (s *coordTransformServer) GCJ02ToBd09(ctx context.Context, p *ct.Point) (*ct.Point, error) {
	z := math.Sqrt(p.Lon*p.Lon+p.Lat*p.Lat) + 0.00002*math.Sin(p.Lat*xPi)
	theta := math.Atan2(p.Lat, p.Lon) + 0.000003*math.Cos(p.Lon*xPi)
	bdLon := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return &ct.Point{
		Lat: bdLat,
		Lon: bdLon,
	}, nil
}
func (s *coordTransformServer) Bd09ToWGS84(ctx context.Context, p *ct.Point) (*ct.Point, error) {
	gcjPnt, err := s.Bd09ToGCJ02(ctx, p)
	if err != nil {
		return nil, err
	}
	return s.GCJ02ToWGS84(ctx, gcjPnt)
}
func (s *coordTransformServer) WGS84ToBd09(ctx context.Context, p *ct.Point) (*ct.Point, error) {
	gcjPnt, err := s.WGS84ToGCJ02(ctx, p)
	if err != nil {
		return nil, err
	}
	return s.GCJ02ToBd09(ctx, gcjPnt)
}

func main() {
	flag.Parse()

	// new coord transform server
	ctServer := NewCoordTransformServer()

	// new grpc server
	s := grpc.NewServer()
	ct.RegisterCoordTransformServer(s, ctServer)
	handleSignal(s)

	// listen port
	l, err := net.Listen("tcp", *host+":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	for {
		err = s.Serve(l)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}
}

func handleSignal(grpcServer *grpc.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		sig := <-sigCh
		log.Printf("Got signal %v to exit.", sig)
		grpcServer.Stop()
	}()
}
