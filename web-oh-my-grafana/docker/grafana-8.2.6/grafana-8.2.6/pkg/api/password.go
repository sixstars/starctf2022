package api

import (
	"errors"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

func SendResetPasswordEmail(c *models.ReqContext, form dtos.SendResetPasswordEmailForm) response.Response {
	if setting.LDAPEnabled || setting.AuthProxyEnabled {
		return response.Error(401, "Not allowed to reset password when LDAP or Auth Proxy is enabled", nil)
	}
	if setting.DisableLoginForm {
		return response.Error(401, "Not allowed to reset password when login form is disabled", nil)
	}

	userQuery := models.GetUserByLoginQuery{LoginOrEmail: form.UserOrEmail}

	if err := bus.Dispatch(&userQuery); err != nil {
		c.Logger.Info("Requested password reset for user that was not found", "user", userQuery.LoginOrEmail)
		return response.Error(200, "Email sent", err)
	}

	emailCmd := models.SendResetPasswordEmailCommand{User: userQuery.Result}
	if err := bus.Dispatch(&emailCmd); err != nil {
		return response.Error(500, "Failed to send email", err)
	}

	return response.Success("Email sent")
}

func ResetPassword(c *models.ReqContext, form dtos.ResetUserPasswordForm) response.Response {
	return response.Success("Failed to change user password")
	query := models.ValidateResetPasswordCodeQuery{Code: form.Code}

	if err := bus.Dispatch(&query); err != nil {
		if errors.Is(err, models.ErrInvalidEmailCode) {
			return response.Error(400, "Invalid or expired reset password code", nil)
		}
		return response.Error(500, "Unknown error validating email code", err)
	}

	if form.NewPassword != form.ConfirmPassword {
		return response.Error(400, "Passwords do not match", nil)
	}

	cmd := models.ChangeUserPasswordCommand{}
	cmd.UserId = query.Result.Id
	var err error
	cmd.NewPassword, err = util.EncodePassword(form.NewPassword, query.Result.Salt)
	if err != nil {
		return response.Error(500, "Failed to encode password", err)
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return response.Error(500, "Failed to change user password", err)
	}

	return response.Success("User password changed")
}
