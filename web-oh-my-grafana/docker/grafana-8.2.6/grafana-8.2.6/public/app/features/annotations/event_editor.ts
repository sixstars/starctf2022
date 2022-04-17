import { cloneDeep, isNumber } from 'lodash';
import { coreModule } from 'app/core/core';
import { AnnotationEvent, dateTime } from '@grafana/data';
import { MetricsPanelCtrl } from '../panel/metrics_panel_ctrl';
import { deleteAnnotation, saveAnnotation, updateAnnotation } from './api';
import { getDashboardQueryRunner } from '../query/state/DashboardQueryRunner/DashboardQueryRunner';

export class EventEditorCtrl {
  // @ts-ignore initialized through Angular not constructor
  panelCtrl: MetricsPanelCtrl;
  // @ts-ignore initialized through Angular not constructor
  event: AnnotationEvent;
  timeRange?: { from: number; to: number };
  form: any;
  close: any;
  timeFormated?: string;

  /** @ngInject */
  constructor() {}

  $onInit() {
    this.event.panelId = this.panelCtrl.panel.editSourceId ?? this.panelCtrl.panel.id; // set correct id if in panel edit
    this.event.dashboardId = this.panelCtrl.dashboard.id;

    // Annotations query returns time as Unix timestamp in milliseconds
    this.event.time = tryEpochToMoment(this.event.time);
    if (this.event.isRegion) {
      this.event.timeEnd = tryEpochToMoment(this.event.timeEnd);
    }

    this.timeFormated = this.panelCtrl.dashboard.formatDate(this.event.time!);
  }

  async save(): Promise<void> {
    if (!this.form.$valid) {
      return;
    }

    const saveModel = cloneDeep(this.event);
    saveModel.time = saveModel.time!.valueOf();
    saveModel.timeEnd = 0;

    if (saveModel.isRegion) {
      saveModel.timeEnd = this.event.timeEnd!.valueOf();

      if (saveModel.timeEnd < saveModel.time) {
        console.log('invalid time');
        return;
      }
    }

    let crudFunction = saveAnnotation;
    if (saveModel.id) {
      crudFunction = updateAnnotation;
    }

    try {
      await crudFunction(saveModel);
    } catch (err) {
      console.log(err);
    } finally {
      this.close();
      getDashboardQueryRunner().run({ dashboard: this.panelCtrl.dashboard, range: this.panelCtrl.range });
    }
  }

  async delete(): Promise<void> {
    try {
      await deleteAnnotation(this.event);
    } catch (err) {
      console.log(err);
    } finally {
      this.close();
      getDashboardQueryRunner().run({ dashboard: this.panelCtrl.dashboard, range: this.panelCtrl.range });
    }
  }
}

function tryEpochToMoment(timestamp: any) {
  if (timestamp && isNumber(timestamp)) {
    const epoch = Number(timestamp);
    return dateTime(epoch);
  } else {
    return timestamp;
  }
}

export function eventEditor() {
  return {
    restrict: 'E',
    controller: EventEditorCtrl,
    bindToController: true,
    controllerAs: 'ctrl',
    templateUrl: 'public/app/features/annotations/partials/event_editor.html',
    scope: {
      panelCtrl: '=',
      event: '=',
      close: '&',
    },
  };
}

coreModule.directive('eventEditor', eventEditor);
