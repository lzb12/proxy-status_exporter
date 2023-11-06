package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/proxy"
)

var (
	proxyUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "proxy_up",
		Help: "1 if the proxy is up, 0 if it's down.",
	},
		[]string{"proxyaddr"},
	)
)

func main() {
	var proxyaddress string

	flag.StringVar(&proxyaddress, "addr", "192.168.10.202:1289", "proxy addr  Please use multiple, divide")
	flag.Parse()

	parts := strings.Split(proxyaddress, ",")

	prometheus.MustRegister(proxyUp)
	var wg sync.WaitGroup
	wg.Add(1)
	for _, part := range parts {
		go monitorProxyStatus(&wg, part)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9095", nil)
	wg.Wait()

}

func monitorProxyStatus(wg *sync.WaitGroup, proxyaddress string) {
	defer wg.Done()
	for {
		isAlive := ProxyStatus(proxyaddress)
		// fmt.Println(isAlive)
		if isAlive {
			proxyUp.WithLabelValues(proxyaddress).Set(1)
		} else {
			proxyUp.WithLabelValues(proxyaddress).Set(0)
		}
		time.Sleep(5 * time.Minute)

	}

}

func ProxyStatus(proxyaddress string) bool {
	dialer, err := proxy.SOCKS5("tcp", proxyaddress, nil, proxy.Direct)
	// message := fmt.Sprintf("%s访问谷歌有问题", proxyaddress)
	if err != nil {
		fmt.Println(err)
		
		return false
	}
	clt := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial:            dialer.Dial,
		},
	}
	resp, err := clt.Get("https://www.google.com")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("curl google failed~")
		return false
	} else {
		fmt.Println(proxyaddress)
		return true
	}

}
