sudo yum -y install golang, postgresql

# Creating go workspace and adding env variables
mkdir -p $HOME/go/{bin,pkg,src}
echo export GOPATH=$HOME/go | sudo tee -a /etc/profile
echo PATH=$PATH:$GOPATH/bin | sudo tee -a /etc/profile

go get -u github.com/Packet-Clearing-House/DNSAuth/DNSAuth/...