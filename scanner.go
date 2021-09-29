package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getFirstAndLastPort(portRange string) (int, int) {
	first, _ := strconv.Atoi(strings.Split(portRange, "-")[0])
	last, _ := strconv.Atoi(strings.Split(portRange, "-")[1])

	return first, last
}

func scanPort(host string, port int, protocol string, ch chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	target := host + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout(protocol, target, 60*time.Second)

	if err != nil {
		// fmt.Println(err)
		return
	}

	fmt.Printf("[+] Port %d is open\n", port)
	defer conn.Close()
	ch <- port
}

func monitorWorker(ch chan int, wg *sync.WaitGroup) {
	wg.Wait()
	close(ch)
}

func scanPorts(host string, protocol string, opts *options) []int {
	portCh := make(chan int)
	wg := &sync.WaitGroup{}

	first, last := getFirstAndLastPort(opts.portRange)

	for port := first; port < last; port++ {
		wg.Add(1)
		go scanPort(host, port, protocol, portCh, wg)
		if port%1024 == 0 {
			time.Sleep(time.Second)
		}
	}

	go monitorWorker(portCh, wg)

	var ports []int
	for port := range portCh {
		ports = append(ports, port)
	}

	return ports
}

type options struct {
	host      string
	portRange string
}

func main() {
	var (
		host        string
		commonPorts bool
		portRange   string
	)

	flag.StringVar(&host, "host", "127.0.0.1", "IP of host to scan")
	flag.BoolVar(&commonPorts, "c", false, "Scan only common ports (top 1,024)")
	flag.BoolVar(&commonPorts, "common", false, "Scan only common ports (top 1,024)")
	flag.StringVar(&portRange, "r", "1-65536", "Range of ports to scan")
	flag.StringVar(&portRange, "range", "1-65536", "Range of ports to scan")

	flag.Parse()

	if commonPorts {
		portRange = "1-1024"
	}

	opts := &options{
		host:      host,
		portRange: portRange,
	}

	ports := scanPorts(host, "tcp", opts)
	fmt.Println(ports)
}
