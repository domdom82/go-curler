package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	loop := flag.Bool("loop", false, "Keep calling the URL forever.")
	insecure := flag.Bool("insecure", false, "Allow insecure server connections.")
	serverName := flag.String("sni", "", "Server name to use for SNI.")
	loopInterval := flag.Duration("interval", 1*time.Second, "Interval if loop is used.")
	tcpKeepalive := flag.Bool("tcpKeepAlive", true, "Use tcp keep-alive.")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <url> \n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.LUTC)
	log.SetOutput(os.Stdout)

	url := flag.Arg(0)
	if url == "" {
		flag.Usage()
	}

	newClient := func() *http.Client {
		dl := &net.Dialer{
			Timeout: 30 * time.Second,
		}

		if *tcpKeepalive {
			dl.KeepAlive = 10 * time.Second
		} else {
			dl.KeepAlive = -1
		}

		tlsConfig := &tls.Config{
			ServerName:         *serverName,
			InsecureSkipVerify: *insecure,
		}

		tr := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dl.DialContext,
			MaxIdleConns:          100,
			TLSClientConfig:       tlsConfig,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		client := &http.Client{Transport: tr}
		return client
	}

	for {
		fmt.Printf("-> GET %s\n", url)
		start := time.Now()
		r, err := newClient().Get(url)
		stop := time.Now()
		duration := stop.Sub(start)
		if err != nil {
			fmt.Printf("%s (%v)\n", err, duration)
		} else {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("%s (%v)\n", err, duration)
			} else {
				fmt.Printf("<- %s (%v)\n", string(data), duration)
			}
		}
		if *loop == false {
			break
		}
		time.Sleep(*loopInterval)
	}
}
