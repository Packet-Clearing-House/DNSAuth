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

### File Format

DNSAuth needs all log files to be gzipped and end in ``.gz``.

## Resolving customer

Customers are defined and stored in the MySQL database. They contain IP range start/end and zone columns. When DNSAuth starts, they are loaded in memory in an interval tree, and used for looking up queries. Given a host IP and a qname of a DNS query:

- first host IP is matched against the customer ranges
- result is further filtered by longest common prefix match of qname on the zone

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

DNSAuth needs a config `dnsauth.toml` file to run:

```
# URL for the Mysql instance to retreive customers
customer-db = "dnsauth:pass@(127.0.0.1)/customers"

# Refreshing interval (hours) of the customer database.
customer-refresh = 10

# The URL of the influx DB instance.
influx-db = "http://127.0.0.1:8086/write?db=authdns"

# The directory DNSAuth should watch for new log files coming in.
watch-dir = "./"

# Action to take after processing a file; one of none, move, or delete.
cleanup-action = "none"

# Path to move processed files when cleanup-action = "move".
# Must not be a sub-directory of watch-dir; no trailing slash.
cleanup-dir = "/tmp"
```

DNSAuth ships with this file as displayed above.  During the set up steps below, you'll copy it to have a local copy which you can customize if needed.


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
mysql -u root -e "CREATE DATABASE customers"
mysql -u root -e "GRANT ALL PRIVILEGES ON customers.* TO 'dnsauth'@'localhost' IDENTIFIED BY 'pass'; GRANT ALL PRIVILEGES ON customers.* TO 'dnsauth'@'127.0.0.1' IDENTIFIED BY 'pass'; FLUSH PRIVILEGES;"
wget https://s3-us-west-2.amazonaws.com/grafana-releases/release/grafana_4.6.2_amd64.deb
apt-get install -y adduser libfontconfig
dpkg -i grafana_4.6.2_amd64.deb
service grafana-server start
update-rc.d grafana-server defaults
add-apt-repository -y ppa:longsleep/golang-backports
apt-get update
apt-get install -y golang-go
```

Note - In a production environment you'll want to not set the database user password to `pass` ;)


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

#### Mysql

Using the previously created `dnsauth` user with the password of `pass`, here's how you would 
load our default database and test customers:

```
mysql -u dnsauth -p -h localhost < $GOPATH/src/github.com/Packet-Clearing-House/DNSAuth/customers.sql
```

This will generate 2 dummy customers "foo", "bar". Here's how to query them for
visual output:

```sql
[customers]> SELECT
  id,
  name,
  INET6_NTOA(TRIM(LEADING CHAR('\0') FROM ip_start)) AS ip_start,
  INET6_NTOA(TRIM(LEADING CHAR('\0') FROM ip_end)) AS ip_end,
  zone
FROM zones;
+----+------+---------------------------+---------------------------+-------------+
| id | name | ip_start                  | ip_end                    | zone        |
+----+------+---------------------------+---------------------------+-------------+
|  1 | foo  | 100.100.100.0             | 100.100.100.255           | auction.com |
|  2 | bar  | fdfe::5a55:caff:fefa:9089 | fdfe::5a55:caff:fefa:9089 | test.com    |
+----+------+---------------------------+---------------------------+-------------+
```

The host IP range is stored in binary format with `ip_start` and `ip_end`
fields. In this example, customer "foo" has host IP range from `100.100.100.0` to
`100.100.100.255` (`100.100.100.0/24`). The range is inclusive - so to designate
a single IP address, `ip_start` and `ip_end` would have the same value, as is the
case with customer "bar". The host IP range supports both IPv4 and IPv6 addresses.

To add more customers, simply `INSERT` more rows. Statements are similar whether
adding IPv4 or IPv6 addresses. For example:

```sql
INSERT INTO zones (id, name, ip_start, ip_end, zone)
VALUES (
  3,
  'new_customer',
  LPAD(INET6_ATON('240.1.44.0'),16,'\0'),
  LPAD(INET6_ATON('240.1.45.255'),16,'\0'),
  'some-domain.test'
), (
  4,
  'another_customer',
  LPAD(INET6_ATON('fdf8:f53b:82e4::52'),16,'\0'),
  LPAD(INET6_ATON('fdf8:f53b:82e4::53'),16,'\0'),
  'ipv6-domain.test'
);
```

#### Run 


Now try running dnsauth. We need to run as `sudo` so that it can bind to a privileged port:

```
cd
cp $GOPATH/src/github.com/Packet-Clearing-House/DNSAuth/DNSAuth/dnsauth.toml .
sudo ./go/bin/DNSAuth -c ./dnsauth.toml 

```

We're using the default `dnsauth.toml` config file. Likely this shouldn't need to change.

Finally, in another terminal, copy a sample file in:

```
cd
cp $GOPATH/src/github.com/Packet-Clearing-House/DNSAuth/DNSAuth/tests/mon-01.xyz.foonet.net_2017-10-17.17-07.dmp.gz .
```

If everything is working, then you should see this after you copy the file:

```bash
2018/05/08 14:26:06 Loading config file...
2018/05/08 14:26:06 OK!
2018/05/08 14:26:06 Initializing customer DB (will be refresh every 24 hours)...
2018/05/08 14:26:06 [CustomerDB] Refreshing list from mysql...
2018/05/08 14:26:06 [Influx] Inserted 1 points in 560.884µsseconds
2017/12/12 06:56:16 Processed dump [mon-01.lga](2017-10-17 17:07:00 +0000 UTC - 2017-10-17 17:10:00.215724 +0000 UTC): 833 lines in (2.876312ms) seconds!

```
#### Running tests

Existing unit tests (customer name resolution) can be run via:

```
cd DNSAuth
go test
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