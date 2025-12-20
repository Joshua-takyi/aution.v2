package websockets

import (
	"encoding/json"
	"time"
)

type NotificationType string

const (
	NotifBidPlaced       NotificationType = "BID_PLACED"
	NotifBidOutbid       NotificationType = "BID_OUTBID"
	NotifAuctionStarted  NotificationType = "AUCTION_STARTED"
	NotifAuctionEnding   NotificationType = "AUCTION_ENDING_SOON"
	NotifAuctionEnded    NotificationType = "AUCTION_ENDED"
	NotifAuctionWon      NotificationType = "AUCTION_WON"
	NotifAuctionLost     NotificationType = "AUCTION_LOST"
	NotifPaymentReminder NotificationType = "PAYMENT_REMINDER"
	NotifSystemMessage   NotificationType = "SYSTEM_MESSAGE"
)

type Notification struct {
	Type      NotificationType       `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Priority  string                 `json:"priority"` // "low", "normal", "high", "urgent"
	Message   string                 `json:"message"`
}

func NewNotification(notifType NotificationType, message string, data map[string]interface{}) Notification {
	return Notification{
		Type:      notifType,
		Timestamp: time.Now(),
		Data:      data,
		Priority:  "normal",
		Message:   message,
	}
}

// ToMessage converts a notification into a raw JSON message for the WebSocket
func (n Notification) ToMessage() ([]byte, error) {
	// Standard format: { "type": "notification", "payload": { ...notification } }
	msg := map[string]interface{}{
		"type":    "notification",
		"payload": n,
	}
	return json.Marshal(msg)
}
