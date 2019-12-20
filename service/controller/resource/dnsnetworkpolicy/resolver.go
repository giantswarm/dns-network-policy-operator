package dnsnetworkpolicy

import (
	"fmt"
	"net"
	"sync"
)

func resolveDomains(domains []string, wg *sync.WaitGroup, ch chan net.IP, resolveAttempts int) {
	defer wg.Done()

	defer close(ch)
	var w sync.WaitGroup
	for _, domain := range domains {
		w.Add(1)
		go resolveDomain(&w, domain, ch, resolveAttempts)
	}
	w.Wait()
}

func resolveDomain(wg *sync.WaitGroup, domain string, ch chan net.IP, resolveAttempts int) {
	var desiredIPs []net.IP
	defer wg.Done()

	for i := 1; i <= resolveAttempts; i++ {

		currentIPs, err := net.LookupIP(domain)
		if err != nil {
			fmt.Printf("%s", err.Error())
		}
		desiredIPs = append(desiredIPs, currentIPs...)
		for _, ip := range desiredIPs {
			ch <- ip
		}
	}
}

func collectResult(wg *sync.WaitGroup, channel chan net.IP) []string {
	defer wg.Done()
	targetIPs := make(map[string]int)
	for ip := range channel {
		var ipType string
		if ip.To4() != nil {
			ipType = "32"
		} else {
			ipType = "128"
		}
		targetIP := fmt.Sprintf("%s/%s", ip.String(), ipType)
		targetIPs[targetIP] = 0
	}

	uniqueIPs := make([]string, 0, len(targetIPs))
	for ip := range targetIPs {
		uniqueIPs = append(uniqueIPs, ip)
	}

	return uniqueIPs
}
