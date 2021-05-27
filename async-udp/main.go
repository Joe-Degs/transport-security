// Asynchronous UDP Server
package main

import (
	"fmt"
	"net"
	"os"
)

const MAX = 10

func main() {
	ServeUDP(":1024", func(conn *net.UDPConn, addr *net.UDPAddr, msg []byte) {
		//raddr, err := conn.ReadFromUDP(make([]byte, 10))
		//if err != nil {
		//	throwErr(err)
		//}
		//fmt.Fprintln(conn, "it totally works!")

		if _, err := conn.WriteToUDP(msg, addr); err != nil {
			throwErr(err)
		}
		fmt.Println("served ", conn.RemoteAddr)
	})
}

func throwErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// Writing a udp server in go was not trivial. It took me days to literally do
// this. It was not trivial at all and ofcourse i did it lazily soo cool...
// Lovely advice from Eleanor McHugh, concurrency is go in incredibly cheap
// so if you might need two of something just do it. Don't think about it too
// much.
// Things i learnt from this;
//	  UDP is connectionless so you just need one connection to talk to millions
//    of clients, amazing right? Performance wise this is way better than tcp
//    but not as reliable as you would want it to be.
//
// Pitfalls, for a long time, i thought it was the ListenUDP function that would
// be blocking. So i put it in a loop and was getting a socket bind yada-yada
// error. It turns out the blocking function is the ReadFromUDP function and
// that should be the one in the loop.
func ServeUDP(addr string, f func(c *net.UDPConn, addr *net.UDPAddr, msg []byte)) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		throwErr(err)
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		throwErr(err)
	}
	fmt.Println("listening on ", conn.LocalAddr())
	defer conn.Close()
	for {
		buf := make([]byte, MAX)
		_, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			throwErr(err)
		}
		go f(conn, raddr, buf)
	}
}
