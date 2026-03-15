package notification

import "fmt"

type Notification struct {
	Title string
	Body  string
	Data  map[string]string
}

func ReportResolved(reportID, address string) *Notification {
	return &Notification{
		Title: "Issue Resolved!",
		Body:  fmt.Sprintf("Your reported issue at %s has been resolved. Thank you for keeping your city clean!", address),
		Data: map[string]string{
			"type":      "report_resolved",
			"report_id": reportID,
		},
	}
}

func BookingAssigned(collectorName, timeSlot string) *Notification {
	return &Notification{
		Title: "Collector Assigned",
		Body:  fmt.Sprintf("%s will arrive during %s for your pickup.", collectorName, timeSlot),
		Data: map[string]string{
			"type": "booking_assigned",
		},
	}
}

func CollectorNearby(eta string) *Notification {
	return &Notification{
		Title: "Collector Nearby",
		Body:  fmt.Sprintf("Your collector is nearby! Estimated arrival: %s", eta),
		Data: map[string]string{
			"type": "collector_nearby",
		},
	}
}

func BadgeUnlocked(badgeName string) *Notification {
	return &Notification{
		Title: "Badge Unlocked!",
		Body:  fmt.Sprintf("Congratulations! You earned the '%s' badge!", badgeName),
		Data: map[string]string{
			"type":  "badge_unlocked",
			"badge": badgeName,
		},
	}
}

func PointsEarned(points int, reason string) *Notification {
	return &Notification{
		Title: "Points Earned!",
		Body:  fmt.Sprintf("You earned +%d points for %s", points, reason),
		Data: map[string]string{
			"type":   "points_earned",
			"points": fmt.Sprintf("%d", points),
		},
	}
}

func LeaderboardRank(rank int) *Notification {
	return &Notification{
		Title: "Leaderboard Update",
		Body:  fmt.Sprintf("You're now ranked #%d on the leaderboard! Keep going!", rank),
		Data: map[string]string{
			"type": "leaderboard_rank",
			"rank": fmt.Sprintf("%d", rank),
		},
	}
}

func WithdrawalSuccess(amount float64) *Notification {
	return &Notification{
		Title: "Withdrawal Successful",
		Body:  fmt.Sprintf("₹%.2f has been transferred to your bank account.", amount),
		Data: map[string]string{
			"type":   "withdrawal_success",
			"amount": fmt.Sprintf("%.2f", amount),
		},
	}
}

func WithdrawalFailed(reason string) *Notification {
	return &Notification{
		Title: "Withdrawal Failed",
		Body:  fmt.Sprintf("Your withdrawal could not be processed: %s", reason),
		Data: map[string]string{
			"type":   "withdrawal_failed",
			"reason": reason,
		},
	}
}

func NewBookingForCollector(address, timeSlot, wasteType string) *Notification {
	return &Notification{
		Title: "New Pickup Assignment",
		Body:  fmt.Sprintf("New %s pickup at %s during %s", wasteType, address, timeSlot),
		Data: map[string]string{
			"type": "new_booking",
		},
	}
}
