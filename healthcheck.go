package frontman

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthCheck runs before any other check to ensure that the host itself and its network are healthly.
// This is useful to confirm a stable internet connection to avoid false alerts due to network outages.
func (fm *Frontman) HealthCheck() error {
	hcfg := fm.Config.HealthChecks
	if len(hcfg.ReferencePingHosts) == 0 {
		return nil
	}
	if hcfg.ReferencePingCount == 0 {
		return nil
	}
	timeout := secToDuration(hcfg.ReferencePingTimeout)
	if timeout == 0 {
		// use the default timeout
		timeout = 500 * time.Millisecond
	}
	failC := make(chan string, len(hcfg.ReferencePingHosts))

	wg := new(sync.WaitGroup)
	for _, addr := range hcfg.ReferencePingHosts {
		p, err := NewPinger(addr)
		if err != nil {
			logrus.WithError(err).Warningln("failed to parse host for ICMP ping")
			continue
		}
		p.Timeout = timeout
		p.Count = hcfg.ReferencePingCount
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			p.Run()
			if p.Statistics().PacketLoss > 0 {
				failC <- addr
			}
		}(addr)
	}
	go func() {
		wg.Wait()
		close(failC)
	}()

	failedHosts := []string{}
	for host := range failC {
		failedHosts = append(failedHosts, host)
	}
	if len(failedHosts) > 0 {
		return fmt.Errorf("host(s) failed to respond to ICMP ping: %s", strings.Join(failedHosts, ", "))
	}
	return nil
}
