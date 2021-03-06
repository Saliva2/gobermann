package dnsspam

import (
	"errors"
	"fmt"
	"time"

	"github.com/brutalgg/cli"
	dga "github.com/brutalgg/gobermann/pkg/domaingeneratingalgorithm"
	"github.com/brutalgg/gobermann/plugins/dga/locky"
	"github.com/brutalgg/gobermann/plugins/dga/monerodownloader"
	"github.com/brutalgg/gobermann/plugins/dga/necurs"
	"github.com/brutalgg/gobermann/plugins/dga/nymaim2"
	"github.com/miekg/dns"
)

// Spammer Configuration for running this module
type spammer struct {
	DryRun    bool
	Burst     int
	Delay     int
	Interval  int
	DNSServer string
	Algorithm string
}

// New return a new spammer
func New(dryrun bool, burst int, delay int, interval int, dnsserver string, algorithm string) *spammer {
	return &spammer{
		Algorithm: algorithm,
		Interval:  interval,
		Burst:     burst,
		Delay:     delay,
		DNSServer: dnsserver,
		DryRun:    dryrun,
	}
}

// Run run DNSSpam according to the config struct
func (s spammer) Run() {

	for {
		dga, err := selectDGA(s.Algorithm)
		if err != nil {
			cli.Fatalln("Empty DGA detected. How did you even hit this message?")
		}
		s.burst(dga)
		cli.Infoln("Waiting for interval...")
		time.Sleep(time.Minute * time.Duration(s.Interval))
	}
}

func (s spammer) burst(d dga.DomainGenerator) {
	cli.Infoln("Starting Burst")
	for i := 0; i < s.Burst; i++ {
		domain := d.GenerateDomain()
		cli.Debugln(domain)
		if !s.DryRun {
			DNSQuery(domain, s.DNSServer)
		}
		time.Sleep(time.Millisecond * time.Duration(s.Delay))
	}
	if s.DryRun {
		cli.Infoln("Dryrun burst has been completed successfully. Turn off Drynrun to send traffic.")
	} else {
		cli.Infoln("Burst Completed Successfully")
	}
}

// CheckAlgorithm Checks if a DGA algorithm is supported
func CheckAlgorithm(y string) error {
	_, err := selectDGA(y)
	if err != nil {
		return err
	}
	return nil
}

func selectDGA(x string) (dga.DomainGenerator, error) {
	switch x {
	case "locky":
		return locky.SeedRNG(1, 1, time.Now()), nil
	case "nymaim2":
		return nymaim2.SeedRNG(time.Now()), nil
	case "necurs":
		return necurs.SeedRNG(0, 9, time.Now()), nil
	case "monero":
		return monerodownloader.SeedRNG(0, time.Now()), nil
	}

	return dga.DefaultGenerator{}, errors.New("using empty generator")
}

// DNSQuery queries a server for a FQDN string
func DNSQuery(f string, server string) error {
	var msg dns.Msg
	fqdn := dns.Fqdn(f)
	msg.SetQuestion(fqdn, dns.TypeA)
	if _, err := dns.Exchange(&msg, fmt.Sprintf("%v:53", server)); err != nil {
		return err
	}
	return nil
}
