package role

type Role int

const (
	User      Role = iota // 0
	Moderator             // 1
	Admin                 // 2
)
