package my_errors

import "errors"

var (
	// DB hotel errors

	DBDriverNotAvailableErr = errors.New("DB driver not available")
	DBIsNotAvailableErr     = errors.New("DB is not available")
	CouldNotCreateHotelErr  = errors.New("couldn't create hotel")
	CouldNotUpdateHotelErr  = errors.New("couldn't update hotel")
	HotelWithIDNotFoundErr  = errors.New("hotel with ID not found")

	// DB room errors

	CouldNotCreateRoomErr = errors.New("couldn't create room")
	CouldNotUpdateRoomErr = errors.New("couldn't update room")
	RoomWithIDNotFoundErr = errors.New("room with ID not found")

	// Service hotel errors

	HotelValidationErr            = errors.New("hotel validation error")
	HotelWithoutNameErr           = errors.New("hotel without name")
	InvalidHotelRatingErr         = errors.New("invalid hotel rating")
	InvalidRoomsNumberErr         = errors.New("invalid rooms number")
	InvalidRoomsOccupiedNumberErr = errors.New("invalid rooms occupier number")
	NilPointerToHotelErr          = errors.New("nil pointer to hotel")

	// Service room errors
	RoomValidationErr      = errors.New("room validation error")
	NilPointerToRoomErr    = errors.New("nil pointer to room")
	InvalidRoomTypeErr     = errors.New("invalid room type")
	InvalidAvailabilityErr = errors.New("invalid availability")
	InvalidRoomNumberErr   = errors.New("invalid room number")
	InvalidRoomCostErr     = errors.New("invalid room cost")

	InvalidIDErr = errors.New("invalid id")
)
