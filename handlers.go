package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// ReqHandler contains the database connection
type ReqHandler struct {
	uzorgStore UzorgStorer
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user User) (string, error) {
	var jwtKey = []byte(os.Getenv("UZORG_JWT_SECRET"))
	expirationTime := time.Now().Add(24 * time.Hour) // Token is valid for 1 day
	claims := &jwt.StandardClaims{
		Subject:   user.UserID,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	return tokenString, err
}

// registerUser handles user registration
func (h *ReqHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequestResponse(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("Error decoding request: %v", err),
		)
		return
	}

	errs := req.Validate()
	if len(errs) > 0 {
		writeValidationErrorResponse(w, errs)
		return
	}

	userID := uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error hashing password: %v", err))
		return
	}

	user := User{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  string(hashedPassword),
	}

	// check that the user does not already exist by email
	_, err = h.uzorgStore.GetUserByEmail(user.Email)
	if err == nil {
		writeBadRequestResponse(w, http.StatusBadRequest, "User with email already exists")
		return
	}

	org := makeUserDefaultOrg(&user)

	err = h.uzorgStore.InsertUserAndDefaultOrg(&user, &org)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error inserting user into database: %v", err))
		return
	}

	// generate token for user
	token, err := GenerateJWT(user)

	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error while generating jwt: %s", err))
		return
	}

	resp := RegisterUserResponse{
		ResponseStatus: ResponseStatus{
			Status:  "success",
			Message: "Registration successful",
		},
		Data: &UserData{
			Token: token,
			User:  &user,
		},
	}

	// Return the created user as response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login handler
func (h *ReqHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}

	user, err := h.uzorgStore.GetUserByEmail(req.Email)
	if err != nil {
		log.Println("Error getting user by email: ", err)
		writeBadRequestResponse(w, http.StatusUnauthorized, "Authentication failed")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Println("Error comparing password: ", err)
		log.Println("Hashed Password: ", user.Password)
		writeBadRequestResponse(w, http.StatusUnauthorized, "Authentication failed")
		return
	}

	token, err := GenerateJWT(user)
	if err != nil {
		log.Println("Error generating token: ", err)
		writeServerErrorResponse(w, "Error generating token")
		return
	}

	resp := LoginResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "Login successful",
		},
		Data: &UserData{
			Token: token,
			User:  &user,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Implement handler for /api/users/:id
func (h *ReqHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	// retrieve userId from context claim
	userID := r.Context().Value("userId").(string)

	if id != userID {
		log.Printf("Requested user id [%s] does not match token user id [%s]", id, userID)
		writeBadRequestResponse(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	user, err := h.uzorgStore.GetUserByID(id)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error getting user: %v", err))
		return
	}

	response := GetUserResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "User retrieved successfully",
		},
		Data: &user,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Implement handler for /api/organisations that retrieves all the orgs that a logged in user belongs to
func (h *ReqHandler) GetOrgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// retrieve userId from context claim
	userID := r.Context().Value("userId").(string)

	orgs, err := h.uzorgStore.GetUserOrgs(userID)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error getting orgs: %v", err))
		return
	}

	response := GetOrgsResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "Organisations retrieved successfully",
		},
		Data: &Organisations{
			Orgs: orgs,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// implement handler for /api/organisations that creates a new organisation for a user
func (h *ReqHandler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequestResponse(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("Error decoding request: %v", err),
		)
		return
	}

	errs := req.Validate()
	if len(errs) > 0 {
		writeValidationErrorResponse(w, errs)
		return
	}

	// retrieve userId from context claim
	userID := r.Context().Value("userId").(string)

	org := Org{
		OrgID:       uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
	}

	err := h.uzorgStore.InsertOrgAndAddUser(&org, userID)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error inserting org: %v", err))
		return
	}

	response := CreateOrgResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "Organisation created successfully",
		},
		Data: &org,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// write handler for /api/organisations/:id that retrieves an organisation by id. only the user that belongs to the organisation can retrieve it
func (h *ReqHandler) GetOrg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	// retrieve userId from context claim
	userID := r.Context().Value("userId").(string)

	// check if user belongs to org
	belongs, err := h.uzorgStore.UserBelongsToOrg(userID, id)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error checking if user belongs to org: %v", err))
		return
	}

	if !belongs {
		writeBadRequestResponse(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	org, err := h.uzorgStore.GetOrg(id)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error getting org: %v", err))
		return
	}

	response := GetOrgResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "Organisation retrieved successfully",
		},
		Data: &org,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// write handler for /api/organisations/:id/users that retrieves all users in an organisation
func (h *ReqHandler) GetOrgUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	// retrieve userId from context claim
	userID := r.Context().Value("userId").(string)

	// check if user belongs to org
	belongs, err := h.uzorgStore.UserBelongsToOrg(userID, id)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error checking if user belongs to org: %v", err))
		return
	}

	if !belongs {
		writeBadRequestResponse(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	users, err := h.uzorgStore.GetOrgUsers(id)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error getting users: %v", err))
		return
	}

	response := GetOrgUsersResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "Users retrieved successfully",
		},
		Data: users,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// write handler for /api/organisations/:id/users that adds a user to an organisation. only the user that belongs to the organisation can add a user
func (h *ReqHandler) AddUserToOrg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID := vars["id"]

	// retrieve userId from context claim
	userID := r.Context().Value("userId").(string)

	var req AddUserToOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequestResponse(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("Error decoding request: %v", err),
		)
		return
	}

	errs := req.Validate()
	if len(errs) > 0 {
		writeValidationErrorResponse(w, errs)
		return
	}

	// check if user exists by id
	user, err := h.uzorgStore.GetUserByID(req.UserID)
	if err != nil {
		writeBadRequestResponse(w, http.StatusBadRequest, "User does not exist")
		return
	}

	err = h.uzorgStore.AddUserToOrg(user.UserID, orgID)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error adding user to org: %v", err))
		return
	}

	// check if user belongs to org
	belongs, err := h.uzorgStore.UserBelongsToOrg(userID, orgID)
	if err != nil {
		writeServerErrorResponse(w, fmt.Sprintf("Error checking if user belongs to org: %v", err))
		return
	}

	if !belongs {
		writeBadRequestResponse(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	response := AddUserToOrgResponse{
		ResponseStatus: ResponseStatus{
			Status:  SuccessStatus,
			Message: "User added to organisation successfully",
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
