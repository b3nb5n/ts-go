package main

import "github.com/google/uuid"

type ComplexType map[string]map[uint16]*uint32

type UserRole = string
const (
	UserRoleDefault UserRole = "viewer"
	UserRoleEditor  UserRole = "editor"
)

type UserEntry struct {
	// Instead of specifying `tstype` we could also declare the typing
	// for uuid.NullUUID in the config file.
	ID uuid.NullUUID `json:"id" tstype:"string | null"`

	Preferences map[string]struct {
		Foo uint32 `json:"foo"`
		// An unknown type without a `tstype` tag or mapping in the config file
		// becomes `any`
		Bar uuid.UUID `json:"bar"`
	} `json:"prefs"`

	MaybeFieldWithStar *string  `json:"address"`
	Nickname           string   `json:"nickname,omitempty"`
	Role               UserRole `json:"role"`

	Complex    ComplexType `json:"complex"`
	unexported bool        // Unexported fields are omitted
	Ignored    bool        `tstype:"-"` // Fields with - are omitted too
}

type ListUsersResponse struct {
	Users []UserEntry `json:"users"`
	X, Y *int
}

type MyIotaType int

const (
	Zero MyIotaType = iota
	One
	Two
	_
	Four
	FourString string = "four"
	_
	AlsoFourString
	Five = 5
	FiveAgain

	Sixteen = iota + 6
	Seventeen
)

const (
	_One = "one"
	_Two = 2
	Three int = 3
)

func Hello(name, lastname string, age *int, other ...string) {
	// return "", nil
}