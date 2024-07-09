package main

type UzorgStorer interface {
	InsertUserAndDefaultOrg(u *User, o *Org) error
	InsertOrgAndAddUser(o *Org, userID string) error
	InsertUser(u *User) error
	AddUserToOrg(userID, orgID string) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(userID string) (User, error)
	InsertOrg(o *Org) error
	GetOrg(orgID string) (Org, error)
	GetUserOrgs(userID string) ([]*Org, error)
	GetOrgUsers(orgID string) ([]*User, error)
	UserBelongsToOrg(userID, orgID string) (bool, error)
}
