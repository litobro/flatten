# flatten
Flatten attempts to provide a minimal CNAME flattening implementation for [CoreDNS](https://coredns.io) that is RFC compliant in accordance with the implementation provided by [cloudflare](https://developers.cloudflare.com/dns/cname-flattening/).

This is a better option than using the `rewrite` plugin as it avoids records such as MX, TXT, SOA and others from being resolved to the rewritten domain.

## Usage
Flatten is implemented using a single line in a Corefile that will resolve the CNAMEs recursively and fallthrough any non-A or AAAA records requested. 

`flatten [FROM] [TO] [DNSIP:PORT]`

- `FROM`: Original requested name
- `TO`: Name to overwrite the A and AAAA records from
- `DNSIP:PORT`: The DNS server and port to resolve the `TO` records from

## Compilation
The easy way to consume this plugin is by adding the following on `plugin.cfg` after the `cache` plugi, and recompile it as detailed on [coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

```
<snip>
cache:cache
flatten:github.com/litobro/flatten
rewrite:rewrite
<snip>
```

After this you can compile coredns by:
```
go generate
go build
```

Or you can use the makefile:
```
make
```

## Implementation
The best docs are to read `flatten.go` and view the implementation. It is fewer than 100 lines of code.

The plugin checks if the request `name` matches the `FROM` defined and is of Type `A` or `AAAA`. It then creates a new response, resolves the records from the defined server, and responds with those addresses in the `RR`. The header name is set back to the `FROM` name to maintain RFC 1034 compliance.

## Example Corefile
```
example.org:53 {
    log
    flatten example.org google.ca 1.1.1.1:53

    forward . 1.1.1.1
}
```

### Example output
Run the server
```
$ ./coredns -conf Corefile

example.org.:53
CoreDNS-1.11.2
linux/amd64, go1.21.8, 8de4531d-dirty
[INFO] plugin/flatten: 127.0.0.1:37237 - [example.org.] flattened to [google.ca.] via 1.1.1.1:53
[INFO] plugin/flatten: 127.0.0.1:36670 - [example.org.] flattened to [google.ca.] via 1.1.1.1:53
```

Make a DNS request to the server
```
$ nslookup example.org 127.0.0.1

Server:         127.0.0.1
Address:        127.0.0.1#53

Name:   example.org
Address: 142.250.217.67
Name:   example.org
Address: 2607:f8b0:400a:80b::2003
Name:   example.org
Address: 142.250.217.67
Name:   example.org
Address: 2607:f8b0:400a:804::2003

$ dig TXT example.org @127.0.0.1

; <<>> DiG 9.18.18-0ubuntu0.22.04.1-Ubuntu <<>> TXT example.org @127.00.0.1
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 63388
;; flags: qr rd ra ad; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
;; QUESTION SECTION:
;example.org.                   IN      TXT

;; ANSWER SECTION:
example.org.            86400   IN      TXT     "v=spf1 -all"
example.org.            86400   IN      TXT     "6r4wtj10lt2hw0zhyhk7cgzzffhjp7fl"

;; Query time: 99 msec
;; SERVER: 127.0.0.1#53(127.00.0.1) (UDP)
;; WHEN: Wed Apr 17 09:42:46 MDT 2024
;; MSG SIZE  rcvd: 131
```

## TODO
 - Write unit tests
 - Allow optional parameters and more parameters
 - Create pre-built docker image with Flatten added to plugins