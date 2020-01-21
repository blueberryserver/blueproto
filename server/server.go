package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/blueberryserver/bluecore/bluegrpc"
	pb "github.com/blueberryserver/blueproto/msg"
)

const (
	port = ":50051"
)

// 서버 인터페이스
type server struct{}

// SayHello 구현
func (s server) SayHello(srv pb.Greeter_SayHelloServer) error {
	// 컨텍스트 획득
	ctx := srv.Context()

	for {
		// 종료 처리
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 데이터 Recv 되면
		req, err := srv.Recv()
		if err == io.EOF {
			// 연결 종료 시
			log.Println("disconnect")
			return nil
		}

		// error 반화 시
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}

		// 획득한 메시지 값 출력
		log.Printf("Received: %v", req.GetName())

		// 응답 생성
		resp := pb.HelloReply{Message: "Hello " + req.GetName()}

		// 응답 전송
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

func main() {
	fmt.Println("Server Start")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 신규 GRPServer 인스턴스 생성
	s := bluegrpc.NewGRPCServer()

	// 서버 객체 획득
	grpcserver := s.GetServer()

	// 서버 등록
	pb.RegisterGreeterServer(grpcserver, server{})

	// 서비스 시작(블록킹)
	if err := grpcserver.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
