-- create users
CREATE TABLE IF NOT EXISTS users(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR(100) UNIQUE NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- create rooms
CREATE TABLE IF NOT EXISTS rooms(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	-- room type: direct or group
	-- room_type VARCHAR(10) NOT NULL CHECK(room_type IN ('direct', 'group')) DEFAULT 'group',
	creator_id UUID NOT NULL REFERENCES users(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- create room members
CREATE TABLE IF NOT EXISTS room_members(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	room_id UUID NOT NULL REFERENCES rooms(id),
	user_id UUID NOT NULL REFERENCES users(id),
	-- role: admin or member
	role VARCHAR(10) NOT NULL CHECK(role IN ('admin', 'member')) DEFAULT 'member',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	UNIQUE(room_id, user_id)
);

-- create messages
CREATE TABLE IF NOT EXISTS messages(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	room_id UUID NOT NULL REFERENCES rooms(id),
	-- remove sender info when message is deleted
	sender_id UUID  REFERENCES users(id) ON DELETE SET NULL,
	sender_username VARCHAR(100) NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index to access room messages in descending order
CREATE INDEX IF NOT EXISTS idx_messages_room_created_at ON messages(room_id, created_at DESC)