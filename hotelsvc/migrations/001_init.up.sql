CREATE TABLE hotels (
                        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

                        name TEXT NOT NULL,
                        address TEXT NOT NULL,
                        rating INTEGER NOT NULL,

                        rooms_num INTEGER NOT NULL DEFAULT 0,
                        rooms_occupied INTEGER NOT NULL DEFAULT 0,

                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE rooms (
                       id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

                       hotel_id BIGINT NOT NULL
                           REFERENCES hotels(id)
                               ON DELETE CASCADE,

                       room_num INTEGER NOT NULL,
                       type TEXT NOT NULL,
                       cost INTEGER NOT NULL,

                       available INTEGER NOT NULL DEFAULT 2,

                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);