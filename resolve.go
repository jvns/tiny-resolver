package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
)

func resolve(name string) net.IP {
	// We always start with a root nameserver
	nameserver := net.ParseIP("198.41.0.4")
	for {
		reply := dnsQuery(name, nameserver)
		if ip := getAnswer(reply); ip != nil {
			// Best case: we get an answer to our query and we're done
			return ip
		} else if nsIP := getGlue(reply); nsIP != nil {
			// Second best: we get a "glue record" with the *IP address* of another nameserver to query
			nameserver = nsIP
		} else if domain := getNS(reply); domain != "" {
			// Third best: we get the *domain name* of another nameserver to query, which we can look up the IP for
			nameserver = resolve(domain)
		} else {
			// If there's no A record we just panic, this is not a very good
			// resolver :)
			panic("something went wrong")
		}
	}
}

func getAnswer(reply *dns.Msg) net.IP {
	for _, record := range reply.Answer {
		if record.Header().Rrtype == dns.TypeA {
			fmt.Println("  ", record)
			return record.(*dns.A).A
		}
	}
	return nil
}

func getGlue(reply *dns.Msg) net.IP {
	for _, record := range reply.Extra {
		if record.Header().Rrtype == dns.TypeA {
			fmt.Println("  ", record)
			return record.(*dns.A).A
		}
	}
	return nil
}

func getNS(reply *dns.Msg) string {
	for _, record := range reply.Ns {
		if record.Header().Rrtype == dns.TypeNS {
			fmt.Println("  ", record)
			return record.(*dns.NS).Ns
		}
	}
	return ""
}

func dnsQuery(name string, server net.IP) *dns.Msg {
	fmt.Printf("dig -r @%s %s\n", server.String(), name)
	msg := new(dns.Msg)
	msg.SetQuestion(name, dns.TypeA)
	c := new(dns.Client)
	reply, _, _ := c.Exchange(msg, server.String()+":53")
	return reply
}

func main() {
	name := os.Args[1]
	if !strings.HasSuffix(name, ".") {
		name = name + "."
	}
	fmt.Println("Result:", resolve(name))
}
