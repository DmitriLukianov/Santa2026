package participant

import "github.com/Masterminds/squirrel"

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func createParticipantQuery() squirrel.InsertBuilder {
	return qb.Insert("participants").
		Columns("event_id", "user_id")
	// id, created_at — пусть БД сама заполняет
}

func getParticipantByIDQuery(id string) squirrel.SelectBuilder {
	return qb.Select("id", "event_id", "user_id", "created_at").
		From("participants").
		Where(squirrel.Eq{"id": id})
}

func getParticipantsByEventQuery(eventID string) squirrel.SelectBuilder {
	return qb.Select(
		"p.id", "p.event_id", "p.user_id", "p.created_at",
		"COALESCE(u.name, '')", "COALESCE(u.email, '')",
	).
		From("participants p").
		LeftJoin("users u ON u.id = p.user_id").
		Where(squirrel.Eq{"p.event_id": eventID}).
		OrderBy("p.created_at ASC")
}

func countParticipantsByEventQuery(eventID string) squirrel.SelectBuilder {
	return qb.Select("COUNT(*)").
		From("participants p").
		Where(squirrel.Eq{"p.event_id": eventID})
}

func deleteParticipantQuery(id string) squirrel.DeleteBuilder {
	return qb.Delete("participants").
		Where(squirrel.Eq{"id": id})
}

func getParticipantByUserAndEventQuery(userID, eventID string) squirrel.SelectBuilder {
	return qb.Select("id", "event_id", "user_id", "created_at").
		From("participants").
		Where(squirrel.Eq{"user_id": userID, "event_id": eventID}).
		Limit(1)
}
