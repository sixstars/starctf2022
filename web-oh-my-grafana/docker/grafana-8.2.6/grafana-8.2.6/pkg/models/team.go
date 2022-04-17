package models

import (
	"errors"
	"time"
)

// Typed errors
var (
	ErrTeamNotFound                         = errors.New("team not found")
	ErrTeamNameTaken                        = errors.New("team name is taken")
	ErrTeamMemberNotFound                   = errors.New("team member not found")
	ErrLastTeamAdmin                        = errors.New("not allowed to remove last admin")
	ErrNotAllowedToUpdateTeam               = errors.New("user not allowed to update team")
	ErrNotAllowedToUpdateTeamInDifferentOrg = errors.New("user not allowed to update team in another org")
)

// Team model
type Team struct {
	Id    int64  `json:"id"`
	OrgId int64  `json:"orgId"`
	Name  string `json:"name"`
	Email string `json:"email"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// ---------------------
// COMMANDS

type CreateTeamCommand struct {
	Name  string `json:"name" binding:"Required"`
	Email string `json:"email"`
	OrgId int64  `json:"-"`

	Result Team `json:"-"`
}

type UpdateTeamCommand struct {
	Id    int64
	Name  string
	Email string
	OrgId int64 `json:"-"`
}

type DeleteTeamCommand struct {
	OrgId int64
	Id    int64
}

type GetTeamByIdQuery struct {
	OrgId        int64
	Id           int64
	SignedInUser *SignedInUser
	HiddenUsers  map[string]struct{}
	Result       *TeamDTO
}

type GetTeamsByUserQuery struct {
	OrgId  int64
	UserId int64      `json:"userId"`
	Result []*TeamDTO `json:"teams"`
}

type SearchTeamsQuery struct {
	Query        string
	Name         string
	Limit        int
	Page         int
	OrgId        int64
	UserIdFilter int64
	SignedInUser *SignedInUser
	HiddenUsers  map[string]struct{}

	Result SearchTeamQueryResult
}

type TeamDTO struct {
	Id          int64          `json:"id"`
	OrgId       int64          `json:"orgId"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	AvatarUrl   string         `json:"avatarUrl"`
	MemberCount int64          `json:"memberCount"`
	Permission  PermissionType `json:"permission"`
}

type SearchTeamQueryResult struct {
	TotalCount int64      `json:"totalCount"`
	Teams      []*TeamDTO `json:"teams"`
	Page       int        `json:"page"`
	PerPage    int        `json:"perPage"`
}

type IsAdminOfTeamsQuery struct {
	SignedInUser *SignedInUser
	Result       bool
}
