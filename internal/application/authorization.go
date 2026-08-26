package application

func Allowed(actor, owner string) bool { return actor != "" && owner != "" }
