package invitation

import "github.com/Masterminds/squirrel"

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func createInvitationQuery() squirrel.InsertBuilder {
	return qb.Insert("invitations").
		Columns("event_id", "token", "expires_at", "created_by")
}

func getInvitationByTokenQuery() squirrel.SelectBuilder {
	return qb.Select(
		"id", "event_id", "token", "expires_at",
		"created_by", "created_at",
	).
		From("invitations")
}

func getActiveInvitationByEventQuery(eventID string) squirrel.SelectBuilder {
	return qb.Select(
		"id", "event_id", "token", "expires_at",
		"created_by", "created_at",
	).
		From("invitations").
		Where(squirrel.Eq{"event_id": eventID}).
		Where("expires_at > NOW()").
		OrderBy("created_at DESC").
		Limit(1)
}
