package definitions

import "errors"

var (
	ErrNotFound                    = errors.New("not found")
	ErrForbidden                   = errors.New("forbidden")
	ErrInvalidUUID                 = errors.New("invalid uuid")
	ErrUnauthorized                = errors.New("unauthorized")
	ErrConflict                    = errors.New("conflict")
	ErrInvalidOAuthCode            = errors.New("invalid oauth code")
	ErrMissingOAuthCode            = errors.New("missing oauth code")
	ErrInvalidUserInput            = errors.New("invalid user input")
	ErrInvalidOAuthUserInfo        = errors.New("invalid oauth user info")
	ErrUserNotFound                = errors.New("user not found")
	ErrEventNotFound               = errors.New("event not found")
	ErrEventAlreadyFinished        = errors.New("event already finished")
	ErrInvalidEventState           = errors.New("invalid event state")
	ErrNotEnoughParticipants       = errors.New("not enough participants")
	ErrDuplicateParticipant        = errors.New("duplicate participant")
	ErrAlreadyParticipating        = errors.New("already participating")
	ErrParticipantNotFound         = errors.New("participant not found")
	ErrWishlistNotFound            = errors.New("wishlist not found")
	ErrInvalidWishlistVisibility   = errors.New("invalid wishlist visibility")
	ErrWishlistVisibilityForbidden = errors.New("wishlist visibility forbidden")
	ErrNotOrganizer                = errors.New("not organizer")
	ErrNotSanta                    = errors.New("not santa")
	ErrAssignmentNotFound          = errors.New("assignment not found")
)
