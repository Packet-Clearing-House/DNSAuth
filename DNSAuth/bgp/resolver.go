package bgp

// This module

import (
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/server"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/packet/bgp"
	"github.com/asergeyev/nradix"
	"errors"
	log "github.com/sirupsen/logrus"
)


type BGPConf struct {
	LocalAs uint32 `cfg:"local-as; required; "`
	RouterId string `cfg:"router-id; required; "`
	PeerAddr string `cfg:"peer-addr; required; "`
	PeerAs uint32 `cfg:"peer-as; required; "`
	ListenPort int32 `cfg:"listen-port; -1; "`
	ConnectPort uint16 `cfg:"connect-port; 179; "`
	RemoveAsFromPath []uint32 `cfg:"remove-as-from-path"`
}

type RIBEntry struct {
	Timestamp int64
	Prefix string
	Path []uint32
}


var server *gobgp.BgpServer
var watcher *gobgp.Watcher
var tree *nradix.Tree

func IsUp() bool {
	if tree == nil {
		return false
	}
	return true
}

func Resolve(ip string) (*RIBEntry, error) {

	if tree == nil {
		return nil, errors.New("RIB table not ready... BGP peer not up!")
	}
	if value, _ := tree.FindCIDR(ip); value != nil {
		entry := value.(*RIBEntry)
		return entry, nil
	}
	return nil, errors.New("Prefix not found!")
}


func Start(conf BGPConf) error {

	//log.SetLevel(log.DebugLevel)

	if server != nil {
		return errors.New("BGP server already started!")
	}


	log.Printf("Starting BGP server: (router-id :%s, local-as: %d, peer-address: %s, remote-as: %d)",
		conf.RouterId, conf.LocalAs,
		conf.PeerAddr, conf.PeerAs)


	global := &config.Global{
		Config: config.GlobalConfig{
			As:       conf.LocalAs,
			RouterId: conf.RouterId,
			Port:     conf.ListenPort, // gobgp won't listen on tcp:179
		},
		//ApplyPolicy: config.ApplyPolicy{
		//	Config: config.ApplyPolicyConfig{
		//		DefaultImportPolicy: config.DEFAULT_POLICY_TYPE_REJECT_ROUTE,
		//		DefaultExportPolicy: config.DEFAULT_POLICY_TYPE_REJECT_ROUTE,
		//	},
		//},
	}

	neighbor := &config.Neighbor{
		Transport: config.Transport{
			Config: config.TransportConfig{
				RemotePort: conf.ConnectPort,
			},
		},
		Config: config.NeighborConfig{
			NeighborAddress: conf.PeerAddr,
			PeerAs: conf.PeerAs,
		},
	}

	server := gobgp.NewBgpServer()
	go server.Serve()

	// start grpc api server. this is not mandatory
	// but you will be able to use `gobgp` cmd with this.
	g := api.NewGrpcServer(server, ":50051")
	go g.Serve()

	if err := server.Start(global); err != nil {
		return err
	}

	if err := server.AddNeighbor(neighbor); err != nil {
		return err
	}

	watcher := server.Watch(
		gobgp.WatchBestPath(true),
		gobgp.WatchUpdate(true),
		gobgp.WatchPeerState(true),
	)


	go func() {
		for {
			select {
			case ev, ok := <-watcher.Event():

				if !ok {
					log.Println("END BGP GOROUTINE!")
					return
				}

				switch msg := ev.(type) {

				case *gobgp.WatchEventUpdate:
					//msg.Is
					if len(msg.PathList) == 0 {
						log.Println("Full table transfer complete: BGP peer up!")
					}

				case *gobgp.WatchEventBestPath:
					for _, path := range msg.PathList {
						if path.IsWithdraw {
							tree.DeleteCIDR(path.GetNlri().String())
							//metricsBgpPrefixes.Dec()
						} else {

							list := path.GetAsList()
							if len(list) == 0 {
								list = []uint32{conf.LocalAs}
							} else  if uint32(len(list)) > 1 {
								cpt := 0
								for _, as := range list {
									for _, ras := range conf.RemoveAsFromPath {
										if ras == as {
											cpt += 1
										}
									}
								}
								list = list[cpt:]
							}

							entry := &RIBEntry{
								path.GetTimestamp().Unix(),
								path.GetNlri().String(),
								list,
							}

							tree.AddCIDR(entry.Prefix, entry)
							//metricsBgpPrefixes.Inc()
						}
					}

				case *gobgp.WatchEventPeerState:
					if msg.State == bgp.BGP_FSM_ESTABLISHED {
						log.Println("BGP peer connected: starting full table transfer...")
						tree = nradix.NewTree(600000)
					} else if msg.State != bgp.BGP_FSM_ESTABLISHED  {
						//metricsBgpFlaps.Inc()
						tree = nil
						log.Println("BGP peer disconnected...")
					}
				}
			}
		}
	}()

	return nil
}

func Stop() {
	log.Printf("Stopping BGP server...")
	watcher.Stop()
	server.Stop()
	log.Printf(" Done!\n")
}

func Reload(conf BGPConf) {
	Stop()
	Start(conf)
}
