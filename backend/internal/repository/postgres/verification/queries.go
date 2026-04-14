package verification

import "github.com/Masterminds/squirrel"

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func saveCodeQuery() squirrel.InsertBuilder {
	return qb.Insert("email_verification_codes").
		Columns("email", "code", "expires_at")
}

func getValidCodeQuery() squirrel.SelectBuilder {
	return qb.Select("id", "email", "code", "expires_at", "used", "created_at").
		From("email_verification_codes").
		Where("used = false").
		Where("expires_at > NOW()")
}

func markAsUsedQuery() squirrel.UpdateBuilder {
	return qb.Update("email_verification_codes").
		Set("used", true).
		Where("used = false")
}

func invalidateCodesQuery(email string) squirrel.UpdateBuilder {
	return qb.Update("email_verification_codes").
		Set("used", true).
		Where("email = ?", email).
		Where("used = false")
}
