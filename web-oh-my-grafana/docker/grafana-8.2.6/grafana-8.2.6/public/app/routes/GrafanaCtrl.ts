// Libraries
// eslint-disable-next-line lodash/import-scope
import _ from 'lodash';
import $ from 'jquery';

// Utils and servies
import { colors } from '@grafana/ui';
import {
  setBackendSrv,
  setDataSourceSrv,
  setLegacyAngularInjector,
  setLocationSrv,
  locationService,
} from '@grafana/runtime';
import config from 'app/core/config';
import coreModule from 'app/core/core_module';
import { profiler } from 'app/core/profiler';
import appEvents from 'app/core/app_events';
import { DatasourceSrv } from 'app/features/plugins/datasource_srv';
import { AngularLoader, setAngularLoader } from 'app/core/services/AngularLoader';

// Types
import { CoreEvents, AppEventEmitter, AppEventConsumer } from 'app/types';
import { setLinkSrv, LinkSrv } from 'app/features/panel/panellinks/link_srv';
import { UtilSrv } from 'app/core/services/util_srv';
import { ContextSrv } from 'app/core/services/context_srv';
import { DashboardSrv, setDashboardSrv } from 'app/features/dashboard/services/DashboardSrv';
import { IRootScopeService, IAngularEvent, auto } from 'angular';
import { AppEvent } from '@grafana/data';
import { backendSrv } from 'app/core/services/backend_srv';
import { initGrafanaLive } from 'app/features/live/live';

export type GrafanaRootScope = IRootScopeService & AppEventEmitter & AppEventConsumer & { colors: string[] };

export class GrafanaCtrl {
  /** @ngInject */
  constructor(
    $scope: any,
    utilSrv: UtilSrv,
    $rootScope: GrafanaRootScope,
    contextSrv: ContextSrv,
    linkSrv: LinkSrv,
    datasourceSrv: DatasourceSrv,
    dashboardSrv: DashboardSrv,
    angularLoader: AngularLoader,
    $injector: auto.IInjectorService
  ) {
    // make angular loader service available to react components
    setAngularLoader(angularLoader);
    setBackendSrv(backendSrv);
    setDataSourceSrv(datasourceSrv);
    setLinkSrv(linkSrv);
    setDashboardSrv(dashboardSrv);
    setLegacyAngularInjector($injector);

    datasourceSrv.init(config.datasources, config.defaultDatasource);

    setLocationSrv(locationService);

    initGrafanaLive();

    $scope.init = () => {
      $scope.contextSrv = contextSrv;
      $scope.appSubUrl = config.appSubUrl;
      $scope._ = _;

      profiler.init(config, $rootScope);
      utilSrv.init();
    };

    $rootScope.colors = colors;

    $rootScope.onAppEvent = function <T>(
      event: AppEvent<T> | string,
      callback: (event: IAngularEvent, ...args: any[]) => void,
      localScope?: any
    ) {
      let unbind;
      if (typeof event === 'string') {
        unbind = $rootScope.$on(event, callback);
      } else {
        unbind = $rootScope.$on(event.name, callback);
      }

      let callerScope = this;
      if (callerScope.$id === 1 && !localScope) {
        console.warn('warning rootScope onAppEvent called without localscope');
      }
      if (localScope) {
        callerScope = localScope;
      }
      callerScope.$on('$destroy', unbind);
    };

    $rootScope.appEvent = <T>(event: AppEvent<T> | string, payload?: T | any) => {
      if (typeof event === 'string') {
        $rootScope.$emit(event, payload);
        appEvents.emit(event, payload);
      } else {
        $rootScope.$emit(event.name, payload);
        appEvents.emit(event, payload);
      }
    };

    $scope.init();
  }
}

/** @ngInject */
export function grafanaAppDirective() {
  return {
    restrict: 'E',
    controller: GrafanaCtrl,
    link: (scope: IRootScopeService & AppEventEmitter, elem: JQuery) => {
      const body = $('body');
      // see https://github.com/zenorocha/clipboard.js/issues/155
      $.fn.modal.Constructor.prototype.enforceFocus = () => {};

      $('.preloader').remove();

      appEvents.on(CoreEvents.toggleSidemenuMobile, () => {
        body.toggleClass('sidemenu-open--xs');
      });

      appEvents.on(CoreEvents.toggleSidemenuHidden, () => {
        body.toggleClass('sidemenu-hidden');
      });

      // handle in active view state class
      let lastActivity = new Date().getTime();
      let activeUser = true;
      const inActiveTimeLimit = 60 * 5000;

      function checkForInActiveUser() {
        if (!activeUser) {
          return;
        }
        // only go to activity low mode on dashboard page
        if (!body.hasClass('page-dashboard')) {
          return;
        }

        if (new Date().getTime() - lastActivity > inActiveTimeLimit) {
          activeUser = false;
          body.addClass('view-mode--inactive');
        }
      }

      function userActivityDetected() {
        lastActivity = new Date().getTime();
        if (!activeUser) {
          activeUser = true;
          body.removeClass('view-mode--inactive');
        }
      }

      // mouse and keyboard is user activity
      body.mousemove(userActivityDetected);
      body.keydown(userActivityDetected);
      // set useCapture = true to catch event here
      document.addEventListener('wheel', userActivityDetected, { capture: true, passive: true });
      // treat tab change as activity
      document.addEventListener('visibilitychange', userActivityDetected);

      // check every 2 seconds
      setInterval(checkForInActiveUser, 2000);

      // handle document clicks that should hide things
      body.click((evt) => {
        const target = $(evt.target);
        if (target.parents().length === 0) {
          return;
        }

        // ensure dropdown menu doesn't impact on z-index
        body.find('.dropdown-menu-open').removeClass('dropdown-menu-open');

        // for stuff that animates, slides out etc, clicking it needs to
        // hide it right away
        const clickAutoHide = target.closest('[data-click-hide]');
        if (clickAutoHide.length) {
          const clickAutoHideParent = clickAutoHide.parent();
          clickAutoHide.detach();
          setTimeout(() => {
            clickAutoHideParent.append(clickAutoHide);
          }, 100);
        }

        // hide popovers
        const popover = elem.find('.popover');
        if (popover.length > 0 && target.parents('.graph-legend').length === 0) {
          popover.hide();
        }
      });
    },
  };
}

coreModule.directive('grafanaApp', grafanaAppDirective);
