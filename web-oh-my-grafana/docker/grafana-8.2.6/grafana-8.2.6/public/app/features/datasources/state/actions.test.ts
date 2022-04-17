import {
  findNewName,
  nameExits,
  InitDataSourceSettingDependencies,
  testDataSource,
  TestDataSourceDependencies,
  getDataSourceUsingUidOrId,
} from './actions';
import { getMockPlugin, getMockPlugins } from '../../plugins/__mocks__/pluginMocks';
import { thunkTester } from 'test/core/thunk/thunkTester';
import {
  initDataSourceSettingsSucceeded,
  initDataSourceSettingsFailed,
  testDataSourceStarting,
  testDataSourceSucceeded,
  testDataSourceFailed,
} from './reducers';
import { initDataSourceSettings } from '../state/actions';
import { ThunkResult, ThunkDispatch } from 'app/types';
import { GenericDataSourcePlugin } from '../settings/PluginSettings';
import { getBackendSrv } from 'app/core/services/backend_srv';
import { BackendSrvRequest, FetchResponse } from '@grafana/runtime';
import { of } from 'rxjs';

jest.mock('app/core/services/backend_srv');
jest.mock('@grafana/runtime', () => ({
  ...((jest.requireActual('@grafana/runtime') as unknown) as object),
  getBackendSrv: jest.fn(),
}));

const getBackendSrvMock = () =>
  ({
    get: jest.fn().mockReturnValue({
      testDatasource: jest.fn().mockReturnValue({
        status: '',
        message: '',
      }),
    }),
    withNoBackendCache: jest.fn().mockImplementationOnce((cb) => cb()),
  } as any);

const failDataSourceTest = async (error: object) => {
  const dependencies: TestDataSourceDependencies = {
    getDatasourceSrv: () =>
      ({
        get: jest.fn().mockReturnValue({
          testDatasource: jest.fn().mockImplementation(() => {
            throw error;
          }),
        }),
      } as any),
    getBackendSrv: getBackendSrvMock,
  };
  const state = {
    testingStatus: {
      message: '',
      status: '',
    },
  };
  const dispatchedActions = await thunkTester(state)
    .givenThunk(testDataSource)
    .whenThunkIsDispatched('Azure Monitor', dependencies);

  return dispatchedActions;
};

describe('getDataSourceUsingUidOrId', () => {
  const uidResponse = {
    ok: true,
    data: {
      id: 111,
      uid: 'abcdefg',
    },
  };

  const idResponse = {
    ok: true,
    data: {
      id: 222,
      uid: 'xyz',
    },
  };

  it('should return UID response data', async () => {
    (getBackendSrv as jest.Mock).mockReturnValueOnce({
      fetch: (options: BackendSrvRequest) => {
        return of(uidResponse as FetchResponse);
      },
    });

    expect(await getDataSourceUsingUidOrId('abcdefg')).toBe(uidResponse.data);
  });

  it('should return ID response data', async () => {
    const uidResponse = {
      ok: false,
    };

    (getBackendSrv as jest.Mock)
      .mockReturnValueOnce({
        fetch: (options: BackendSrvRequest) => {
          return of(uidResponse as FetchResponse);
        },
      })
      .mockReturnValueOnce({
        fetch: (options: BackendSrvRequest) => {
          return of(idResponse as FetchResponse);
        },
      });

    expect(await getDataSourceUsingUidOrId(222)).toBe(idResponse.data);
  });

  it('should return empty response data', async () => {
    // @ts-ignore
    delete window.location;
    window.location = {} as Location;

    const uidResponse = {
      ok: false,
    };

    (getBackendSrv as jest.Mock)
      .mockReturnValueOnce({
        fetch: (options: BackendSrvRequest) => {
          return of(uidResponse as FetchResponse);
        },
      })
      .mockReturnValueOnce({
        fetch: (options: BackendSrvRequest) => {
          return of(idResponse as FetchResponse);
        },
      });

    expect(await getDataSourceUsingUidOrId('222')).toStrictEqual({});
    expect(window.location.href).toBe('/datasources/edit/xyz');
  });
});

describe('Name exists', () => {
  const plugins = getMockPlugins(5);

  it('should be true', () => {
    const name = 'pretty cool plugin-1';

    expect(nameExits(plugins, name)).toEqual(true);
  });

  it('should be false', () => {
    const name = 'pretty cool plugin-6';

    expect(nameExits(plugins, name));
  });
});

describe('Find new name', () => {
  it('should create a new name', () => {
    const plugins = getMockPlugins(5);
    const name = 'pretty cool plugin-1';

    expect(findNewName(plugins, name)).toEqual('pretty cool plugin-6');
  });

  it('should create new name without suffix', () => {
    const plugin = getMockPlugin();
    plugin.name = 'prometheus';
    const plugins = [plugin];
    const name = 'prometheus';

    expect(findNewName(plugins, name)).toEqual('prometheus-1');
  });

  it('should handle names that end with -', () => {
    const plugin = getMockPlugin();
    const plugins = [plugin];
    const name = 'pretty cool plugin-';

    expect(findNewName(plugins, name)).toEqual('pretty cool plugin-');
  });
});

describe('initDataSourceSettings', () => {
  describe('when pageId is missing', () => {
    it('then initDataSourceSettingsFailed should be dispatched', async () => {
      const dispatchedActions = await thunkTester({}).givenThunk(initDataSourceSettings).whenThunkIsDispatched('');

      expect(dispatchedActions).toEqual([initDataSourceSettingsFailed(new Error('Invalid ID'))]);
    });
  });

  describe('when pageId is a valid', () => {
    it('then initDataSourceSettingsSucceeded should be dispatched', async () => {
      const dataSource = { type: 'app' };
      const dataSourceMeta = { id: 'some id' };
      const dependencies: InitDataSourceSettingDependencies = {
        loadDataSource: jest.fn((): ThunkResult<void> => (dispatch: ThunkDispatch, getState) => dataSource) as any,
        loadDataSourceMeta: jest.fn((): ThunkResult<void> => (dispatch: ThunkDispatch, getState) => {}),
        getDataSource: jest.fn().mockReturnValue(dataSource),
        getDataSourceMeta: jest.fn().mockReturnValue(dataSourceMeta),
        importDataSourcePlugin: jest.fn().mockReturnValue({} as GenericDataSourcePlugin),
      };
      const state = {
        dataSourceSettings: {},
        dataSources: {},
      };
      const dispatchedActions = await thunkTester(state)
        .givenThunk(initDataSourceSettings)
        .whenThunkIsDispatched(256, dependencies);

      expect(dispatchedActions).toEqual([initDataSourceSettingsSucceeded({} as GenericDataSourcePlugin)]);
      expect(dependencies.loadDataSource).toHaveBeenCalledTimes(1);
      expect(dependencies.loadDataSource).toHaveBeenCalledWith(256);

      expect(dependencies.loadDataSourceMeta).toHaveBeenCalledTimes(1);
      expect(dependencies.loadDataSourceMeta).toHaveBeenCalledWith(dataSource);

      expect(dependencies.getDataSource).toHaveBeenCalledTimes(1);
      expect(dependencies.getDataSource).toHaveBeenCalledWith({}, 256);

      expect(dependencies.getDataSourceMeta).toHaveBeenCalledTimes(1);
      expect(dependencies.getDataSourceMeta).toHaveBeenCalledWith({}, 'app');

      expect(dependencies.importDataSourcePlugin).toHaveBeenCalledTimes(1);
      expect(dependencies.importDataSourcePlugin).toHaveBeenCalledWith(dataSourceMeta);
    });
  });

  describe('when plugin loading fails', () => {
    it('then initDataSourceSettingsFailed should be dispatched', async () => {
      const dataSource = { type: 'app' };
      const dependencies: InitDataSourceSettingDependencies = {
        loadDataSource: jest.fn((): ThunkResult<void> => (dispatch: ThunkDispatch, getState) => dataSource) as any,
        loadDataSourceMeta: jest.fn().mockImplementation(() => {
          throw new Error('Error loading plugin');
        }),
        getDataSource: jest.fn(),
        getDataSourceMeta: jest.fn(),
        importDataSourcePlugin: jest.fn(),
      };
      const state = {
        dataSourceSettings: {},
        dataSources: {},
      };
      const dispatchedActions = await thunkTester(state)
        .givenThunk(initDataSourceSettings)
        .whenThunkIsDispatched(301, dependencies);

      expect(dispatchedActions).toEqual([initDataSourceSettingsFailed(new Error('Error loading plugin'))]);
      expect(dependencies.loadDataSource).toHaveBeenCalledTimes(1);
      expect(dependencies.loadDataSource).toHaveBeenCalledWith(301);

      expect(dependencies.loadDataSourceMeta).toHaveBeenCalledTimes(1);
      expect(dependencies.loadDataSourceMeta).toHaveBeenCalledWith(dataSource);
    });
  });
});

describe('testDataSource', () => {
  describe('when a datasource is tested', () => {
    it('then testDataSourceStarting and testDataSourceSucceeded should be dispatched', async () => {
      const dependencies: TestDataSourceDependencies = {
        getDatasourceSrv: () =>
          ({
            get: jest.fn().mockReturnValue({
              testDatasource: jest.fn().mockReturnValue({
                status: '',
                message: '',
              }),
            }),
          } as any),
        getBackendSrv: getBackendSrvMock,
      };
      const state = {
        testingStatus: {
          status: '',
          message: '',
        },
      };
      const dispatchedActions = await thunkTester(state)
        .givenThunk(testDataSource)
        .whenThunkIsDispatched('Azure Monitor', dependencies);

      expect(dispatchedActions).toEqual([testDataSourceStarting(), testDataSourceSucceeded(state.testingStatus)]);
    });

    it('then testDataSourceFailed should be dispatched', async () => {
      const dependencies: TestDataSourceDependencies = {
        getDatasourceSrv: () =>
          ({
            get: jest.fn().mockReturnValue({
              testDatasource: jest.fn().mockImplementation(() => {
                throw new Error('Error testing datasource');
              }),
            }),
          } as any),
        getBackendSrv: getBackendSrvMock,
      };
      const result = {
        message: 'Error testing datasource',
      };
      const state = {
        testingStatus: {
          message: '',
          status: '',
        },
      };
      const dispatchedActions = await thunkTester(state)
        .givenThunk(testDataSource)
        .whenThunkIsDispatched('Azure Monitor', dependencies);

      expect(dispatchedActions).toEqual([testDataSourceStarting(), testDataSourceFailed(result)]);
    });

    it('then testDataSourceFailed should be dispatched with response error message', async () => {
      const result = {
        message: 'Error testing datasource',
      };
      const dispatchedActions = await failDataSourceTest({
        message: 'Error testing datasource',
        data: { message: 'Response error message' },
        statusText: 'Bad Request',
      });
      expect(dispatchedActions).toEqual([testDataSourceStarting(), testDataSourceFailed(result)]);
    });

    it('then testDataSourceFailed should be dispatched with response data message', async () => {
      const result = {
        message: 'Response error message',
      };
      const dispatchedActions = await failDataSourceTest({
        data: { message: 'Response error message' },
        statusText: 'Bad Request',
      });
      expect(dispatchedActions).toEqual([testDataSourceStarting(), testDataSourceFailed(result)]);
    });

    it('then testDataSourceFailed should be dispatched with response statusText', async () => {
      const result = {
        message: 'HTTP error Bad Request',
      };
      const dispatchedActions = await failDataSourceTest({ data: {}, statusText: 'Bad Request' });
      expect(dispatchedActions).toEqual([testDataSourceStarting(), testDataSourceFailed(result)]);
    });
  });
});
