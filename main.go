package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var port int
var index int = 1

func init() {
	flag.IntVar(&port, "port", 8000, "Port number: -port 800")
	flag.Parse()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Exit trapped...")
		os.Exit(1)
	}()
}

func main() {
	addr := "localhost:" + strconv.Itoa(port)
	go startConnect(addr)
	select {}
}

func startConnect(addr string) {
	con, err := waitForServer(addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer con.Close()

	log.Println("Got connection")

	go startReading(con)

	if _, err := io.Copy(os.Stdout, con); err != nil {
		log.Fatal(err)
	}

	log.Println("Finished sending to peer")
}

func startReading(con net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter text: ")
		text, _ := reader.ReadString('\n')
		con.Write([]byte(text))
	}
}

func waitForServer(addr string) (net.Conn, error) {
	const timeout = 1 * time.Minute
	deadline := time.Now().Add(timeout)

	for tries := 0; time.Now().Before(deadline); tries++ {
		con, err := net.Dial("tcp", addr)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second << uint(tries))
			waitForServer(addr)
		} else {
			return con, nil
		}
	}

	return nil, errors.New("Unable to contact server after " + string(timeout))
}
