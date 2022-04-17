import { getDefaultTimeRange, LoadingState } from '@grafana/data';

import { variableAdapters } from '../adapters';
import { createQueryVariableAdapter } from './adapter';
import { reduxTester } from '../../../../test/core/redux/reduxTester';
import { getRootReducer, RootReducerType } from '../state/helpers';
import { QueryVariableModel, VariableHide, VariableRefresh, VariableSort } from '../types';
import { ALL_VARIABLE_TEXT, ALL_VARIABLE_VALUE, toVariablePayload } from '../state/types';
import {
  addVariable,
  changeVariableProp,
  setCurrentVariableValue,
  variableStateCompleted,
  variableStateFailed,
  variableStateFetching,
} from '../state/sharedReducer';
import {
  changeQueryVariableDataSource,
  changeQueryVariableQuery,
  flattenQuery,
  hasSelfReferencingQuery,
  initQueryVariableEditor,
  updateQueryVariableOptions,
} from './actions';
import { updateVariableOptions } from './reducer';
import {
  addVariableEditorError,
  changeVariableEditorExtended,
  removeVariableEditorError,
  setIdInEditor,
} from '../editor/reducer';
import { LegacyVariableQueryEditor } from '../editor/LegacyVariableQueryEditor';
import { expect } from 'test/lib/common';
import { updateOptions } from '../state/actions';
import { notifyApp } from '../../../core/reducers/appNotification';
import { silenceConsoleOutput } from '../../../../test/core/utils/silenceConsoleOutput';
import { getTimeSrv, setTimeSrv, TimeSrv } from '../../dashboard/services/TimeSrv';
import { setVariableQueryRunner, VariableQueryRunner } from './VariableQueryRunner';
import { setDataSourceSrv } from '@grafana/runtime';

const mocks: Record<string, any> = {
  datasource: {
    metricFindQuery: jest.fn().mockResolvedValue([]),
  },
  dataSourceSrv: {
    get: (name: string) => Promise.resolve(mocks[name]),
    getList: jest.fn().mockReturnValue([]),
  },
  pluginLoader: {
    importDataSourcePlugin: jest.fn().mockResolvedValue({ components: {} }),
  },
};

setDataSourceSrv(mocks.dataSourceSrv as any);

jest.mock('../../plugins/plugin_loader', () => ({
  importDataSourcePlugin: () => mocks.pluginLoader.importDataSourcePlugin(),
}));

jest.mock('../../templating/template_srv', () => ({
  replace: jest.fn().mockReturnValue(''),
}));

describe('query actions', () => {
  let originalTimeSrv: TimeSrv;

  beforeEach(() => {
    originalTimeSrv = getTimeSrv();
    setTimeSrv(({
      timeRange: jest.fn().mockReturnValue(getDefaultTimeRange()),
    } as unknown) as TimeSrv);
    setVariableQueryRunner(new VariableQueryRunner());
  });

  afterEach(() => {
    setTimeSrv(originalTimeSrv);
  });

  variableAdapters.setInit(() => [createQueryVariableAdapter()]);

  describe('when updateQueryVariableOptions is dispatched for variable without both tags and includeAll', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: false });
      const optionsMetrics = [createMetric('A'), createMetric('B')];

      mockDatasourceMetrics(variable, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(updateQueryVariableOptions(toVariablePayload(variable)), true);

      const option = createOption('A');
      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsShouldEqual(
        updateVariableOptions(toVariablePayload(variable, update)),
        setCurrentVariableValue(toVariablePayload(variable, { option }))
      );
    });
  });

  describe('when updateQueryVariableOptions is dispatched for variable with includeAll but without tags', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const optionsMetrics = [createMetric('A'), createMetric('B')];

      mockDatasourceMetrics(variable, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(updateQueryVariableOptions(toVariablePayload(variable)), true);

      const option = createOption(ALL_VARIABLE_TEXT, ALL_VARIABLE_VALUE);
      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [updateOptions, setCurrentAction] = actions;
        const expectedNumberOfActions = 2;

        expect(updateOptions).toEqual(updateVariableOptions(toVariablePayload(variable, update)));
        expect(setCurrentAction).toEqual(setCurrentVariableValue(toVariablePayload(variable, { option })));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when updateQueryVariableOptions is dispatched for variable open in editor', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const optionsMetrics = [createMetric('A'), createMetric('B')];

      mockDatasourceMetrics(variable, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenActionIsDispatched(setIdInEditor({ id: variable.id }))
        .whenAsyncActionIsDispatched(updateQueryVariableOptions(toVariablePayload(variable)), true);

      const option = createOption(ALL_VARIABLE_TEXT, ALL_VARIABLE_VALUE);
      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [clearErrors, updateOptions, setCurrentAction] = actions;
        const expectedNumberOfActions = 3;

        expect(clearErrors).toEqual(removeVariableEditorError({ errorProp: 'update' }));
        expect(updateOptions).toEqual(updateVariableOptions(toVariablePayload(variable, update)));
        expect(setCurrentAction).toEqual(setCurrentVariableValue(toVariablePayload(variable, { option })));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when updateQueryVariableOptions is dispatched for variable with searchFilter', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const optionsMetrics = [createMetric('A'), createMetric('B')];

      mockDatasourceMetrics(variable, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenActionIsDispatched(setIdInEditor({ id: variable.id }))
        .whenAsyncActionIsDispatched(updateQueryVariableOptions(toVariablePayload(variable), 'search'), true);

      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [clearErrors, updateOptions] = actions;
        const expectedNumberOfActions = 2;

        expect(clearErrors).toEqual(removeVariableEditorError({ errorProp: 'update' }));
        expect(updateOptions).toEqual(updateVariableOptions(toVariablePayload(variable, update)));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when updateQueryVariableOptions is dispatched and fails for variable open in editor', () => {
    silenceConsoleOutput();
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const error = { message: 'failed to fetch metrics' };

      mocks[variable.datasource!].metricFindQuery = jest.fn(() => Promise.reject(error));

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenActionIsDispatched(setIdInEditor({ id: variable.id }))
        .whenAsyncActionIsDispatched(updateOptions(toVariablePayload(variable)), true);

      tester.thenDispatchedActionsPredicateShouldEqual((dispatchedActions) => {
        const expectedNumberOfActions = 5;

        expect(dispatchedActions[0]).toEqual(variableStateFetching(toVariablePayload(variable)));
        expect(dispatchedActions[1]).toEqual(removeVariableEditorError({ errorProp: 'update' }));
        expect(dispatchedActions[2]).toEqual(addVariableEditorError({ errorProp: 'update', errorText: error.message }));
        expect(dispatchedActions[3]).toEqual(
          variableStateFailed(toVariablePayload(variable, { error: { message: 'failed to fetch metrics' } }))
        );
        expect(dispatchedActions[4].type).toEqual(notifyApp.type);
        expect(dispatchedActions[4].payload.title).toEqual('Templating [0]');
        expect(dispatchedActions[4].payload.text).toEqual('Error updating options: failed to fetch metrics');
        expect(dispatchedActions[4].payload.severity).toEqual('error');

        return dispatchedActions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when initQueryVariableEditor is dispatched', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const testMetricSource = { name: 'test', value: 'test', meta: {} };
      const editor = {};

      mocks.dataSourceSrv.getList = jest.fn().mockReturnValue([testMetricSource]);
      mocks.pluginLoader.importDataSourcePlugin = jest.fn().mockResolvedValue({
        components: { VariableQueryEditor: editor },
      });

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(initQueryVariableEditor(toVariablePayload(variable)), true);

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [setDatasource, setEditor] = actions;
        const expectedNumberOfActions = 2;

        expect(setDatasource).toEqual(
          changeVariableEditorExtended({ propName: 'dataSource', propValue: mocks['datasource'] })
        );
        expect(setEditor).toEqual(changeVariableEditorExtended({ propName: 'VariableQueryEditor', propValue: editor }));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when initQueryVariableEditor is dispatched and metricsource without value is available', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const testMetricSource = { name: 'test', value: (null as unknown) as string, meta: {} };
      const editor = {};

      mocks.dataSourceSrv.getList = jest.fn().mockReturnValue([testMetricSource]);
      mocks.pluginLoader.importDataSourcePlugin = jest.fn().mockResolvedValue({
        components: { VariableQueryEditor: editor },
      });

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(initQueryVariableEditor(toVariablePayload(variable)), true);

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [setDatasource, setEditor] = actions;
        const expectedNumberOfActions = 2;

        expect(setDatasource).toEqual(
          changeVariableEditorExtended({ propName: 'dataSource', propValue: mocks['datasource'] })
        );
        expect(setEditor).toEqual(changeVariableEditorExtended({ propName: 'VariableQueryEditor', propValue: editor }));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when initQueryVariableEditor is dispatched and no metric sources was found', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ includeAll: true });
      const editor = {};

      mocks.dataSourceSrv.getList = jest.fn().mockReturnValue([]);
      mocks.pluginLoader.importDataSourcePlugin = jest.fn().mockResolvedValue({
        components: { VariableQueryEditor: editor },
      });

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(initQueryVariableEditor(toVariablePayload(variable)), true);

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [setDatasource, setEditor] = actions;
        const expectedNumberOfActions = 2;

        expect(setDatasource).toEqual(
          changeVariableEditorExtended({ propName: 'dataSource', propValue: mocks['datasource'] })
        );
        expect(setEditor).toEqual(changeVariableEditorExtended({ propName: 'VariableQueryEditor', propValue: editor }));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when initQueryVariableEditor is dispatched and variable dont have datasource', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ datasource: undefined });

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(initQueryVariableEditor(toVariablePayload(variable)), true);

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [setDatasource] = actions;
        const expectedNumberOfActions = 1;

        expect(setDatasource).toEqual(changeVariableEditorExtended({ propName: 'dataSource', propValue: undefined }));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when changeQueryVariableDataSource is dispatched', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ datasource: 'other' });
      const editor = {};

      mocks.pluginLoader.importDataSourcePlugin = jest.fn().mockResolvedValue({
        components: { VariableQueryEditor: editor },
      });

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(changeQueryVariableDataSource(toVariablePayload(variable), 'datasource'), true);

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [updateDatasource, updateEditor] = actions;
        const expectedNumberOfActions = 2;

        expect(updateDatasource).toEqual(
          changeVariableEditorExtended({ propName: 'dataSource', propValue: mocks.datasource })
        );
        expect(updateEditor).toEqual(
          changeVariableEditorExtended({ propName: 'VariableQueryEditor', propValue: editor })
        );

        return actions.length === expectedNumberOfActions;
      });
    });

    describe('and data source type changed', () => {
      it('then correct actions are dispatched', async () => {
        const variable = createVariable({ datasource: 'other' });
        const editor = {};
        const preloadedState: any = { templating: { editor: { extended: { dataSource: { type: 'previous' } } } } };

        mocks.pluginLoader.importDataSourcePlugin = jest.fn().mockResolvedValue({
          components: { VariableQueryEditor: editor },
        });

        const tester = await reduxTester<RootReducerType>({ preloadedState })
          .givenRootReducer(getRootReducer())
          .whenActionIsDispatched(
            addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable }))
          )
          .whenAsyncActionIsDispatched(changeQueryVariableDataSource(toVariablePayload(variable), 'datasource'), true);

        tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
          const [changeVariable, updateDatasource, updateEditor] = actions;
          const expectedNumberOfActions = 3;

          expect(changeVariable).toEqual(
            changeVariableProp(toVariablePayload(variable, { propName: 'query', propValue: '' }))
          );
          expect(updateDatasource).toEqual(
            changeVariableEditorExtended({ propName: 'dataSource', propValue: mocks.datasource })
          );
          expect(updateEditor).toEqual(
            changeVariableEditorExtended({ propName: 'VariableQueryEditor', propValue: editor })
          );

          return actions.length === expectedNumberOfActions;
        });
      });
    });
  });

  describe('when changeQueryVariableDataSource is dispatched and editor is not configured', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ datasource: 'other' });
      const editor = LegacyVariableQueryEditor;

      mocks.pluginLoader.importDataSourcePlugin = jest.fn().mockResolvedValue({
        components: {},
      });

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(changeQueryVariableDataSource(toVariablePayload(variable), 'datasource'), true);

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [updateDatasource, updateEditor] = actions;
        const expectedNumberOfActions = 2;

        expect(updateDatasource).toEqual(
          changeVariableEditorExtended({ propName: 'dataSource', propValue: mocks.datasource })
        );
        expect(updateEditor).toEqual(
          changeVariableEditorExtended({ propName: 'VariableQueryEditor', propValue: editor })
        );

        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('when changeQueryVariableQuery is dispatched', () => {
    it('then correct actions are dispatched', async () => {
      const optionsMetrics = [createMetric('A'), createMetric('B')];
      const variable = createVariable({ datasource: 'datasource', includeAll: true });

      const query = '$datasource';
      const definition = 'depends on datasource variable';

      mockDatasourceMetrics({ ...variable, query }, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(changeQueryVariableQuery(toVariablePayload(variable), query, definition), true);

      const option = createOption(ALL_VARIABLE_TEXT, ALL_VARIABLE_VALUE);
      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsShouldEqual(
        removeVariableEditorError({ errorProp: 'query' }),
        changeVariableProp(toVariablePayload(variable, { propName: 'query', propValue: query })),
        changeVariableProp(toVariablePayload(variable, { propName: 'definition', propValue: definition })),
        variableStateFetching(toVariablePayload(variable)),
        updateVariableOptions(toVariablePayload(variable, update)),
        setCurrentVariableValue(toVariablePayload(variable, { option })),
        variableStateCompleted(toVariablePayload(variable))
      );
    });
  });

  describe('when changeQueryVariableQuery is dispatched for variable without tags', () => {
    it('then correct actions are dispatched', async () => {
      const optionsMetrics = [createMetric('A'), createMetric('B')];
      const variable = createVariable({ datasource: 'datasource', includeAll: true });

      const query = '$datasource';
      const definition = 'depends on datasource variable';

      mockDatasourceMetrics({ ...variable, query }, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(changeQueryVariableQuery(toVariablePayload(variable), query, definition), true);

      const option = createOption(ALL_VARIABLE_TEXT, ALL_VARIABLE_VALUE);
      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsShouldEqual(
        removeVariableEditorError({ errorProp: 'query' }),
        changeVariableProp(toVariablePayload(variable, { propName: 'query', propValue: query })),
        changeVariableProp(toVariablePayload(variable, { propName: 'definition', propValue: definition })),
        variableStateFetching(toVariablePayload(variable)),
        updateVariableOptions(toVariablePayload(variable, update)),
        setCurrentVariableValue(toVariablePayload(variable, { option })),
        variableStateCompleted(toVariablePayload(variable))
      );
    });
  });

  describe('when changeQueryVariableQuery is dispatched for variable without tags and all', () => {
    it('then correct actions are dispatched', async () => {
      const optionsMetrics = [createMetric('A'), createMetric('B')];
      const variable = createVariable({ datasource: 'datasource', includeAll: false });
      const query = '$datasource';
      const definition = 'depends on datasource variable';

      mockDatasourceMetrics({ ...variable, query }, optionsMetrics);

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(changeQueryVariableQuery(toVariablePayload(variable), query, definition), true);

      const option = createOption('A');
      const update = { results: optionsMetrics, templatedRegex: '' };

      tester.thenDispatchedActionsShouldEqual(
        removeVariableEditorError({ errorProp: 'query' }),
        changeVariableProp(toVariablePayload(variable, { propName: 'query', propValue: query })),
        changeVariableProp(toVariablePayload(variable, { propName: 'definition', propValue: definition })),
        variableStateFetching(toVariablePayload(variable)),
        updateVariableOptions(toVariablePayload(variable, update)),
        setCurrentVariableValue(toVariablePayload(variable, { option })),
        variableStateCompleted(toVariablePayload(variable))
      );
    });
  });

  describe('when changeQueryVariableQuery is dispatched with invalid query', () => {
    it('then correct actions are dispatched', async () => {
      const variable = createVariable({ datasource: 'datasource', includeAll: false });
      const query = `$${variable.name}`;
      const definition = 'depends on datasource variable';

      const tester = await reduxTester<RootReducerType>()
        .givenRootReducer(getRootReducer())
        .whenActionIsDispatched(addVariable(toVariablePayload(variable, { global: false, index: 0, model: variable })))
        .whenAsyncActionIsDispatched(changeQueryVariableQuery(toVariablePayload(variable), query, definition), true);

      const errorText = 'Query cannot contain a reference to itself. Variable: $' + variable.name;

      tester.thenDispatchedActionsPredicateShouldEqual((actions) => {
        const [editorError] = actions;
        const expectedNumberOfActions = 1;

        expect(editorError).toEqual(addVariableEditorError({ errorProp: 'query', errorText }));
        return actions.length === expectedNumberOfActions;
      });
    });
  });

  describe('hasSelfReferencingQuery', () => {
    it('when called with a string', () => {
      const query = '$query';
      const name = 'query';

      expect(hasSelfReferencingQuery(name, query)).toBe(true);
    });

    it('when called with an array', () => {
      const query = ['$query'];
      const name = 'query';

      expect(hasSelfReferencingQuery(name, query)).toBe(true);
    });

    it('when called with a simple object', () => {
      const query = { a: '$query' };
      const name = 'query';

      expect(hasSelfReferencingQuery(name, query)).toBe(true);
    });

    it('when called with a complex object', () => {
      const query = {
        level2: {
          level3: {
            query: 'query3',
            refId: 'C',
            num: 2,
            bool: true,
            arr: [
              { query: 'query4', refId: 'D', num: 4, bool: true },
              {
                query: 'query5',
                refId: 'E',
                num: 5,
                bool: true,
                arr: [{ query: '$query', refId: 'F', num: 6, bool: true }],
              },
            ],
          },
          query: 'query2',
          refId: 'B',
          num: 1,
          bool: false,
        },
        query: 'query1',
        refId: 'A',
        num: 0,
        bool: true,
        arr: [
          { query: 'query7', refId: 'G', num: 7, bool: true },
          {
            query: 'query8',
            refId: 'H',
            num: 8,
            bool: true,
            arr: [{ query: 'query9', refId: 'I', num: 9, bool: true }],
          },
        ],
      };
      const name = 'query';

      expect(hasSelfReferencingQuery(name, query)).toBe(true);
    });

    it('when called with a number', () => {
      const query = 1;
      const name = 'query';

      expect(hasSelfReferencingQuery(name, query)).toBe(false);
    });
  });

  describe('flattenQuery', () => {
    it('when called with a complex object', () => {
      const query = {
        level2: {
          level3: {
            query: '${query3}',
            refId: 'C',
            num: 2,
            bool: true,
            arr: [
              { query: '${query4}', refId: 'D', num: 4, bool: true },
              {
                query: '${query5}',
                refId: 'E',
                num: 5,
                bool: true,
                arr: [{ query: '${query6}', refId: 'F', num: 6, bool: true }],
              },
            ],
          },
          query: '${query2}',
          refId: 'B',
          num: 1,
          bool: false,
        },
        query: '${query1}',
        refId: 'A',
        num: 0,
        bool: true,
        arr: [
          { query: '${query7}', refId: 'G', num: 7, bool: true },
          {
            query: '${query8}',
            refId: 'H',
            num: 8,
            bool: true,
            arr: [{ query: '${query9}', refId: 'I', num: 9, bool: true }],
          },
        ],
      };

      expect(flattenQuery(query)).toEqual({
        query: '${query1}',
        refId: 'A',
        num: 0,
        bool: true,
        level2_query: '${query2}',
        level2_refId: 'B',
        level2_num: 1,
        level2_bool: false,
        level2_level3_query: '${query3}',
        level2_level3_refId: 'C',
        level2_level3_num: 2,
        level2_level3_bool: true,
        level2_level3_arr_0_query: '${query4}',
        level2_level3_arr_0_refId: 'D',
        level2_level3_arr_0_num: 4,
        level2_level3_arr_0_bool: true,
        level2_level3_arr_1_query: '${query5}',
        level2_level3_arr_1_refId: 'E',
        level2_level3_arr_1_num: 5,
        level2_level3_arr_1_bool: true,
        level2_level3_arr_1_arr_0_query: '${query6}',
        level2_level3_arr_1_arr_0_refId: 'F',
        level2_level3_arr_1_arr_0_num: 6,
        level2_level3_arr_1_arr_0_bool: true,
        arr_0_query: '${query7}',
        arr_0_refId: 'G',
        arr_0_num: 7,
        arr_0_bool: true,
        arr_1_query: '${query8}',
        arr_1_refId: 'H',
        arr_1_num: 8,
        arr_1_bool: true,
        arr_1_arr_0_query: '${query9}',
        arr_1_arr_0_refId: 'I',
        arr_1_arr_0_num: 9,
        arr_1_arr_0_bool: true,
      });
    });
  });
});

function mockDatasourceMetrics(variable: QueryVariableModel, optionsMetrics: any[]) {
  const metrics: Record<string, any[]> = {
    [variable.query]: optionsMetrics,
  };

  const { metricFindQuery } = mocks[variable.datasource!];

  metricFindQuery.mockReset();
  metricFindQuery.mockImplementation((query: string) => Promise.resolve(metrics[query] ?? []));
}

function createVariable(extend?: Partial<QueryVariableModel>): QueryVariableModel {
  return {
    type: 'query',
    id: '0',
    global: false,
    current: createOption(''),
    options: [],
    query: 'options-query',
    name: 'Constant',
    label: '',
    hide: VariableHide.dontHide,
    skipUrlSync: false,
    index: 0,
    datasource: 'datasource',
    definition: '',
    sort: VariableSort.alphabeticalAsc,
    refresh: VariableRefresh.onDashboardLoad,
    regex: '',
    multi: true,
    includeAll: true,
    state: LoadingState.NotStarted,
    error: null,
    description: null,
    ...(extend ?? {}),
  };
}

function createOption(text: string, value?: string) {
  const metric = createMetric(text);
  return {
    ...metric,
    value: value ?? metric.text,
    selected: false,
  };
}

function createMetric(value: string) {
  return {
    text: value,
  };
}
