package sqlstore

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
)

func init() {
	bus.AddHandler("sql", UpdateTeam)
	bus.AddHandler("sql", DeleteTeam)
	bus.AddHandler("sql", SearchTeams)
	bus.AddHandler("sql", GetTeamById)
	bus.AddHandler("sql", GetTeamsByUser)

	bus.AddHandler("sql", UpdateTeamMember)
	bus.AddHandler("sql", RemoveTeamMember)
	bus.AddHandler("sql", GetTeamMembers)
	bus.AddHandler("sql", IsAdminOfTeams)
}

func getFilteredUsers(signedInUser *models.SignedInUser, hiddenUsers map[string]struct{}) []string {
	filteredUsers := make([]string, 0, len(hiddenUsers))
	if signedInUser == nil || signedInUser.IsGrafanaAdmin {
		return filteredUsers
	}

	for u := range hiddenUsers {
		if u == signedInUser.Login {
			continue
		}
		filteredUsers = append(filteredUsers, u)
	}

	return filteredUsers
}

func getTeamMemberCount(filteredUsers []string) string {
	if len(filteredUsers) > 0 {
		return `(SELECT COUNT(*) FROM team_member
			INNER JOIN ` + dialect.Quote("user") + ` ON team_member.user_id = ` + dialect.Quote("user") + `.id
			WHERE team_member.team_id = team.id AND ` + dialect.Quote("user") + `.login NOT IN (?` +
			strings.Repeat(",?", len(filteredUsers)-1) + ")" +
			`) AS member_count `
	}

	return "(SELECT COUNT(*) FROM team_member WHERE team_member.team_id = team.id) AS member_count "
}

func getTeamSearchSQLBase(filteredUsers []string) string {
	return `SELECT
		team.id AS id,
		team.org_id,
		team.name AS name,
		team.email AS email,
		team_member.permission, ` +
		getTeamMemberCount(filteredUsers) +
		` FROM team AS team
		INNER JOIN team_member ON team.id = team_member.team_id AND team_member.user_id = ? `
}

func getTeamSelectSQLBase(filteredUsers []string) string {
	return `SELECT
		team.id as id,
		team.org_id,
		team.name as name,
		team.email as email, ` +
		getTeamMemberCount(filteredUsers) +
		` FROM team as team `
}

func (ss *SQLStore) CreateTeam(name, email string, orgID int64) (models.Team, error) {
	team := models.Team{
		Name:    name,
		Email:   email,
		OrgId:   orgID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	err := ss.WithTransactionalDbSession(context.Background(), func(sess *DBSession) error {
		if isNameTaken, err := isTeamNameTaken(orgID, name, 0, sess); err != nil {
			return err
		} else if isNameTaken {
			return models.ErrTeamNameTaken
		}

		_, err := sess.Insert(&team)
		return err
	})
	return team, err
}

func UpdateTeam(cmd *models.UpdateTeamCommand) error {
	return inTransaction(func(sess *DBSession) error {
		if isNameTaken, err := isTeamNameTaken(cmd.OrgId, cmd.Name, cmd.Id, sess); err != nil {
			return err
		} else if isNameTaken {
			return models.ErrTeamNameTaken
		}

		team := models.Team{
			Name:    cmd.Name,
			Email:   cmd.Email,
			Updated: time.Now(),
		}

		sess.MustCols("email")

		affectedRows, err := sess.ID(cmd.Id).Update(&team)

		if err != nil {
			return err
		}

		if affectedRows == 0 {
			return models.ErrTeamNotFound
		}

		return nil
	})
}

// DeleteTeam will delete a team, its member and any permissions connected to the team
func DeleteTeam(cmd *models.DeleteTeamCommand) error {
	return inTransaction(func(sess *DBSession) error {
		if _, err := teamExists(cmd.OrgId, cmd.Id, sess); err != nil {
			return err
		}

		deletes := []string{
			"DELETE FROM team_member WHERE org_id=? and team_id = ?",
			"DELETE FROM team WHERE org_id=? and id = ?",
			"DELETE FROM dashboard_acl WHERE org_id=? and team_id = ?",
		}

		for _, sql := range deletes {
			_, err := sess.Exec(sql, cmd.OrgId, cmd.Id)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func teamExists(orgID int64, teamID int64, sess *DBSession) (bool, error) {
	if res, err := sess.Query("SELECT 1 from team WHERE org_id=? and id=?", orgID, teamID); err != nil {
		return false, err
	} else if len(res) != 1 {
		return false, models.ErrTeamNotFound
	}

	return true, nil
}

func isTeamNameTaken(orgId int64, name string, existingId int64, sess *DBSession) (bool, error) {
	var team models.Team
	exists, err := sess.Where("org_id=? and name=?", orgId, name).Get(&team)
	if err != nil {
		return false, nil
	}

	if exists && existingId != team.Id {
		return true, nil
	}

	return false, nil
}

func SearchTeams(query *models.SearchTeamsQuery) error {
	query.Result = models.SearchTeamQueryResult{
		Teams: make([]*models.TeamDTO, 0),
	}
	queryWithWildcards := "%" + query.Query + "%"

	var sql bytes.Buffer
	params := make([]interface{}, 0)

	filteredUsers := getFilteredUsers(query.SignedInUser, query.HiddenUsers)
	if query.UserIdFilter > 0 {
		sql.WriteString(getTeamSearchSQLBase(filteredUsers))
		for _, user := range filteredUsers {
			params = append(params, user)
		}
		params = append(params, query.UserIdFilter)
	} else {
		sql.WriteString(getTeamSelectSQLBase(filteredUsers))
		for _, user := range filteredUsers {
			params = append(params, user)
		}
	}

	sql.WriteString(` WHERE team.org_id = ?`)
	params = append(params, query.OrgId)

	if query.Query != "" {
		sql.WriteString(` and team.name ` + dialect.LikeStr() + ` ?`)
		params = append(params, queryWithWildcards)
	}

	if query.Name != "" {
		sql.WriteString(` and team.name = ?`)
		params = append(params, query.Name)
	}

	sql.WriteString(` order by team.name asc`)

	if query.Limit != 0 {
		offset := query.Limit * (query.Page - 1)
		sql.WriteString(dialect.LimitOffset(int64(query.Limit), int64(offset)))
	}

	if err := x.SQL(sql.String(), params...).Find(&query.Result.Teams); err != nil {
		return err
	}

	team := models.Team{}
	countSess := x.Table("team")
	if query.Query != "" {
		countSess.Where(`name `+dialect.LikeStr()+` ?`, queryWithWildcards)
	}

	if query.Name != "" {
		countSess.Where("name=?", query.Name)
	}

	count, err := countSess.Count(&team)
	query.Result.TotalCount = count

	return err
}

func GetTeamById(query *models.GetTeamByIdQuery) error {
	var sql bytes.Buffer
	params := make([]interface{}, 0)

	filteredUsers := getFilteredUsers(query.SignedInUser, query.HiddenUsers)
	sql.WriteString(getTeamSelectSQLBase(filteredUsers))
	for _, user := range filteredUsers {
		params = append(params, user)
	}

	sql.WriteString(` WHERE team.org_id = ? and team.id = ?`)
	params = append(params, query.OrgId, query.Id)

	var team models.TeamDTO
	exists, err := x.SQL(sql.String(), params...).Get(&team)

	if err != nil {
		return err
	}

	if !exists {
		return models.ErrTeamNotFound
	}

	query.Result = &team
	return nil
}

// GetTeamsByUser is used by the Guardian when checking a users' permissions
func GetTeamsByUser(query *models.GetTeamsByUserQuery) error {
	query.Result = make([]*models.TeamDTO, 0)

	var sql bytes.Buffer

	sql.WriteString(getTeamSelectSQLBase([]string{}))
	sql.WriteString(` INNER JOIN team_member on team.id = team_member.team_id`)
	sql.WriteString(` WHERE team.org_id = ? and team_member.user_id = ?`)

	err := x.SQL(sql.String(), query.OrgId, query.UserId).Find(&query.Result)
	return err
}

// AddTeamMember adds a user to a team
func (ss *SQLStore) AddTeamMember(userID, orgID, teamID int64, isExternal bool, permission models.PermissionType) error {
	return ss.WithTransactionalDbSession(context.Background(), func(sess *DBSession) error {
		if res, err := sess.Query("SELECT 1 from team_member WHERE org_id=? and team_id=? and user_id=?",
			orgID, teamID, userID); err != nil {
			return err
		} else if len(res) == 1 {
			return models.ErrTeamMemberAlreadyAdded
		}

		if _, err := teamExists(orgID, teamID, sess); err != nil {
			return err
		}

		entity := models.TeamMember{
			OrgId:      orgID,
			TeamId:     teamID,
			UserId:     userID,
			External:   isExternal,
			Created:    time.Now(),
			Updated:    time.Now(),
			Permission: permission,
		}

		_, err := sess.Insert(&entity)
		return err
	})
}

func getTeamMember(sess *DBSession, orgId int64, teamId int64, userId int64) (models.TeamMember, error) {
	rawSQL := `SELECT * FROM team_member WHERE org_id=? and team_id=? and user_id=?`
	var member models.TeamMember
	exists, err := sess.SQL(rawSQL, orgId, teamId, userId).Get(&member)

	if err != nil {
		return member, err
	}
	if !exists {
		return member, models.ErrTeamMemberNotFound
	}

	return member, nil
}

// UpdateTeamMember updates a team member
func UpdateTeamMember(cmd *models.UpdateTeamMemberCommand) error {
	return inTransaction(func(sess *DBSession) error {
		member, err := getTeamMember(sess, cmd.OrgId, cmd.TeamId, cmd.UserId)
		if err != nil {
			return err
		}

		if cmd.ProtectLastAdmin {
			_, err := isLastAdmin(sess, cmd.OrgId, cmd.TeamId, cmd.UserId)
			if err != nil {
				return err
			}
		}

		if cmd.Permission != models.PERMISSION_ADMIN { // make sure we don't get invalid permission levels in store
			cmd.Permission = 0
		}

		member.Permission = cmd.Permission
		_, err = sess.Cols("permission").Where("org_id=? and team_id=? and user_id=?", cmd.OrgId, cmd.TeamId, cmd.UserId).Update(member)

		return err
	})
}

// RemoveTeamMember removes a member from a team
func RemoveTeamMember(cmd *models.RemoveTeamMemberCommand) error {
	return inTransaction(func(sess *DBSession) error {
		if _, err := teamExists(cmd.OrgId, cmd.TeamId, sess); err != nil {
			return err
		}

		if cmd.ProtectLastAdmin {
			_, err := isLastAdmin(sess, cmd.OrgId, cmd.TeamId, cmd.UserId)
			if err != nil {
				return err
			}
		}

		var rawSQL = "DELETE FROM team_member WHERE org_id=? and team_id=? and user_id=?"
		res, err := sess.Exec(rawSQL, cmd.OrgId, cmd.TeamId, cmd.UserId)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if rows == 0 {
			return models.ErrTeamMemberNotFound
		}

		return err
	})
}

func isLastAdmin(sess *DBSession, orgId int64, teamId int64, userId int64) (bool, error) {
	rawSQL := "SELECT user_id FROM team_member WHERE org_id=? and team_id=? and permission=?"
	userIds := []*int64{}
	err := sess.SQL(rawSQL, orgId, teamId, models.PERMISSION_ADMIN).Find(&userIds)
	if err != nil {
		return false, err
	}

	isAdmin := false
	for _, adminId := range userIds {
		if userId == *adminId {
			isAdmin = true
			break
		}
	}

	if isAdmin && len(userIds) == 1 {
		return true, models.ErrLastTeamAdmin
	}

	return false, err
}

// GetTeamMembers return a list of members for the specified team
func GetTeamMembers(query *models.GetTeamMembersQuery) error {
	query.Result = make([]*models.TeamMemberDTO, 0)
	sess := x.Table("team_member")
	sess.Join("INNER", x.Dialect().Quote("user"), fmt.Sprintf("team_member.user_id=%s.id", x.Dialect().Quote("user")))

	// Join with only most recent auth module
	authJoinCondition := `(
		SELECT id from user_auth
			WHERE user_auth.user_id = team_member.user_id
			ORDER BY user_auth.created DESC `
	authJoinCondition = "user_auth.id=" + authJoinCondition + dialect.Limit(1) + ")"
	sess.Join("LEFT", "user_auth", authJoinCondition)

	if query.OrgId != 0 {
		sess.Where("team_member.org_id=?", query.OrgId)
	}
	if query.TeamId != 0 {
		sess.Where("team_member.team_id=?", query.TeamId)
	}
	if query.UserId != 0 {
		sess.Where("team_member.user_id=?", query.UserId)
	}
	if query.External {
		sess.Where("team_member.external=?", dialect.BooleanStr(true))
	}
	sess.Cols(
		"team_member.org_id",
		"team_member.team_id",
		"team_member.user_id",
		"user.email",
		"user.name",
		"user.login",
		"team_member.external",
		"team_member.permission",
		"user_auth.auth_module",
	)
	sess.Asc("user.login", "user.email")

	err := sess.Find(&query.Result)
	return err
}

func IsAdminOfTeams(query *models.IsAdminOfTeamsQuery) error {
	builder := &SQLBuilder{}
	builder.Write("SELECT COUNT(team.id) AS count FROM team INNER JOIN team_member ON team_member.team_id = team.id WHERE team.org_id = ? AND team_member.user_id = ? AND team_member.permission = ?", query.SignedInUser.OrgId, query.SignedInUser.UserId, models.PERMISSION_ADMIN)

	type teamCount struct {
		Count int64
	}

	resp := make([]*teamCount, 0)
	if err := x.SQL(builder.GetSQLString(), builder.params...).Find(&resp); err != nil {
		return err
	}

	query.Result = len(resp) > 0 && resp[0].Count > 0

	return nil
}
