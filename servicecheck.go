package frontman

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (fm *Frontman) runServiceCheck(check ServiceCheck) (map[string]interface{}, error) {
	var done = make(chan struct{})
	var err error
	var results map[string]interface{}
	go func() {
		ipaddr, resolveErr := resolveIPAddrWithTimeout(check.Check.Connect, timeoutDNSResolve)
		if resolveErr != nil {
			err = fmt.Errorf("resolve ip error: %s", resolveErr.Error())
			logrus.Debugf("serviceCheck: ResolveIPAddr error: %s", resolveErr.Error())
			done <- struct{}{}
			return
		}

		switch check.Check.Protocol {
		case ProtocolICMP:
			results, err = fm.runPing(ipaddr)
			if err != nil {
				logrus.Debugf("serviceCheck: %s: %s", check.UUID, err.Error())
			}
		case ProtocolTCP:
			port, _ := check.Check.Port.Int64()

			results, err = fm.runTCPCheck(&net.TCPAddr{IP: ipaddr.IP, Port: int(port)}, check.Check.Connect, check.Check.Service)
			if err != nil {
				logrus.Debugf("serviceCheck: %s: %s", check.UUID, err.Error())
			}
		case ProtocolSSL:
			port, _ := check.Check.Port.Int64()

			results, err = fm.runSSLCheck(&net.TCPAddr{IP: ipaddr.IP, Port: int(port)}, check.Check.Connect, check.Check.Service)
			if err != nil {
				logrus.Debugf("serviceCheck: %s: %s", check.UUID, err.Error())
			}
		case "":
			logrus.Info("serviceCheck: missing check.protocol")
			err = errors.New("Missing check.protocol")
		default:
			logrus.Errorf("serviceCheck: unknown check.protocol: '%s'", check.Check.Protocol)
			err = errors.New("Unknown check.protocol")
		}
		done <- struct{}{}
	}()

	// Warning: do not rely on serviceCheckEmergencyTimeout as it leak goroutines(until it will be finished)
	// instead use individual timeouts inside all checks
	select {
	case <-done:
		return results, err
	case <-time.After(serviceCheckEmergencyTimeout):
		logrus.Errorf("serviceCheck: %s got unexpected timeout after %.0fs", check.UUID, serviceCheckEmergencyTimeout.Seconds())
		return nil, fmt.Errorf("got unexpected timeout")
	}
}

func resolveIPAddrWithTimeout(addr string, timeout time.Duration) (*net.IPAddr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ipAddrs, err := net.DefaultResolver.LookupIPAddr(ctx, addr)
	if err != nil {
		return nil, err
	}

	if len(ipAddrs) == 0 {
		return nil, errors.New("can't resolve host")
	}

	ipAddr := ipAddrs[0]
	return &ipAddr, nil
}

func (checkList *ServiceCheckList) Check(fm *Frontman, wg *sync.WaitGroup, resultsChan chan<- Result, succeed *int) {
	for _, check := range checkList.Checks {
		wg.Add(1)
		go func(check ServiceCheck) {
			defer wg.Done()

			if check.UUID == "" {
				// in case checkUuid is missing we can ignore this item
				logrus.Info("serviceCheck: missing checkUuid key")
				return
			}

			res := Result{
				CheckType: "serviceCheck",
				CheckUUID: check.UUID,
				Timestamp: time.Now().Unix(),
			}

			res.Check = check.Check

			if check.Check.Connect == "" {
				logrus.Info("serviceCheck: missing data.connect key")
				res.Message = "Missing data.connect key"
			} else {
				var err error
				res.Measurements, err = fm.runServiceCheck(check)
				if err != nil {
					recovered := false
					if fm.Config.FailureConfirmation > 0 {
						logrus.Debugf("serviceCheck failed, retrying up to %d times: %s: %s", fm.Config.FailureConfirmation, check.UUID, err.Error())

						for i := 1; i <= fm.Config.FailureConfirmation; i++ {
							time.Sleep(time.Duration(fm.Config.FailureConfirmationDelay*1000) * time.Millisecond)
							logrus.Debugf("Retry %d for failed check %s", i, check.UUID)
							res.Measurements, err = fm.runServiceCheck(check)
							if err == nil {
								recovered = true
								break
							}
						}
					}
					if !recovered && fm.Config.AskNeigbors {
						logrus.Debug("asking neighbors...")

						var responses []http.Response

						for _, neighbor := range fm.Config.Neighbors {
							logrus.Debug("asking neighbor", neighbor.Name)
							url, err := url.Parse(neighbor.URL)
							if err != nil {
								logrus.Warnf("Invalid neighbor url in config: '%s': %s", neighbor.URL, err.Error())
								continue
							}
							url.Path = path.Join(url.Path, "check")
							logrus.Debug("connecting to ", url.String())

							// ask neighbor
							client := &http.Client{}
							req, _ := http.NewRequest("GET", url.String(), nil)
							res, err := client.Do(req)
							if err != nil {
								logrus.Warnf("Failed to ask neighbor: %s", err.Error())
							} else {
								responses = append(responses, *res)
							}
						}

						if len(responses) > 0 {
							// XXX pick "fastest result" and send it back
							spew.Dump(responses)

							// XXX create a new result message with fastest result + group_measurements with all responses
							// XXX attach new messagew to result: "message": "Check failed locally and on 2 neigbors but succeded on Frontman EU"
						}
					}
					if !recovered {
						logrus.Debugf("serviceCheck: %s: %s", check.UUID, err.Error())
						res.Message = err.Error()
					}
				}

				if res.Message == nil {
					*succeed++
				}
			}

			resultsChan <- res
		}(check)
	}
}