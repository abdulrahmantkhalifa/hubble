package auth

type declineAuth struct{}

func NewDeclineAllModule() AuthorizationModule {
	return declineAuth{}
}

func (auth declineAuth) Connect(r ConnectionRequest) (bool, error) {
	return false, nil
}

type acceptAuth struct{}

func NewAcceptAllModule() AuthorizationModule {
	return acceptAuth{}
}

func (auth acceptAuth) Connect(r ConnectionRequest) (bool, error) {
	return true, nil
}
