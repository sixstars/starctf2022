import config from '../../core/config';
import { extend } from 'lodash';
import coreModule from 'app/core/core_module';
import { rangeUtil } from '@grafana/data';
import { AccessControlAction, UserPermission } from 'app/types';

export class User {
  id: number;
  isGrafanaAdmin: any;
  isSignedIn: any;
  orgRole: any;
  orgId: number;
  orgName: string;
  login: string;
  orgCount: number;
  timezone: string;
  fiscalYearStartMonth: number;
  helpFlags1: number;
  lightTheme: boolean;
  hasEditPermissionInFolders: boolean;
  email?: string;
  permissions?: UserPermission;

  constructor() {
    this.id = 0;
    this.isGrafanaAdmin = false;
    this.isSignedIn = false;
    this.orgRole = '';
    this.orgId = 0;
    this.orgName = '';
    this.login = '';
    this.orgCount = 0;
    this.timezone = '';
    this.fiscalYearStartMonth = 0;
    this.helpFlags1 = 0;
    this.lightTheme = false;
    this.hasEditPermissionInFolders = false;
    this.email = undefined;
    if (config.bootData.user) {
      extend(this, config.bootData.user);
    }
  }
}

export class ContextSrv {
  pinned: any;
  version: any;
  user: User;
  isSignedIn: any;
  isGrafanaAdmin: any;
  isEditor: any;
  sidemenuSmallBreakpoint = false;
  hasEditPermissionInFolders: boolean;
  minRefreshInterval: string;

  constructor() {
    if (!config.bootData) {
      config.bootData = { user: {}, settings: {} };
    }

    this.user = new User();
    this.isSignedIn = this.user.isSignedIn;
    this.isGrafanaAdmin = this.user.isGrafanaAdmin;
    this.isEditor = this.hasRole('Editor') || this.hasRole('Admin');
    this.hasEditPermissionInFolders = this.user.hasEditPermissionInFolders;
    this.minRefreshInterval = config.minRefreshInterval;
  }

  /**
   * Indicate the user has been logged out
   */
  setLoggedOut() {
    this.user.isSignedIn = false;
    this.isSignedIn = false;
  }

  hasRole(role: string) {
    return this.user.orgRole === role;
  }

  // Checks whether user has required permission
  hasPermission(action: AccessControlAction | string): boolean {
    // Fallback if access control disabled
    if (!config.featureToggles['accesscontrol']) {
      return true;
    }

    return !!this.user.permissions?.[action];
  }

  isGrafanaVisible() {
    return !!(document.visibilityState === undefined || document.visibilityState === 'visible');
  }

  // checks whether the passed interval is longer than the configured minimum refresh rate
  isAllowedInterval(interval: string) {
    if (!config.minRefreshInterval) {
      return true;
    }
    return rangeUtil.intervalToMs(interval) >= rangeUtil.intervalToMs(config.minRefreshInterval);
  }

  getValidInterval(interval: string) {
    if (!this.isAllowedInterval(interval)) {
      return config.minRefreshInterval;
    }
    return interval;
  }

  hasAccessToExplore() {
    if (config.featureToggles['accesscontrol']) {
      return this.hasPermission(AccessControlAction.DataSourcesExplore);
    }
    return (this.isEditor || config.viewersCanEdit) && config.exploreEnabled;
  }
}

let contextSrv = new ContextSrv();
export { contextSrv };

export const setContextSrv = (override: ContextSrv) => {
  if (process.env.NODE_ENV !== 'test') {
    throw new Error('contextSrv can be only overriden in test environment');
  }
  contextSrv = override;
};

coreModule.factory('contextSrv', () => {
  return contextSrv;
});
