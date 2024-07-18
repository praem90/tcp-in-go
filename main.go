package main

import (
	"fmt"
	"net"
)

func main() {
    tcp, err := net.Listen("tcp", "127.0.0.1:8990")

    if err != nil {
        panic(err)
    }
    fmt.Println("Listening and waiting for clients")

    for {
        conn, err := tcp.Accept()

        if err != nil {
            panic(err)
        }

        go onConnect(conn);
    }
}

func onConnect(conn net.Conn) {
    name := make([]byte, 1024)

    fmt.Printf("We got a new connection from %s", conn.RemoteAddr().String())
    conn.Write([]byte(fmt.Sprintf("Hello %s, What is your name?", conn.RemoteAddr().String())))

    for {
        if bytesReceived, err := conn.Read(name); err == nil && bytesReceived > 0 {
            conn.Write([]byte(fmt.Sprintf("Hi %s, Nice to meet you!!", name)))
        }
    }
}
