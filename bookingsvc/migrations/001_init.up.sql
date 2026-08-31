CREATE TABLE bookings (
                          id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

                          user_id BIGINT NOT NULL,
                          hotel_id BIGINT NOT NULL,
                          room_id BIGINT NOT NULL,

                          date_from DATE NOT NULL,
                          date_to DATE NOT NULL,

                          status TEXT NOT NULL DEFAULT 'active'
);