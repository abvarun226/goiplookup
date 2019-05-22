# Geo IP Lookup

[![Go Report Card](https://goreportcard.com/badge/github.com/abvarun226/geoiplookup)](https://goreportcard.com/report/github.com/abvarun226/geoiplookup)

This library maps an IP address to a country

It basically uses the Regional Internet Registries (APNIC, ARIN, LACNIC, RIPE NCC and AFRINIC) to populate a database with the IP address blocks allocated to a country. It then uses this database, which is refereshed on a daily basis, to lookup country to which an IP address belongs to.