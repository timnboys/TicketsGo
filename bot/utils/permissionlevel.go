package utils

type PermissionLevel int

const(
	Everyone PermissionLevel = iota
	Support
	Admin
)
