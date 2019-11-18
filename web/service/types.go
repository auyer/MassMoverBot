package service

// SimpleGuild struct stores a simpler version a a Discordgo.Guild object
// This is used to send guild information over the internet
type SimpleGuild struct {
	// The ID of the guild.
	ID string `json:"id"`

	// The name of the guild. (2–100 characters)
	Name string `json:"name"`

	// The guild's icon.
	Icon string `json:"icon"`

	// The number of members in the guild.
	// This field is only present in GUILD_CREATE events and websocket
	// update events, and thus is only present in state-cached guilds.
	MemberCount int `json:"member_count"`
}

// A User stores all data for an individual Discord user.
type User struct {
	// The ID of the user.
	ID string `json:"id"`

	// The user's username.
	Username string `json:"username"`

	// The hash of the user's avatar. Use Session.UserAvatar
	// to retrieve the avatar itself.
	// Avatar string `json:"avatar"`
}

// A Channel holds all data related to an individual Discord channel.
type Channel struct {
	// The ID of the channel.
	ID string `json:"id"`

	// The ID of the guild to which the channel belongs, if it is in a guild.
	// Else, this ID is empty (e.g. DM channels).
	GuildID string `json:"guild_id"`

	// The name of the channel.
	Name string `json:"name"`

	// The position of the channel, used for sorting in client.
	Position int `json:"position"`

	Users []*User
}
