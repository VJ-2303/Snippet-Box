package models

import "errors"

var ErrNoRecord = errors.New("models: no matching record found")
var ErrInvalidCredentials = errors.New("models: invalid creddentials")
var ErrDuplicateEmail = errors.New("models: duplicate email")
