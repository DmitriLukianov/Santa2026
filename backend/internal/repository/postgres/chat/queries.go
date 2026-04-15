package chat

import "github.com/Masterminds/squirrel"

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// createMessageQuery — DB-first (убрали id и created_at)
func createMessageQuery() squirrel.InsertBuilder {
	return qb.Insert("messages").
		Columns("event_id", "sender_id", "receiver_id", "content")
}

func getMessagesByPairQuery(eventID, user1ID, user2ID string) squirrel.SelectBuilder {
	return qb.Select("id", "event_id", "sender_id", "receiver_id", "content", "created_at").
		From("messages").
		Where(squirrel.Eq{"event_id": eventID}).
		Where(
			squirrel.Or{
				squirrel.And{
					squirrel.Eq{"sender_id": user1ID},
					squirrel.Eq{"receiver_id": user2ID},
				},
				squirrel.And{
					squirrel.Eq{"sender_id": user2ID},
					squirrel.Eq{"receiver_id": user1ID},
				},
			},
		).
		OrderBy("created_at ASC")
}

func countMessagesByPairQuery(eventID, user1ID, user2ID string) squirrel.SelectBuilder {
	return qb.Select("COUNT(*)").
		From("messages").
		Where(squirrel.Eq{"event_id": eventID}).
		Where(
			squirrel.Or{
				squirrel.And{
					squirrel.Eq{"sender_id": user1ID},
					squirrel.Eq{"receiver_id": user2ID},
				},
				squirrel.And{
					squirrel.Eq{"sender_id": user2ID},
					squirrel.Eq{"receiver_id": user1ID},
				},
			},
		)
}
