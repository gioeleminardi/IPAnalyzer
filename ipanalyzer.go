package main

import (
	"os"
	"fmt"
	"strings"
	"errors"
	"strconv"
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
	"os/signal"
	"syscall"
	"sort"
)

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

func main() {

	ipList, err := parseIPRange(os.Args[1])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p := fastping.NewPinger()
	results := make(map[string]*response)
	for _, v := range ipList {
		ra, err := net.ResolveIPAddr("ip4:icmp", v)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		results[ra.String()] = nil
		p.AddIPAddr(ra)
	}

	onRecv, onIdle := make(chan *response), make(chan bool)
	p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &response{addr: addr, rtt: t}
	}
	p.OnIdle = func() {
		onIdle <- true
	}
	p.MaxRTT = 10*time.Second

	p.RunLoop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	var keys []string
loop:
	for {
		select {
		case <- c:
			fmt.Println("get interrupted")
			break loop
		case res := <- onRecv:
			if _, ok := results[res.addr.String()]; ok {
				results[res.addr.String()] = res
			}
		case <- onIdle:
			for k := range results {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if results[k] == nil {
					fmt.Printf("%s: unreacheable\n", k)
				} else {
					fmt.Printf("%s: %v\n", k, results[k].rtt)
				}
				results[k] = nil
			}
		case <- p.Done():
			if err := p.Err(); err != nil {
				fmt.Println("ping failed: ", err)
			}
			break loop
		}
	}
	signal.Stop(c)
	p.Stop()

}

func parseIPRange(ipRangeArg string) (ipList []string, err error) {
	if !strings.Contains(ipRangeArg, "-") {
		return nil, errors.New("The range notation is invalid.")
	}

	slices := strings.Split(ipRangeArg, "-")

	if (len(slices) != 2) {
		return nil, errors.New("Error while splitting the ipRangeArg string")
	}

	//Calculate the network address of the ip (ex. 192.168.0.1 -> 192.168.0 )
	netAddr := strings.Split(slices[0], ".")
	netAddrStr := strings.Join(netAddr[:3], ".")

	initialHost, err := strconv.Atoi(netAddr[3])

	if err != nil {
		return nil, err
	}

	finalHost, err := strconv.Atoi(slices[1])

	if err != nil {
		return nil, err
	}

	var singleAddr []string

	for ; initialHost != finalHost+1; initialHost++ {
		singleAddr = []string{netAddrStr, strconv.Itoa(initialHost)}
		ipList = append(ipList, strings.Join(singleAddr, "."))
	}

	//fmt.Printf("%s\n", netAddrStr)
	//fmt.Printf("initialHost: %d\n", initialHost)
	//fmt.Printf("finalHost: %d\n", finalHost)
	//fmt.Printf("ipList: %q", ipList)

	return ipList, nil
}
