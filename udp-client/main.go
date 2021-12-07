package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"time"
)

const MAX = 1024

func main() {
	SendUDPRequest("localhost:1024", 1000,
		func(c context.Context, conn *net.UDPConn, id int, ch chan<- struct{}) {
			fmt.Printf("goroutine %d running...\n", id)
			ctx, cancel := context.WithDeadline(c, time.Now().Add(3*time.Second))
			defer cancel()
			buf := make([]byte, MAX)

			// read random characters into buffer
			if n, err := rand.Read(buf); err != nil {
				printErr(fmt.Errorf("random reading error: %w %d\n", err, n))
				ch <- struct{}{}
				return
			}

			done := make(chan error)
			go func() {
				if n, err := conn.Write(buf); err != nil {
					printErr(fmt.Errorf("id: %d, writing to connection error, %w\n", id, err))
					done <- err
					ch <- struct{}{}
					return
				} else {
					fmt.Printf("goroutine %d | wrote %d bytes to server\n", id, n)
				}

				// use the same buffer to recieve from conn
				if n, _, err := conn.ReadFrom(buf); err != nil {
					printErr(fmt.Errorf("id: %d, reading from connection error: %w\n", id, err))
					done <- err
				} else {
					fmt.Printf("goroutine %d | read %d bytes from server\n", id, n)
					done <- nil
				}
				close(done)
			}()

			select {
			case <-ctx.Done():
				printErr(fmt.Errorf("goroutine %d cancelled | reason: %w\n", id, ctx.Err()))
				ch <- struct{}{}
				return
			case err := <-done:
				if err != nil {
					printErr(fmt.Errorf("goroutine %d quiting due to error... %w\n", id, err))
				} else {
					fmt.Printf("goroutine %d exiting successfully\n", id)
				}
				ch <- struct{}{}
				return
			}
		})
}

// client reads arbitrary characters from systems random device and sends it as
// a message to the server. This is done fast and concurrently.

// SendUDPRequest launches N gouroutines sending requests concurrently to the
// the server at listening at addr.
func SendUDPRequest(addr string, N int, f func(context.Context, *net.UDPConn, int, chan<- struct{})) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		printErr(err)
		return
	}

	// to use the same connection to send all the request or use just the one.
	//
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		printErr(err)
		return
	}
	ctx := context.Background()
	done := make(chan struct{})
	// ohh yes we have to wait on all the goroutines to finish before we quit
	// this function. Shit!
	for i := 0; i < N; i++ {
		go f(ctx, conn, i, done)
	}

	// wait for all goroutines to finish or quit
	var counter int
	for _ = range done {
		counter++
		if counter == N {
			fmt.Println("done!")
			return
		}
	}
}

func printErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	//os.Exit(1)
}
