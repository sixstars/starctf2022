import { DataQueryRequest, DataSourceInstanceSettings, ScopedVars, MetricFindValue } from '@grafana/data';
import { getTemplateSrv, DataSourceWithBackend } from '@grafana/runtime';
import { isString } from 'lodash';

import TimegrainConverter from '../time_grain_converter';
import { AzureDataSourceJsonData, AzureMonitorQuery, AzureQueryType, DatasourceValidationResult } from '../types';
import { routeNames } from '../utils/common';
import ResponseParser from './response_parser';

export interface LogAnalyticsColumn {
  text: string;
  value: string;
}

export default class AppInsightsDatasource extends DataSourceWithBackend<AzureMonitorQuery, AzureDataSourceJsonData> {
  resourcePath: string;
  version = 'beta';
  applicationId: string;
  logAnalyticsColumns: { [key: string]: LogAnalyticsColumn[] } = {};

  constructor(instanceSettings: DataSourceInstanceSettings<AzureDataSourceJsonData>) {
    super(instanceSettings);
    this.applicationId = instanceSettings.jsonData.appInsightsAppId || '';

    this.resourcePath = `${routeNames.appInsights}/${this.version}/apps/${this.applicationId}`;
  }

  isConfigured(): boolean {
    return !!this.applicationId && this.applicationId.length > 0;
  }

  createRawQueryRequest(item: any, options: DataQueryRequest<AzureMonitorQuery>, target: AzureMonitorQuery) {
    if (item.xaxis && !item.timeColumn) {
      item.timeColumn = item.xaxis;
    }

    if (item.yaxis && !item.valueColumn) {
      item.valueColumn = item.yaxis;
    }

    if (item.spliton && !item.segmentColumn) {
      item.segmentColumn = item.spliton;
    }

    return {
      type: 'timeSeriesQuery',
      raw: false,
      appInsights: {
        rawQuery: true,
        rawQueryString: getTemplateSrv().replace(item.rawQueryString, options.scopedVars),
        timeColumn: item.timeColumn,
        valueColumn: item.valueColumn,
        segmentColumn: item.segmentColumn,
      },
    };
  }

  applyTemplateVariables(target: AzureMonitorQuery, scopedVars: ScopedVars): AzureMonitorQuery {
    const item = target.appInsights;

    if (!item) {
      return target;
    }

    const old: any = item;
    // fix for timeGrainUnit which is a deprecated/removed field name
    if (old.timeGrainCount) {
      item.timeGrain = TimegrainConverter.createISO8601Duration(old.timeGrainCount, item.timeGrainUnit);
    } else if (item.timeGrain && item.timeGrainUnit && item.timeGrain !== 'auto') {
      item.timeGrain = TimegrainConverter.createISO8601Duration(item.timeGrain, item.timeGrainUnit);
    }

    // migration for non-standard names
    if (old.groupBy && !item.dimension) {
      item.dimension = [old.groupBy];
    }
    if (old.filter && !item.dimensionFilter) {
      item.dimensionFilter = old.filter;
    }

    // Migrate single dimension string to array
    if (isString(item.dimension)) {
      if (item.dimension === 'None') {
        item.dimension = [];
      } else {
        item.dimension = [item.dimension as string];
      }
    }
    if (!item.dimension) {
      item.dimension = [];
    }

    const templateSrv = getTemplateSrv();

    return {
      refId: target.refId,
      queryType: AzureQueryType.ApplicationInsights,
      appInsights: {
        timeGrain: templateSrv.replace((item.timeGrain || '').toString(), scopedVars),
        metricName: templateSrv.replace(item.metricName, scopedVars),
        aggregation: templateSrv.replace(item.aggregation, scopedVars),
        dimension: item.dimension.map((d) => templateSrv.replace(d, scopedVars)),
        dimensionFilter: templateSrv.replace(item.dimensionFilter, scopedVars),
        alias: item.alias,
      },
    };
  }

  /**
   * This is named differently than DataSourceApi.metricFindQuery
   * because it's not exposed to Grafana like the main AzureMonitorDataSource.
   * And some of the azure internal data sources return null in this function, which the
   * external interface does not support
   */
  metricFindQueryInternal(query: string): Promise<MetricFindValue[]> | null {
    const appInsightsMetricNameQuery = query.match(/^AppInsightsMetricNames\(\)/i);
    if (appInsightsMetricNameQuery) {
      return this.getMetricNames();
    }

    const appInsightsGroupByQuery = query.match(/^AppInsightsGroupBys\(([^\)]+?)(,\s?([^,]+?))?\)/i);
    if (appInsightsGroupByQuery) {
      const metricName = appInsightsGroupByQuery[1];
      return this.getGroupBys(getTemplateSrv().replace(metricName));
    }

    return null;
  }

  testDatasource(): Promise<DatasourceValidationResult> {
    const path = `${this.resourcePath}/metrics/metadata`;
    return this.getResource(path)
      .then<DatasourceValidationResult>((response: any) => {
        return {
          status: 'success',
          message: 'Successfully queried the Application Insights service.',
          title: 'Success',
        };
      })
      .catch((error: any) => {
        let message = 'Application Insights: ';
        message += error.statusText ? error.statusText + ': ' : '';

        if (error.data && error.data.error && error.data.error.code === 'PathNotFoundError') {
          message += 'Invalid Application Id for Application Insights service.';
        } else if (error.data && error.data.error) {
          message += error.data.error.code + '. ' + error.data.error.message;
        } else {
          message += 'Cannot connect to Application Insights REST API.';
        }

        return {
          status: 'error',
          message: message,
        };
      });
  }

  getMetricNames() {
    const path = `${this.resourcePath}/metrics/metadata`;
    return this.getResource(path).then(ResponseParser.parseMetricNames);
  }

  getMetricMetadata(metricName: string) {
    const path = `${this.resourcePath}/metrics/metadata`;
    return this.getResource(path).then((result: any) => {
      return new ResponseParser(result).parseMetadata(metricName);
    });
  }

  getGroupBys(metricName: string) {
    return this.getMetricMetadata(metricName).then((result: any) => {
      return new ResponseParser(result).parseGroupBys();
    });
  }

  getQuerySchema() {
    const path = `${this.resourcePath}/query/schema`;
    return this.getResource(path).then((result: any) => {
      const schema = new ResponseParser(result).parseQuerySchema();
      // console.log(schema);
      return schema;
    });
  }
}
