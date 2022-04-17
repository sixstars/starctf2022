package datasources

import (
	"os"
	"testing"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	logger log.Logger = log.New("fake.log")

	twoDatasourcesConfig            = "testdata/two-datasources"
	twoDatasourcesConfigPurgeOthers = "testdata/insert-two-delete-two"
	doubleDatasourcesConfig         = "testdata/double-default"
	allProperties                   = "testdata/all-properties"
	versionZero                     = "testdata/version-0"
	brokenYaml                      = "testdata/broken-yaml"
	multipleOrgsWithDefault         = "testdata/multiple-org-default"
	withoutDefaults                 = "testdata/appliedDefaults"
	invalidAccess                   = "testdata/invalid-access"

	fakeRepo *fakeRepository
)

func TestDatasourceAsConfig(t *testing.T) {
	Convey("Testing datasource as configuration", t, func() {
		fakeRepo = &fakeRepository{}
		bus.ClearBusHandlers()
		bus.AddHandler("test", mockDelete)
		bus.AddHandler("test", mockInsert)
		bus.AddHandler("test", mockUpdate)
		bus.AddHandler("test", mockGet)
		bus.AddHandler("test", mockGetOrg)

		Convey("apply default values when missing", func() {
			dc := newDatasourceProvisioner(logger)
			err := dc.applyChanges(withoutDefaults)
			if err != nil {
				t.Fatalf("applyChanges return an error %v", err)
			}

			So(len(fakeRepo.inserted), ShouldEqual, 1)
			So(fakeRepo.inserted[0].OrgId, ShouldEqual, 1)
			So(fakeRepo.inserted[0].Access, ShouldEqual, "proxy")
		})

		Convey("One configured datasource", func() {
			Convey("no datasource in database", func() {
				dc := newDatasourceProvisioner(logger)
				err := dc.applyChanges(twoDatasourcesConfig)
				if err != nil {
					t.Fatalf("applyChanges return an error %v", err)
				}

				So(len(fakeRepo.deleted), ShouldEqual, 0)
				So(len(fakeRepo.inserted), ShouldEqual, 2)
				So(len(fakeRepo.updated), ShouldEqual, 0)
			})

			Convey("One datasource in database with same name", func() {
				fakeRepo.loadAll = []*models.DataSource{
					{Name: "Graphite", OrgId: 1, Id: 1},
				}

				Convey("should update one datasource", func() {
					dc := newDatasourceProvisioner(logger)
					err := dc.applyChanges(twoDatasourcesConfig)
					if err != nil {
						t.Fatalf("applyChanges return an error %v", err)
					}

					So(len(fakeRepo.deleted), ShouldEqual, 0)
					So(len(fakeRepo.inserted), ShouldEqual, 1)
					So(len(fakeRepo.updated), ShouldEqual, 1)
				})
			})

			Convey("Two datasources with is_default", func() {
				dc := newDatasourceProvisioner(logger)
				err := dc.applyChanges(doubleDatasourcesConfig)
				Convey("should raise error", func() {
					So(err, ShouldEqual, ErrInvalidConfigToManyDefault)
				})
			})
		})

		Convey("Multiple datasources in different organizations with isDefault in each organization", func() {
			dc := newDatasourceProvisioner(logger)
			err := dc.applyChanges(multipleOrgsWithDefault)
			Convey("should not raise error", func() {
				So(err, ShouldBeNil)
				So(len(fakeRepo.inserted), ShouldEqual, 4)
				So(fakeRepo.inserted[0].IsDefault, ShouldBeTrue)
				So(fakeRepo.inserted[0].OrgId, ShouldEqual, 1)
				So(fakeRepo.inserted[2].IsDefault, ShouldBeTrue)
				So(fakeRepo.inserted[2].OrgId, ShouldEqual, 2)
			})
		})

		Convey("Two configured datasource and purge others ", func() {
			Convey("two other datasources in database", func() {
				fakeRepo.loadAll = []*models.DataSource{
					{Name: "old-graphite", OrgId: 1, Id: 1},
					{Name: "old-graphite2", OrgId: 1, Id: 2},
				}

				Convey("should have two new datasources", func() {
					dc := newDatasourceProvisioner(logger)
					err := dc.applyChanges(twoDatasourcesConfigPurgeOthers)
					if err != nil {
						t.Fatalf("applyChanges return an error %v", err)
					}

					So(len(fakeRepo.deleted), ShouldEqual, 2)
					So(len(fakeRepo.inserted), ShouldEqual, 2)
					So(len(fakeRepo.updated), ShouldEqual, 0)
				})
			})
		})

		Convey("Two configured datasource and purge others = false", func() {
			Convey("two other datasources in database", func() {
				fakeRepo.loadAll = []*models.DataSource{
					{Name: "Graphite", OrgId: 1, Id: 1},
					{Name: "old-graphite2", OrgId: 1, Id: 2},
				}

				Convey("should have two new datasources", func() {
					dc := newDatasourceProvisioner(logger)
					err := dc.applyChanges(twoDatasourcesConfig)
					if err != nil {
						t.Fatalf("applyChanges return an error %v", err)
					}

					So(len(fakeRepo.deleted), ShouldEqual, 0)
					So(len(fakeRepo.inserted), ShouldEqual, 1)
					So(len(fakeRepo.updated), ShouldEqual, 1)
				})
			})
		})

		Convey("broken yaml should return error", func() {
			reader := &configReader{}
			_, err := reader.readConfig(brokenYaml)
			So(err, ShouldNotBeNil)
		})

		Convey("invalid access should warn about invalid value and return 'proxy'", func() {
			reader := &configReader{log: logger}
			configs, err := reader.readConfig(invalidAccess)
			So(err, ShouldBeNil)
			So(configs[0].Datasources[0].Access, ShouldEqual, models.DS_ACCESS_PROXY)
		})

		Convey("skip invalid directory", func() {
			cfgProvider := &configReader{log: log.New("test logger")}
			cfg, err := cfgProvider.readConfig("./invalid-directory")
			if err != nil {
				t.Fatalf("readConfig return an error %v", err)
			}

			So(len(cfg), ShouldEqual, 0)
		})

		Convey("can read all properties from version 1", func() {
			_ = os.Setenv("TEST_VAR", "name")
			cfgProvider := &configReader{log: log.New("test logger")}
			cfg, err := cfgProvider.readConfig(allProperties)
			_ = os.Unsetenv("TEST_VAR")
			if err != nil {
				t.Fatalf("readConfig return an error %v", err)
			}

			So(len(cfg), ShouldEqual, 3)

			dsCfg := cfg[0]

			So(dsCfg.APIVersion, ShouldEqual, 1)

			validateDatasourceV1(dsCfg)
			validateDeleteDatasources(dsCfg)

			dsCount := 0
			delDsCount := 0

			for _, c := range cfg {
				dsCount += len(c.Datasources)
				delDsCount += len(c.DeleteDatasources)
			}

			So(dsCount, ShouldEqual, 2)
			So(delDsCount, ShouldEqual, 1)
		})

		Convey("can read all properties from version 0", func() {
			cfgProvider := &configReader{log: log.New("test logger")}
			cfg, err := cfgProvider.readConfig(versionZero)
			if err != nil {
				t.Fatalf("readConfig return an error %v", err)
			}

			So(len(cfg), ShouldEqual, 1)

			dsCfg := cfg[0]

			So(dsCfg.APIVersion, ShouldEqual, 0)

			validateDatasource(dsCfg)
			validateDeleteDatasources(dsCfg)
		})
	})
}

func validateDeleteDatasources(dsCfg *configs) {
	So(len(dsCfg.DeleteDatasources), ShouldEqual, 1)
	deleteDs := dsCfg.DeleteDatasources[0]
	So(deleteDs.Name, ShouldEqual, "old-graphite3")
	So(deleteDs.OrgID, ShouldEqual, 2)
}

func validateDatasource(dsCfg *configs) {
	ds := dsCfg.Datasources[0]
	So(ds.Name, ShouldEqual, "name")
	So(ds.Type, ShouldEqual, "type")
	So(ds.Access, ShouldEqual, models.DS_ACCESS_PROXY)
	So(ds.OrgID, ShouldEqual, 2)
	So(ds.URL, ShouldEqual, "url")
	So(ds.User, ShouldEqual, "user")
	So(ds.Password, ShouldEqual, "password")
	So(ds.Database, ShouldEqual, "database")
	So(ds.BasicAuth, ShouldBeTrue)
	So(ds.BasicAuthUser, ShouldEqual, "basic_auth_user")
	So(ds.BasicAuthPassword, ShouldEqual, "basic_auth_password")
	So(ds.WithCredentials, ShouldBeTrue)
	So(ds.IsDefault, ShouldBeTrue)
	So(ds.Editable, ShouldBeTrue)
	So(ds.Version, ShouldEqual, 10)

	So(len(ds.JSONData), ShouldBeGreaterThan, 2)
	So(ds.JSONData["graphiteVersion"], ShouldEqual, "1.1")
	So(ds.JSONData["tlsAuth"], ShouldEqual, true)
	So(ds.JSONData["tlsAuthWithCACert"], ShouldEqual, true)

	So(len(ds.SecureJSONData), ShouldBeGreaterThan, 2)
	So(ds.SecureJSONData["tlsCACert"], ShouldEqual, "MjNOcW9RdkbUDHZmpco2HCYzVq9dE+i6Yi+gmUJotq5CDA==")
	So(ds.SecureJSONData["tlsClientCert"], ShouldEqual, "ckN0dGlyMXN503YNfjTcf9CV+GGQneN+xmAclQ==")
	So(ds.SecureJSONData["tlsClientKey"], ShouldEqual, "ZkN4aG1aNkja/gKAB1wlnKFIsy2SRDq4slrM0A==")
}

func validateDatasourceV1(dsCfg *configs) {
	validateDatasource(dsCfg)
	ds := dsCfg.Datasources[0]
	So(ds.UID, ShouldEqual, "test_uid")
}

type fakeRepository struct {
	inserted []*models.AddDataSourceCommand
	deleted  []*models.DeleteDataSourceCommand
	updated  []*models.UpdateDataSourceCommand

	loadAll []*models.DataSource
}

func mockDelete(cmd *models.DeleteDataSourceCommand) error {
	fakeRepo.deleted = append(fakeRepo.deleted, cmd)
	return nil
}

func mockUpdate(cmd *models.UpdateDataSourceCommand) error {
	fakeRepo.updated = append(fakeRepo.updated, cmd)
	return nil
}

func mockInsert(cmd *models.AddDataSourceCommand) error {
	fakeRepo.inserted = append(fakeRepo.inserted, cmd)
	return nil
}

func mockGet(cmd *models.GetDataSourceQuery) error {
	for _, v := range fakeRepo.loadAll {
		if cmd.Name == v.Name && cmd.OrgId == v.OrgId {
			cmd.Result = v
			return nil
		}
	}

	return models.ErrDataSourceNotFound
}

func mockGetOrg(_ *models.GetOrgByIdQuery) error {
	return nil
}
