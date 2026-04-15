package wishlist

import (
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var qb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Все SELECT-запросы выбирают 6 колонок: id, participant_id, user_id, visibility, created_at, updated_at

func createParticipantWishlistQuery(participantID uuid.UUID, visibility string) squirrel.InsertBuilder {
	return qb.Insert("wishlists").
		Columns("participant_id", "visibility").
		Values(participantID, visibility).
		Suffix("RETURNING id, participant_id, user_id, visibility, created_at, updated_at")
}

func createPersonalWishlistQuery(userID uuid.UUID) squirrel.InsertBuilder {
	return qb.Insert("wishlists").
		Columns("user_id", "visibility").
		Values(userID, "public").
		Suffix("RETURNING id, participant_id, user_id, visibility, created_at, updated_at")
}

func createWishlistItemQuery() squirrel.InsertBuilder {
	return qb.Insert("wishlist_items").
		Columns("wishlist_id", "title", "link", "image_url", "price")
	// id, created_at — БД сама
}

func getWishlistByParticipantQuery(participantID string) squirrel.SelectBuilder {
	return qb.Select("id", "participant_id", "user_id", "visibility", "created_at", "updated_at").
		From("wishlists").
		Where(squirrel.Eq{"participant_id": participantID})
}

func getWishlistByUserIDQuery(userID uuid.UUID) squirrel.SelectBuilder {
	return qb.Select("id", "participant_id", "user_id", "visibility", "created_at", "updated_at").
		From("wishlists").
		Where(squirrel.Eq{"user_id": userID})
}

func getWishlistItemsQuery(wishlistID string) squirrel.SelectBuilder {
	return qb.Select("id", "wishlist_id", "title", "link", "image_url", "price", "created_at").
		From("wishlist_items").
		Where(squirrel.Eq{"wishlist_id": wishlistID}).
		OrderBy("created_at ASC")
}

func countWishlistItemsQuery(wishlistID string) squirrel.SelectBuilder {
	return qb.Select("COUNT(*)").
		From("wishlist_items").
		Where(squirrel.Eq{"wishlist_id": wishlistID})
}

func updateWishlistItemQuery(itemID string) squirrel.UpdateBuilder {
	return qb.Update("wishlist_items").
		Where(squirrel.Eq{"id": itemID})
}

func deleteWishlistItemQuery(itemID string) squirrel.DeleteBuilder {
	return qb.Delete("wishlist_items").
		Where(squirrel.Eq{"id": itemID})
}

func getWishlistItemByIDQuery(itemID uuid.UUID) squirrel.SelectBuilder {
	return qb.Select(
		"id", "wishlist_id", "title", "link", "image_url", "price", "created_at",
	).
		From("wishlist_items").
		Where(squirrel.Eq{"id": itemID})
}

func getWishlistByIDQuery(wishlistID uuid.UUID) squirrel.SelectBuilder {
	return qb.Select("id", "participant_id", "user_id", "visibility", "created_at", "updated_at").
		From("wishlists").
		Where(squirrel.Eq{"id": wishlistID})
}
