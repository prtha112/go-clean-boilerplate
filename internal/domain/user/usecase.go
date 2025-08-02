package user

type UseCase interface {
	Login(user *User) error
}
