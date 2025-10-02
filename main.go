package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func main() {
	client, err := NewClient(rand.Intn(10000))
	if err != nil {
		panic(fmt.Sprintf("Unable to init client: %s", err.Error()))
	}

	client.Run()
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
	client, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return Client{}, nil
	}

	return Client{Id: id, Conn: client}, nil
}

// Run main client loop
func (c *Client) Run() error {
	defer c.Conn.Close()

	stopCast := make(chan bool)
	go c.Broadcast(stopCast)

	// Try to read from Socket
	msg := make([]byte, 64)
	n, addr, err := c.Conn.ReadFromUDP(msg)
	if err != nil {
		return err
	}

	peerId, err := strconv.Atoi(string(msg[:n]))
	if err != nil {
		return err
	}

	// Stop broadcasting when connection is made
	stopCast <- true

	// Check to see which
	if peerId < c.Id {
		c.Initiate(addr)
	} else {
		c.Respond(addr)
	}

	return nil
}

func (c *Client) Initiate(addr *net.UDPAddr) error {
	// Send ACK to peer
	msg := make([]byte, 64)
	msg = fmt.Appendf(msg, "ACK %d", c.Id)
	_, err := c.Conn.WriteToUDP(msg, addr)
	if err != nil {
		return err
	}

	// Receive ACK from peer
	clear(msg)
	_, _, err = c.Conn.ReadFromUDP(msg)
	if err != nil {
		return err
	}

	// Send data to peer
	clear(msg)
	msg = fmt.Appendf(msg, "secret message from client %d", c.Id)
	_, err = c.Conn.WriteToUDP(msg, addr)
	if err != nil {
		return err
	}

	// Receive data from peer
	clear(msg)
	n, _, err := c.Conn.ReadFromUDP(msg)
	if err != nil {
		return err
	}
	fmt.Printf("Received: \"%s\"\n", string(msg[:n]))

	return nil
}

func (c *Client) Respond(addr *net.UDPAddr) error {
	// Receive ACK from peer
	msg := make([]byte, 64)
	_, _, err := c.Conn.ReadFromUDP(msg)
	if err != nil {
		return err
	}

	// Send ACK to peer
	clear(msg)
	msg = fmt.Appendf(msg, "ACK %d", c.Id)
	_, err = c.Conn.WriteToUDP(msg, addr)
	if err != nil {
		return err
	}

	// Receive data from peer
	clear(msg)
	n, _, err := c.Conn.ReadFromUDP(msg)
	if err != nil {
		return err
	}
	fmt.Printf("Received: \"%s\"\n", string(msg[:n]))

	// Send data to peer
	clear(msg)
	msg = fmt.Appendf(msg, "secret message from client %d", c.Id)
	_, err = c.Conn.WriteToUDP(msg, addr)
	if err != nil {
		return err
	}

	return nil
}

// Broadcast data until stop flag is received
func (c *Client) Broadcast(stopCast chan bool) {
	broadcastAddr := &net.UDPAddr{IP: net.IPv4bcast, Port: DefaultPort}
	for {
		select {
		case stop := <-stopCast:
			if stop {
				return
			}
		default:
			data := []byte(strconv.Itoa(c.Id))
			_, err := c.Conn.WriteToUDP(data, broadcastAddr)
			if err != nil {
				fmt.Printf("error: %s", err.Error())
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}
