package main

import (
	"fmt"
	"log"
	"net"
)

// Taken from https://stackoverflow.com/a/37382208
func main() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Printf("%v", localAddr.IP)
}
