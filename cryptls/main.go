package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
)

func main() {
	Listen(":443", ConfigTLS("scert", "skey"), func(c *tls.Conn) {
		fmt.Fprintln(c, "hello world")
	})
}

// Listen uses a tls connection with underlying tcp connection
// for communication
func Listen(a string, conf *tls.Config, f func(*tls.Conn)) {
	if listener, err := tls.Listen("tcp", a, conf); err == nil {
		for {
			if connection, err := listener.Accept(); err == nil {
				go func(c *tls.Conn) {
					defer c.Close()
					f(c)
				}(connection.(*tls.Conn))
			}
		}
	}
}

func ConfigTLS(c, k string) (r *tls.Config) {
	if cert, e := tls.LoadX509KeyPair(c, k); e == nil {
		r = &tls.Config{
			Certificates: []tls.Certificate{cert},
			Rand:         rand.Reader,
		}
	}
	return
}
