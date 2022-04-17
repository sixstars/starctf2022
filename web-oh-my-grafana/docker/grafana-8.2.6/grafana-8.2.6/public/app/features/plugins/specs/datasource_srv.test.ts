import { DatasourceSrv } from 'app/features/plugins/datasource_srv';
import { DataSourceInstanceSettings, DataSourcePlugin } from '@grafana/data';

// Datasource variable $datasource with current value 'BBB'
const templateSrv: any = {
  getVariables: () => [
    {
      type: 'datasource',
      name: 'datasource',
      current: {
        value: 'BBB',
      },
    },
    {
      type: 'datasource',
      name: 'datasourceDefault',
      current: {
        value: 'default',
      },
    },
  ],
  replace: (v: string) => {
    let result = v.replace('${datasource}', 'BBB');
    result = result.replace('${datasourceDefault}', 'default');
    return result;
  },
};

class TestDataSource {
  constructor(public instanceSettings: DataSourceInstanceSettings) {}
}

jest.mock('../plugin_loader', () => ({
  importDataSourcePlugin: () => {
    return Promise.resolve(new DataSourcePlugin(TestDataSource as any));
  },
}));

describe('datasource_srv', () => {
  const dataSourceSrv = new DatasourceSrv({} as any, {} as any, templateSrv);
  const dataSourceInit = {
    mmm: {
      type: 'test-db',
      name: 'mmm',
      uid: 'uid-code-mmm',
      meta: { metrics: true, annotations: true } as any,
    },
    '-- Grafana --': {
      type: 'grafana',
      name: '-- Grafana --',
      meta: { builtIn: true, metrics: true, id: 'grafana' },
    },
    '-- Dashboard --': {
      type: 'dashboard',
      name: '-- Dashboard --',
      meta: { builtIn: true, metrics: true, id: 'dashboard' },
    },
    '-- Mixed --': {
      type: 'test-db',
      name: '-- Mixed --',
      meta: { builtIn: true, metrics: true, id: 'mixed' },
    },
    ZZZ: {
      type: 'test-db',
      name: 'ZZZ',
      uid: 'uid-code-ZZZ',
      meta: { metrics: true },
    },
    aaa: {
      type: 'test-db',
      name: 'aaa',
      uid: 'uid-code-aaa',
      meta: { metrics: true },
    },
    BBB: {
      type: 'test-db',
      name: 'BBB',
      uid: 'uid-code-BBB',
      meta: { metrics: true },
      isDefault: true,
    },
    Jaeger: {
      type: 'jaeger-db',
      name: 'Jaeger',
      uid: 'uid-code-Jaeger',
      meta: { tracing: true, id: 'jaeger' },
    },
    CannotBeQueried: {
      type: 'no-query',
      name: 'no-query',
      uid: 'no-query',
      meta: { id: 'no-query' },
    },
  };

  describe('Given a list of data sources', () => {
    beforeEach(() => {
      dataSourceSrv.init(dataSourceInit as any, 'BBB');
    });

    describe('when getting data source class instance', () => {
      it('should load plugin and create instance and set meta', async () => {
        const ds = (await dataSourceSrv.get('mmm')) as any;
        expect(ds.meta).toBe(dataSourceInit.mmm.meta);
        expect(ds.instanceSettings).toBe(dataSourceInit.mmm);

        // validate that it caches instance
        const ds2 = await dataSourceSrv.get('mmm');
        expect(ds).toBe(ds2);
      });

      it('should be able to load data source using uid as well', async () => {
        const dsByUid = await dataSourceSrv.get('uid-code-mmm');
        const dsByName = await dataSourceSrv.get('mmm');
        expect(dsByUid.meta).toBe(dsByName.meta);
        expect(dsByUid).toBe(dsByName);
      });
    });

    describe('when getting instance settings', () => {
      it('should work by name or uid', () => {
        expect(dataSourceSrv.getInstanceSettings('mmm')).toBe(dataSourceSrv.getInstanceSettings('uid-code-mmm'));
      });

      it('should work with variable', () => {
        const ds = dataSourceSrv.getInstanceSettings('${datasource}');
        expect(ds?.name).toBe('${datasource}');
        expect(ds?.uid).toBe('uid-code-BBB');
      });

      it('should not set isDefault when being fetched via variable', () => {
        const ds = dataSourceSrv.getInstanceSettings('${datasource}');
        expect(ds?.isDefault).toBe(false);
      });

      it('should work with variable', () => {
        const ds = dataSourceSrv.getInstanceSettings('${datasourceDefault}');
        expect(ds?.name).toBe('${datasourceDefault}');
        expect(ds?.uid).toBe('uid-code-BBB');
      });
    });

    describe('when getting external metric sources', () => {
      it('should return list of explore sources', () => {
        const externalSources = dataSourceSrv.getExternal();
        expect(externalSources.length).toBe(6);
      });
    });

    it('Should by default filter out data sources that cannot be queried', () => {
      const list = dataSourceSrv.getList({});
      expect(list.find((x) => x.name === 'no-query')).toBeUndefined();
      const all = dataSourceSrv.getList({ all: true });
      expect(all.find((x) => x.name === 'no-query')).toBeDefined();
    });

    it('Can get list of data sources with variables: true', () => {
      const list = dataSourceSrv.getList({ metrics: true, variables: true });
      expect(list[0].name).toBe('${datasourceDefault}');
      expect(list[1].name).toBe('${datasource}');
    });

    it('Can get list of data sources with tracing: true', () => {
      const list = dataSourceSrv.getList({ tracing: true });
      expect(list[0].name).toBe('Jaeger');
    });

    it('Can get list of data sources with annotation: true', () => {
      const list = dataSourceSrv.getList({ annotations: true });
      expect(list[0].name).toBe('mmm');
    });

    it('Can get get list and filter by pluginId', () => {
      const list = dataSourceSrv.getList({ pluginId: 'jaeger' });
      expect(list[0].name).toBe('Jaeger');
      expect(list.length).toBe(1);
    });

    it('Can get list  of data sources with metrics: true, builtIn: true, mixed: true', () => {
      expect(dataSourceSrv.getList({ metrics: true, dashboard: true, mixed: true })).toMatchInlineSnapshot(`
        Array [
          Object {
            "meta": Object {
              "metrics": true,
            },
            "name": "aaa",
            "type": "test-db",
            "uid": "uid-code-aaa",
          },
          Object {
            "isDefault": true,
            "meta": Object {
              "metrics": true,
            },
            "name": "BBB",
            "type": "test-db",
            "uid": "uid-code-BBB",
          },
          Object {
            "meta": Object {
              "annotations": true,
              "metrics": true,
            },
            "name": "mmm",
            "type": "test-db",
            "uid": "uid-code-mmm",
          },
          Object {
            "meta": Object {
              "metrics": true,
            },
            "name": "ZZZ",
            "type": "test-db",
            "uid": "uid-code-ZZZ",
          },
          Object {
            "meta": Object {
              "builtIn": true,
              "id": "mixed",
              "metrics": true,
            },
            "name": "-- Mixed --",
            "type": "test-db",
          },
          Object {
            "meta": Object {
              "builtIn": true,
              "id": "dashboard",
              "metrics": true,
            },
            "name": "-- Dashboard --",
            "type": "dashboard",
          },
          Object {
            "meta": Object {
              "builtIn": true,
              "id": "grafana",
              "metrics": true,
            },
            "name": "-- Grafana --",
            "type": "grafana",
          },
        ]
      `);
    });
  });
});
