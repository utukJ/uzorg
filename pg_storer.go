package main

import "database/sql"

type UzorgPgStorer struct {
	db *sql.DB
}

func (ups *UzorgPgStorer) InsertUserAndDefaultOrg(u *User, o *Org) error {
	// Begin a transaction
	tx, err := ups.db.Begin()
	if err != nil {
		return err
	}

	// Insert user
	_, err = tx.Exec(
		"INSERT INTO users (user_id, first_name, last_name, email, phone, password) VALUES ($1, $2, $3, $4, $5, $6)",
		u.UserID,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		u.Password,
	)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}

	// Insert default org
	_, err = tx.Exec(
		"INSERT INTO orgs (org_id, name, description) VALUES ($1, $2, $3)",
		o.OrgID,
		o.Name,
		o.Description,
	)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}

	// Insert into org_users to link the user with the default org
	_, err = tx.Exec(
		"INSERT INTO org_users (user_id, org_id) VALUES ($1, $2)",
		u.UserID,
		o.OrgID,
	)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	return err
}

// InsertUser inserts a user into the database
func (ups *UzorgPgStorer) InsertUser(u *User) error {
	// Insert user into the database
	_, err := ups.db.Exec(
		"INSERT INTO users (user_id, first_name, last_name, email, phone, password) VALUES ($1, $2, $3, $4, $5, $6)",
		u.UserID,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		u.Password,
	)
	return err
}

// AddUserToOrg adds a user to an organisation
func (ups *UzorgPgStorer) AddUserToOrg(userID, orgID string) error {
	_, err := ups.db.Exec(
		"INSERT INTO org_users (user_id, org_id) VALUES ($1, $2)",
		userID, orgID,
	)
	return err
}

// GetUsersByOrgID retrieves all users belonging to a specific organisation
func (ups *UzorgPgStorer) GetOrgUsers(orgID string) ([]*User, error) {
	rows, err := ups.db.Query(
		"SELECT u.user_id, u.first_name, u.last_name, u.email, u.phone, u.password FROM users u INNER JOIN org_users ou ON u.user_id = ou.user_id WHERE ou.org_id = $1",
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (ups *UzorgPgStorer) GetUserByEmail(email string) (User, error) {
	var user User
	err := ups.db.QueryRow(
		"SELECT user_id, first_name, last_name, email, phone, password FROM users WHERE email = $1",
		email,
	).Scan(&user.UserID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password)
	return user, err
}

func (ups *UzorgPgStorer) GetUserByID(userID string) (User, error) {
	var user User
	err := ups.db.QueryRow(
		"SELECT user_id, first_name, last_name, email, phone, password FROM users WHERE user_id = $1",
		userID,
	).Scan(&user.UserID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.Password)
	return user, err
}

func (ups *UzorgPgStorer) InsertOrg(o *Org) error {
	_, err := ups.db.Exec(
		"INSERT INTO orgs (org_id, name, description) VALUES ($1, $2, $3)",
		o.OrgID,
		o.Name,
		o.Description,
	)
	return err
}

// GetUserOrgs retrieves all organisations that a user belongs to
func (ups *UzorgPgStorer) GetUserOrgs(userID string) ([]*Org, error) {
	rows, err := ups.db.Query(
		"SELECT o.org_id, o.name, o.description FROM orgs o INNER JOIN org_users ou ON o.org_id = ou.org_id WHERE ou.user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*Org
	for rows.Next() {
		var org Org
		if err := rows.Scan(&org.OrgID, &org.Name, &org.Description); err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}

// getorg retrieves an org by ID
func (ups *UzorgPgStorer) GetOrg(orgID string) (Org, error) {
	var org Org
	err := ups.db.QueryRow(
		"SELECT org_id, name, description FROM orgs WHERE org_id = $1",
		orgID,
	).Scan(&org.OrgID, &org.Name, &org.Description)
	return org, err
}

// check if user belongs to an organisation
func (ups *UzorgPgStorer) UserBelongsToOrg(userID, orgID string) (bool, error) {
	var count int
	err := ups.db.QueryRow(
		"SELECT COUNT(*) FROM org_users WHERE user_id = $1 AND org_id = $2",
		userID, orgID,
	).Scan(&count)
	return count > 0, err
}

// InsertOrgAndAddUser inserts an organisation and adds a user to it
func (ups *UzorgPgStorer) InsertOrgAndAddUser(o *Org, userID string) error {
	// Begin a transaction
	tx, err := ups.db.Begin()
	if err != nil {
		return err
	}

	// Insert org
	_, err = tx.Exec(
		"INSERT INTO orgs (org_id, name, description) VALUES ($1, $2, $3)",
		o.OrgID,
		o.Name,
		o.Description,
	)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}

	// Insert into org_users to link the user with the org
	_, err = tx.Exec(
		"INSERT INTO org_users (user_id, org_id) VALUES ($1, $2)",
		userID,
		o.OrgID,
	)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	return err
}
