package splenda

var schema = []string{
	// The users table; one entry per named user irrespective of game.
	"CREATE TABLE users (" +
		"id varchar(256) PRIMARY KEY, " +
		"hash varchar(256) NOT NULL" +
		")",

	// Enumeration of game states.
	"CREATE TYPE game_state AS ENUM (" +
		"'play', 'picknoble', 'gameover'" +
		")",
	// Enumeration of colors.
	"CREATE TYPE color AS ENUM (" +
		"'red', 'blue', 'green', 'white', 'black', 'wild'" +
		")",

	// The games table; one entry per active game.
	"CREATE TABLE games (" +
		"id varchar(256) PRIMARY KEY, " +
		"ts integer NOT NULL, " +
		"state game_state NOT NULL, " +
		"current varchar(256) NOT NULL REFERENCES users" +
		")",
	// How many of each type of coin are in the bank.
	"CREATE TABLE game_coins (" +
		"game_id varchar(256) REFERENCES games ON DELETE CASCADE, " +
		"color color, " +
		"count int NOT NULL, " +
		"PRIMARY KEY (game_id, color)" +
		")",
	// Which nobles are currently on the table.
	"CREATE TABLE game_nobles (" +
		"game_id varchar(256) REFERENCES games ON DELETE CASCADE, " +
		"index integer, " +
		"noble_id varchar(256) NOT NULL, " +
		"PRIMARY KEY (game_id, index)" +
		")",
	// Which cards are currently on the table.
	"CREATE TABLE game_cards (" +
		"game_id varchar(256) REFERENCES games ON DELETE CASCADE, " +
		"tier integer, " +
		"index integer, " +
		"card_id varchar(256) NOT NULL, " +
		"PRIMARY KEY (game_id, tier, index)" +
		")",
	// Which cards are currently in the deck.
	"CREATE TABLE game_decks (" +
		"game_id varchar(256) REFERENCES games ON DELETE CASCADE, " +
		"tier integer, " +
		"index integer, " +
		"card_id varchar(256) NOT NULL, " +
		"PRIMARY KEY (game_id, tier, index)" +
		")",

	// The players table; one entry for each user for each game they're in.
	"CREATE TABLE players (" +
		"game_id varchar(256) REFERENCES games ON DELETE CASCADE, " +
		"user_id varchar(256) REFERENCES users ON DELETE RESTRICT, " +
		"index integer NOT NULL, " +
		"PRIMARY KEY (game_id, user_id)" +
		")",
	// How many coins the player owns.
	"CREATE TABLE player_coins (" +
		"game_id varchar(256), " +
		"user_id varchar(256), " +
		"color color, " +
		"count integer NOT NULL, " +
		"PRIMARY KEY (game_id, user_id, color), " +
		"FOREIGN KEY (game_id, user_id) REFERENCES players ON DELETE CASCADE" +
		")",
	// Which nobles the player owns.
	"CREATE TABLE player_nobles (" +
		"game_id varchar(256), " +
		"user_id varchar(256), " +
		"noble_id varchar(256), " +
		"PRIMARY KEY (game_id, user_id, noble_id), " +
		"FOREIGN KEY (game_id, user_id) REFERENCES players ON DELETE CASCADE" +
		")",
	// Which cards the player owns (or has reserved)
	"CREATE TABLE player_cards (" +
		"game_id varchar(256), " +
		"user_id varchar(256), " +
		"card_id varchar(256), " +
		"reserved boolean NOT NULL DEFAULT FALSE, " +
		"PRIMARY KEY (game_id, user_id, card_id), " +
		"FOREIGN KEY (game_id, user_id) REFERENCES players ON DELETE CASCADE" +
		")",
}
