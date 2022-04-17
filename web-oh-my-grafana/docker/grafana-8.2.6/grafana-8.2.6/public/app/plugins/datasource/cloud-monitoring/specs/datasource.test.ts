import { of, throwError } from 'rxjs';
import { DataSourceInstanceSettings, toUtc } from '@grafana/data';

import CloudMonitoringDataSource from '../datasource';
import { metricDescriptors } from './testData';
import { TemplateSrv } from 'app/features/templating/template_srv';
import { CloudMonitoringOptions } from '../types';
import { backendSrv } from 'app/core/services/backend_srv'; // will use the version in __mocks__
import { TimeSrv } from 'app/features/dashboard/services/TimeSrv';
import { CustomVariableModel } from '../../../../features/variables/types';
import { initialCustomVariableModelState } from '../../../../features/variables/custom/reducer';
import { createFetchResponse } from 'test/helpers/createFetchResponse';

jest.mock('@grafana/runtime', () => ({
  ...((jest.requireActual('@grafana/runtime') as unknown) as object),
  getBackendSrv: () => backendSrv,
}));

type Args = { response?: any; throws?: boolean; templateSrv?: TemplateSrv };

function getTestcontext({ response = {}, throws = false, templateSrv = new TemplateSrv() }: Args = {}) {
  jest.clearAllMocks();

  const instanceSettings = ({
    jsonData: {
      defaultProject: 'testproject',
    },
  } as unknown) as DataSourceInstanceSettings<CloudMonitoringOptions>;

  const timeSrv = {} as TimeSrv;

  const fetchMock = jest.spyOn(backendSrv, 'fetch');

  throws
    ? fetchMock.mockImplementation(() => throwError(response))
    : fetchMock.mockImplementation(() => of(createFetchResponse(response)));

  const ds = new CloudMonitoringDataSource(instanceSettings, templateSrv, timeSrv);

  return { ds };
}

describe('CloudMonitoringDataSource', () => {
  describe('when performing testDataSource', () => {
    describe('and call to cloud monitoring api succeeds', () => {
      it('should return successfully', async () => {
        const { ds } = getTestcontext();

        const result = await ds.testDatasource();

        expect(result.status).toBe('success');
      });
    });

    describe('and a list of metricDescriptors are returned', () => {
      it('should return status success', async () => {
        const { ds } = getTestcontext({ response: metricDescriptors });

        const result = await ds.testDatasource();

        expect(result.status).toBe('success');
      });
    });

    describe('and call to cloud monitoring api fails with 400 error', () => {
      it('should return error status and a detailed error message', async () => {
        const response = {
          statusText: 'Bad Request',
          data: {
            error: { code: 400, message: 'Field interval.endTime had an invalid value' },
          },
        };
        const { ds } = getTestcontext({ response, throws: true });

        const result = await ds.testDatasource();

        expect(result.status).toEqual('error');
        expect(result.message).toBe(
          'Google Cloud Monitoring: Bad Request: 400. Field interval.endTime had an invalid value'
        );
      });
    });
  });

  describe('When performing query', () => {
    describe('and no time series data is returned', () => {
      it('should return a list of datapoints', async () => {
        const options = {
          range: {
            from: toUtc('2017-08-22T20:00:00Z'),
            to: toUtc('2017-08-22T23:59:00Z'),
          },
          rangeRaw: {
            from: 'now-4h',
            to: 'now',
          },
          targets: [
            {
              refId: 'A',
            },
          ],
        };

        const response: any = {
          results: {
            A: {
              refId: 'A',
              meta: {
                rawQuery: 'arawquerystring',
              },
              series: null,
              tables: null,
            },
          },
        };

        const { ds } = getTestcontext({ response });

        await expect(ds.query(options as any)).toEmitValuesWith((received) => {
          const results = received[0];
          expect(results.data.length).toBe(0);
        });
      });
    });
  });

  describe('when performing getMetricTypes', () => {
    describe('and call to cloud monitoring api succeeds', () => {
      it('should return successfully', async () => {
        const response = {
          metricDescriptors: [
            {
              displayName: 'test metric name 1',
              type: 'compute.googleapis.com/instance/cpu/test-metric-type-1',
              description: 'A description',
            },
            {
              type: 'logging.googleapis.com/user/logbased-metric-with-no-display-name',
            },
          ],
        };
        const { ds } = getTestcontext({ response });

        const result = await ds.getMetricTypes('proj');

        expect(result.length).toBe(2);
        expect(result[0].service).toBe('compute.googleapis.com');
        expect(result[0].serviceShortName).toBe('compute');
        expect(result[0].type).toBe('compute.googleapis.com/instance/cpu/test-metric-type-1');
        expect(result[0].displayName).toBe('test metric name 1');
        expect(result[0].description).toBe('A description');
        expect(result[1].type).toBe('logging.googleapis.com/user/logbased-metric-with-no-display-name');
        expect(result[1].displayName).toBe('logging.googleapis.com/user/logbased-metric-with-no-display-name');
      });
    });
  });

  describe('when interpolating a template variable for the filter', () => {
    describe('and is single value variable', () => {
      it('should replace the variable with the value', () => {
        const templateSrv = initTemplateSrv('filtervalue1');
        const { ds } = getTestcontext({ templateSrv });
        const interpolated = ds.interpolateFilters(['resource.label.zone', '=~', '${test}'], {});

        expect(interpolated.length).toBe(3);
        expect(interpolated[2]).toBe('filtervalue1');
      });
    });

    describe('and is single value variable for the label part', () => {
      it('should replace the variable with the value and not with regex formatting', () => {
        const templateSrv = initTemplateSrv('resource.label.zone');
        const { ds } = getTestcontext({ templateSrv });
        const interpolated = ds.interpolateFilters(['${test}', '=~', 'europe-north-1a'], {});

        expect(interpolated.length).toBe(3);
        expect(interpolated[0]).toBe('resource.label.zone');
      });
    });

    describe('and is multi value variable', () => {
      it('should replace the variable with a regex expression', () => {
        const templateSrv = initTemplateSrv(['filtervalue1', 'filtervalue2'], true);
        const { ds } = getTestcontext({ templateSrv });
        const interpolated = ds.interpolateFilters(['resource.label.zone', '=~', '[[test]]'], {});

        expect(interpolated[2]).toBe('(filtervalue1|filtervalue2)');
      });
    });
  });

  describe('when interpolating a template variable for group bys', () => {
    describe('and is single value variable', () => {
      it('should replace the variable with the value', () => {
        const templateSrv = initTemplateSrv('groupby1');
        const { ds } = getTestcontext({ templateSrv });
        const interpolated = ds.interpolateGroupBys(['[[test]]'], {});

        expect(interpolated.length).toBe(1);
        expect(interpolated[0]).toBe('groupby1');
      });
    });

    describe('and is multi value variable', () => {
      it('should replace the variable with an array of group bys', () => {
        const templateSrv = initTemplateSrv(['groupby1', 'groupby2'], true);
        const { ds } = getTestcontext({ templateSrv });
        const interpolated = ds.interpolateGroupBys(['[[test]]'], {});

        expect(interpolated.length).toBe(2);
        expect(interpolated[0]).toBe('groupby1');
        expect(interpolated[1]).toBe('groupby2');
      });
    });
  });
});

function initTemplateSrv(values: any, multi = false) {
  const templateSrv = new TemplateSrv();
  const test: CustomVariableModel = {
    ...initialCustomVariableModelState,
    id: 'test',
    name: 'test',
    current: { value: values, text: Array.isArray(values) ? values.toString() : values, selected: true },
    options: [{ value: values, text: Array.isArray(values) ? values.toString() : values, selected: false }],
    multi,
  };
  templateSrv.init([test]);
  return templateSrv;
}
