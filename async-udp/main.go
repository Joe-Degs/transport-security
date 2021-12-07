// Asynchronous UDP Server
package main

import (
	"fmt"
	"net"
	"os"
)

const MAX = 1024

func main() {
	ServeUDP(":1024", func(conn *net.UDPConn, addr *net.UDPAddr, msg []byte, id int) {
		if n, err := conn.WriteToUDP(msg, addr); err != nil {
			throwErr(err)
			return
		} else {
			// interesting error i got here. Because I am using the ReadFromUDP,
			// and WriteToUDP to recieve and send messages, the connection is
			// not bound to any single remote host so the conn.RemoteAddr
			// returns a nil interface which i didn't realise at first. took
			// some doing to debug. So i was getting a nil pointer dereference
			// error and was wondering why?. So i tried casting the Addr
			// interface returned to a *net.IPAddr and saw i was trying to cast
			// from nil.
			fmt.Printf("%d) read and wrote %d bytes to %s\n", id, n, addr)
		}
	})
}

func throwErr(err error) {
	fmt.Fprintln(os.Stderr, err)
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
func ServeUDP(addr string, f func(c *net.UDPConn, addr *net.UDPAddr, msg []byte, id int)) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		throwErr(err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		throwErr(err)
		os.Exit(1)
	}
	fmt.Println("listening on ", conn.LocalAddr())
	defer conn.Close()
	buf := make([]byte, MAX)
	var counter int
	for {
		_, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			throwErr(err)
			continue
		}
		counter++
		go f(conn, raddr, buf[:], counter)
	}
}
