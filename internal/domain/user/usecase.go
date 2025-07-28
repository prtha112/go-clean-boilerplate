package user

type UseCase interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}