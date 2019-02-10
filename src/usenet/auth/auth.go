package auth

type Authenticator interface {
	CheckLogin(user, password string) (bool, error)
}
