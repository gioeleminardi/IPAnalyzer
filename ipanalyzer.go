package main

import (
	"os"
	"fmt"
	"strings"
	"errors"
	"strconv"
)

func main() {

	parseIPRange(os.Args[1])

	//p := fastping.NewPinger()
	//ra, err := net.ResolveIPAddr("ip4:icmp", os.Args[1])
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//p.AddIPAddr(ra)
	//p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
	//	fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	//}
	//p.OnIdle = func() {
	//	fmt.Println("finish")
	//}
	//err = p.Run()
	//if err != nil {
	//	fmt.Println(err)
	//}
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
