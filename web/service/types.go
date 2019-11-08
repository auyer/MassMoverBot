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
