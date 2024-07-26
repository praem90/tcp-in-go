package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Client struct {
    Name string
    Conn *net.Conn
}

var Clients []Client
var scanner *bufio.Scanner

func main() {
    tcp, err := net.Listen("tcp", "127.0.0.1:8990")

    if err != nil {
        panic(err)
    }
    fmt.Println("Listening and waiting for clients")

    scanner = bufio.NewScanner(os.Stdin)

    go func() {
        for {
            conn, err := tcp.Accept()

            if err != nil {
                panic(err)
            }

            go onConnect(conn);
        }
    }()

    fmt.Println("What do you want to do?")
    for scanner.Scan() {
        fmt.Println("What do you want to do?")

        cmd := scanner.Text()

        switch {
        case cmd == "help":
            fmt.Println("List of commands: \nlist, use <client id from the list> \nhelp")
        case cmd == "list":
            fmt.Printf("List of clients attached: %d\n", len(Clients))
            for i, client := range(Clients) {
                fmt.Printf(" %d) %s  from %s\n", i+1, client.Name, (*client.Conn).RemoteAddr().String())
            }
        case regexp.MustCompile(`^use [0-9]+`).MatchString(cmd):
            fmt.Printf("List of clients attached: %d\n", len(Clients))
            for i, client := range(Clients) {
                fmt.Printf(" %d) %s  from %s\n", i+1, client.Name, (*client.Conn).RemoteAddr().String())
            }

            if index, err := strconv.Atoi(strings.TrimLeft(cmd,"use ")); err == nil {
                if index > len(Clients) {
                    println("Invlaid index")
                    continue
                }

                useClient(index - 1)
            } else {
                println(err.Error())
            }


        default:
            fmt.Println("Try valid commands")
        }

    }
}

func onConnect(conn net.Conn) {
    client := Client{
        Conn: &conn,
    }

    conn.Write([]byte("Hey, Who are you?\n"));
    name := make([]byte, 100)

    if bytesReceived, err := conn.Read(name); err == nil && bytesReceived > 0 {
        client.Name = string(name[:bytesReceived-2])
        fmt.Printf("%s has been joined\n", client.Name)
        conn.Write([]byte(fmt.Sprintf("Hi %s, Nice to meet you!!\n", client.Name)))
        Clients = append(Clients, client)
        return
    }

    conn.Write([]byte("Sorry, I could not recognize you. :(\n"))

    conn.Close()
}

func useClient(index int) {
    client := Clients[index]
    fmt.Printf("Interactive with the client %s\n", client.Name)

    dir := os.TempDir()
    connScanner := bufio.NewScanner(*client.Conn)

    for scanner.Scan() {

        cmd := scanner.Text()

        switch {
        case cmd == "help":
            fmt.Println("Available commands: \n ls \n cd <dir> \n get <file>")
        case cmd == "ls":
            (*client.Conn).Write([]byte("ls"))

            for {
                connScanner.Scan()
                txt := connScanner.Text()
                if (txt == "") {
                    break
                }
                fmt.Println(txt)
            }
        case regexp.MustCompile(`^get `).MatchString(cmd):
            (*client.Conn).Write([]byte(cmd))
            connScanner.Scan()
            filename := connScanner.Text()
            fmt.Printf("received filename %s\n", filename)
            connScanner.Scan()
            if size, err := strconv.Atoi(connScanner.Text()); err == nil {
                fmt.Printf("Creating tmp file: %s \n", filepath.Join(dir, filename))
                file, err := os.Create(filepath.Join(dir, filename))
                if err != nil {
                    println(err.Error())
                    continue
                }
                buf := make([]byte, 1024)
                chunk := size/1024
                for i:=0; i<=chunk; i++ {
                    if read, err := (*client.Conn).Read(buf); read > 0 && err == nil {
                        fmt.Printf("Read chunk %d of %d", i, chunk)
                        file.WriteAt(buf, int64(i * 1024))
                    } else {
                        println("Could not read file buffer")
                    }
                }
                file.Close()
            }
        case regexp.MustCompile(`^dir `).MatchString(cmd):
            dir = strings.TrimLeft(cmd, "dir ")
        case regexp.MustCompile(`^cd `).MatchString(cmd):
            (*client.Conn).Write([]byte(cmd))
        case cmd == "exit" || cmd == "q":
            break
        default:
            fmt.Println("Try help to know available options")
        }
    }
}
