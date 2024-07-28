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
    tcp, err := net.Listen("tcp", ":8990")

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
    name := make([]byte, 200)

    if bytesReceived, err := conn.Read(name); err == nil && bytesReceived > 0 {
        client.Name = string(name[:bytesReceived])
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
    connReader := bufio.NewReader(*client.Conn)

    for scanner.Scan() {
        cmd := scanner.Text()

        switch {
        case cmd == "help":
            fmt.Println("Available commands: \n ls \n cd <dir> \n get <file>")
        case cmd == "ls":
            (*client.Conn).Write([]byte("ls\n"))

            for {
                txt, err := connReader.ReadString('\n')
                if err != nil {
                    println(err.Error())
                    break
                }
                // connScanner.Scan()
                // txt := connScanner.Text()
                if (txt == "" || txt == "EOF\n") {
                    fmt.Println("End of file list")
                    break
                }
                if (txt == "exit") {
                    (*client.Conn).Close()
                }
                fmt.Println(txt)
            }
            fmt.Println("End of ls")
        case regexp.MustCompile(`^get `).MatchString(cmd):
            (*client.Conn).Write([]byte(fmt.Sprintf("%s\n", cmd)))
            connScanner.Scan()
            filename := connScanner.Text()
            fmt.Printf("received filename %s\n", filename)
            connScanner.Scan()
            sizeBuf := connScanner.Text()
            fmt.Printf("File size string %s", sizeBuf)
            if size, err := strconv.Atoi(sizeBuf); err == nil {
                fmt.Printf("Creating tmp file: %s for the size %d\n", filepath.Join(dir, filename), size)
                file, err := os.Create(filepath.Join(dir, filename))
                if err != nil {
                    println(err.Error())
                    continue
                }
                buf := make([]byte, 1024)
                received := 0
                i := 0
                for {
                    i++
                    if read, err := (*client.Conn).Read(buf); read > 0 && err == nil {
                        // fmt.Printf("Last 6 bytes %s", string(buf[read-6:read]))
                        // if (string(buf[read-6:read]) == "EOFEOF") {
                        //     fmt.Print("\nGot the EOFEOF\n")
                        //     break
                        // }

                        received += read

                        fmt.Printf("Read %d chunk %d, total size: %d\n", read, i, received)
                        file.Write(buf[0:read])

                        if received >= size {
                            fmt.Println("Received all the data")
                            (*client.Conn).Write([]byte("ACK\n"))
                            break
                        }
                    } else {
                        println("Could not read file buffer")
                        break
                    }
                }
                fmt.Println("\nFile saved successfully")
                file.Close()
            } else {
                (*client.Conn).Write([]byte("ERR"))
                println(err.Error())
                println(sizeBuf)
            }
        case regexp.MustCompile(`^dir `).MatchString(cmd):
            dir = strings.TrimLeft(cmd, "dir ")
        case regexp.MustCompile(`^cd `).MatchString(cmd):
            msg := fmt.Sprintf("%s\n", cmd)
            fmt.Println(msg)
            (*client.Conn).Write([]byte(msg))
        case cmd == "exit" || cmd == "q":
            return
        default:
            fmt.Println("Try help to know available options")
        }

        fmt.Println("Waiting for input")
    }
}
