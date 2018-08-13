package proto

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Packet-Clearing-House/DNSAuth/libs/utils"
)

type Connector interface {
	Serve(func(conn *ProtoConn)) error
}

type Dialer struct {
	//TagsAdd string
	//TagsPass string
	Tag  string
	Addr string
	stop chan bool
}

func (d *Dialer) Close() {
	log.Println("oué")
	close(d.stop)
	log.Println("oué")
}

func (d *Dialer) Serve(onNewConn func(pc *ProtoConn)) error {
	log.Println("Serving")
	d.stop = make(chan bool)
	log.Println(d)
	//log.Println(d.Addr)
	_, _, err := utils.SplitHostPort(d.Addr)

	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-d.stop:
				return
			default:
				conn, err := net.Dial("tcp", d.Addr)
				if err != nil {
					log.Println(err)
					time.Sleep(time.Second)
				} else {
					_, port, _ := net.SplitHostPort(d.Addr)
					pc := NewProtoConn(conn, port)
					go onNewConn(pc)

					select {
					case <-d.stop:
						pc.Close()
						return
					case <-pc.closed:
						break
					}
				}
			}
		}
	}()
	return nil
}

type Listener struct {
	Tag       string
	Addr      string
	ACL       []string
	stop      chan bool
	stopGroup sync.WaitGroup
	ln        *utils.ACListener
}

func (l *Listener) Stop() error {
	close(l.stop)
	l.ln.Close()
	l.stopGroup.Wait()
	log.Println("Waiting over")
	return nil

}

func (l *Listener) Serve(onNewConn func(pc *ProtoConn)) error {

	l.stop = make(chan bool)
	l.stopGroup = sync.WaitGroup{}
	_, _, err := utils.SplitHostPort(l.Addr)

	if err != nil {
		return err
	}

	acl, err := utils.ParseACLFromStrings(l.ACL)
	if err != nil {
		return err
	}

	l.ln, err = utils.ACLListen("tcp", l.Addr, acl)
	if err != nil {
		return err
	}

	go func() {
		l.stopGroup.Add(1)
		defer l.stopGroup.Done()
		for {
			conn, err := l.ln.Accept()
			if err != nil {
				select {
				case <-l.stop:
					return
				default:
				}
				log.Println(err)
			} else {

				_, port, _ := net.SplitHostPort(l.Addr)
				pc := NewProtoConn(conn, port)
				log.Println("ADEDESS")
				log.Println(pc.LocalAddr())
				log.Println(pc.RemoteAddr())
				go func() {
					l.stopGroup.Add(1)
					defer l.stopGroup.Done()
					go onNewConn(pc)
					select {
					case <-pc.closed:
					case <-l.stop:
						pc.Close()
					}
					log.Println("Exit")
				}()
			}
		}
	}()

	//ip := net.ParseIP(l.Addr)
	//if ip == nil {
	//	return errors.New("Wrong ip address")
	//}
	//
	//listener, err := net.Listen("tcp", l.Addr)
	//if err != nil {
	//	return err
	//}
	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		log.Println(err)
	//
	//	}
	//}
	return nil
}

//type ConnHandler func(conn onNewConn, ) onNewConn

type onNewConn func(conn ProtoConn)
type onNewMsg func(msg []byte)

type ProtoConn struct {
	OriginPort string
	net.Conn
	closed chan bool
}

func NewProtoConn(conn net.Conn, originPort string) *ProtoConn {
	return &ProtoConn{originPort, conn, make(chan bool)}
}

func (p ProtoConn) Close() {
	p.Conn.Close()
	close(p.closed)
}

func (p ProtoConn) Wait() {
	<-p.closed
}

func (p ProtoConn) ReadMsg() ([]byte, error) {
	datalen := make([]byte, 2)

	log.Println("ee")
	// Reading the 2 first bytes (size of the following message)
	if _, err := io.ReadFull(p, datalen); err != nil {
		return nil, err
	}

	log.Println("First")

	// Translating those 2 bytes and reading the message
	len := binary.BigEndian.Uint16(datalen)

	data := make([]byte, len)
	log.Println("Waiting for ", len)
	if _, err := io.ReadFull(p, data); err != nil {
		return nil, err
	}
	log.Println("e")

	return data, nil
}

func (p ProtoConn) WriteMsg(msg []byte) error {
	len := uint16(len(msg))
	lenbuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenbuf, len)
	if _, err := p.Write(lenbuf); err != nil {
		return err
	}
	if _, err := p.Write(msg); err != nil {
		return err
	}
	return nil
}

//func (p *ProtoConn) WriteFrom(input chan []byte) error {
//	for item := range input {
//		if err := p.WriteMsg(item); err != nil {
//			return err
//		}
//	}
//	return nil
//}

//

//func (p *ProtoConn) ReadMsgForever(onNewMsg onNewMsg) {
//	p.stop = make(chan bool)
//	for {
//		if msg, err := p.ReadMsg(); err != nil {
//			select{
//			case <-p.stop:
//				return
//			default:
//			}
//			fmt.Println("ERROR reading message from " + utils.GetHost(p.LocalAddr()), " to " + utils.GetHost(p.RemoteAddr()))
//			fmt.Println("Closing connection...")
//			return
//		} else {
//			err := onNewMsg(msg)
//			if err != nil {
//				return
//			}
//		}
//	}
//}
