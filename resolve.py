import dns.message
import dns.query
import dns.rdatatype
import sys

def resolve(domain):
    # Start at the root nameserver
    nameserver = "198.41.0.4"
    while True:
        reply = query(domain, nameserver)
        ip = get_answer(reply)
        if ip:
			# Best case: we get an answer to our query and we're done
            return ip
        nameserver_ip = get_glue(reply)
        if nameserver_ip:
			# Second best: we get a "glue record" with the *IP address* of another nameserver to query
            nameserver = nameserver_ip
        else:
            # Otherwise: we get the *domain name* of another nameserver to query, which we can look up the IP for
            nameserver_domain = get_nameserver(reply)
            nameserver = resolve(nameserver_domain)

def query(name, nameserver):
    query = dns.message.make_query(name, 'A')
    return dns.query.udp(query, nameserver)

def get_answer(reply):
    for record in reply.answer:
        if record.rdtype == dns.rdatatype.A:
            return record[0].address

def get_glue(reply):
    for record in reply.additional:
        if record.rdtype == dns.rdatatype.A:
            return record[0].address

def get_nameserver(reply):
    for record in reply.authority:
        if record.rdtype == dns.rdatatype.NS:
            return record[0].target

print(resolve(sys.argv[1]))
