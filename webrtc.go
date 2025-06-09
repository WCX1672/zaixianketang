package live

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type WebRTCManager struct {
	connections map[string]*webrtc.PeerConnection
	tracks      map[string]*webrtc.TrackLocalStaticRTP
	stunServer  string
	mutex       sync.RWMutex
}

func NewWebRTCManager(stunServer string) *WebRTCManager {
	return &WebRTCManager{
		connections: make(map[string]*webrtc.PeerConnection),
		tracks:      make(map[string]*webrtc.TrackLocalStaticRTP),
		stunServer:  stunServer,
	}
}

func ServeStream(manager *WebRTCManager, w http.ResponseWriter, r *http.Request, room, user string) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 创建PeerConnection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{manager.stunServer}},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println("PeerConnection error:", err)
		return
	}
	defer peerConnection.Close()

	// 保存连接
	manager.mutex.Lock()
	manager.connections[user] = peerConnection
	manager.mutex.Unlock()

	// 处理信令
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var signal map[string]interface{}
		if err := json.Unmarshal(msg, &signal); err != nil {
			log.Println("Signal decode error:", err)
			continue
		}

		switch signal["type"] {
		case "offer":
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(signal["sdp"].(string)), &offer); err != nil {
				log.Println("Offer decode error:", err)
				continue
			}

			if err := peerConnection.SetRemoteDescription(offer); err != nil {
				log.Println("SetRemoteDescription error:", err)
				continue
			}

			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				log.Println("CreateAnswer error:", err)
				continue
			}

			if err = peerConnection.SetLocalDescription(answer); err != nil {
				log.Println("SetLocalDescription error:", err)
				continue
			}

			answerBytes, err := json.Marshal(answer)
			if err != nil {
				log.Println("Answer marshal error:", err)
				continue
			}

			response := map[string]interface{}{
				"type": "answer",
				"sdp":  string(answerBytes),
			}

			if err := conn.WriteJSON(response); err != nil {
				log.Println("WriteJSON error:", err)
			}

		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(signal["candidate"].(string)), &candidate); err != nil {
				log.Println("Candidate decode error:", err)
				continue
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println("AddICECandidate error:", err)
			}
		}
	}
}
