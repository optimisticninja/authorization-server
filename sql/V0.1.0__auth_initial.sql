CREATE TABLE IF NOT EXISTS client (
	id						text NOT NULL PRIMARY KEY,
	secret				text NOT NULL,
	extra					text NOT NULL,
	redirect_uri	text NOT NULL
);

CREATE TABLE IF NOT EXISTS authorize (
	client				text NOT NULL references client(id),
	code					text NOT NULL,
	expires_in		int NOT NULL,
	scope					text NOT NULL,
	redirect_uri  text NOT NULL,
	state					text NOT NULL,
	extra					text NOT NULL,
	created_at		timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS access (
	client				text NOT NULL references client(id),
	authorize			text NOT NULL,
	previous			text NOT NULL,
	access_token	text NOT NULL PRIMARY KEY,
	refresh_token	text NOT NULL,
	expires_in		int NOT NULL,
	scope					text NOT NULL,
	redirect_uri	text NOT NULL,
	extra					text NOT NULL,
	created_at		timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh (
	token					text NOT NULL PRIMARY KEY,
	access				text NOT NULL
);
