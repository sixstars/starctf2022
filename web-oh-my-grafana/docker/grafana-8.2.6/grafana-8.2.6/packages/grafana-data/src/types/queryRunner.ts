import { Observable } from 'rxjs';
import { DataQuery, DatasourceRef } from './query';
import { DataSourceApi } from './datasource';
import { PanelData } from './panel';
import { ScopedVars } from './ScopedVars';
import { TimeRange, TimeZone } from './time';

/**
 * Describes the options used when triggering a query via the {@link QueryRunner}.
 *
 * @internal
 */
export interface QueryRunnerOptions {
  datasource: DatasourceRef | DataSourceApi | null;
  queries: DataQuery[];
  panelId?: number;
  dashboardId?: number;
  timezone: TimeZone;
  timeRange: TimeRange;
  timeInfo?: string; // String description of time range for display
  maxDataPoints: number;
  minInterval: string | undefined | null;
  scopedVars?: ScopedVars;
  cacheTimeout?: string;
  app?: string;
}

/**
 * Describes the QueryRunner that can used to exectue queries in e.g. app plugins.
 * QueryRunner instances can be created via the {@link @grafana/runtime#createQueryRunner | createQueryRunner}.
 *
 * @internal
 */
export interface QueryRunner {
  get(): Observable<PanelData>;
  run(options: QueryRunnerOptions): void;
  cancel(): void;
  destroy(): void;
}
