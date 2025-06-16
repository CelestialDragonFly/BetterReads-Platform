package data

type User struct {
	ID           *string
	Username     *string
	FirstName    *string
	LastName     *string
	Email        *string
	ProfilePhoto *string
	CreatedAt    *string
}

func (u *User) GetID() string {
	if u == nil || u.ID == nil {
		return ""
	}
	return *u.ID
}

func (u *User) GetUsername() string {
	if u == nil || u.Username == nil {
		return ""
	}
	return *u.Username
}

func (u *User) GetFirstName() string {
	if u == nil || u.FirstName == nil {
		return ""
	}
	return *u.FirstName
}

func (u *User) GetLastName() string {
	if u == nil || u.LastName == nil {
		return ""
	}
	return *u.LastName
}

func (u *User) GetEmail() string {
	if u == nil || u.Email == nil {
		return ""
	}
	return *u.Email
}

func (u *User) GetProfilePhoto() string {
	if u == nil || u.ProfilePhoto == nil {
		return ""
	}
	return *u.ProfilePhoto
}

func (u *User) GetCreatedAt() string {
	if u == nil || u.CreatedAt == nil {
		return ""
	}
	return *u.CreatedAt
}
