import { cloneDeep, extend, isString } from 'lodash';
import {
  dateMath,
  dateTime,
  getDefaultTimeRange,
  isDateTime,
  rangeUtil,
  RawTimeRange,
  TimeRange,
  toUtc,
} from '@grafana/data';
import { DashboardModel } from '../state/DashboardModel';
import { getShiftedTimeRange, getZoomedTimeRange } from 'app/core/utils/timePicker';
import { config } from 'app/core/config';
import { getRefreshFromUrl } from '../utils/getRefreshFromUrl';
import { locationService } from '@grafana/runtime';
import { ShiftTimeEvent, ShiftTimeEventPayload, ZoomOutEvent } from '../../../types/events';
import { contextSrv, ContextSrv } from 'app/core/services/context_srv';
import appEvents from 'app/core/app_events';

export class TimeSrv {
  time: any;
  refreshTimer: any;
  refresh: any;
  previousAutoRefresh: any;
  oldRefresh: string | null | undefined;
  dashboard?: DashboardModel;
  timeAtLoad: any;
  private autoRefreshBlocked?: boolean;

  constructor(private contextSrv: ContextSrv) {
    // default time
    this.time = getDefaultTimeRange().raw;
    this.refreshDashboard = this.refreshDashboard.bind(this);

    appEvents.subscribe(ZoomOutEvent, (e) => {
      this.zoomOut(e.payload);
    });

    appEvents.subscribe(ShiftTimeEvent, (e) => {
      this.shiftTime(e.payload);
    });

    document.addEventListener('visibilitychange', () => {
      if (this.autoRefreshBlocked && document.visibilityState === 'visible') {
        this.autoRefreshBlocked = false;
        this.refreshDashboard();
      }
    });
  }

  init(dashboard: DashboardModel) {
    this.dashboard = dashboard;
    this.time = dashboard.time;
    this.refresh = dashboard.refresh;

    this.initTimeFromUrl();
    this.parseTime();

    // remember time at load so we can go back to it
    this.timeAtLoad = cloneDeep(this.time);

    const range = rangeUtil.convertRawToRange(
      this.time,
      this.dashboard?.getTimezone(),
      this.dashboard?.fiscalYearStartMonth
    );

    if (range.to.isBefore(range.from)) {
      this.setTime(
        {
          from: range.raw.to,
          to: range.raw.from,
        },
        false
      );
    }

    if (this.refresh) {
      this.setAutoRefresh(this.refresh);
    }
  }

  getValidIntervals(intervals: string[]): string[] {
    if (!this.contextSrv.minRefreshInterval) {
      return intervals;
    }

    return intervals.filter((str) => str !== '').filter(this.contextSrv.isAllowedInterval);
  }

  private parseTime() {
    // when absolute time is saved in json it is turned to a string
    if (isString(this.time.from) && this.time.from.indexOf('Z') >= 0) {
      this.time.from = dateTime(this.time.from).utc();
    }
    if (isString(this.time.to) && this.time.to.indexOf('Z') >= 0) {
      this.time.to = dateTime(this.time.to).utc();
    }
  }

  private parseUrlParam(value: any) {
    if (value.indexOf('now') !== -1) {
      return value;
    }
    if (value.length === 8) {
      const utcValue = toUtc(value, 'YYYYMMDD');
      if (utcValue.isValid()) {
        return utcValue;
      }
    } else if (value.length === 15) {
      const utcValue = toUtc(value, 'YYYYMMDDTHHmmss');
      if (utcValue.isValid()) {
        return utcValue;
      }
    }

    if (!isNaN(value)) {
      const epoch = parseInt(value, 10);
      return toUtc(epoch);
    }

    return null;
  }

  private getTimeWindow(time: string, timeWindow: string) {
    const valueTime = parseInt(time, 10);
    let timeWindowMs;

    if (timeWindow.match(/^\d+$/) && parseInt(timeWindow, 10)) {
      // when time window specified in ms
      timeWindowMs = parseInt(timeWindow, 10);
    } else {
      timeWindowMs = rangeUtil.intervalToMs(timeWindow);
    }

    return {
      from: toUtc(valueTime - timeWindowMs / 2),
      to: toUtc(valueTime + timeWindowMs / 2),
    };
  }

  private initTimeFromUrl() {
    const params = locationService.getSearch();

    if (params.get('time') && params.get('time.window')) {
      this.time = this.getTimeWindow(params.get('time')!, params.get('time.window')!);
    }

    if (params.get('from')) {
      this.time.from = this.parseUrlParam(params.get('from')!) || this.time.from;
    }

    if (params.get('to')) {
      this.time.to = this.parseUrlParam(params.get('to')!) || this.time.to;
    }

    // if absolute ignore refresh option saved to dashboard
    if (params.get('to') && params.get('to')!.indexOf('now') === -1) {
      this.refresh = false;
      if (this.dashboard) {
        this.dashboard.refresh = false;
      }
    }

    let paramsJSON: Record<string, string> = {};
    params.forEach(function (value, key) {
      paramsJSON[key] = value;
    });

    // but if refresh explicitly set then use that
    this.refresh = getRefreshFromUrl({
      params: paramsJSON,
      currentRefresh: this.refresh,
      refreshIntervals: Array.isArray(this.dashboard?.timepicker?.refresh_intervals)
        ? this.dashboard?.timepicker?.refresh_intervals
        : undefined,
      isAllowedIntervalFn: this.contextSrv.isAllowedInterval,
      minRefreshInterval: config.minRefreshInterval,
    });
  }

  updateTimeRangeFromUrl() {
    const params = locationService.getSearch();

    if (params.get('left')) {
      return; // explore handles this;
    }

    const urlRange = this.timeRangeForUrl();
    const from = params.get('from');
    const to = params.get('to');

    // check if url has time range
    if (from && to) {
      // is it different from what our current time range?
      if (from !== urlRange.from || to !== urlRange.to) {
        // issue update
        this.initTimeFromUrl();
        this.setTime(this.time, true);
      }
    } else if (this.timeHasChangedSinceLoad()) {
      this.setTime(this.timeAtLoad, true);
    }
  }

  private timeHasChangedSinceLoad() {
    return this.timeAtLoad && (this.timeAtLoad.from !== this.time.from || this.timeAtLoad.to !== this.time.to);
  }

  setAutoRefresh(interval: any) {
    if (this.dashboard) {
      this.dashboard.refresh = interval;
    }

    this.stopAutoRefresh();

    const currentUrlState = locationService.getSearchObject();

    if (!interval) {
      // Clear URL state
      if (currentUrlState.refresh) {
        locationService.partial({ refresh: null }, true);
      }

      return;
    }

    const validInterval = this.contextSrv.getValidInterval(interval);
    const intervalMs = rangeUtil.intervalToMs(validInterval);

    this.refreshTimer = setTimeout(() => {
      this.startNextRefreshTimer(intervalMs);
      this.refreshDashboard();
    }, intervalMs);

    const refresh = this.contextSrv.getValidInterval(interval);

    if (currentUrlState.refresh !== refresh) {
      locationService.partial({ refresh }, true);
    }
  }

  refreshDashboard() {
    this.dashboard?.timeRangeUpdated(this.timeRange());
  }

  private startNextRefreshTimer(afterMs: number) {
    this.refreshTimer = setTimeout(() => {
      this.startNextRefreshTimer(afterMs);
      if (this.contextSrv.isGrafanaVisible()) {
        this.refreshDashboard();
      } else {
        this.autoRefreshBlocked = true;
      }
    }, afterMs);
  }

  stopAutoRefresh() {
    clearTimeout(this.refreshTimer);
  }

  // store dashboard refresh value and pause auto-refresh in some places
  // i.e panel edit
  pauseAutoRefresh() {
    this.previousAutoRefresh = this.dashboard?.refresh;
    this.setAutoRefresh('');
  }

  // resume auto-refresh based on old dashboard refresh property
  resumeAutoRefresh() {
    this.setAutoRefresh(this.previousAutoRefresh);
  }

  setTime(time: RawTimeRange, fromRouteUpdate?: boolean) {
    extend(this.time, time);

    // disable refresh if zoom in or zoom out
    if (isDateTime(time.to)) {
      this.oldRefresh = this.dashboard?.refresh || this.oldRefresh;
      this.setAutoRefresh(false);
    } else if (this.oldRefresh && this.oldRefresh !== this.dashboard?.refresh) {
      this.setAutoRefresh(this.oldRefresh);
      this.oldRefresh = null;
    }

    // update url
    if (fromRouteUpdate !== true) {
      const urlRange = this.timeRangeForUrl();
      const urlParams = locationService.getSearch();

      const from = urlParams.get('from');
      const to = urlParams.get('to');

      if (from && to && from === urlRange.from.toString() && to === urlRange.to.toString()) {
        return;
      }

      urlParams.set('from', urlRange.from.toString());
      urlParams.set('to', urlRange.to.toString());

      locationService.push({
        ...locationService.getLocation(),
        search: urlParams.toString(),
      });
    }

    this.refreshDashboard();
  }

  timeRangeForUrl = () => {
    const range = this.timeRange().raw;

    if (isDateTime(range.from)) {
      range.from = range.from.valueOf().toString();
    }
    if (isDateTime(range.to)) {
      range.to = range.to.valueOf().toString();
    }

    return range;
  };

  timeRange(): TimeRange {
    // make copies if they are moment  (do not want to return out internal moment, because they are mutable!)
    const raw = {
      from: isDateTime(this.time.from) ? dateTime(this.time.from) : this.time.from,
      to: isDateTime(this.time.to) ? dateTime(this.time.to) : this.time.to,
    };

    const timezone = this.dashboard ? this.dashboard.getTimezone() : undefined;

    return {
      from: dateMath.parse(raw.from, false, timezone, this.dashboard?.fiscalYearStartMonth)!,
      to: dateMath.parse(raw.to, true, timezone, this.dashboard?.fiscalYearStartMonth)!,
      raw: raw,
    };
  }

  zoomOut(factor: number) {
    const range = this.timeRange();
    const { from, to } = getZoomedTimeRange(range, factor);

    this.setTime({ from: toUtc(from), to: toUtc(to) });
  }

  shiftTime(direction: ShiftTimeEventPayload) {
    const range = this.timeRange();
    const { from, to } = getShiftedTimeRange(direction, range);

    this.setTime({
      from: toUtc(from),
      to: toUtc(to),
    });
  }
}

let singleton: TimeSrv | undefined;

export function setTimeSrv(srv: TimeSrv) {
  singleton = srv;
}

export function getTimeSrv(): TimeSrv {
  if (!singleton) {
    singleton = new TimeSrv(contextSrv);
  }

  return singleton;
}
