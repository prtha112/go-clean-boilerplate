package user

type Repository interface {
	GetByUsernameAndPassword(username, password string) (*User, error)
}
