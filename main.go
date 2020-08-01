package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var (
	openPorts   uint16
	closedPorts uint16
	startTime   int64

	openPortsMutex   sync.RWMutex
	closedPortsMutex sync.RWMutex
)

func TCPScan(wg *sync.WaitGroup, workerID int, IP string, startPort int, endPort int, waitTimeout time.Duration) {
	defer wg.Done()

	var workerOpenPorts uint16 = 0
	var workerClosedPorts uint16 = 0
	//fmt.Printf("Worker %v started - scanning ports %v - %v\n", workerID, startPort, endPort)
	for i := startPort; i <= endPort; i++ {
		conn, err := net.DialTimeout("tcp", IP+":"+strconv.Itoa(i), 1*time.Second)
		if err != nil {
			fmt.Printf("\r\033[2KPort %v closed (worker: %v)", i, workerID)
			workerClosedPorts++
		} else {
			fmt.Printf("\r\033[2KPort %v open\n", i)
			conn.Close()
			workerOpenPorts++
		}
	}
	openPortsMutex.Lock()
	openPorts += workerOpenPorts
	openPortsMutex.Unlock()

	closedPortsMutex.Lock()
	closedPorts += workerClosedPorts
	closedPortsMutex.Unlock()
}

func IPFormat(IP string) string {
	for i := 0; i < len(IP); i++ {
		switch IP[i] {
		case '.':
			return ("IPv4")
		case ':':
			return ("IPv6")
		}
	}
	return ""
}

func checkValidIP(IP string) {
	if net.ParseIP(IP) == nil {
		fmt.Printf("Invalid IP Address: %s\n", IP)
		os.Exit(1)
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func handleSIGTERM() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("\n\nPort scan cancelled\n")
		stats()
		os.Exit(0)
	}()
}

func stats() {
	td := time.Now()
	endTime := td.Unix()
	timeTaken := endTime - startTime
	fmt.Printf("\r\033[2K\n%v Open port(s). Scanned %v ports in %v seconds.\n", openPorts, (openPorts + closedPorts), timeTaken)
}

func main() {
	handleSIGTERM()

	usage := fmt.Sprintf("Usage: %s -s <start port> -e <end port> -c <worker count> -w <wait timeout> <ip address>\n", filepath.Base(os.Args[0]))

	argLength := len(os.Args[1:])
	if argLength < 1 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	//set defaults
	startPort := flag.Int("s", 1, "start port")
	endPort := flag.Int("e", 1024, "end port")
	maxConcurrency := flag.Int("c", 5, "Worker count")
	waitTimeout := flag.Int64("w", 1, "Wait timeout")

	flag.Parse()
	IP := flag.Arg(0)

	checkValidIP(IP)

	if *startPort < 0 {
		*startPort = 0
	}
	if *startPort > 65535 {
		*startPort = 65535
	}

	if *endPort < 0 {
		*endPort = 0
	}
	if *endPort > 65535 {
		*endPort = 65535
	}

	if *startPort > *endPort {
		fmt.Printf("Start port (-s) %v should be lower than end port (-e) %v\n", *startPort, *endPort)
		os.Exit(1)
	}

	if *maxConcurrency > 100 {
		*maxConcurrency = 100
		fmt.Printf("Warning - Worker count lowered to 100 as results unpredictable when using too many\n")
	}

	portRange := []int{}

	for i := *startPort; i <= *endPort; i++ {
		portRange = append(portRange, i)
	}

	if (*endPort - *startPort) < *maxConcurrency {
		*maxConcurrency = *endPort - *startPort
	}

	chunkSize := int(math.Ceil(float64(len(portRange)) / float64(*maxConcurrency)))

	fmt.Printf("TCP Connect Port Scanning %s (%s) Ports %v - %v\n\n", IP, IPFormat(IP), *startPort, *endPort)
	td := time.Now()
	startTime = td.Unix()

	workerID := 1
	var wg sync.WaitGroup
	for i := 0; i < len(portRange); i += chunkSize {
		batch := portRange[i:min(i+chunkSize, len(portRange))]
		wg.Add(1)
		go TCPScan(&wg, workerID, IP, batch[0], batch[len(batch)-1], time.Duration(*waitTimeout))
		workerID += 1
	}

	wg.Wait()
	stats()
}
