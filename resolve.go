package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jvns/resolve/logger"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

type Resolver struct {
	nameserver    net.IP
	maxIterations int
	logger        *zap.Logger
}

const (
	MaxIterations = 5
	Nameserver    = "198.41.0.4"
)

// New creates new instance of Resolver
func New(nameserver *string, maxIterations *int, logger *zap.Logger) *Resolver {
	desiredMaxIterations := MaxIterations
	if maxIterations != nil {
		desiredMaxIterations = *maxIterations
	}

	// We will use prefered nameserver if available
	// Otherwise we will fallback to a root nameserver
	var desiredNameserver net.IP
	if nameserver == nil {
		desiredNameserver = net.ParseIP(Nameserver)
	} else {
		if *nameserver == "" {
			desiredNameserver = net.ParseIP(Nameserver)
		} else {
			desiredNameserver = net.ParseIP(*nameserver)
		}
	}

	return &Resolver{
		nameserver:    desiredNameserver,
		maxIterations: desiredMaxIterations,
		logger:        logger,
	}
}

// Resolve performs all the logic
func (r *Resolver) Resolve(target string) (ip *net.IP, err error) {
	var (
		nameserver = r.nameserver
		it         int
	)

	for it < r.maxIterations {
		it++

		if r.logger != nil {
			r.logger.Debug("Attempt to perform dns query",
				zap.String("nameserver", nameserver.String()),
			)
		}

		reply := r.dnsQuery(target, nameserver)

		if res := r.getAnswer(reply); res != nil {
			// Best case: we get an answer to our query and we're done
			ip = &res
			return
		}

		if nsIP := r.getGlue(reply); nsIP != nil {
			// Second best: we get a "glue record" with the *IP address* of another nameserver to query
			nameserver = nsIP
			continue
		}

		if domain := r.getNS(reply); domain != "" {
			// Third best: we get the *domain name* of another nameserver to query, which we can look up the IP for
			var res *net.IP
			res, err = r.Resolve(domain)
			if err != nil {
				return
			}

			nameserver = *res
			continue
		}

		err = errors.New("something went wrong")
		return
	}

	err = errors.New("max iterations reached")
	return
}

// getAnswer parses answer from dns server
func (r *Resolver) getAnswer(reply *dns.Msg) net.IP {
	for _, record := range reply.Answer {
		if record.Header().Rrtype == dns.TypeA {
			fmt.Println("  ", record)

			return record.(*dns.A).A
		}
	}

	return nil
}

// getGlue gets dns glue record
func (r *Resolver) getGlue(reply *dns.Msg) net.IP {
	for _, record := range reply.Extra {
		if record.Header().Rrtype == dns.TypeA {
			fmt.Println("  ", record)

			return record.(*dns.A).A
		}
	}

	return nil
}

// getNS gets NS dns record
func (r *Resolver) getNS(reply *dns.Msg) string {
	for _, record := range reply.Ns {
		if record.Header().Rrtype == dns.TypeNS {
			fmt.Println("  ", record)

			return record.(*dns.NS).Ns
		}
	}

	return ""
}

// dnsQuery performs dns query to desired dns server
func (r *Resolver) dnsQuery(name string, server net.IP) *dns.Msg {
	fmt.Printf("dig -r @%s %s\n", server.String(), name)

	msg := new(dns.Msg)
	msg.SetQuestion(name, dns.TypeA)
	c := new(dns.Client)
	reply, _, _ := c.Exchange(msg, server.String()+":53")

	return reply
}

func main() {
	var (
		target     *string
		nameserver *string
	)

	target = flag.String("target", "", "desired record for resolving")
	nameserver = flag.String("nameserver", "", "desired nameserver used for resolving")

	flag.Parse()

	if *target == "" {
		logger.Log.Error("Target is required")
		os.Exit(1)
	}

	if !strings.HasSuffix(*target, ".") {
		*target = *target + "."
	}

	log := logger.Log.WithOptions(zap.Fields(
		zap.String("target", *target),
	))

	r := New(nameserver, nil, log)
	ip, err := r.Resolve(*target)

	if err != nil {
		log.Error("Error in resolving",
			zap.Error(err),
		)
	} else {
		log.Info("Got result",
			zap.String("ip", ip.String()),
		)
	}
}
