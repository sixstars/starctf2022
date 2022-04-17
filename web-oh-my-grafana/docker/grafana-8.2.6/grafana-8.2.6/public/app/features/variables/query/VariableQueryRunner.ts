import { merge, Observable, of, Subject, throwError, Unsubscribable } from 'rxjs';
import { catchError, filter, finalize, first, mergeMap, takeUntil } from 'rxjs/operators';
import {
  CoreApp,
  DataQuery,
  DataQueryRequest,
  DataSourceApi,
  getDefaultTimeRange,
  LoadingState,
  ScopedVars,
} from '@grafana/data';

import { VariableIdentifier } from '../state/types';
import { getVariable } from '../state/selectors';
import { QueryVariableModel, VariableRefresh } from '../types';
import { StoreState, ThunkDispatch } from '../../../types';
import { dispatch, getState } from '../../../store/store';
import { getTemplatedRegex } from '../utils';
import { v4 as uuidv4 } from 'uuid';
import { getTimeSrv } from '../../dashboard/services/TimeSrv';
import { QueryRunners } from './queryRunners';
import { runRequest } from '../../query/state/runRequest';
import { toMetricFindValues, updateOptionsState, validateVariableSelection } from './operators';

interface UpdateOptionsArgs {
  identifier: VariableIdentifier;
  datasource: DataSourceApi;
  searchFilter?: string;
}

export interface UpdateOptionsResults {
  state: LoadingState;
  identifier: VariableIdentifier;
  error?: any;
  cancelled?: boolean;
}

interface VariableQueryRunnerArgs {
  dispatch: ThunkDispatch;
  getState: () => StoreState;
  getVariable: typeof getVariable;
  getTemplatedRegex: typeof getTemplatedRegex;
  getTimeSrv: typeof getTimeSrv;
  queryRunners: QueryRunners;
  runRequest: typeof runRequest;
}

export class VariableQueryRunner {
  private readonly updateOptionsRequests: Subject<UpdateOptionsArgs>;
  private readonly updateOptionsResults: Subject<UpdateOptionsResults>;
  private readonly cancelRequests: Subject<{ identifier: VariableIdentifier }>;
  private readonly subscription: Unsubscribable;

  constructor(
    private dependencies: VariableQueryRunnerArgs = {
      dispatch,
      getState,
      getVariable,
      getTemplatedRegex,
      getTimeSrv,
      queryRunners: new QueryRunners(),
      runRequest,
    }
  ) {
    this.updateOptionsRequests = new Subject<UpdateOptionsArgs>();
    this.updateOptionsResults = new Subject<UpdateOptionsResults>();
    this.cancelRequests = new Subject<{ identifier: VariableIdentifier }>();
    this.onNewRequest = this.onNewRequest.bind(this);
    this.subscription = this.updateOptionsRequests.subscribe(this.onNewRequest);
  }

  queueRequest(args: UpdateOptionsArgs): void {
    this.updateOptionsRequests.next(args);
  }

  getResponse(identifier: VariableIdentifier): Observable<UpdateOptionsResults> {
    return this.updateOptionsResults.asObservable().pipe(filter((result) => result.identifier === identifier));
  }

  cancelRequest(identifier: VariableIdentifier): void {
    this.cancelRequests.next({ identifier });
  }

  destroy(): void {
    this.subscription.unsubscribe();
  }

  private onNewRequest(args: UpdateOptionsArgs): void {
    const { datasource, identifier, searchFilter } = args;
    try {
      const {
        dispatch,
        runRequest,
        getTemplatedRegex: getTemplatedRegexFunc,
        getVariable,
        queryRunners,
        getTimeSrv,
        getState,
      } = this.dependencies;

      const beforeUid = getState().templating.transaction.uid;

      this.updateOptionsResults.next({ identifier, state: LoadingState.Loading });

      const variable = getVariable<QueryVariableModel>(identifier.id, getState());
      const timeSrv = getTimeSrv();
      const runnerArgs = { variable, datasource, searchFilter, timeSrv, runRequest };
      const runner = queryRunners.getRunnerForDatasource(datasource);
      const target = runner.getTarget({ datasource, variable });
      const request = this.getRequest(variable, args, target);

      runner
        .runRequest(runnerArgs, request)
        .pipe(
          filter(() => {
            // Lets check if we started another batch during the execution of the observable. If so we just want to abort the rest.
            const afterUid = getState().templating.transaction.uid;
            return beforeUid === afterUid;
          }),
          first((data) => data.state === LoadingState.Done || data.state === LoadingState.Error),
          mergeMap((data) => {
            if (data.state === LoadingState.Error) {
              return throwError(data.error);
            }

            return of(data);
          }),
          toMetricFindValues(),
          updateOptionsState({ variable, dispatch, getTemplatedRegexFunc }),
          validateVariableSelection({ variable, dispatch, searchFilter }),
          takeUntil(
            merge(this.updateOptionsRequests, this.cancelRequests).pipe(
              filter((args) => {
                let cancelRequest = false;

                if (args.identifier.id === identifier.id) {
                  cancelRequest = true;
                  this.updateOptionsResults.next({ identifier, state: LoadingState.Loading, cancelled: cancelRequest });
                }

                return cancelRequest;
              })
            )
          ),
          catchError((error) => {
            if (error.cancelled) {
              return of({});
            }

            this.updateOptionsResults.next({ identifier, state: LoadingState.Error, error });
            return throwError(error);
          }),
          finalize(() => {
            this.updateOptionsResults.next({ identifier, state: LoadingState.Done });
          })
        )
        .subscribe();
    } catch (error) {
      this.updateOptionsResults.next({ identifier, state: LoadingState.Error, error });
    }
  }

  private getRequest(variable: QueryVariableModel, args: UpdateOptionsArgs, target: DataQuery) {
    const { searchFilter } = args;
    const variableAsVars = { variable: { text: variable.current.text, value: variable.current.value } };
    const searchFilterScope = { searchFilter: { text: searchFilter, value: searchFilter } };
    const searchFilterAsVars = searchFilter ? searchFilterScope : {};
    const scopedVars = { ...searchFilterAsVars, ...variableAsVars } as ScopedVars;
    const range =
      variable.refresh === VariableRefresh.onTimeRangeChanged
        ? this.dependencies.getTimeSrv().timeRange()
        : getDefaultTimeRange();

    const request: DataQueryRequest = {
      app: CoreApp.Dashboard,
      requestId: uuidv4(),
      timezone: '',
      range,
      interval: '',
      intervalMs: 0,
      targets: [target],
      scopedVars,
      startTime: Date.now(),
    };

    return request;
  }
}

let singleton: VariableQueryRunner;

export function setVariableQueryRunner(runner: VariableQueryRunner): void {
  singleton = runner;
}

export function getVariableQueryRunner(): VariableQueryRunner {
  return singleton;
}
