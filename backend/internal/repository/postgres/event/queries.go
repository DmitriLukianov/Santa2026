package event

import "github.com/Masterminds/squirrel"

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func getEventQuery() squirrel.SelectBuilder {
	return psql.Select(
		"events.id",
		"events.title",
		"events.organizer_notes",
		"events.organizer_id",
		"events.start_date",
		"events.draw_date",
		"events.status",
		"events.created_at",
		"events.updated_at",
	).
		From("events").
		Where("events.deleted_at IS NULL")
}

func listEventsQuery() squirrel.SelectBuilder {
	return getEventQuery()
}

func createEventQuery() squirrel.InsertBuilder {
	return psql.Insert("events").
		Columns(
			"title", "organizer_notes",
			"organizer_id", "start_date", "draw_date",
			"status",
		)
}

func updateEventQuery() squirrel.UpdateBuilder {
	return psql.Update("events")
}

func deleteEventQuery() squirrel.DeleteBuilder {
	return psql.Delete("events")
}

func getEventsForUserQuery() squirrel.SelectBuilder {
	return getEventQuery().
		Distinct().
		LeftJoin("participants p ON events.id = p.event_id").
		OrderBy("events.created_at DESC")
}
