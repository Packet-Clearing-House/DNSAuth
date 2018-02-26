# DNSAuth

## About

DNSAuth is a Golang application for ingesting DNS server statistics into an InfuxDB instance. The assumption is that you are running a DNS server and you get the raw statics from it either by having it log in the DNSAuth format or by pulling it off the wire via a packet capture (pcap) and then distill the pcap to the DNSAuth format.  It also assumes you want to have customer correlated to each entry in InfluxDB.

## Components

This repo contains 3 different main directories:
  * DNSAuth which is the main piece of software. It looks at a directory for new files coming in (through rsync atm) and then aggregate all the DNS queries to minute buckets that are then forwarded to influxdb.
  * API that allows to interact with the customer database
  * GUI contains a simple GUI implementation to display information about customers


## Logs

### File Format

This is a sample log from a DNS server that DNSAuth reads:

```
Q 192.0.2.10 203.0.113.254 0 0 15 www.domain.com. 44
R 192.0.2.10 203.0.113.254 0 0 15 www.domain.com. 582 0
```

Breaking this down, we can label the fields 1 through 9:

```
R   192.0.2.10    203.0.113.254 0    0   15  www.domain.com.     582     0- 
```

And then the labels translate to: 
1. Query or Response:  flag for query or response
1. IP: source of host making the query
1. Server: nameserver IP, used to determine customer
1. Protocol; 0=UDP, 1=TCP
1. Operation Code; 0=Query, 4=Notify, 5=Update, etc.
1. Query Type; 1=A, 2=NS, 5=CNAME, 6=SOA, 12=PTR, etc
1. Query string: zone being queried
1. Size: packet size in bytes
1. Response: If field 1 was an R; 0=NOERROR, 3=NXDOMAIN, 2=SERVFAIL, etc.

Note that that DNSAuth assumes all lines come in pairs of a Query and then Response line. The query line will always have a ``NULL`` for field 9. 


### File Names

DNSAuth assumes these facts about the file name:
* A three letter pop is used to denote which location the DNS server is running
* the three leter pop is part of the hostname who's format is ``subdomain.domain.tld``
* A UTC based time stamp is included in the file name in ``YEAR-MONTH-DAY.HOUR-SECOND`` 
* The file name is prefaced by ``SZC_`` followed by ``mon-01`` where ``01`` may be any zero padded number up to 10
* the file's suffix will be ``.dmp.gz``

An example of this for a pop in lga (New York) from Feb 25th, 2018 at 5:32am would be:

```
SZC_mon-01.lga.example.com_2018-02-25.05-32.dmp.gz
```

This file is included in the repository for example purposes.

### Fie Format

DNSAuth needs all log files to be gzipped and end in ``.gz``.

## Resolving customer

Given the server IP (field 3 from above), DNSAuth will query a postgres database to try try and find a matching customer.  It assumes that each customer row in the table has a CIDR formatted IP and will try to find the server IP in the that CIDR block.

If no customer is found, then "Unknown" is written to the database. See the "Installation" section below for further details about configuring customer rows.

## InfluxDB Rows

DNSAuth writes stats in 1 minute buckets with the following fields:

* pop - point of presence
* time - in 1 minute buckets
* direction - query or response
* qtypestr - query type (eg A, NS etc.)
* rcodestr - response coe (eg NXDOMAIN, SERVFAIL etc.) 
* name - resolved via IP from local postgres DB
* originAs - optional, the AS the client's IP is from
* prefix -  optional, the prefix the client's IP is from
* protocol - UDP or TCP
* version - IPv4 or IPv6 

## Installation and Running


### Root vs local user

All installation should be done by a user with sudo. As well, you can run the entire app with root or sudo.

### Prerequisites

You need to install the following before DNSAuth will work:

* [Influxdb](https://www.docs.influxdata.com/influxdb/v0.9/introduction/installation/)
* [Grafana](http://docs.grafana.org/installation/)
* [go-lang](https://golang.org/doc/install), ideally >=1.9
* [postgres](https://wiki.postgresql.org/wiki/Detailed_installation_guides)
* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

For Ubuntu, installing all the packages looks like this:

```
apt-get update
apt-get upgrade
curl -sL https://repos.influxdata.com/influxdb.key | sudo apt-key add -
source /etc/lsb-release
echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
apt-get update && sudo apt-get install influxdb
service influxdb start
apt-get install postgresql postgresql-contrib
wget https://s3-us-west-2.amazonaws.com/grafana-releases/release/grafana_4.6.2_amd64.deb
apt-get install -y adduser libfontconfig
dpkg -i grafana_4.6.2_amd64.deb
service grafana-server start
update-rc.d grafana-server defaults
add-apt-repository ppa:longsleep/golang-backports
apt-get update
apt-get install golang-go 
```


#### Set up Go and run go get

Create go directories and run `git go`:

``` 
echo export GOPATH=$HOME/go | sudo tee -a /etc/profile
echo PATH=$PATH:$GOPATH/bin | sudo tee -a /etc/profile
mkdir -p $HOME/go/{bin,pkg,src}
env GIT_TERMINAL_PROMPT=1 go get -u github.com/Packet-Clearing-House/DNSAuth/...
```

#### Postgres user and data

Launch postgres CLI via `sudo -u postgres psql postgres` and then run this code:

```
DROP TABLE ns_customers;
CREATE TABLE ns_customers(
   ip TEXT PRIMARY KEY NOT NULL,
   name TEXT,
   asn BOOL,
   prefix BOOL
);
INSERT INTO ns_customers VALUES ('203.0.113.254/24', 'Foo', true, true);
INSERT INTO ns_customers VALUES ('2001:DB8::/32', 'Bar', true, true);
INSERT INTO ns_customers VALUES ('198.51.100.3/24', 'Bash', true, true);

CREATE USER "user" WITH PASSWORD 'password';
grant select on ns_customers to "user";
```

This will generate 3 dummy customers "Foo", "Bar" and "Bash". Create rows with your real customers when deploying to production.sh

#### Set up influxdb

After running `influx`, create the database:

```bash
CREATE DATABASE authdns
```

#### create dirs, clone and run

You'll need a log file directory created:

```bash
sudo mkdir -p /home/user/count
sudo chmod -R 777 /home/user
```

Then clone the repo  `cd;git clone git@github.com:Packet-Clearing-House/DNSAuth.git`

and then try running dnsauth. We need to run as `sudo` so that it can bind to a privileged port:

```
cd
sudo ./go/bin/DNSAuth -c DNSAuth/DNSAuth/dnsauth.toml
```

We're using the default `DNSAuth/DNSAuth/dnsauth.toml` config file. Likely this shouldn't  need to change.

Finally, in another terminal, copy a sample file in:

```
cp DNSAuth/test/SZC_mon-01.lga.example.com_2018-02-25.05-32.dmp.gz /home/user/count/
```

If everything is working, then you should see this after you copy  the file:

```bash
 sudo ./go/bin/DNSAuth -c dnsauth/DNSAuth/dnsauth.toml 
2017/12/12 06:55:46 Loading config file...
2017/12/12 06:55:46 OK!
2017/12/12 06:55:46 Getting customer list from postgres...
2017/12/12 06:55:46 OK!
2017/12/12 06:55:46 Starting BGP Resolver...
INFO[0000] Starting BGP server: (router-id :1.199.71.1, local-as: 1234, peer-address: 1.199.71.1, remote-as: 5678) 
INFO[0000] Add a peer configuration for:11.206.206.245     Topic=Peer
2017/12/12 06:55:46 OK!
2017/12/12 06:55:46 Pushing metrics!!
2017/12/12 06:55:46 Influx pusher inserted 1 points!
2017/12/12 06:55:46 Took 417.687µsseconds

2017/12/12 06:56:16 Processed dump [mon-01.lga](2017-10-17 17:07:00 +0000 UTC - 2017-10-17 17:10:00.215724 +0000 UTC): 833 lines in (2.876312ms) seconds!

```