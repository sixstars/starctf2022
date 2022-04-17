import { cloneDeep, flattenDeep } from 'lodash';
import { lastValueFrom, Observable, of } from 'rxjs';
import { map, mergeMap } from 'rxjs/operators';
import {
  AnnotationEvent,
  AppEvents,
  CoreApp,
  DataQueryRequest,
  DataSourceApi,
  rangeUtil,
  ScopedVars,
} from '@grafana/data';
import { getBackendSrv, getDataSourceSrv } from '@grafana/runtime';

import coreModule from 'app/core/core_module';
import { dedupAnnotations } from './events_processing';
import { DashboardModel } from '../dashboard/state';
import { appEvents } from 'app/core/core';
import { getTimeSrv } from '../dashboard/services/TimeSrv';
import { AnnotationQueryOptions, AnnotationQueryResponse } from './types';
import { standardAnnotationSupport } from './standardAnnotationSupport';
import { runRequest } from '../query/state/runRequest';
import { RefreshEvent } from 'app/types/events';
import { deleteAnnotation, saveAnnotation, updateAnnotation } from './api';

let counter = 100;
function getNextRequestId() {
  return 'AQ' + counter++;
}

export class AnnotationsSrv {
  globalAnnotationsPromise: any;
  alertStatesPromise: any;
  datasourcePromises: any;

  init(dashboard: DashboardModel) {
    // always clearPromiseCaches when loading new dashboard
    this.clearPromiseCaches();
    // clear promises on refresh events
    dashboard.events.subscribe(RefreshEvent, this.clearPromiseCaches.bind(this));
  }

  clearPromiseCaches() {
    this.globalAnnotationsPromise = null;
    this.alertStatesPromise = null;
    this.datasourcePromises = null;
  }

  getAnnotations(options: AnnotationQueryOptions) {
    return Promise.all([this.getGlobalAnnotations(options), this.getAlertStates(options)])
      .then((results) => {
        // combine the annotations and flatten results
        let annotations: AnnotationEvent[] = flattenDeep(results[0]);
        // when in edit mode we need to use this function to get the saved ID
        let panelFilterId = options.panel.getSavedId();

        // filter out annotations that do not belong to requesting panel
        annotations = annotations.filter((item) => {
          // if event has panel ID and query is of type dashboard then panel and requesting panel ID must match
          if (item.panelId && item.source.type === 'dashboard') {
            return item.panelId === panelFilterId;
          }
          return true;
        });

        annotations = dedupAnnotations(annotations);

        // look for alert state for this panel
        const alertState: any = results[1].find((res: any) => res.panelId === panelFilterId);

        return {
          annotations: annotations,
          alertState: alertState,
        };
      })
      .catch((err) => {
        if (err.cancelled) {
          return [];
        }

        if (!err.message && err.data && err.data.message) {
          err.message = err.data.message;
        }

        console.error('AnnotationSrv.query error', err);
        appEvents.emit(AppEvents.alertError, ['Annotation Query Failed', err.message || err]);
        return [];
      });
  }

  getAlertStates(options: any) {
    if (!options.dashboard.id) {
      return Promise.resolve([]);
    }

    // ignore if no alerts
    if (options.panel && !options.panel.alert) {
      return Promise.resolve([]);
    }

    if (options.range.raw.to !== 'now') {
      return Promise.resolve([]);
    }

    if (this.alertStatesPromise) {
      return this.alertStatesPromise;
    }

    this.alertStatesPromise = getBackendSrv().get(
      '/api/alerts/states-for-dashboard',
      {
        dashboardId: options.dashboard.id,
      },
      `get-alert-states-${options.dashboard.id}`
    );

    return this.alertStatesPromise;
  }

  getGlobalAnnotations(options: AnnotationQueryOptions) {
    const dashboard = options.dashboard;

    if (this.globalAnnotationsPromise) {
      return this.globalAnnotationsPromise;
    }

    const range = getTimeSrv().timeRange();
    const promises = [];
    const dsPromises = [];

    for (const annotation of dashboard.annotations.list) {
      if (!annotation.enable) {
        continue;
      }

      if (annotation.snapshotData) {
        return this.translateQueryResult(annotation, annotation.snapshotData);
      }
      const datasourcePromise = getDataSourceSrv().get(annotation.datasource);
      dsPromises.push(datasourcePromise);
      promises.push(
        datasourcePromise
          .then((datasource: DataSourceApi) => {
            // Use the legacy annotationQuery unless annotation support is explicitly defined
            if (datasource.annotationQuery && !datasource.annotations) {
              return datasource.annotationQuery({
                range,
                rangeRaw: range.raw,
                annotation: annotation,
                dashboard: dashboard,
              });
            }
            // Note: future annotation lifecycle will use observables directly
            return lastValueFrom(executeAnnotationQuery(options, datasource, annotation)).then((res) => {
              return res.events ?? [];
            });
          })
          .then((results) => {
            // store response in annotation object if this is a snapshot call
            if (dashboard.snapshot) {
              annotation.snapshotData = cloneDeep(results);
            }
            // translate result
            return this.translateQueryResult(annotation, results);
          })
      );
    }
    this.datasourcePromises = Promise.all(dsPromises);
    this.globalAnnotationsPromise = Promise.all(promises);
    return this.globalAnnotationsPromise;
  }

  saveAnnotationEvent(annotation: AnnotationEvent) {
    this.globalAnnotationsPromise = null;
    return saveAnnotation(annotation);
  }

  updateAnnotationEvent(annotation: AnnotationEvent) {
    this.globalAnnotationsPromise = null;
    return updateAnnotation(annotation);
  }

  deleteAnnotationEvent(annotation: AnnotationEvent) {
    this.globalAnnotationsPromise = null;
    return deleteAnnotation(annotation);
  }

  translateQueryResult(annotation: any, results: any) {
    // if annotation has snapshotData
    // make clone and remove it
    if (annotation.snapshotData) {
      annotation = cloneDeep(annotation);
      delete annotation.snapshotData;
    }

    for (const item of results) {
      item.source = annotation;
      item.color = annotation.iconColor;
      item.type = annotation.name;
      item.isRegion = item.timeEnd && item.time !== item.timeEnd;
    }

    return results;
  }
}

export function executeAnnotationQuery(
  options: AnnotationQueryOptions,
  datasource: DataSourceApi,
  savedJsonAnno: any
): Observable<AnnotationQueryResponse> {
  const processor = {
    ...standardAnnotationSupport,
    ...datasource.annotations,
  };

  const annotation = processor.prepareAnnotation!(savedJsonAnno);
  if (!annotation) {
    return of({});
  }

  const query = processor.prepareQuery!(annotation);
  if (!query) {
    return of({});
  }

  // No more points than pixels
  const maxDataPoints = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;

  // Add interval to annotation queries
  const interval = rangeUtil.calculateInterval(options.range, maxDataPoints, datasource.interval);

  const scopedVars: ScopedVars = {
    __interval: { text: interval.interval, value: interval.interval },
    __interval_ms: { text: interval.intervalMs.toString(), value: interval.intervalMs },
    __annotation: { text: annotation.name, value: annotation },
  };

  const queryRequest: DataQueryRequest = {
    startTime: Date.now(),
    requestId: getNextRequestId(),
    range: options.range,
    maxDataPoints,
    scopedVars,
    ...interval,
    app: CoreApp.Dashboard,

    timezone: options.dashboard.timezone,

    targets: [
      {
        ...query,
        refId: 'Anno',
      },
    ],
  };

  return runRequest(datasource, queryRequest).pipe(
    mergeMap((panelData) => {
      if (!panelData.series) {
        return of({ panelData, events: [] });
      }

      return processor.processEvents!(annotation, panelData.series).pipe(map((events) => ({ panelData, events })));
    })
  );
}

coreModule.service('annotationsSrv', AnnotationsSrv);
