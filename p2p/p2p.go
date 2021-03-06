package p2p

import (
	"fmt"
	"net/http"

	"github.com/chiwon99881/chyocoin/blockchain"
	"github.com/chiwon99881/chyocoin/utils"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// Upgrade is function of router
func Upgrade(rw http.ResponseWriter, r *http.Request) {
	openPort := r.URL.Query().Get("openPort")
	ip := utils.Splitter(r.RemoteAddr, ":", 0)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		if openPort == "" || ip == "" {
			return false
		}
		return true
	}
	fmt.Printf("%s wants an upgrade\n", openPort)
	// 3000 -> 4000 으로 가는 conn
	conn, err := upgrader.Upgrade(rw, r, nil)
	utils.HandleError(err)
	initPeer(conn, ip, openPort)
}

// AddPeer is function of p2p
func AddPeer(address, port, openPort string, broadcast bool) {
	fmt.Printf("%s wants to connect to port %s\n", openPort, port)
	// Dial은 해당 URL을 call하면 새로운 connection을 만들어 준다.
	// 즉, Port가 4000인 node가 이 function을 call하여 Dial을 실행하면 저 URL(ws://%s:%s/ws)에 대한 새로운 peer을 만들고
	// 그 만들어진 peer는 해당 URL에 대한 request가 실행되고 그 request는 위 Upgrade function을 호출하는 request handler를 request한다.
	// 그러면 http -> ws로 upgrade가 되고 connection이 생겨나게 된다. 즉, 4000과 3000이 각자의 connection이 생겨 연결된다.
	//즉, 다시 말해 Dial function은 위 Upgrade function을 다른 노드로부터 실행시키는 trigger이다.

	//이 function은 Port 4000이 Port 3000에게 연결을 하고 싶다는 function이고 그 연결을 Dial을 통해 3000이 위 Upgrade function을 실행할 수 있게 한다.

	// 4000 -> 3000 으로 가는 conn
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws?openPort=%s", address, port, openPort), nil)
	utils.HandleError(err)
	p := initPeer(conn, address, port)
	if broadcast {
		broadcastNewPeer(p)
	}
	sendNewestBlock(p)
}

// BroadcastNewBlock is function of broadcast new block for all peers.
func BroadcastNewBlock(b *blockchain.Block) {
	Peers.m.Lock()
	Peers.m.Unlock()
	for _, p := range Peers.v {
		notifyNewBlock(b, p)
	}
}

// BroadcastNewTx is function of broadcast new tx for all peers.
func BroadcastNewTx(tx *blockchain.Tx) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v {
		notifyNewTx(tx, p)
	}
}

func broadcastNewPeer(newPeer *peer) {
	for key, p := range Peers.v {
		if key != newPeer.key {
			payload := fmt.Sprintf("%s:%s", newPeer.key, p.port)
			notifyNewPeer(payload, p)
		}
	}
}
