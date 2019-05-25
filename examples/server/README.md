# Using goiplookup in HTTP server

Bolt allows only one process to lock the db file. So, both populate data and ip lookup has to be done in the same process that holds the lock. Otherwise, one process will be stuck waiting for the other process, which holds the lock, to complete.

## How to run this server?
```
$ go build .
$ ./server
```

## Sample Output
```
$ curl -s "http://localhost:8085/iplookup?ip=2001:4c0:0:0:0:0:0:0&ip=49.206.13.16&ip=216.58.196.174&ip=3.91.28.69" | jq .
[
  {
    "country": "IN",
    "ip": "49.206.13.16",
    "version": "ipv4"
  },
  {
    "country": "US",
    "ip": "3.91.28.69",
    "version": "ipv4"
  },
  {
    "country": "US",
    "ip": "216.58.196.174",
    "version": "ipv4"
  },
  {
    "country": "CA",
    "ip": "2001:4c0:0:0:0:0:0:0",
    "version": "ipv6"
  }
]
```