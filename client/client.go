package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/blueberryserver/blueproto/msg"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// 서버 주소 설정
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// 신규 연결 인스턴스 생성
	c := pb.NewGreeterClient(conn)

	// 명령 인자 체크
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	// 타임 아웃 2000ms 설정
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	// 연결 인스턴스로 부터 SayHelloClient 획득
	sayhelloClient, err := c.SayHello(ctx)
	if err != nil {
		log.Println(err)
	}

	// 파라미터 함께 SayHello 요청
	go func() {
		err := sayhelloClient.Send(&pb.HelloRequest{Name: name})
		if err != nil {
			log.Println(err)
		}
	}()

	// 응답 받기 완료 까지 대기 채널
	done := make(chan bool)

	// SayHello 응답
	go func() {
		m, err := sayhelloClient.Recv()
		if err == io.EOF {
			log.Println(err)
			return
		}
		// 받은 응답 메시지 출력
		log.Printf("Greeting: %s", m.GetMessage())

		// 연결 종료
		if err := sayhelloClient.CloseSend(); err != nil {
			log.Println(err)
		}

		// 대기 200ms
		<-time.After(time.Millisecond * 200)

		// 응답 대기 완료
		done <- true
	}()

	// 응답 완료 여부및 타임 아웃 체크
	go func() {
		select {
		// 응답 완료 시 리턴
		case <-done:
			return
		// 타임 아웃 시
		case <-ctx.Done():
			// 애러 코드 출력
			if err := ctx.Err(); err != nil {
				log.Println(err)
			}
		}
	}()

	<-done
}
