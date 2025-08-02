package user

type Repository interface {
	GetByID(id int) (*User, error)
	Create(user *User) error
	GetByUsernameAndPassword(username, password string) (*User, error)
}
