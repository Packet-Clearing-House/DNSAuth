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

Given the qname of a DNS query, the resolution happens through zone names that are retreive from a mysql database. Each customer will be assocated with one or more zone names which then will be used in a radix tree in order to get the longest prefix match from the qname.

If no customer is found, then "Unknown" is written to the database. See the "Installation" section below for further details about configuring customer rows.

## InfluxDB Rows

DNSAuth writes stats in 1 minute buckets with the following fields:

* pop - point of presence
* time - in 1 minute buckets
* direction - query or response
* qtypestr - query type (eg A, NS etc.)
* rcodestr - response coe (eg NXDOMAIN, SERVFAIL etc.) 
* name - resolved via zone name from local postgres DB
* originAs - optional, the AS the client's IP is from
* prefix -  optional, the prefix the client's IP is from
* protocol - UDP or TCP
* version - IPv4 or IPv6 


## Config file

DNSAuth needs a config file to run. It contains multiple fields:

```
# URL for the Mysql instance to retreive customers
customer-db = "root:pass@(127.0.0.1)/customers"

# Refreshing interval (hours) of the customer database.
customer-refresh = 10

# The URL of the influx DB instance.
influx-db = "http://127.0.0.1:8086/write?db=authdns"

# The directory DNSAuth should watch for new log files coming in.
watch-dir = "./"

```


## Installation and Running


### Root vs local user

All installation should be done by a user with sudo. As well, you can run the entire app with root or sudo.

### Prerequisites

You need to install the following before DNSAuth will work:

* [Influxdb](https://www.docs.influxdata.com/influxdb/v0.9/introduction/installation/)
* [Grafana](http://docs.grafana.org/installation/)
* [go-lang](https://golang.org/doc/install), ideally >=1.9
* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* [MySQL](https://dev.mysql.com/downloads/)

For Ubuntu, installing all the packages looks like this:

```
apt-get update
apt-get upgrade -y
curl -sL https://repos.influxdata.com/influxdb.key | sudo apt-key add -
source /etc/lsb-release
echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
apt-get update && sudo apt-get install -y influxdb mariadb-server
service influxdb start
service mysql start
mysql -u root -e "GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' IDENTIFIED BY 'pass' WITH GRANT OPTION; FLUSH PRIVILEGES;"
wget https://s3-us-west-2.amazonaws.com/grafana-releases/release/grafana_4.6.2_amd64.deb
apt-get install -y adduser libfontconfig
dpkg -i grafana_4.6.2_amd64.deb
service grafana-server start
update-rc.d grafana-server defaults
add-apt-repository -y ppa:longsleep/golang-backports
apt-get update
apt-get install -y golang-go
```

Note - In a production environment you'll want to not set the root password to "pass" ;)


#### Set up Go and run go get

Create go directories and environment: 

``` 
echo export GOPATH=$HOME/go | sudo tee -a /etc/profile
echo PATH=$PATH:$GOPATH/bin | sudo tee -a /etc/profile
mkdir -p $HOME/go/{bin,pkg,src}
source /etc/profile
```

then run `go get`:

```
env GIT_TERMINAL_PROMPT=1 go get -u github.com/Packet-Clearing-House/DNSAuth/...
```

#### Set up influxdb

After running `influx`, create the database:

```bash
CREATE DATABASE authdns
```

#### Clone 

Clone the repo with:

```
cd
git clone https://github.com/Packet-Clearing-House/DNSAuth.git
```

#### Mysql

Assuming you're running MySQL locally with the root password of `pass`, here's how you would 
load our default database and test customers:

```
cd
mysql -u root -p -h localhost < DNSAuth/customers.sql
```

This will generate 2 dummy customers "foo", "bar". Now be sure that go has access to the 
driver by installing it:


#### Run 


Now try running dnsauth. We need to run as `sudo` so that it can bind to a privileged port:

```
cd
sudo ./go/bin/DNSAuth -c DNSAuth/DNSAuth/dnsauth.toml 

```

We're using the default `DNSAuth/dnsauth.toml` config file. Likely this shouldn't need to change.

Finally, in another terminal, copy a sample file in:

```
cd
cp DNSAuth/test/SZC_mon-01.lga.example.com_2018-02-25.05-32.dmp.gz  ./
```

If everything is working, then you should see this after you copy the file:

```bash
2018/05/08 14:26:06 Loading config file...
2018/05/08 14:26:06 OK!
2018/05/08 14:26:06 Initializing customer DB (will be refresh every 24 hours)...
2018/05/08 14:26:06 BGP lookups will be ignored, no BGP config provided.
2018/05/08 14:26:06 [CustomerDB] Refreshing list from mysql...
2018/05/08 14:26:06 [Influx] Inserted 1 points in 560.884µsseconds
2017/12/12 06:56:16 Processed dump [mon-01.lga](2017-10-17 17:07:00 +0000 UTC - 2017-10-17 17:10:00.215724 +0000 UTC): 833 lines in (2.876312ms) seconds!

```

#### Building from a branch

The `go get` command above will build DNSAuth binary from the master branch. 
If you need to build from a branch instead: then you'll need to clone the repo 
within the correct path: `$GOPATH/src/github.com/Packet-Clearing-House/DNSAuth`.

Then checkout the branch you need and run `go install`.  So if you wanted to checkout a branch called
`test-branch`, you'd run this:

```
cd $GOPATH/src/github.com/Packet-Clearing-House/DNSAuth
git checkout test-branch
cd DNSAuth
go get ./...
go install
```