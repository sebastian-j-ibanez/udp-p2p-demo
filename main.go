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
	time.Sleep(time.Millisecond * 100) // Wait for broadcast to fully stop

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
	broadcastIP, err := getBroadcastAddr()
	if err != nil {
		fmt.Printf("error getting broadcast address: %s\n", err.Error())
		return
	}
	broadcastAddr := &net.UDPAddr{IP: broadcastIP, Port: DefaultPort}
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for {
		select {
		case <-stopCast:
			return
		case <-ticker.C:
			data := []byte(strconv.Itoa(c.Id))
			_, err := c.Conn.WriteToUDP(data, broadcastAddr)
			if err != nil {
				fmt.Printf("error: %s", err.Error())
			}
		}
	}
}

// getBroadcastAddr finds the broadcast address for the local network
func getBroadcastAddr() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue // Skip IPv6
			}

			// Calculate broadcast address: IP | ^Mask
			broadcast := make(net.IP, len(ip))
			for i := range ip {
				broadcast[i] = ip[i] | ^ipNet.Mask[i]
			}
			return broadcast, nil
		}
	}

	return nil, fmt.Errorf("no suitable network interface found")
}
