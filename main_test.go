package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// type UzorgTestStorer struct {
// 	users []User
// 	orgs  []Org
// }

// func (uts *UzorgTestStorer) GetOrgUsers(orgID string) ([]*User, error) {
// 	var orgUsers []*User
// 	for _, user := range uts.users {
// 		if user.OrgID == orgID {
// 			orgUsers = append(orgUsers, &user)
// 		}
// 	}
// 	return orgUsers, nil
// }

// func (uts *UzorgTestStorer) AddUserToOrg(userID string, orgID string) error {
// 	// Implement the logic to add a user to an organisation
// 	return nil
// }

// func (uts *UzorgTestStorer) GetOrg(orgID string) (Org, error) {
// 	for _, org := range uts.orgs {
// 		if org.OrgID == orgID {
// 			return org, nil
// 		}
// 	}
// 	return Org{}, errors.New("Org not found")
// }

// // GetUserOrgs retrieves the organisations associated with a user from the test store
// func (uts *UzorgTestStorer) GetUserOrgs(userID string) ([]*Org, error) {
// 	var userOrgs []*Org
// 	for _, org := range uts.orgs {
// 		if org.UserID == userID {
// 			userOrgs = append(userOrgs, &org)
// 		}
// 	}
// 	return userOrgs, nil
// }

// // GetUserByID retrieves a user by ID from the test store
// func (uts *UzorgTestStorer) GetUserByID(userID string) (User, error) {
// 	for _, user := range uts.users {
// 		if user.UserID == userID {
// 			return user, nil
// 		}
// 	}
// 	return User{}, errors.New("User not found")
// }

// // InsertUser inserts a user into the test store
// func (uts *UzorgTestStorer) InsertUser(u *User) error {
// 	uts.users = append(uts.users, *u)
// 	return nil
// }

// // GetUserByEmail retrieves a user by email from the test store
// func (uts *UzorgTestStorer) GetUserByEmail(email string) (User, error) {
// 	for _, user := range uts.users {
// 		if user.Email == email {
// 			return user, nil
// 		}
// 	}
// 	return User{}, errors.New("User not found")
// }

// // InsertOrg inserts an org into the test store
// func (uts *UzorgTestStorer) InsertOrg(o *Org) error {
// 	uts.orgs = append(uts.orgs, *o)
// 	return nil
// }

// // InsertUserAndDefaultOrg inserts a user and a default org into the test store
// func (uts *UzorgTestStorer) InsertUserAndDefaultOrg(u *User, o *Org) error {
// 	uts.users = append(uts.users, *u)
// 	uts.orgs = append(uts.orgs, *o)
// 	return nil
// }

// // test create user handler
// func TestCreateUserHandler(t *testing.T) {
// 	// GIVEN
// 	store := &UzorgTestStorer{}

// 	handler := &ReqHandler{uzorgStore: store}

// 	reqBody := RegisterUserRequest{
// 		FirstName: "John",
// 		LastName:  "Doe",
// 		Email:     "eml@eml.eml",
// 		Password:  "password",
// 		Phone:     "+1234567890",
// 	}

// 	var buf bytes.Buffer
// 	json.NewEncoder(&buf).Encode(reqBody)

// 	req := httptest.NewRequest("POST", "/register", &buf)
// 	rec := httptest.NewRecorder()

// 	// WHEN
// 	handler.registerUser(rec, req)

// 	// THEN
// 	if rec.Code != http.StatusCreated {
// 		t.Fatalf(
// 			"Expected status code %d, got %d with message %s",
// 			http.StatusCreated,
// 			rec.Code,
// 			rec.Result().Status,
// 		)
// 	}

// 	var user User
// 	err := json.NewDecoder(rec.Body).Decode(&user)
// 	if err != nil {
// 		t.Fatalf("Error decoding response body: %v", err)
// 	}

// 	if len(store.orgs) != 1 {
// 		t.Fatalf("Expected 1 org in store, got %d", len(store.orgs))
// 	}

// 	// check that the org was created with the user
// 	if store.orgs[0].UserID != user.UserID {
// 		t.Fatalf(
// 			"Expected org to be created with userID %s, got org with userID %s",
// 			user.UserID,
// 			store.orgs[0].UserID,
// 		)
// 	}

// 	// check org name
// 	if store.orgs[0].Name != user.FirstName+"'s Organisation" {
// 		t.Errorf(
// 			"Expected org name to be %s's Organisation, got %s",
// 			user.FirstName,
// 			store.orgs[0].Name,
// 		)
// 	}
// }

// // test create duplicate email fails
// func TestCreateUserDuplicateEmail(t *testing.T) {
// 	// initialize the test store
// 	store := &UzorgTestStorer{}

// 	// insert a user into the store
// 	store.users = append(store.users, User{
// 		UserID:    "1",
// 		FirstName: "John",
// 		LastName:  "Doe",
// 		Email:     "same@email.com",
// 		Phone:     "+1234567890",
// 	})

// 	// create a request handler
// 	handler := &ReqHandler{uzorgStore: store}

// 	// create a request body
// 	reqBody := RegisterUserRequest{
// 		FirstName: "Jane",
// 		LastName:  "Doe",
// 		Email:     "same@email.com",
// 		Password:  "password",
// 		Phone:     "+1234567890",
// 	}

// 	// create a request
// 	req := httptest.NewRequest("POST", "/register", nil)

// 	// create a request body
// 	var buf bytes.Buffer
// 	json.NewEncoder(&buf).Encode(reqBody)
// 	req.Body = io.NopCloser(&buf)

// 	// create a response recorder
// 	rec := httptest.NewRecorder()

// 	// call the handler
// 	handler.registerUser(rec, req)

// 	if rec.Code != http.StatusConflict {
// 		t.Errorf("Expected status code %d, got %d", http.StatusConflict, rec.Code)
// 	}

// 	if rec.Body.String() != "User with email already exists\n" {
// 		t.Errorf(
// 			"Expected response body 'User with email already exists', got %s",
// 			rec.Body.String(),
// 		)
// 	}

// 	if len(store.users) != 1 {
// 		t.Errorf("Expected 1 user in store, got %d", len(store.users))
// 	}
// }
