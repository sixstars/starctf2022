import AzureMonitorDatasource from '../datasource';

import { TemplateSrv } from 'app/features/templating/template_srv';
import { DataSourceInstanceSettings } from '@grafana/data';
import { AzureDataSourceJsonData, DatasourceValidationResult } from '../types';

const templateSrv = new TemplateSrv();

jest.mock('@grafana/runtime', () => ({
  ...((jest.requireActual('@grafana/runtime') as unknown) as object),
  getTemplateSrv: () => templateSrv,
}));

interface TestContext {
  instanceSettings: DataSourceInstanceSettings<AzureDataSourceJsonData>;
  ds: AzureMonitorDatasource;
}

describe('AzureMonitorDatasource', () => {
  const ctx: TestContext = {} as TestContext;

  beforeEach(() => {
    jest.clearAllMocks();
    ctx.instanceSettings = ({
      name: 'test',
      url: 'http://azuremonitor.com',
      jsonData: { subscriptionId: '9935389e-9122-4ef9-95f9-1513dd24753f', cloudName: 'azuremonitor' },
    } as unknown) as DataSourceInstanceSettings<AzureDataSourceJsonData>;
    ctx.ds = new AzureMonitorDatasource(ctx.instanceSettings);
  });

  describe('When performing testDatasource', () => {
    describe('and an error is returned', () => {
      const error = {
        data: {
          error: {
            code: 'InvalidApiVersionParameter',
            message: `An error message.`,
          },
        },
        status: 400,
        statusText: 'Bad Request',
      };

      beforeEach(() => {
        ctx.instanceSettings.jsonData.azureAuthType = 'msi';
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockRejectedValue(error);
      });

      it('should return error status and a detailed error message', () => {
        return ctx.ds.azureMonitorDatasource.testDatasource().then((result: DatasourceValidationResult) => {
          expect(result.status).toEqual('error');
          expect(result.message).toEqual('Azure Monitor: Bad Request: InvalidApiVersionParameter. An error message.');
        });
      });
    });

    describe('and a list of resource groups is returned', () => {
      const response = {
        value: [{ name: 'grp1' }, { name: 'grp2' }],
      };

      beforeEach(() => {
        ctx.instanceSettings.jsonData.tenantId = 'xxx';
        ctx.instanceSettings.jsonData.clientId = 'xxx';
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockResolvedValue({ data: response, status: 200 });
      });

      it('should return success status', () => {
        return ctx.ds.azureMonitorDatasource.testDatasource().then((result: DatasourceValidationResult) => {
          expect(result.status).toEqual('success');
        });
      });
    });
  });

  describe('When performing metricFindQuery', () => {
    describe('with a subscriptions query', () => {
      const response = {
        value: [
          { displayName: 'Primary', subscriptionId: 'sub1' },
          { displayName: 'Secondary', subscriptionId: 'sub2' },
        ],
      };

      beforeEach(() => {
        ctx.instanceSettings.jsonData.azureAuthType = 'msi';
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockResolvedValue(response);
      });

      it('should return a list of subscriptions', async () => {
        const results = await ctx.ds.metricFindQuery('subscriptions()');
        expect(results.length).toBe(2);
        expect(results[0].text).toBe('Primary');
        expect(results[0].value).toBe('sub1');
        expect(results[1].text).toBe('Secondary');
        expect(results[1].value).toBe('sub2');
      });
    });

    describe('with a resource groups query', () => {
      const response = {
        value: [{ name: 'grp1' }, { name: 'grp2' }],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockResolvedValue(response);
      });

      it('should return a list of resource groups', async () => {
        const results = await ctx.ds.metricFindQuery('ResourceGroups()');
        expect(results.length).toBe(2);
        expect(results[0].text).toBe('grp1');
        expect(results[0].value).toBe('grp1');
        expect(results[1].text).toBe('grp2');
        expect(results[1].value).toBe('grp2');
      });
    });

    describe('with a resource groups query that specifies a subscription id', () => {
      const response = {
        value: [{ name: 'grp1' }, { name: 'grp2' }],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          expect(path).toContain('11112222-eeee-4949-9b2d-9106972f9123');
          return Promise.resolve(response);
        });
      });

      it('should return a list of resource groups', async () => {
        const results = await ctx.ds.metricFindQuery('ResourceGroups(11112222-eeee-4949-9b2d-9106972f9123)');
        expect(results.length).toBe(2);
        expect(results[0].text).toBe('grp1');
        expect(results[0].value).toBe('grp1');
        expect(results[1].text).toBe('grp2');
        expect(results[1].value).toBe('grp2');
      });
    });

    describe('with namespaces query', () => {
      const response = {
        value: [
          {
            name: 'test',
            type: 'Microsoft.Network/networkInterfaces',
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
          expect(path).toBe(basePath + '/nodesapp/resources?api-version=2018-01-01');
          return Promise.resolve(response);
        });
      });

      it('should return a list of namespaces', async () => {
        const results = await ctx.ds.metricFindQuery('Namespaces(nodesapp)');
        expect(results.length).toEqual(1);
        expect(results[0].text).toEqual('Network interface');
        expect(results[0].value).toEqual('Microsoft.Network/networkInterfaces');
      });
    });

    describe('with namespaces query that specifies a subscription id', () => {
      const response = {
        value: [
          {
            name: 'test',
            type: 'Microsoft.Network/networkInterfaces',
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/11112222-eeee-4949-9b2d-9106972f9123/resourceGroups';
          expect(path).toBe(basePath + '/nodesapp/resources?api-version=2018-01-01');
          return Promise.resolve(response);
        });
      });

      it('should return a list of namespaces', async () => {
        const results = await ctx.ds.metricFindQuery('namespaces(11112222-eeee-4949-9b2d-9106972f9123, nodesapp)');
        expect(results.length).toEqual(1);
        expect(results[0].text).toEqual('Network interface');
        expect(results[0].value).toEqual('Microsoft.Network/networkInterfaces');
      });
    });

    describe('with resource names query', () => {
      const response = {
        value: [
          {
            name: 'Failure Anomalies - nodeapp',
            type: 'microsoft.insights/alertrules',
          },
          {
            name: 'nodeapp',
            type: 'microsoft.insights/components',
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
          expect(path).toBe(basePath + '/nodeapp/resources?api-version=2018-01-01');
          return Promise.resolve(response);
        });
      });

      it('should return a list of resource names', async () => {
        const results = await ctx.ds.metricFindQuery('resourceNames(nodeapp, microsoft.insights/components )');
        expect(results.length).toEqual(1);
        expect(results[0].text).toEqual('nodeapp');
        expect(results[0].value).toEqual('nodeapp');
      });
    });

    describe('with resource names query and that specifies a subscription id', () => {
      const response = {
        value: [
          {
            name: 'Failure Anomalies - nodeapp',
            type: 'microsoft.insights/alertrules',
          },
          {
            name: 'nodeapp',
            type: 'microsoft.insights/components',
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/11112222-eeee-4949-9b2d-9106972f9123/resourceGroups';
          expect(path).toBe(basePath + '/nodeapp/resources?api-version=2018-01-01');
          return Promise.resolve(response);
        });
      });

      it('should return a list of resource names', () => {
        return ctx.ds
          .metricFindQuery(
            'resourceNames(11112222-eeee-4949-9b2d-9106972f9123, nodeapp, microsoft.insights/components )'
          )
          .then((results: any) => {
            expect(results.length).toEqual(1);
            expect(results[0].text).toEqual('nodeapp');
            expect(results[0].value).toEqual('nodeapp');
          });
      });
    });

    describe('with metric names query', () => {
      const response = {
        value: [
          {
            name: {
              value: 'Percentage CPU',
              localizedValue: 'Percentage CPU',
            },
          },
          {
            name: {
              value: 'UsedCapacity',
              localizedValue: 'Used capacity',
            },
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
          expect(path).toBe(
            basePath +
              '/nodeapp/providers/microsoft.insights/components/rn/providers/microsoft.insights/' +
              'metricdefinitions?api-version=2018-01-01&metricnamespace=default'
          );
          return Promise.resolve(response);
        });
      });

      it('should return a list of metric names', async () => {
        const results = await ctx.ds.metricFindQuery(
          'Metricnames(nodeapp, microsoft.insights/components, rn, default)'
        );
        expect(results.length).toEqual(2);
        expect(results[0].text).toEqual('Percentage CPU');
        expect(results[0].value).toEqual('Percentage CPU');

        expect(results[1].text).toEqual('Used capacity');
        expect(results[1].value).toEqual('UsedCapacity');
      });
    });

    describe('with metric names query and specifies a subscription id', () => {
      const response = {
        value: [
          {
            name: {
              value: 'Percentage CPU',
              localizedValue: 'Percentage CPU',
            },
          },
          {
            name: {
              value: 'UsedCapacity',
              localizedValue: 'Used capacity',
            },
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/11112222-eeee-4949-9b2d-9106972f9123/resourceGroups';
          expect(path).toBe(
            basePath +
              '/nodeapp/providers/microsoft.insights/components/rn/providers/microsoft.insights/' +
              'metricdefinitions?api-version=2018-01-01&metricnamespace=default'
          );
          return Promise.resolve(response);
        });
      });

      it('should return a list of metric names', async () => {
        const results = await ctx.ds.metricFindQuery(
          'Metricnames(11112222-eeee-4949-9b2d-9106972f9123, nodeapp, microsoft.insights/components, rn, default)'
        );
        expect(results.length).toEqual(2);
        expect(results[0].text).toEqual('Percentage CPU');
        expect(results[0].value).toEqual('Percentage CPU');

        expect(results[1].text).toEqual('Used capacity');
        expect(results[1].value).toEqual('UsedCapacity');
      });
    });

    describe('with metric namespace query', () => {
      const response = {
        value: [
          {
            name: 'Microsoft.Compute-virtualMachines',
            properties: {
              metricNamespaceName: 'Microsoft.Compute/virtualMachines',
            },
          },
          {
            name: 'Telegraf-mem',
            properties: {
              metricNamespaceName: 'Telegraf/mem',
            },
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
          expect(path).toBe(
            basePath +
              '/nodeapp/providers/Microsoft.Compute/virtualMachines/rn/providers/microsoft.insights/metricNamespaces?api-version=2017-12-01-preview'
          );
          return Promise.resolve(response);
        });
      });

      it('should return a list of metric names', async () => {
        const results = await ctx.ds.metricFindQuery('Metricnamespace(nodeapp, Microsoft.Compute/virtualMachines, rn)');
        expect(results.length).toEqual(2);
        expect(results[0].text).toEqual('Microsoft.Compute-virtualMachines');
        expect(results[0].value).toEqual('Microsoft.Compute/virtualMachines');

        expect(results[1].text).toEqual('Telegraf-mem');
        expect(results[1].value).toEqual('Telegraf/mem');
      });
    });

    describe('with metric namespace query and specifies a subscription id', () => {
      const response = {
        value: [
          {
            name: 'Microsoft.Compute-virtualMachines',
            properties: {
              metricNamespaceName: 'Microsoft.Compute/virtualMachines',
            },
          },
          {
            name: 'Telegraf-mem',
            properties: {
              metricNamespaceName: 'Telegraf/mem',
            },
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/11112222-eeee-4949-9b2d-9106972f9123/resourceGroups';
          expect(path).toBe(
            basePath +
              '/nodeapp/providers/Microsoft.Compute/virtualMachines/rn/providers/microsoft.insights/metricNamespaces?api-version=2017-12-01-preview'
          );
          return Promise.resolve(response);
        });
      });

      it('should return a list of metric namespaces', async () => {
        const results = await ctx.ds.metricFindQuery(
          'Metricnamespace(11112222-eeee-4949-9b2d-9106972f9123, nodeapp, Microsoft.Compute/virtualMachines, rn)'
        );
        expect(results.length).toEqual(2);
        expect(results[0].text).toEqual('Microsoft.Compute-virtualMachines');
        expect(results[0].value).toEqual('Microsoft.Compute/virtualMachines');

        expect(results[1].text).toEqual('Telegraf-mem');
        expect(results[1].value).toEqual('Telegraf/mem');
      });
    });
  });

  describe('When performing getSubscriptions', () => {
    const response = {
      value: [
        {
          id: '/subscriptions/99999999-cccc-bbbb-aaaa-9106972f9572',
          subscriptionId: '99999999-cccc-bbbb-aaaa-9106972f9572',
          tenantId: '99999999-aaaa-bbbb-cccc-51c4f982ec48',
          displayName: 'Primary Subscription',
          state: 'Enabled',
          subscriptionPolicies: {
            locationPlacementId: 'Public_2014-09-01',
            quotaId: 'PayAsYouGo_2014-09-01',
            spendingLimit: 'Off',
          },
          authorizationSource: 'RoleBased',
        },
      ],
      count: {
        type: 'Total',
        value: 1,
      },
    };

    beforeEach(() => {
      ctx.instanceSettings.jsonData.azureAuthType = 'msi';
      ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockResolvedValue(response);
    });

    it('should return list of subscriptions', () => {
      return ctx.ds.getSubscriptions().then((results: Array<{ text: string; value: string }>) => {
        expect(results.length).toEqual(1);
        expect(results[0].text).toEqual('Primary Subscription');
        expect(results[0].value).toEqual('99999999-cccc-bbbb-aaaa-9106972f9572');
      });
    });
  });

  describe('When performing getResourceGroups', () => {
    const response = {
      value: [{ name: 'grp1' }, { name: 'grp2' }],
    };

    beforeEach(() => {
      ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockResolvedValue(response);
    });

    it('should return list of Resource Groups', () => {
      return ctx.ds.getResourceGroups('subscriptionId').then((results: Array<{ text: string; value: string }>) => {
        expect(results.length).toEqual(2);
        expect(results[0].text).toEqual('grp1');
        expect(results[0].value).toEqual('grp1');
        expect(results[1].text).toEqual('grp2');
        expect(results[1].value).toEqual('grp2');
      });
    });
  });

  describe('When performing getMetricDefinitions', () => {
    const response = {
      value: [
        {
          name: 'test',
          type: 'Microsoft.Network/networkInterfaces',
        },
        {
          location: 'northeurope',
          name: 'northeur',
          type: 'Microsoft.Compute/virtualMachines',
        },
        {
          location: 'westcentralus',
          name: 'us',
          type: 'Microsoft.Compute/virtualMachines',
        },
        {
          name: 'IHaveNoMetrics',
          type: 'IShouldBeFilteredOut',
        },
        {
          name: 'storageTest',
          type: 'Microsoft.Storage/storageAccounts',
        },
      ],
    };

    beforeEach(() => {
      ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
        const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
        expect(path).toBe(basePath + '/nodesapp/resources?api-version=2018-01-01');
        return Promise.resolve(response);
      });
    });

    it('should return list of Metric Definitions with no duplicates and no unsupported namespaces', () => {
      return ctx.ds
        .getMetricDefinitions('9935389e-9122-4ef9-95f9-1513dd24753f', 'nodesapp')
        .then((results: Array<{ text: string; value: string }>) => {
          expect(results.length).toEqual(7);
          expect(results[0].text).toEqual('Network interface');
          expect(results[0].value).toEqual('Microsoft.Network/networkInterfaces');
          expect(results[1].text).toEqual('Virtual machine');
          expect(results[1].value).toEqual('Microsoft.Compute/virtualMachines');
          expect(results[2].text).toEqual('Storage account');
          expect(results[2].value).toEqual('Microsoft.Storage/storageAccounts');
          expect(results[3].text).toEqual('Microsoft.Storage/storageAccounts/blobServices');
          expect(results[3].value).toEqual('Microsoft.Storage/storageAccounts/blobServices');
          expect(results[4].text).toEqual('Microsoft.Storage/storageAccounts/fileServices');
          expect(results[4].value).toEqual('Microsoft.Storage/storageAccounts/fileServices');
          expect(results[5].text).toEqual('Microsoft.Storage/storageAccounts/tableServices');
          expect(results[5].value).toEqual('Microsoft.Storage/storageAccounts/tableServices');
          expect(results[6].text).toEqual('Microsoft.Storage/storageAccounts/queueServices');
          expect(results[6].value).toEqual('Microsoft.Storage/storageAccounts/queueServices');
        });
    });
  });

  describe('When performing getResourceNames', () => {
    describe('and there are no special cases', () => {
      const response = {
        value: [
          {
            name: 'Failure Anomalies - nodeapp',
            type: 'microsoft.insights/alertrules',
          },
          {
            name: 'nodeapp',
            type: 'microsoft.insights/components',
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
          expect(path).toBe(basePath + '/nodeapp/resources?api-version=2018-01-01');
          return Promise.resolve(response);
        });
      });

      it('should return list of Resource Names', () => {
        return ctx.ds
          .getResourceNames('9935389e-9122-4ef9-95f9-1513dd24753f', 'nodeapp', 'microsoft.insights/components')
          .then((results: Array<{ text: string; value: string }>) => {
            expect(results.length).toEqual(1);
            expect(results[0].text).toEqual('nodeapp');
            expect(results[0].value).toEqual('nodeapp');
          });
      });

      it('should return ignore letter case', () => {
        return ctx.ds
          .getResourceNames('9935389e-9122-4ef9-95f9-1513dd24753f', 'nodeapp', 'microsoft.insights/Components')
          .then((results: Array<{ text: string; value: string }>) => {
            expect(results.length).toEqual(1);
            expect(results[0].text).toEqual('nodeapp');
            expect(results[0].value).toEqual('nodeapp');
          });
      });
    });

    describe('and the metric definition is blobServices', () => {
      const response = {
        value: [
          {
            name: 'Failure Anomalies - nodeapp',
            type: 'microsoft.insights/alertrules',
          },
          {
            name: 'storagetest',
            type: 'Microsoft.Storage/storageAccounts',
          },
        ],
      };

      beforeEach(() => {
        ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
          const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups';
          expect(path).toBe(basePath + '/nodeapp/resources?api-version=2018-01-01');
          return Promise.resolve(response);
        });
      });

      it('should return list of Resource Names', () => {
        return ctx.ds
          .getResourceNames(
            '9935389e-9122-4ef9-95f9-1513dd24753f',
            'nodeapp',
            'Microsoft.Storage/storageAccounts/blobServices'
          )
          .then((results: Array<{ text: string; value: string }>) => {
            expect(results.length).toEqual(1);
            expect(results[0].text).toEqual('storagetest/default');
            expect(results[0].value).toEqual('storagetest/default');
          });
      });
    });
  });

  describe('When performing getMetricNames', () => {
    const response = {
      value: [
        {
          name: {
            value: 'UsedCapacity',
            localizedValue: 'Used capacity',
          },
          unit: 'CountPerSecond',
          primaryAggregationType: 'Total',
          supportedAggregationTypes: ['None', 'Average', 'Minimum', 'Maximum', 'Total', 'Count'],
          metricAvailabilities: [
            { timeGrain: 'PT1H', retention: 'P93D' },
            { timeGrain: 'PT6H', retention: 'P93D' },
            { timeGrain: 'PT12H', retention: 'P93D' },
            { timeGrain: 'P1D', retention: 'P93D' },
          ],
        },
        {
          name: {
            value: 'FreeCapacity',
            localizedValue: 'Free capacity',
          },
          unit: 'CountPerSecond',
          primaryAggregationType: 'Average',
          supportedAggregationTypes: ['None', 'Average'],
          metricAvailabilities: [
            { timeGrain: 'PT1H', retention: 'P93D' },
            { timeGrain: 'PT6H', retention: 'P93D' },
            { timeGrain: 'PT12H', retention: 'P93D' },
            { timeGrain: 'P1D', retention: 'P93D' },
          ],
        },
      ],
    };

    beforeEach(() => {
      ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
        const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups/nodeapp';
        const expected =
          basePath +
          '/providers/microsoft.insights/components/resource1' +
          '/providers/microsoft.insights/metricdefinitions?api-version=2018-01-01&metricnamespace=default';
        expect(path).toBe(expected);
        return Promise.resolve(response);
      });
    });

    it('should return list of Metric Definitions', () => {
      return ctx.ds
        .getMetricNames(
          '9935389e-9122-4ef9-95f9-1513dd24753f',
          'nodeapp',
          'microsoft.insights/components',
          'resource1',
          'default'
        )
        .then((results: Array<{ text: string; value: string }>) => {
          expect(results.length).toEqual(2);
          expect(results[0].text).toEqual('Used capacity');
          expect(results[0].value).toEqual('UsedCapacity');
          expect(results[1].text).toEqual('Free capacity');
          expect(results[1].value).toEqual('FreeCapacity');
        });
    });
  });

  describe('When performing getMetricMetadata', () => {
    const response = {
      value: [
        {
          name: {
            value: 'UsedCapacity',
            localizedValue: 'Used capacity',
          },
          unit: 'CountPerSecond',
          primaryAggregationType: 'Total',
          supportedAggregationTypes: ['None', 'Average', 'Minimum', 'Maximum', 'Total', 'Count'],
          metricAvailabilities: [
            { timeGrain: 'PT1H', retention: 'P93D' },
            { timeGrain: 'PT6H', retention: 'P93D' },
            { timeGrain: 'PT12H', retention: 'P93D' },
            { timeGrain: 'P1D', retention: 'P93D' },
          ],
        },
        {
          name: {
            value: 'FreeCapacity',
            localizedValue: 'Free capacity',
          },
          unit: 'CountPerSecond',
          primaryAggregationType: 'Average',
          supportedAggregationTypes: ['None', 'Average'],
          metricAvailabilities: [
            { timeGrain: 'PT1H', retention: 'P93D' },
            { timeGrain: 'PT6H', retention: 'P93D' },
            { timeGrain: 'PT12H', retention: 'P93D' },
            { timeGrain: 'P1D', retention: 'P93D' },
          ],
        },
      ],
    };

    beforeEach(() => {
      ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
        const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups/nodeapp';
        const expected =
          basePath +
          '/providers/microsoft.insights/components/resource1' +
          '/providers/microsoft.insights/metricdefinitions?api-version=2018-01-01&metricnamespace=default';
        expect(path).toBe(expected);
        return Promise.resolve(response);
      });
    });

    it('should return Aggregation metadata for a Metric', () => {
      return ctx.ds
        .getMetricMetadata(
          '9935389e-9122-4ef9-95f9-1513dd24753f',
          'nodeapp',
          'microsoft.insights/components',
          'resource1',
          'default',
          'UsedCapacity'
        )
        .then((results) => {
          expect(results.primaryAggType).toEqual('Total');
          expect(results.supportedAggTypes.length).toEqual(6);
          expect(results.supportedTimeGrains.length).toEqual(5); // 4 time grains from the API + auto
        });
    });
  });

  describe('When performing getMetricMetadata on metrics with dimensions', () => {
    const response = {
      value: [
        {
          name: {
            value: 'Transactions',
            localizedValue: 'Transactions',
          },
          unit: 'Count',
          primaryAggregationType: 'Total',
          supportedAggregationTypes: ['None', 'Average', 'Minimum', 'Maximum', 'Total', 'Count'],
          isDimensionRequired: false,
          dimensions: [
            {
              value: 'ResponseType',
              localizedValue: 'Response type',
            },
            {
              value: 'GeoType',
              localizedValue: 'Geo type',
            },
            {
              value: 'ApiName',
              localizedValue: 'API name',
            },
          ],
        },
        {
          name: {
            value: 'FreeCapacity',
            localizedValue: 'Free capacity',
          },
          unit: 'CountPerSecond',
          primaryAggregationType: 'Average',
          supportedAggregationTypes: ['None', 'Average'],
        },
      ],
    };

    beforeEach(() => {
      ctx.ds.azureMonitorDatasource.getResource = jest.fn().mockImplementation((path: string) => {
        const basePath = 'azuremonitor/subscriptions/9935389e-9122-4ef9-95f9-1513dd24753f/resourceGroups/nodeapp';
        const expected =
          basePath +
          '/providers/microsoft.insights/components/resource1' +
          '/providers/microsoft.insights/metricdefinitions?api-version=2018-01-01&metricnamespace=default';
        expect(path).toBe(expected);
        return Promise.resolve(response);
      });
    });

    it('should return dimensions for a Metric that has dimensions', () => {
      return ctx.ds
        .getMetricMetadata(
          '9935389e-9122-4ef9-95f9-1513dd24753f',
          'nodeapp',
          'microsoft.insights/components',
          'resource1',
          'default',
          'Transactions'
        )
        .then((results: any) => {
          expect(results.dimensions).toMatchInlineSnapshot(`
            Array [
              Object {
                "label": "Response type",
                "value": "ResponseType",
              },
              Object {
                "label": "Geo type",
                "value": "GeoType",
              },
              Object {
                "label": "API name",
                "value": "ApiName",
              },
            ]
          `);
        });
    });

    it('should return an empty array for a Metric that does not have dimensions', () => {
      return ctx.ds
        .getMetricMetadata(
          '9935389e-9122-4ef9-95f9-1513dd24753f',
          'nodeapp',
          'microsoft.insights/components',
          'resource1',
          'default',
          'FreeCapacity'
        )
        .then((results: any) => {
          expect(results.dimensions.length).toEqual(0);
        });
    });
  });
});
