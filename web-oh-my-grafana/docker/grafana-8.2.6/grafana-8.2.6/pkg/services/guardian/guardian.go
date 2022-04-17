package guardian

import (
	"errors"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

var (
	ErrGuardianPermissionExists = errors.New("permission already exists")
	ErrGuardianOverride         = errors.New("you can only override a permission to be higher")
)

// DashboardGuardian to be used for guard against operations without access on dashboard and acl
type DashboardGuardian interface {
	CanSave() (bool, error)
	CanEdit() (bool, error)
	CanView() (bool, error)
	CanAdmin() (bool, error)
	HasPermission(permission models.PermissionType) (bool, error)
	CheckPermissionBeforeUpdate(permission models.PermissionType, updatePermissions []*models.DashboardAcl) (bool, error)

	// GetAcl returns ACL.
	GetAcl() ([]*models.DashboardAclInfoDTO, error)

	// GetACLWithoutDuplicates returns ACL and strips any permission
	// that already has an inherited permission with higher or equal
	// permission.
	GetACLWithoutDuplicates() ([]*models.DashboardAclInfoDTO, error)
	GetHiddenACL(*setting.Cfg) ([]*models.DashboardAcl, error)
}

type dashboardGuardianImpl struct {
	user   *models.SignedInUser
	dashId int64
	orgId  int64
	acl    []*models.DashboardAclInfoDTO
	teams  []*models.TeamDTO
	log    log.Logger
}

// New factory for creating a new dashboard guardian instance
var New = func(dashId int64, orgId int64, user *models.SignedInUser) DashboardGuardian {
	return &dashboardGuardianImpl{
		user:   user,
		dashId: dashId,
		orgId:  orgId,
		log:    log.New("dashboard.permissions"),
	}
}

func (g *dashboardGuardianImpl) CanSave() (bool, error) {
	return g.HasPermission(models.PERMISSION_EDIT)
}

func (g *dashboardGuardianImpl) CanEdit() (bool, error) {
	if setting.ViewersCanEdit {
		return g.HasPermission(models.PERMISSION_VIEW)
	}

	return g.HasPermission(models.PERMISSION_EDIT)
}

func (g *dashboardGuardianImpl) CanView() (bool, error) {
	return g.HasPermission(models.PERMISSION_VIEW)
}

func (g *dashboardGuardianImpl) CanAdmin() (bool, error) {
	return g.HasPermission(models.PERMISSION_ADMIN)
}

func (g *dashboardGuardianImpl) HasPermission(permission models.PermissionType) (bool, error) {
	if g.user.OrgRole == models.ROLE_ADMIN {
		return g.logHasPermissionResult(permission, true, nil)
	}

	acl, err := g.GetAcl()
	if err != nil {
		return g.logHasPermissionResult(permission, false, err)
	}

	result, err := g.checkAcl(permission, acl)
	return g.logHasPermissionResult(permission, result, err)
}

func (g *dashboardGuardianImpl) logHasPermissionResult(permission models.PermissionType, hasPermission bool, err error) (bool, error) {
	if err != nil {
		return hasPermission, err
	}

	if hasPermission {
		g.log.Debug("User granted access to execute action", "userId", g.user.UserId, "orgId", g.orgId, "uname", g.user.Login, "dashId", g.dashId, "action", permission)
	} else {
		g.log.Debug("User denied access to execute action", "userId", g.user.UserId, "orgId", g.orgId, "uname", g.user.Login, "dashId", g.dashId, "action", permission)
	}

	return hasPermission, err
}

func (g *dashboardGuardianImpl) checkAcl(permission models.PermissionType, acl []*models.DashboardAclInfoDTO) (bool, error) {
	orgRole := g.user.OrgRole
	teamAclItems := []*models.DashboardAclInfoDTO{}

	for _, p := range acl {
		// user match
		if !g.user.IsAnonymous && p.UserId > 0 {
			if p.UserId == g.user.UserId && p.Permission >= permission {
				return true, nil
			}
		}

		// role match
		if p.Role != nil {
			if *p.Role == orgRole && p.Permission >= permission {
				return true, nil
			}
		}

		// remember this rule for later
		if p.TeamId > 0 {
			teamAclItems = append(teamAclItems, p)
		}
	}

	// do we have team rules?
	if len(teamAclItems) == 0 {
		return false, nil
	}

	// load teams
	teams, err := g.getTeams()
	if err != nil {
		return false, err
	}

	// evaluate team rules
	for _, p := range acl {
		for _, ug := range teams {
			if ug.Id == p.TeamId && p.Permission >= permission {
				return true, nil
			}
		}
	}

	return false, nil
}

func (g *dashboardGuardianImpl) CheckPermissionBeforeUpdate(permission models.PermissionType, updatePermissions []*models.DashboardAcl) (bool, error) {
	acl := []*models.DashboardAclInfoDTO{}
	adminRole := models.ROLE_ADMIN
	everyoneWithAdminRole := &models.DashboardAclInfoDTO{DashboardId: g.dashId, UserId: 0, TeamId: 0, Role: &adminRole, Permission: models.PERMISSION_ADMIN}

	// validate that duplicate permissions don't exists
	for _, p := range updatePermissions {
		aclItem := &models.DashboardAclInfoDTO{DashboardId: p.DashboardID, UserId: p.UserID, TeamId: p.TeamID, Role: p.Role, Permission: p.Permission}
		if aclItem.IsDuplicateOf(everyoneWithAdminRole) {
			return false, ErrGuardianPermissionExists
		}

		for _, a := range acl {
			if a.IsDuplicateOf(aclItem) {
				return false, ErrGuardianPermissionExists
			}
		}

		acl = append(acl, aclItem)
	}

	existingPermissions, err := g.GetAcl()
	if err != nil {
		return false, err
	}

	// validate overridden permissions to be higher
	for _, a := range acl {
		for _, existingPerm := range existingPermissions {
			if !existingPerm.Inherited {
				continue
			}

			if a.IsDuplicateOf(existingPerm) && a.Permission <= existingPerm.Permission {
				return false, ErrGuardianOverride
			}
		}
	}

	if g.user.OrgRole == models.ROLE_ADMIN {
		return true, nil
	}

	return g.checkAcl(permission, existingPermissions)
}

// GetAcl returns dashboard acl
func (g *dashboardGuardianImpl) GetAcl() ([]*models.DashboardAclInfoDTO, error) {
	if g.acl != nil {
		return g.acl, nil
	}

	query := models.GetDashboardAclInfoListQuery{DashboardID: g.dashId, OrgID: g.orgId}
	if err := bus.Dispatch(&query); err != nil {
		return nil, err
	}

	g.acl = query.Result
	return g.acl, nil
}

func (g *dashboardGuardianImpl) GetACLWithoutDuplicates() ([]*models.DashboardAclInfoDTO, error) {
	acl, err := g.GetAcl()
	if err != nil {
		return nil, err
	}

	nonInherited := []*models.DashboardAclInfoDTO{}
	inherited := []*models.DashboardAclInfoDTO{}
	for _, aclItem := range acl {
		if aclItem.Inherited {
			inherited = append(inherited, aclItem)
		} else {
			nonInherited = append(nonInherited, aclItem)
		}
	}

	result := []*models.DashboardAclInfoDTO{}
	for _, nonInheritedAclItem := range nonInherited {
		duplicate := false
		for _, inheritedAclItem := range inherited {
			if nonInheritedAclItem.IsDuplicateOf(inheritedAclItem) && nonInheritedAclItem.Permission <= inheritedAclItem.Permission {
				duplicate = true
				break
			}
		}

		if !duplicate {
			result = append(result, nonInheritedAclItem)
		}
	}

	result = append(inherited, result...)

	return result, nil
}

func (g *dashboardGuardianImpl) getTeams() ([]*models.TeamDTO, error) {
	if g.teams != nil {
		return g.teams, nil
	}

	query := models.GetTeamsByUserQuery{OrgId: g.orgId, UserId: g.user.UserId}
	err := bus.Dispatch(&query)

	g.teams = query.Result
	return query.Result, err
}

func (g *dashboardGuardianImpl) GetHiddenACL(cfg *setting.Cfg) ([]*models.DashboardAcl, error) {
	hiddenACL := make([]*models.DashboardAcl, 0)
	if g.user.IsGrafanaAdmin {
		return hiddenACL, nil
	}

	existingPermissions, err := g.GetAcl()
	if err != nil {
		return hiddenACL, err
	}

	for _, item := range existingPermissions {
		if item.Inherited || item.UserLogin == g.user.Login {
			continue
		}

		if _, hidden := cfg.HiddenUsers[item.UserLogin]; hidden {
			hiddenACL = append(hiddenACL, &models.DashboardAcl{
				OrgID:       item.OrgId,
				DashboardID: item.DashboardId,
				UserID:      item.UserId,
				TeamID:      item.TeamId,
				Role:        item.Role,
				Permission:  item.Permission,
				Created:     item.Created,
				Updated:     item.Updated,
			})
		}
	}
	return hiddenACL, nil
}

// nolint:unused
type FakeDashboardGuardian struct {
	DashId                           int64
	OrgId                            int64
	User                             *models.SignedInUser
	CanSaveValue                     bool
	CanEditValue                     bool
	CanViewValue                     bool
	CanAdminValue                    bool
	HasPermissionValue               bool
	CheckPermissionBeforeUpdateValue bool
	CheckPermissionBeforeUpdateError error
	GetAclValue                      []*models.DashboardAclInfoDTO
	GetHiddenAclValue                []*models.DashboardAcl
}

func (g *FakeDashboardGuardian) CanSave() (bool, error) {
	return g.CanSaveValue, nil
}

func (g *FakeDashboardGuardian) CanEdit() (bool, error) {
	return g.CanEditValue, nil
}

func (g *FakeDashboardGuardian) CanView() (bool, error) {
	return g.CanViewValue, nil
}

func (g *FakeDashboardGuardian) CanAdmin() (bool, error) {
	return g.CanAdminValue, nil
}

func (g *FakeDashboardGuardian) HasPermission(permission models.PermissionType) (bool, error) {
	return g.HasPermissionValue, nil
}

func (g *FakeDashboardGuardian) CheckPermissionBeforeUpdate(permission models.PermissionType, updatePermissions []*models.DashboardAcl) (bool, error) {
	return g.CheckPermissionBeforeUpdateValue, g.CheckPermissionBeforeUpdateError
}

func (g *FakeDashboardGuardian) GetAcl() ([]*models.DashboardAclInfoDTO, error) {
	return g.GetAclValue, nil
}

func (g *FakeDashboardGuardian) GetACLWithoutDuplicates() ([]*models.DashboardAclInfoDTO, error) {
	return g.GetAcl()
}

func (g *FakeDashboardGuardian) GetHiddenACL(cfg *setting.Cfg) ([]*models.DashboardAcl, error) {
	return g.GetHiddenAclValue, nil
}

// nolint:unused
func MockDashboardGuardian(mock *FakeDashboardGuardian) {
	New = func(dashId int64, orgId int64, user *models.SignedInUser) DashboardGuardian {
		mock.OrgId = orgId
		mock.DashId = dashId
		mock.User = user
		return mock
	}
}
