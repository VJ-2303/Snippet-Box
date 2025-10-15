package main

type contextKey string

const AuthenticatedUserContextkey = contextKey("User")

type AuthenticatedUserContext struct {
	isAuthenticated bool
	userID          int
}
