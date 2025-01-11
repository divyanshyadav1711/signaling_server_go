
package server

import (
	"encoding/json"
	"net/http"
	"signaling_server/pkg/matching"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleConnection manages WebSocket connections.
func HandleConnection(w http.ResponseWriter, r *http.Request, log *logrus.Logger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("WebSocket upgrade failed:", err)
		return
	}
	user := matching.NewUser(conn)

	defer func() {
		matching.RemoveUser(user)
		matching.HandlePeerDisconnect(user)
		conn.Close()
	}()

	log.Infof("User %s connected", user.ID)
	matching.QueueUser(user)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Infof("User %s disconnected: %v", user.ID, err)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Warnf("Failed to parse message from user %s: %v", user.ID, err)
			continue
		}

		switch msg["type"].(string) {
		case "offer":
			handleOffer(user, msg["data"].(string), log)
		case "answer":
			handleAnswer(user, msg["data"].(string), log)
		case "ice-candidate":
			handleIceCandidate(user, msg["data"].(string), log)
		default:
			log.Warnf("Unknown message type from user %s: %v", user.ID, msg["type"])
			sendError(conn, "Unknown message type")
		}
	}
}

func handleOffer(user *matching.User, offer string, log *logrus.Logger) {
	log.Infof("Received offer from user %s", user.ID)
	peer := matching.GetPeer(user)
	if peer == nil {
		log.Warnf("No peer found for user %s", user.ID)
		sendError(user.Conn, "No peer available to process your offer")
		return
	}

	err := peer.Conn.WriteJSON(map[string]interface{}{
		"type": "offer",
		"data": offer,
		"from": user.ID,
	})
	if err != nil {
		log.Errorf("Failed to send offer to peer %s: %v", peer.ID, err)
	}
}

func handleAnswer(user *matching.User, answer string, log *logrus.Logger) {
	log.Infof("Received answer from user %s", user.ID)
	peer := matching.GetPeer(user)
	if peer == nil {
		log.Warnf("No peer found for user %s", user.ID)
		sendError(user.Conn, "No peer available to process your answer")
		return
	}

	err := peer.Conn.WriteJSON(map[string]interface{}{
		"type": "answer",
		"data": answer,
		"from": user.ID,
	})
	if err != nil {
		log.Errorf("Failed to send answer to peer %s: %v", peer.ID, err)
	}
}

func handleIceCandidate(user *matching.User, candidate string, log *logrus.Logger) {
	log.Infof("Received ICE candidate from user %s", user.ID)
	peer := matching.GetPeer(user)
	if peer == nil {
		log.Warnf("No peer found for user %s", user.ID)
		sendError(user.Conn, "No peer available to process your ICE candidate")
		return
	}

	err := peer.Conn.WriteJSON(map[string]interface{}{
		"type": "ice-candidate",
		"data": candidate,
		"from": user.ID,
	})
	if err != nil {
		log.Errorf("Failed to send ICE candidate to peer %s: %v", peer.ID, err)
	}
}

// sendError sends an error message to the user.
func sendError(conn *websocket.Conn, message string) {
	_ = conn.WriteJSON(map[string]interface{}{
		"type":    "error",
		"message": message,
	})
}
