package main

import (
	"os"
	"fmt"
	"strings"
	"errors"
	"strconv"
	"net"
	"time"
	"os/signal"
	"syscall"
	"sort"
	"flag"
	"os/user"
	"os/exec"
	"runtime"
	"github.com/fatih/color"
	"github.com/tatsushid/go-fastping"
)

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

var clear map[string]func()

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	CallClear()

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	user, err := user.Current()
	if user.Username != "root" {
		fmt.Println("Need root permissions. Abort")
		os.Exit(1)
	}

	var maxTime time.Duration
	flag.DurationVar(&maxTime, "t", 1000000000, "set the time interval from one ping to another (specify unit like ns, ms, s)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options] hostname\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	ipRange := flag.Arg(0)
	if len(ipRange) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	ipList, err := parseIPRange(ipRange)

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
	p.MaxRTT = maxTime

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
			CallClear()
			for k := range results {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if results[k] == nil {
					fmt.Printf("%s: %s\n", k, green("[free]"))
				} else {
					fmt.Printf("%s: %s %v\n", k, red("[taken]"), results[k].rtt)
				}
				results[k] = nil
			}
			keys = nil
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

	return ipList, nil
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok { //if we defined a clear func for that platform:
		value()  //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
