package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	icmpEchoReply   = 0
	icmpEchoRequest = 8
	maxDelay        = 1048576
)

type ICMPHeader struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	ID       uint16
	Seq      uint16
}

func calculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	return uint16(^sum)
}

func handleIP(conn net.PacketConn, addr net.Addr, packets <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	var delay int = 1 // Начальная задержка

	for buffer := range packets {
		var icmpHeader ICMPHeader
		reader := bytes.NewReader(buffer)
		if err := binary.Read(reader, binary.BigEndian, &icmpHeader); err != nil {
			log.Printf("icmp header err: %v", err)
			continue
		}

		if icmpHeader.Type != icmpEchoRequest {
			continue
		}

		log.Printf("Idle %d msec. IP: %s", delay, addr.String())
		time.Sleep(time.Duration(delay) * time.Millisecond)

		icmpHeader.Type = icmpEchoReply
		icmpHeader.Checksum = 0
		responseBuffer := new(bytes.Buffer)
		if err := binary.Write(responseBuffer, binary.BigEndian, &icmpHeader); err != nil {
			log.Printf("icmp header err: %v", err)
			continue
		}

		responseBuffer.Write(buffer[8:])
		data := responseBuffer.Bytes()

		icmpHeader.Checksum = calculateChecksum(data)
		binary.BigEndian.PutUint16(data[2:4], icmpHeader.Checksum)

		_, err := conn.WriteTo(data, addr)
		if err != nil {
			log.Printf("icmp send err: %v", err)
		} else {
			log.Printf("icmp reply sent to %s", addr.String())
		}

		if delay >= maxDelay {
			delay = 1
		} else {
			delay *= 2
		}
	}
}

func main() {
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Fatalf("Socket error: %v", err)
	}
	defer conn.Close()

	log.Println("icmp_exp started")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Shutting down...")
		conn.Close()
		os.Exit(0)
	}()

	buffer := make([]byte, 65536)

	var (
		lock    sync.Mutex
		wg      sync.WaitGroup
		ipChans = make(map[string]chan []byte)
	)

	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Printf("I/O err: %v", err)
			continue
		}

		lock.Lock()
		ch, exists := ipChans[addr.String()]
		if !exists {
			ch = make(chan []byte, 100)
			ipChans[addr.String()] = ch
			wg.Add(1)
			go handleIP(conn, addr, ch, &wg)
		}
		lock.Unlock()

		packet := make([]byte, n)
		copy(packet, buffer[:n])
		ch <- packet
	}
}
