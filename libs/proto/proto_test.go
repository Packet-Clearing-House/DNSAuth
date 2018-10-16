package proto_test

import (
	"encoding/binary"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Packet-Clearing-House/DNSAuth/DNSAuth/libs/dnsdist"
	. "github.com/Packet-Clearing-House/DNSAuth/DNSAuth/libs/proto"
	"github.com/golang/protobuf/proto"
)

var from = "172.10.0.3"
var msgtype = dnsdist.PBDNSMessage_DNSResponseType
var qname = "test"
var qtype = uint32(0)
var qclass = uint32(0)

var msg = dnsdist.PBDNSMessage{
	Type: &msgtype,
	From: net.ParseIP(from),
	Question: &dnsdist.PBDNSMessage_DNSQuestion{
		QName:  &qname,
		QType:  &qtype,
		QClass: &qclass,
	},
}

func TestProtoDialer(t *testing.T) {

	log.Println("asdasds")
	log.Println("here")

	received := make(chan bool)

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		} else {
			log.Println("Connected")

			msgb, err := proto.Marshal(&msg)
			if err != nil {
				t.Fatal(err)
			}
			log.Println("Writing ", len(msgb))
			len := uint16(len(msgb))
			lenbuf := make([]byte, 2)
			binary.BigEndian.PutUint16(lenbuf, len)
			conn.Write(lenbuf)
			conn.Write(msgb)

			if mr := <-received; mr != true {
				t.Fatal("Coul")
			}
			log.Println("OWJFG?")

			//conn.Write(lenbuf)
			//conn.Write(msgb)
			//
			//<-received
			conn.Close()
		}
		listener.Close()
		log.Println("OWJ")
	}()

	time.Sleep(time.Second * 2)
	t.Log("here")
	log.Println("asdasds")

	d := Dialer{Addr: "0.0.0.0:8080"}
	//log.Println(d)
	err = d.Serve(func(conn *ProtoConn) {
		msg, err := conn.ReadMsg()
		if err != nil {
			t.Fatal(err)
		} else {
			log.Println(msg)
		}
	})

	if err != nil {
		t.Fatal(err)
	}

	log.Println("END?")

	received <- true
	d.Close()

}

func TestProtoListener(t *testing.T) {

	l := Listener{
		Tag:  "",
		Addr: "0.0.0.0:8080",
		ACL:  []string{"127.0.0.1", "::1"},
	}

	//ChanConn1 := make(chan bool)
	log.Println("aasd")

	var cpt = 0

	err := l.Serve(func(conn *ProtoConn) {
		log.Println("Connnected!")
		cpt += 1

		ms, err := conn.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(ms)
		buf, _ := proto.Marshal(&msg)
		conn.WriteMsg(buf)
		conn.Close()

	})

	if err != nil {
		log.Fatal(err)
	}

	// Trying first to connect once and exchange messages
	if conn, err := net.Dial("tcp", "127.0.0.1:8080"); err != nil {
		t.Fatal(err)
	} else {
		//ChanConn1 <-true
		pc := NewProtoConn(conn)
		log.Println("ADEDESS")
		log.Println(pc.LocalAddr())
		log.Println(pc.RemoteAddr())

		buf, _ := proto.Marshal(&msg)
		pc.WriteMsg(buf)
		msg, err := pc.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(msg)
		conn.Close()
	}

	l.Stop()
	l.Serve(func(conn *ProtoConn) {})

	//time.Sleep(time.Second * 5)
	if cpt != 3 {
		t.Fatal("Excecting 3 incoming connections, got only ", cpt)
	}

}
