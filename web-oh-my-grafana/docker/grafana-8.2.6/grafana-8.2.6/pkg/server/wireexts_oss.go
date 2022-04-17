//go:build wireinject && oss
// +build wireinject,oss

package server

import (
	"github.com/google/wire"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/server/backgroundsvcs"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/accesscontrol/ossaccesscontrol"
	"github.com/grafana/grafana/pkg/services/auth"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/encryption"
	"github.com/grafana/grafana/pkg/services/encryption/ossencryption"
	"github.com/grafana/grafana/pkg/services/licensing"
	"github.com/grafana/grafana/pkg/services/login"
	"github.com/grafana/grafana/pkg/services/login/authinfoservice"
	"github.com/grafana/grafana/pkg/services/provisioning"
	"github.com/grafana/grafana/pkg/services/searchusers"
	"github.com/grafana/grafana/pkg/services/sqlstore/migrations"
	"github.com/grafana/grafana/pkg/services/validations"
	"github.com/grafana/grafana/pkg/setting"
)

var wireExtsBasicSet = wire.NewSet(
	auth.ProvideUserAuthTokenService,
	wire.Bind(new(models.UserTokenService), new(*auth.UserAuthTokenService)),
	wire.Bind(new(models.UserTokenBackgroundService), new(*auth.UserAuthTokenService)),
	licensing.ProvideService,
	wire.Bind(new(models.Licensing), new(*licensing.OSSLicensingService)),
	setting.ProvideProvider,
	wire.Bind(new(setting.Provider), new(*setting.OSSImpl)),
	ossaccesscontrol.ProvideService,
	wire.Bind(new(accesscontrol.RoleRegistry), new(*ossaccesscontrol.OSSAccessControlService)),
	wire.Bind(new(accesscontrol.AccessControl), new(*ossaccesscontrol.OSSAccessControlService)),
	validations.ProvideValidator,
	wire.Bind(new(models.PluginRequestValidator), new(*validations.OSSPluginRequestValidator)),
	provisioning.ProvideService,
	wire.Bind(new(provisioning.ProvisioningService), new(*provisioning.ProvisioningServiceImpl)),
	backgroundsvcs.ProvideBackgroundServiceRegistry,
	wire.Bind(new(registry.BackgroundServiceRegistry), new(*backgroundsvcs.BackgroundServiceRegistry)),
	datasources.ProvideCacheService,
	wire.Bind(new(datasources.CacheService), new(*datasources.CacheServiceImpl)),
	migrations.ProvideOSSMigrations,
	wire.Bind(new(registry.DatabaseMigrator), new(*migrations.OSSMigrations)),
	authinfoservice.ProvideOSSUserProtectionService,
	wire.Bind(new(login.UserProtectionService), new(*authinfoservice.OSSUserProtectionImpl)),
	ossencryption.ProvideService,
	wire.Bind(new(encryption.Service), new(*ossencryption.Service)),
	searchusers.ProvideUsersService,
	wire.Bind(new(searchusers.Service), new(*searchusers.OSSService)),
)

var wireExtsSet = wire.NewSet(
	wireSet,
	wireExtsBasicSet,
)

var wireExtsTestSet = wire.NewSet(
	wireTestSet,
	wireExtsBasicSet,
)
