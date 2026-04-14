package response

import "secret-santa-backend/internal/entity"

func EventToResponse(e *entity.Event) EventResponse {
	if e == nil {
		return EventResponse{}
	}
	return EventResponse{
		ID:             e.ID.String(),
		Title:          e.Title,
		OrganizerNotes: e.OrganizerNotes,
		OrganizerID:    e.OrganizerID.String(),
		StartDate:      e.StartDate,
		DrawDate:       e.DrawDate,
		Status:         string(e.Status),
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func EventsToResponse(events []entity.Event) []EventResponse {
	if events == nil {
		return []EventResponse{}
	}
	resp := make([]EventResponse, len(events))
	for i := range events {
		resp[i] = EventToResponse(&events[i])
	}
	return resp
}

func UserToResponse(u *entity.User) UserResponse {
	if u == nil {
		return UserResponse{}
	}
	return UserResponse{
		ID:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
	}
}

func ParticipantToResponse(p *entity.Participant) ParticipantResponse {
	if p == nil {
		return ParticipantResponse{}
	}
	return ParticipantResponse{
		ID:        p.ID.String(),
		EventID:   p.EventID.String(),
		UserID:    p.UserID.String(),
		CreatedAt: p.CreatedAt,
		UserName:  p.UserName,
		UserEmail: p.UserEmail,
	}
}

func ParticipantsToResponse(participants []entity.Participant) []ParticipantResponse {
	if participants == nil {
		return nil
	}
	resp := make([]ParticipantResponse, len(participants))
	for i := range participants {
		resp[i] = ParticipantToResponse(&participants[i])
	}
	return resp
}

func WishlistToResponse(w *entity.Wishlist) WishlistResponse {
	if w == nil {
		return WishlistResponse{}
	}
	participantID := ""
	if w.ParticipantID != nil {
		participantID = w.ParticipantID.String()
	}
	return WishlistResponse{
		ID:            w.ID.String(),
		ParticipantID: participantID,
		Visibility:    w.Visibility,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

func WishlistItemToResponse(item *entity.WishlistItem) WishlistItemResponse {
	if item == nil {
		return WishlistItemResponse{}
	}
	return WishlistItemResponse{
		ID:        item.ID.String(),
		Title:     item.Title,
		Link:      item.Link,
		ImageURL:  item.ImageURL,
		Price:     item.Price,
		CreatedAt: item.CreatedAt,
	}
}

func WishlistItemsToResponse(items []entity.WishlistItem) []WishlistItemResponse {
	if items == nil {
		return []WishlistItemResponse{}
	}
	resp := make([]WishlistItemResponse, len(items))
	for i := range items {
		resp[i] = WishlistItemToResponse(&items[i])
	}
	return resp
}

func AssignmentToResponse(a *entity.Assignment) AssignmentResponse {
	if a == nil {
		return AssignmentResponse{}
	}
	return AssignmentResponse{
		ID:           a.ID.String(),
		EventID:      a.EventID.String(),
		GiverID:      a.GiverID.String(),
		ReceiverID:   a.ReceiverID.String(),
		ReceiverName: a.ReceiverName,
		CreatedAt:    a.CreatedAt,
	}
}

func AssignmentsToResponse(assignments []entity.Assignment) []AssignmentResponse {
	if assignments == nil {
		return nil
	}
	resp := make([]AssignmentResponse, len(assignments))
	for i := range assignments {
		resp[i] = AssignmentToResponse(&assignments[i])
	}
	return resp
}
