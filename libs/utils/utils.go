package utils

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
)

type ACL struct {
	cidrs []*net.IPNet
	ips   []*net.IP
}

func (acl ACL) Check(ip net.IP) bool {
	for _, ipnet := range acl.cidrs {
		if ipnet.Contains(ip) {
			return true
		}
	}
	for _, ipacl := range acl.ips {
		if ip.Equal(*ipacl) {
			return true
		}
	}
	return false
}

type ACLError struct {
	Addr string
}

func (e ACLError) Error() string {
	return "IP (" + e.Addr + ") not allowed, rejecting connection!"
}

func (e ACLError) Timeout() bool {
	return false
}
func (e ACLError) Temporary() bool {
	return true
}

type ACListener struct {
	net.Listener
	acl *ACL // Pointer to be hashable (when passing ACListener to shttp server)
}

func (l ACListener) Accept() (net.Conn, error) {

	var addrStr string
	conn, err := l.Listener.Accept()
	if err == nil {
		addrStr, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
		if l.acl.Check(net.ParseIP(addrStr)) {
			return conn, nil
		} else {
			conn.Close()
			err = ACLError{addrStr}
		}
	}
	return nil, err
}

func (l ACListener) Close() error {
	return l.Listener.Close()
}
func (l ACListener) Addr() net.Addr {
	return l.Listener.Addr()
}

func ACLListen(nett, laddr string, acl ACL) (*ACListener, error) {

	listener, err := net.Listen(nett, laddr)
	if err == nil {
		return &ACListener{
			listener,
			&acl,
		}, nil
	}
	return nil, err
}

func ParseACLFromStrings(aclStr []string) (acl ACL, err error) {

	//acl = make([]*net.IPNet, len(aclStr))
	acl = ACL{}

	for _, item := range aclStr {
		_, ipnet, err := net.ParseCIDR(item)
		if err != nil {
			ip := net.ParseIP(item)
			if ip != nil {

				acl.ips = append(acl.ips, &ip)
			}
		} else {
			acl.cidrs = append(acl.cidrs, ipnet)
		}
	}
	return acl, nil
}

func HTTPBasicAuth(handler http.Handler, login, pass string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		_, password, _ := r.BasicAuth()

		if password != pass {
			log.Println("[ERROR] Unauthorized HTTP access: wrong password!")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}
}

const (
	minTCPPort = 0
	maxTCPPort = 65535
)

// Will split host and port from a string in the format host:port
// Return errors if ip address is invalid, port number out of range
// or if the host:port format is not respected.
// If the ip address if left empty (:9090), "0.0.0.0" will be used as default.
func SplitHostPort(hostPort string) (addr string, port int, err error) {
	addr, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", 0, err
	}
	if addr == "" {
		addr = "0.0.0.0"
	}
	if net.ParseIP(addr) == nil {
		return "", 0, errors.New(addr + " is not a correct IP address")
	} else if port, err := strconv.Atoi(portStr); err != nil || port < minTCPPort || port > maxTCPPort {
		return "", 0, errors.New(portStr + " is not a correct port number")
	} else {
		return addr, port, nil
	}
}
