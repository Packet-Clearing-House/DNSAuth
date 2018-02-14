# DNSAuth

## About

This repo contains 3 different main directories:
  * DNSAuth which is the main piece of software. It looks at a directory for new files coming in (through rsync atm) and then aggregate all the DNS queries to minute buckets that are then forwarded to influxdb.
  * API that allows to interact with the customer database
  * GUI contains a simple GUI implementation to display information about customers


### Prerequisites

You need to install the following:

* [Influxdb](https://www.docs.influxdata.com/influxdb/v0.9/introduction/installation/)
* [Grafana](http://docs.grafana.org/installation/)
* [go-lang](https://golang.org/doc/install), ideally >=1.9
* [postgres](https://wiki.postgresql.org/wiki/Detailed_installation_guides)
* [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)


As well, you'll need a read directory created and set two entries in `/etc/profile`:

```bash
sudo mkdir -p /home/jtodd/count
sudo chmod -R 777 /home/jtodd
```

For Ubuntu, installing all the packages looks like this (run **as root**):

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

**As your user**,  create go directories and, while connected to VPN, run `git go`.  Enter your bitbucket/stash username when prompted:

``` 
echo export GOPATH=$HOME/go | sudo tee -a /etc/profile
echo PATH=$PATH:$GOPATH/bin | sudo tee -a /etc/profile
mkdir -p $HOME/go/{bin,pkg,src}
env GIT_TERMINAL_PROMPT=1 go get -u github.com/Packet-Clearing-House/DNSAuth/...
```

Todo - ``go get`` to github isn't tested.  Need to test and update docs if needed.

#### Postgres user and data

Launch postgres CLI via `sudo -u postgres psql postgres` and then run:

```
DROP TABLE ns_customers;
CREATE TABLE ns_customers(
   ip TEXT PRIMARY KEY NOT NULL,
   name TEXT,
   asn BOOL,
   prefix BOOL
);
INSERT INTO ns_customers VALUES ('1.199.71.00/24', 'Foo', true, true);
INSERT INTO ns_customers VALUES ('caec:cec6:c4ef:bb7b::/48', 'Bar', true, true);
INSERT INTO ns_customers VALUES ('11.206.206.0/24', 'Bash', true, true);

CREATE USER "user" WITH PASSWORD 'password';
grant select on ns_customers to "user";
```

#### Set up influxdb

After running `influx`, create the database:

```bash
CREATE DATABASE authdns
```
#### clone and run

clone the repo  `cd;git clone git@github.com:Packet-Clearing-House/DNSAuth.git`

and then try running dnsauth. We need to run as `sudo` so that it can bind to a privileged port:

```
cd
sudo ./go/bin/DNSAuth -c dnsauth/DNSAuth/dnsauth.toml
```

We're using the default `dnsauth/DNSAuth/dnsauth.toml` config file. Likely this shouldn't  need to change.

Finally, in another terminal, copy a sample file in:

```
cp dnsauth/mon-01.sample.net_2017-10-17.17-07.dmp.gz /home/jtodd/count/
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

2017/12/12 06:56:16 Processed dump [mon-01-foo](2017-10-17 17:07:00 +0000 UTC - 2017-10-17 17:10:00.215724 +0000 UTC): 833 lines in (2.876312ms) seconds!

```