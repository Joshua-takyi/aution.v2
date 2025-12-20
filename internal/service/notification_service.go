package service

import (
	"github.com/joshua-takyi/auction/internal/websockets"
)

type NotificationService struct {
	wsManager *websockets.Manager
}

func NewNotificationService(wsManager *websockets.Manager) *NotificationService {
	return &NotificationService{
		wsManager: wsManager,
	}
}

func (s *NotificationService) NotifyBidPlaced(roomID string, bidderID string, amount string) {
	notif := websockets.NewNotification(
		websockets.NotifBidPlaced,
		"A new bid of "+amount+" has been placed!",
		map[string]interface{}{
			"bidderId": bidderID,
			"amount":   amount,
		},
	)
	s.wsManager.BroadcastNotificationToRoom(roomID, notif)
}

func (s *NotificationService) NotifyOutbid(userID string, roomID string, newAmount string) {
	notif := websockets.NewNotification(
		websockets.NotifBidOutbid,
		"You have been outbid! The new price is "+newAmount,
		map[string]interface{}{
			"roomId":    roomID,
			"newAmount": newAmount,
		},
	)
	notif.Priority = "high"
	s.wsManager.SendNotificationToUser(userID, notif)
}

func (s *NotificationService) NotifyAuctionWon(userID string, auctionTitle string) {
	notif := websockets.NewNotification(
		websockets.NotifAuctionWon,
		"Congratulations! You won the auction for "+auctionTitle,
		map[string]interface{}{
			"title": auctionTitle,
		},
	)
	notif.Priority = "high"
	s.wsManager.SendNotificationToUser(userID, notif)
}
