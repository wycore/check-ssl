package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"math"
	"net"
	"os"
	"runtime/debug"
	"syscall"
	"time"
)

const (
	OK       = iota
	Warning  = iota
	Critical = iota
	Unknown  = iota
)

var exitCode int = OK
var lookupTimeout, connectionTimeout, warningValidity, criticalValidity time.Duration
var warningFlag, criticalFlag uint
var version string
var printVersion bool

func updateExitCode(newCode int) (changed bool) {
	if newCode > exitCode {
		exitCode = newCode
		return true

	}
	return false
}

func main() {
	defer catchPanic()

	var host string
	var ips []net.IP

	flag.StringVar(&host, "host", "", "the domain name of the host to check")
	flag.DurationVar(&lookupTimeout, "lookup-timeout", 10*time.Second, "timeout for DNS lookups - see: https://golang.org/pkg/time/#ParseDuration")
	flag.DurationVar(&connectionTimeout, "connection-timeout", 30*time.Second, "timeout connection - see: https://golang.org/pkg/time/#ParseDuration")
	flag.UintVar(&warningFlag, "w", 30, "warning validity in days")
	flag.UintVar(&criticalFlag, "c", 14, "critical validity in days")
	flag.BoolVar(&printVersion, "V", false, "print version and exit")
	flag.Parse()

	log.SetLevel(log.InfoLevel)
	if printVersion {
		log.Infof("Version: %s", version)
		os.Exit(0)
	}

	if host == "" {
		flag.Usage()
		log.Error("-host is required")
		os.Exit(Critical)
	}
	if warningFlag < criticalFlag {
		log.Warn("-c is higher than -w, i guess thats a bad i idea")
		updateExitCode(Warning)
	}

	warningValidity = time.Duration(warningFlag) * 24 * time.Hour
	criticalValidity = time.Duration(criticalFlag) * 24 * time.Hour

	ips = lookupIPWithTimeout(host, lookupTimeout)
	log.Debugf("lookup result: %v", ips)

	for _, ip := range ips {
		dialer := net.Dialer{Timeout: connectionTimeout, Deadline: time.Now().Add(connectionTimeout + 5*time.Second)}
		connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:443", ip), &tls.Config{ServerName: host})
		if err != nil {
			// catch missing ipv6 connectivity
			// if the ip is ipv6 and the resulting error is "no route to host", the record is skipped
			// otherwise the check will switch to critical
			if ip.To4() == nil {
				switch err.(type) {
				case *net.OpError:
					// https://stackoverflow.com/questions/38764084/proper-way-to-handle-missing-ipv6-connectivity
					if err.(*net.OpError).Err.(*os.SyscallError).Err == syscall.EHOSTUNREACH {
						log.Infof("%-15s - ignoring unreachable IPv6 address", ip)
						continue
					}
				}
			}
			log.Errorf("%s: %s", ip, err)
			updateExitCode(Critical)
			continue
		}
		// rembember the checked certs based on their Signature
		checkedCerts := make(map[string]struct{})
		// loop to all certs we get
		// there might be multiple chains, as there may be one or more CAs present on the current system, so we have multiple possible chains
		for _, chain := range connection.ConnectionState().VerifiedChains {
			for _, cert := range chain {
				if _, checked := checkedCerts[string(cert.Signature)]; checked {
					continue
				}
				checkedCerts[string(cert.Signature)] = struct{}{}
				// filter out CA certificates
				if cert.IsCA {
					log.Debugf("%-15s - ignoring CA certificate %s", ip, cert.Subject.CommonName)
					continue
				}

				var certificateStatus int
				remainingValidity := cert.NotAfter.Sub(time.Now())
				if remainingValidity < criticalValidity {
					certificateStatus = Critical
				} else if remainingValidity < warningValidity {
					certificateStatus = Warning
				} else {
					certificateStatus = OK
				}
				updateExitCode(certificateStatus)
				logWithSeverity(certificateStatus, "%-15s - %s valid until %s (%s)", ip, cert.Subject.CommonName, cert.NotAfter, formatDuration(remainingValidity))
			}
		}
		connection.Close()
	}
	os.Exit(exitCode)
}

func lookupIPWithTimeout(host string, timeout time.Duration) []net.IP {
	timer := time.NewTimer(timeout)

	ch := make(chan []net.IP, 1)
	go func() {
		r, err := net.LookupIP(host)
		if err != nil {
			log.Fatal(err)
		}
		ch <- r
	}()
	select {
	case ips := <-ch:
		return ips
	case <-timer.C:
		log.Errorf("timeout resolving %s", host)
		updateExitCode(Critical)
	}
	return make([]net.IP, 0)
}

func catchPanic() {
	if r := recover(); r != nil {
		log.Errorf("Panic: %+v", r)
		log.Error(string(debug.Stack()[:]))
		os.Exit(Critical)
	}
}

func formatDuration(in time.Duration) string {
	var daysPart, hoursPart, minutesPart, secondsPart string

	days := math.Floor(in.Hours() / 24)
	hoursRemaining := math.Mod(in.Hours(), 24)
	if days > 0 {
		daysPart = fmt.Sprintf("%.fd", days)
	} else {
		daysPart = ""
	}

	hours, hoursRemaining := math.Modf(hoursRemaining)
	minutesRemaining := hoursRemaining * 60
	if hours > 0 {
		hoursPart = fmt.Sprintf("%.fh", hours)
	} else {
		hoursPart = ""
	}

	if minutesRemaining > 0 {
		minutesPart = fmt.Sprintf("%.fm", minutesRemaining)
	}

	_, minutesRemaining = math.Modf(minutesRemaining)
	secondsRemaining := minutesRemaining * 60
	if secondsRemaining > 0 {
		secondsPart = fmt.Sprintf("%.fs", secondsRemaining)
	}

	return fmt.Sprintf("%s %s %s %s", daysPart, hoursPart, minutesPart, secondsPart)
}

func logWithSeverity(severity int, format string, args ...interface{}) {
	switch severity {
	case OK:
		log.Infof(format, args...)
	case Warning:
		log.Warnf(format, args...)
	case Critical:
		log.Errorf(format, args...)
	default:
		log.Panicf("Invalid severity %d", severity)
	}
}