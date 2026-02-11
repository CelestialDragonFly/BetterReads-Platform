package data

import "time"

type User struct {
	ID              string
	Username        string
	FirstName       string
	LastName        string
	Email           string
	ProfilePhotoURL string
	CreatedAt       time.Time
}

func (u *User) GetID() string {
	if u == nil {
		return ""
	}
	return u.ID
}

func (u *User) GetUsername() string {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *User) GetFirstName() string {
	if u == nil {
		return ""
	}
	return u.FirstName
}

func (u *User) GetLastName() string {
	if u == nil {
		return ""
	}
	return u.LastName
}

func (u *User) GetEmail() string {
	if u == nil {
		return ""
	}
	return u.Email
}

func (u *User) GetProfilePhotoURL() string {
	if u == nil {
		return ""
	}
	return u.ProfilePhotoURL
}

func (u *User) GetCreatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.CreatedAt
}
