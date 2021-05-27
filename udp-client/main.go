package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"time"
)

const MAX = 10

func main() {
	SendUDPRequest("localhost:1024", 1000, func(c context.Context, conn *net.UDPConn, id int, ch chan<- int) {
		fmt.Printf("goroutine %d running...\n", id)
		ctx, cancel := context.WithDeadline(c, time.Now().Add(3*time.Second))
		defer cancel()
		buf := make([]byte, MAX)

		// read random characters into buffer
		if n, err := rand.Read(buf); err != nil || n != MAX {
			quit(fmt.Errorf("random reading error: %w %d\n", err, n))
			ch <- 1
			return
		}

		done := make(chan error, 1)
		go func() {
			if n, err := conn.Write(buf); err != nil || n != MAX {
				done <- err
				quit(fmt.Errorf("id: %d, writing to connection error, %w\n", id, err))
				return
			}

			// use the same buffer to recieve from conn
			if n, _, err := conn.ReadFrom(buf); err != nil || n != MAX {
				done <- err
				quit(fmt.Errorf("id: %d, reading from connection error: %w\n", id, err))
				return
			} else {
				fmt.Printf("goroutine %d | msg: %s\n", id, string(buf))
			}
			done <- nil
		}()

		var err error
		select {
		case <-ctx.Done():
			fmt.Printf("goroutine %d cancelled\n", id)
			err = ctx.Err()
		case err = <-done:
			if err != nil {
				quit(fmt.Errorf("goroutine %d quiting... %w\n", id, err))
			}
			ch <- 1
			return
		}
		ch <- 1
		return
	})
}

// client reads arbitrary characters from systems random device and sends it as
// a message to the server. This is done fast and concurrently.

// SendUDPRequest launches N gouroutines sending requests concurrently to the
// the server at listening at addr.
func SendUDPRequest(addr string, N int, f func(context.Context, *net.UDPConn, int, chan<- int)) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		quit(err)
		return
	}

	// to use the same connection to send all the request or use just the one.
	//
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		quit(err)
		return
	}
	ctx := context.Background()
	done := make(chan int, 1)
	// ohh yes we have to wait on all the goroutines to finish before we quit
	// this function. Shit!
	for i := 0; i < N; i++ {
		go f(ctx, conn, i, done)
	}

	// wait for all goroutines to finish or quit
	var counter int
	select {
	case <-done:
		counter++
		if counter == N {
			fmt.Println("done!")
			return
		}
	}
}

func quit(err error) {
	fmt.Fprintln(os.Stderr, err)
	//os.Exit(1)
}
