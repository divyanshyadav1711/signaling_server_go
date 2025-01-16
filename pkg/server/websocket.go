package server

import (
	"encoding/json"
	"net/http"
	"signaling_server/pkg/matching"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

//	var upgrader = websocket.Upgrader{
//	    CheckOrigin: func(r *http.Request) bool {
//	        origin := r.Header.Get("Origin")
//	        allowedOrigins := []string{"http://localhost:3000", "http://example.com"} // Add trusted origins
//	        for _, o := range allowedOrigins {
//	            if origin == o {
//	                return true
//	            }
//	        }
//	        return false
//	    },
//	}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development. Restrict in production.
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

	sendConnectData(user, log)
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
			log.Info("Received offer from user %s %s", user.ID, msg["offer"])
			handleOffer(user, msg["offer"].(map[string]interface{}), log)
		case "answer":
			handleAnswer(user, msg["answer"].(map[string]interface{}), log)
		case "ice-candidate":
			handleIceCandidate(user, msg["candidate"].(map[string]interface{}), log)
		case "next":
			handleNext(user, log)
		case "exit":
			handleExit(user, log)
		case "chat":
			handleChat(user, msg["data"].(string), log)
		default:
			log.Warnf("Unknown message type from user %s: %v", user.ID, msg["type"])
			sendError(conn, "Unknown message type")
		}
	}
	log.Infof("Queue size is %d ", matching.GetQueueSize())
}

// sendConnectData sends the connect-data message to the client.
func sendConnectData(user *matching.User, log *logrus.Logger) {
	message := map[string]interface{}{
		"type":        "connect-data",
		"userId":      user.ID,
		"isInitiator": true, // You can modify this based on your matching logic
	}

	err := user.Conn.WriteJSON(message)
	if err != nil {
		log.Errorf("Failed to send connect-data to user %s: %v", user.ID, err)
	}
}

func handleNext(user *matching.User, log *logrus.Logger) {
	peer := matching.GetPeer(user)
	// remove peer
	if peer != nil {
		// Notify the peer that the user has disconnected
		matching.NotifyDiconnectionToPeer(peer)
		matching.RemovePeer(peer)
		matching.QueueUser(peer)
	}
	// remove user
	matching.RemovePeer(user)
	matching.QueueUser(user)

	log.Infof("User %s requested next", user.ID)
}

func handleExit(user *matching.User, log *logrus.Logger) {
	log.Infof("User %s exited", user.ID)
	matching.HandleDisconnect(user)
}

func handleOffer(user *matching.User, offer map[string]interface{}, log *logrus.Logger) {
	log.Infof("Received offer from user %s", user.ID)
	peer := matching.GetPeer(user)
	if peer == nil {
		log.Warnf("No peer found for user %s", user.ID)
		sendError(user.Conn, "No peer available to process your offer")
		return
	}

	err := peer.Conn.WriteJSON(map[string]interface{}{
		"type":  "offer",
		"offer": offer,
		"from":  user.ID,
	})
	if err != nil {
		log.Errorf("Failed to send offer to peer %s: %v", peer.ID, err)
	}
}

func handleAnswer(user *matching.User, answer map[string]interface{}, log *logrus.Logger) {
	log.Infof("Received answer from user %s", user.ID)
	peer := matching.GetPeer(user)
	if peer == nil {
		log.Warnf("No peer found for user %s", user.ID)
		sendError(user.Conn, "No peer available to process your answer")
		return
	}

	err := peer.Conn.WriteJSON(map[string]interface{}{
		"type":   "answer",
		"answer": answer,
		"from":   user.ID,
	})
	if err != nil {
		log.Errorf("Failed to send answer to peer %s: %v", peer.ID, err)
	}
}

func handleChat(user *matching.User, data string, log *logrus.Logger) {
    peer := matching.GetPeer(user)

    if peer != nil {
        
        message := map[string]interface{}{
            "type": "chat",
            "data": data,  
        }

        
        jsonMessage, err := json.Marshal(message)
        if err != nil {
            log.Errorf("Failed to marshal chat message: %v", err)
            return
        }

        
        err = peer.Conn.WriteMessage(websocket.TextMessage, jsonMessage)
        if err != nil {
            log.Errorf("Failed to send chat message to peer %s: %v", peer.ID, err)
        }
    }
}

func handleIceCandidate(user *matching.User, candidate map[string]interface{}, log *logrus.Logger) {
	log.Infof("Received ICE candidate from user %s", user.ID)
	peer := matching.GetPeer(user)
	if peer == nil {
		log.Warnf("No peer found for user %s", user.ID)
		sendError(user.Conn, "No peer available to process your ICE candidate")
		return
	}

	err := peer.Conn.WriteJSON(map[string]interface{}{
		"type":      "ice-candidate",
		"candidate": candidate,
		"from":      user.ID,
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
