package main

import (
	"fmt"
	"math/rand"
	"net"
)

func main() {
	a, err := NewClient(rand.Intn(100))
	if err != nil {
		panic(fmt.Sprintf("Unable to init client: %s", err.Error()))
	}

	a.RunClient()
}

const DefaultPort = 50000

type Client struct {
	Id   int
	Conn *net.UDPConn
}

func NewClient(id int) (Client, error) {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: DefaultPort,
	}
	client, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return Client{}, nil
	}

	return Client{Id: id, Conn: client}, nil
}

// Run main client loop
func (c *Client) RunClient() error {
	defer c.Conn.Close()

	stopCast := make(chan bool)
	go c.Broadcast(stopCast)

	// Try to read from Socket
	var msg string
	_, addr, err := c.Conn.ReadFromUDP([]byte(msg))
	if err != nil {
		return err
	}

	// Stop broadcasting and close socket when
	// connection is made.
	stopCast <- true
	c.Conn.Close()

	c.Conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	// Send ACK to peer
	msg = fmt.Sprintf("ACK %d", c.Id)
	_, err = c.Conn.Write([]byte(msg))
	if err != nil {
		return err
	}

	// Receive ACK from peer
	msg = ""
	_, err = c.Conn.Read([]byte(msg))
	if err != nil {
		return err
	}

	// Send data to peer
	msg = "secret message"
	_, err = c.Conn.Write([]byte(msg))
	if err != nil {
		return err
	}

	// Receive data from peer
	msg = ""
	_, err = c.Conn.Read([]byte(msg))
	if err != nil {
		return err
	}
	fmt.Printf("Received: \"%s\"\n", msg)

	return nil
}

// Broadcast data until stop flag is received
func (c *Client) Broadcast(stopCast chan bool) {
	for {
		select {
		case stop := <-stopCast:
			if stop {
				return
			}
		default:
			data := fmt.Sprintf("CLT %d", c.Id)
			_, err := c.Conn.Write([]byte(data))
			if err != nil {
				fmt.Printf("error: %s", err.Error())
			}
		}
	}
}
