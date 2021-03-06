package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ernado/stun"
)

func main() {
	var (
		addr *net.UDPAddr
		err  error
	)

	fmt.Println("START")
	for i := 0; i < 10; i++ {
		addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("stun-server:%d", stun.DefaultPort))
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 300 * time.Duration(i))
	}
	if err != nil {
		log.Println("too many attempts to resolve:", err)
	}

	client := new(stun.Client)
	fmt.Println("DIALING", addr)
	if err = client.Dial(addr); err != nil {
		log.Fatalln("failed to client.Dial:", err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		if err := client.ReadUntilClosed(); err != nil {
			log.Fatalln("read until closed loop:", err)
		}
		wg.Done()
	}()

	laddr := client.LocalAddr()
	fmt.Println("LISTEN ON", laddr)
	request, err := stun.Build(stun.BindingRequest, stun.TransactionID)
	if err != nil {
		log.Fatalln("failed to build:", err)
	}
	if err := client.Do(request, func(response *stun.Message) error {
		if response.Type != stun.BindingSuccess {
			log.Fatalln("bad message", response)
		}
		var xorMapped stun.XORMappedAddress
		if err = response.Parse(&xorMapped); err != nil {
			log.Fatalln("failed to parse xor mapped address:", err)
		}
		if laddr.String() != xorMapped.String() {
			log.Fatalln(laddr, "!=", xorMapped)
		}
		fmt.Println("OK", response, "GOT", xorMapped)
		return nil
	}); err != nil {
		log.Fatalln("failed to Do:", err)
	}
	client.Close()
	wg.Wait()
}
