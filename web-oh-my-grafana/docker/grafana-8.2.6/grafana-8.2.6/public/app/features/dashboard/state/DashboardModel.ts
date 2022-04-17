// Libaries
import {
  cloneDeep,
  defaults as _defaults,
  each,
  filter,
  find,
  findIndex,
  indexOf,
  isEqual,
  map,
  maxBy,
  pull,
  some,
} from 'lodash';
// Constants
import { DEFAULT_ANNOTATION_COLOR } from '@grafana/ui';
import { GRID_CELL_HEIGHT, GRID_CELL_VMARGIN, GRID_COLUMN_COUNT, REPEAT_DIR_VERTICAL } from 'app/core/constants';
// Utils & Services
import { contextSrv } from 'app/core/services/context_srv';
import sortByKeys from 'app/core/utils/sort_by_keys';
// Types
import { GridPos, PanelModel } from './PanelModel';
import { DashboardMigrator } from './DashboardMigrator';
import {
  AnnotationQuery,
  AppEvent,
  DashboardCursorSync,
  dateTimeFormat,
  dateTimeFormatTimeAgo,
  DateTimeInput,
  EventBusExtended,
  EventBusSrv,
  PanelModel as IPanelModel,
  TimeRange,
  TimeZone,
  UrlQueryValue,
} from '@grafana/data';
import { CoreEvents, DashboardMeta, KioskMode } from 'app/types';
import { GetVariables, getVariables } from 'app/features/variables/state/selectors';
import { variableAdapters } from 'app/features/variables/adapters';
import { onTimeRangeUpdated } from 'app/features/variables/state/actions';
import { dispatch } from '../../../store/store';
import { isAllVariable } from '../../variables/utils';
import { DashboardPanelsChangedEvent, RefreshEvent, RenderEvent, TimeRangeUpdatedEvent } from 'app/types/events';
import { getTimeSrv } from '../services/TimeSrv';
import { mergePanels, PanelMergeInfo } from '../utils/panelMerge';
import { isOnTheSameGridRow } from './utils';

export interface CloneOptions {
  saveVariables?: boolean;
  saveTimerange?: boolean;
  message?: string;
}

export type DashboardLinkType = 'link' | 'dashboards';

export interface DashboardLink {
  icon: string;
  title: string;
  tooltip: string;
  type: DashboardLinkType;
  url: string;
  asDropdown: boolean;
  tags: any[];
  searchHits?: any[];
  targetBlank: boolean;
  keepTime: boolean;
  includeVars: boolean;
}

export class DashboardModel {
  id: any;
  uid: string;
  title: string;
  autoUpdate: any;
  description: any;
  tags: any;
  style: any;
  timezone: any;
  editable: any;
  graphTooltip: DashboardCursorSync;
  time: any;
  liveNow: boolean;
  private originalTime: any;
  timepicker: any;
  templating: { list: any[] };
  private originalTemplating: any;
  annotations: { list: AnnotationQuery[] };
  refresh: any;
  snapshot: any;
  schemaVersion: number;
  version: number;
  revision: number;
  links: DashboardLink[];
  gnetId: any;
  panels: PanelModel[];
  panelInEdit?: PanelModel;
  panelInView?: PanelModel;
  fiscalYearStartMonth?: number;
  private hasChangesThatAffectsAllPanels: boolean;

  // ------------------
  // not persisted
  // ------------------

  // repeat process cycles
  iteration?: number;
  declare meta: DashboardMeta;
  events: EventBusExtended;

  static nonPersistedProperties: { [str: string]: boolean } = {
    events: true,
    meta: true,
    panels: true, // needs special handling
    templating: true, // needs special handling
    originalTime: true,
    originalTemplating: true,
    originalLibraryPanels: true,
    panelInEdit: true,
    panelInView: true,
    getVariablesFromState: true,
    formatDate: true,
    hasChangesThatAffectsAllPanels: true,
  };

  constructor(data: any, meta?: DashboardMeta, private getVariablesFromState: GetVariables = getVariables) {
    if (!data) {
      data = {};
    }

    this.events = new EventBusSrv();
    this.id = data.id || null;
    this.uid = data.uid || null;
    this.revision = data.revision;
    this.title = data.title ?? 'No Title';
    this.autoUpdate = data.autoUpdate;
    this.description = data.description;
    this.tags = data.tags ?? [];
    this.style = data.style ?? 'dark';
    this.timezone = data.timezone ?? '';
    this.editable = data.editable !== false;
    this.graphTooltip = data.graphTooltip || 0;
    this.time = data.time ?? { from: 'now-6h', to: 'now' };
    this.timepicker = data.timepicker ?? {};
    this.liveNow = Boolean(data.liveNow);
    this.templating = this.ensureListExist(data.templating);
    this.annotations = this.ensureListExist(data.annotations);
    this.refresh = data.refresh;
    this.snapshot = data.snapshot;
    this.schemaVersion = data.schemaVersion ?? 0;
    this.fiscalYearStartMonth = data.fiscalYearStartMonth ?? 0;
    this.version = data.version ?? 0;
    this.links = data.links ?? [];
    this.gnetId = data.gnetId || null;
    this.panels = map(data.panels ?? [], (panelData: any) => new PanelModel(panelData));
    this.formatDate = this.formatDate.bind(this);

    this.resetOriginalVariables(true);
    this.resetOriginalTime();

    this.initMeta(meta);
    this.updateSchema(data);

    this.addBuiltInAnnotationQuery();
    this.sortPanelsByGridPos();
    this.hasChangesThatAffectsAllPanels = false;
  }

  addBuiltInAnnotationQuery() {
    let found = false;
    for (const item of this.annotations.list) {
      if (item.builtIn === 1) {
        found = true;
        break;
      }
    }

    if (found) {
      return;
    }

    this.annotations.list.unshift({
      datasource: '-- Grafana --',
      name: 'Annotations & Alerts',
      type: 'dashboard',
      iconColor: DEFAULT_ANNOTATION_COLOR,
      enable: true,
      hide: true,
      builtIn: 1,
    });
  }

  private initMeta(meta?: DashboardMeta) {
    meta = meta || {};

    meta.canShare = meta.canShare !== false;
    meta.canSave = meta.canSave !== false;
    meta.canStar = meta.canStar !== false;
    meta.canEdit = meta.canEdit !== false;
    meta.showSettings = meta.canEdit;
    meta.canMakeEditable = meta.canSave && !this.editable;
    meta.hasUnsavedFolderChange = false;

    if (!this.editable) {
      meta.canEdit = false;
      meta.canDelete = false;
      meta.canSave = false;
    }

    this.meta = meta;
  }

  // cleans meta data and other non persistent state
  getSaveModelClone(options?: CloneOptions): DashboardModel {
    const defaults = _defaults(options || {}, {
      saveVariables: true,
      saveTimerange: true,
    });

    // make clone
    let copy: any = {};
    for (const property in this) {
      if (DashboardModel.nonPersistedProperties[property] || !this.hasOwnProperty(property)) {
        continue;
      }

      copy[property] = cloneDeep(this[property]);
    }

    this.updateTemplatingSaveModelClone(copy, defaults);

    if (!defaults.saveTimerange) {
      copy.time = this.originalTime;
    }

    // get panel save models
    copy.panels = this.getPanelSaveModels();

    //  sort by keys
    copy = sortByKeys(copy);
    copy.getVariables = () => {
      return copy.templating.list;
    };

    return copy;
  }

  /**
   * This will load a new dashboard, but keep existing panels unchanged
   *
   * This function can be used to implement:
   * 1. potentially faster loading dashboard loading
   * 2. dynamic dashboard behavior
   * 3. "live" dashboard editing
   *
   * @internal and experimental
   */
  updatePanels(panels: IPanelModel[]): PanelMergeInfo {
    const info = mergePanels(this.panels, panels ?? []);
    if (info.changed) {
      this.panels = info.panels ?? [];
      this.sortPanelsByGridPos();
      this.events.publish(new DashboardPanelsChangedEvent());
    }
    return info;
  }

  private getPanelSaveModels() {
    return this.panels
      .filter((panel: PanelModel) => {
        if (this.isSnapshotTruthy()) {
          return true;
        }
        if (panel.type === 'add-panel') {
          return false;
        }
        // skip repeated panels in the saved model
        if (panel.repeatPanelId) {
          return false;
        }
        // skip repeated rows in the saved model
        if (panel.repeatedByRow) {
          return false;
        }
        return true;
      })
      .map((panel: PanelModel) => {
        // If we save while editing we should include the panel in edit mode instead of the
        // unmodified source panel
        if (this.panelInEdit && this.panelInEdit.editSourceId === panel.id) {
          const saveModel = this.panelInEdit.getSaveModel();
          // while editing a panel we modify its id, need to restore it here
          saveModel.id = this.panelInEdit.editSourceId;
          return saveModel;
        }

        return panel.getSaveModel();
      })
      .map((model: any) => {
        if (this.isSnapshotTruthy()) {
          return model;
        }
        // Clear any scopedVars from persisted mode. This cannot be part of getSaveModel as we need to be able to copy
        // panel models with preserved scopedVars, for example when going into edit mode.
        delete model.scopedVars;

        // Clear any repeated panels from collapsed rows
        if (model.type === 'row' && model.panels && model.panels.length > 0) {
          model.panels = model.panels
            .filter((rowPanel: PanelModel) => !rowPanel.repeatPanelId)
            .map((model: PanelModel) => {
              delete model.scopedVars;
              return model;
            });
        }

        return model;
      });
  }

  private updateTemplatingSaveModelClone(
    copy: any,
    defaults: { saveTimerange: boolean; saveVariables: boolean } & CloneOptions
  ) {
    const originalVariables = this.originalTemplating;
    const currentVariables = this.getVariablesFromState();

    copy.templating = {
      list: currentVariables.map((variable) =>
        variableAdapters.get(variable.type).getSaveModel(variable, defaults.saveVariables)
      ),
    };

    if (!defaults.saveVariables) {
      for (let i = 0; i < copy.templating.list.length; i++) {
        const current = copy.templating.list[i];
        const original: any = find(originalVariables, { name: current.name, type: current.type });

        if (!original) {
          continue;
        }

        if (current.type === 'adhoc') {
          copy.templating.list[i].filters = original.filters;
        } else {
          copy.templating.list[i].current = original.current;
        }
      }
    }
  }

  timeRangeUpdated(timeRange: TimeRange) {
    this.events.publish(new TimeRangeUpdatedEvent(timeRange));
    dispatch(onTimeRangeUpdated(timeRange));
  }

  startRefresh() {
    this.events.publish(new RefreshEvent());

    if (this.panelInEdit) {
      this.panelInEdit.refresh();
      return;
    }

    for (const panel of this.panels) {
      if (!this.otherPanelInFullscreen(panel)) {
        panel.refresh();
      }
    }
  }

  render() {
    this.events.publish(new RenderEvent());
    for (const panel of this.panels) {
      panel.render();
    }
  }

  panelInitialized(panel: PanelModel) {
    const lastResult = panel.getQueryRunner().getLastResult();

    if (!this.otherPanelInFullscreen(panel) && !lastResult) {
      panel.refresh();
    }
  }

  otherPanelInFullscreen(panel: PanelModel) {
    return (this.panelInEdit || this.panelInView) && !(panel.isViewing || panel.isEditing);
  }

  initEditPanel(sourcePanel: PanelModel): PanelModel {
    getTimeSrv().pauseAutoRefresh();
    this.panelInEdit = sourcePanel.getEditClone();
    return this.panelInEdit;
  }

  initViewPanel(panel: PanelModel) {
    this.panelInView = panel;
    panel.setIsViewing(true);
  }

  exitViewPanel(panel: PanelModel) {
    this.panelInView = undefined;
    panel.setIsViewing(false);
    this.refreshIfChangeAffectsAllPanels();
  }

  exitPanelEditor() {
    this.panelInEdit!.destroy();
    this.panelInEdit = undefined;
    this.refreshIfChangeAffectsAllPanels();
    getTimeSrv().resumeAutoRefresh();
  }

  setChangeAffectsAllPanels() {
    if (this.panelInEdit || this.panelInView) {
      this.hasChangesThatAffectsAllPanels = true;
    }
  }

  private refreshIfChangeAffectsAllPanels() {
    if (!this.hasChangesThatAffectsAllPanels) {
      return;
    }

    this.hasChangesThatAffectsAllPanels = false;
    this.startRefresh();
  }

  private ensureListExist(data: any) {
    if (!data) {
      data = {};
    }
    if (!data.list) {
      data.list = [];
    }
    return data;
  }

  getNextPanelId() {
    let max = 0;

    for (const panel of this.panels) {
      if (panel.id > max) {
        max = panel.id;
      }

      if (panel.collapsed) {
        for (const rowPanel of panel.panels) {
          if (rowPanel.id > max) {
            max = rowPanel.id;
          }
        }
      }
    }

    return max + 1;
  }

  forEachPanel(callback: (panel: PanelModel, index: number) => void) {
    for (let i = 0; i < this.panels.length; i++) {
      callback(this.panels[i], i);
    }
  }

  getPanelById(id: number): PanelModel | null {
    if (this.panelInEdit && this.panelInEdit.id === id) {
      return this.panelInEdit;
    }

    for (const panel of this.panels) {
      if (panel.id === id) {
        return panel;
      }
    }

    return null;
  }

  canEditPanel(panel?: PanelModel | null): boolean | undefined | null {
    return Boolean(this.meta.canEdit && panel && !panel.repeatPanelId && panel.type !== 'row');
  }

  canEditPanelById(id: number): boolean | undefined | null {
    return this.canEditPanel(this.getPanelById(id));
  }

  addPanel(panelData: any) {
    panelData.id = this.getNextPanelId();

    this.panels.unshift(new PanelModel(panelData));

    this.sortPanelsByGridPos();

    this.events.publish(new DashboardPanelsChangedEvent());
  }

  sortPanelsByGridPos() {
    this.panels.sort((panelA, panelB) => {
      if (panelA.gridPos.y === panelB.gridPos.y) {
        return panelA.gridPos.x - panelB.gridPos.x;
      } else {
        return panelA.gridPos.y - panelB.gridPos.y;
      }
    });
  }

  clearUnsavedChanges() {
    for (const panel of this.panels) {
      panel.configRev = 0;
    }
  }

  hasUnsavedChanges() {
    for (const panel of this.panels) {
      if (panel.hasChanged) {
        console.log('Panel has changed', panel);
        return true;
      }
    }
    return false;
  }

  cleanUpRepeats() {
    if (this.isSnapshotTruthy() || !this.hasVariables()) {
      return;
    }

    this.iteration = (this.iteration || new Date().getTime()) + 1;
    const panelsToRemove = [];

    // cleanup scopedVars
    for (const panel of this.panels) {
      delete panel.scopedVars;
    }

    for (let i = 0; i < this.panels.length; i++) {
      const panel = this.panels[i];
      if ((!panel.repeat || panel.repeatedByRow) && panel.repeatPanelId && panel.repeatIteration !== this.iteration) {
        panelsToRemove.push(panel);
      }
    }

    // remove panels
    pull(this.panels, ...panelsToRemove);
    panelsToRemove.map((p) => p.destroy());
    this.sortPanelsByGridPos();
    this.events.publish(new DashboardPanelsChangedEvent());
  }

  processRepeats() {
    if (this.isSnapshotTruthy() || !this.hasVariables()) {
      return;
    }

    this.cleanUpRepeats();

    this.iteration = (this.iteration || new Date().getTime()) + 1;

    for (let i = 0; i < this.panels.length; i++) {
      const panel = this.panels[i];
      if (panel.repeat) {
        this.repeatPanel(panel, i);
      }
    }

    this.sortPanelsByGridPos();
    this.events.publish(new DashboardPanelsChangedEvent());
  }

  cleanUpRowRepeats(rowPanels: PanelModel[]) {
    const panelsToRemove = [];
    for (let i = 0; i < rowPanels.length; i++) {
      const panel = rowPanels[i];
      if (!panel.repeat && panel.repeatPanelId) {
        panelsToRemove.push(panel);
      }
    }
    pull(rowPanels, ...panelsToRemove);
    pull(this.panels, ...panelsToRemove);
  }

  processRowRepeats(row: PanelModel) {
    if (this.isSnapshotTruthy() || !this.hasVariables()) {
      return;
    }

    let rowPanels = row.panels;
    if (!row.collapsed) {
      const rowPanelIndex = findIndex(this.panels, (p: PanelModel) => p.id === row.id);
      rowPanels = this.getRowPanels(rowPanelIndex);
    }

    this.cleanUpRowRepeats(rowPanels);

    for (let i = 0; i < rowPanels.length; i++) {
      const panel = rowPanels[i];
      if (panel.repeat) {
        const panelIndex = findIndex(this.panels, (p: PanelModel) => p.id === panel.id);
        this.repeatPanel(panel, panelIndex);
      }
    }
  }

  getPanelRepeatClone(sourcePanel: PanelModel, valueIndex: number, sourcePanelIndex: number) {
    // if first clone return source
    if (valueIndex === 0) {
      return sourcePanel;
    }

    const m = sourcePanel.getSaveModel();
    m.id = this.getNextPanelId();
    const clone = new PanelModel(m);

    // insert after source panel + value index
    this.panels.splice(sourcePanelIndex + valueIndex, 0, clone);

    clone.repeatIteration = this.iteration;
    clone.repeatPanelId = sourcePanel.id;
    clone.repeat = undefined;

    if (this.panelInView?.id === clone.id) {
      clone.setIsViewing(true);
      this.panelInView = clone;
    }

    return clone;
  }

  getRowRepeatClone(sourceRowPanel: PanelModel, valueIndex: number, sourcePanelIndex: number) {
    // if first clone return source
    if (valueIndex === 0) {
      if (!sourceRowPanel.collapsed) {
        const rowPanels = this.getRowPanels(sourcePanelIndex);
        sourceRowPanel.panels = rowPanels;
      }
      return sourceRowPanel;
    }

    const clone = new PanelModel(sourceRowPanel.getSaveModel());
    // for row clones we need to figure out panels under row to clone and where to insert clone
    let rowPanels: PanelModel[], insertPos: number;
    if (sourceRowPanel.collapsed) {
      rowPanels = cloneDeep(sourceRowPanel.panels);
      clone.panels = rowPanels;
      // insert copied row after preceding row
      insertPos = sourcePanelIndex + valueIndex;
    } else {
      rowPanels = this.getRowPanels(sourcePanelIndex);
      clone.panels = map(rowPanels, (panel: PanelModel) => panel.getSaveModel());
      // insert copied row after preceding row's panels
      insertPos = sourcePanelIndex + (rowPanels.length + 1) * valueIndex;
    }
    this.panels.splice(insertPos, 0, clone);

    this.updateRepeatedPanelIds(clone);
    return clone;
  }

  repeatPanel(panel: PanelModel, panelIndex: number) {
    const variable: any = this.getPanelRepeatVariable(panel);
    if (!variable) {
      return;
    }

    if (panel.type === 'row') {
      this.repeatRow(panel, panelIndex, variable);
      return;
    }

    const selectedOptions = this.getSelectedVariableOptions(variable);

    const maxPerRow = panel.maxPerRow || 4;
    let xPos = 0;
    let yPos = panel.gridPos.y;

    for (let index = 0; index < selectedOptions.length; index++) {
      const option = selectedOptions[index];
      let copy;

      copy = this.getPanelRepeatClone(panel, index, panelIndex);
      copy.scopedVars = copy.scopedVars || {};
      copy.scopedVars[variable.name] = option;

      if (panel.repeatDirection === REPEAT_DIR_VERTICAL) {
        if (index > 0) {
          yPos += copy.gridPos.h;
        }
        copy.gridPos.y = yPos;
      } else {
        // set width based on how many are selected
        // assumed the repeated panels should take up full row width
        copy.gridPos.w = Math.max(GRID_COLUMN_COUNT / selectedOptions.length, GRID_COLUMN_COUNT / maxPerRow);
        copy.gridPos.x = xPos;
        copy.gridPos.y = yPos;

        xPos += copy.gridPos.w;

        // handle overflow by pushing down one row
        if (xPos + copy.gridPos.w > GRID_COLUMN_COUNT) {
          xPos = 0;
          yPos += copy.gridPos.h;
        }
      }
    }

    // Update gridPos for panels below
    const yOffset = yPos - panel.gridPos.y;
    if (yOffset > 0) {
      const panelBelowIndex = panelIndex + selectedOptions.length;
      for (let i = panelBelowIndex; i < this.panels.length; i++) {
        if (isOnTheSameGridRow(panel, this.panels[i])) {
          continue;
        }

        this.panels[i].gridPos.y += yOffset;
      }
    }
  }

  repeatRow(panel: PanelModel, panelIndex: number, variable: any) {
    const selectedOptions = this.getSelectedVariableOptions(variable);
    let yPos = panel.gridPos.y;

    function setScopedVars(panel: PanelModel, variableOption: any) {
      panel.scopedVars = panel.scopedVars || {};
      panel.scopedVars[variable.name] = variableOption;
    }

    for (let optionIndex = 0; optionIndex < selectedOptions.length; optionIndex++) {
      const option = selectedOptions[optionIndex];
      const rowCopy = this.getRowRepeatClone(panel, optionIndex, panelIndex);
      setScopedVars(rowCopy, option);

      const rowHeight = this.getRowHeight(rowCopy);
      const rowPanels = rowCopy.panels || [];
      let panelBelowIndex;

      if (panel.collapsed) {
        // For collapsed row just copy its panels and set scoped vars and proper IDs
        each(rowPanels, (rowPanel: PanelModel, i: number) => {
          setScopedVars(rowPanel, option);
          if (optionIndex > 0) {
            this.updateRepeatedPanelIds(rowPanel, true);
          }
        });
        rowCopy.gridPos.y += optionIndex;
        yPos += optionIndex;
        panelBelowIndex = panelIndex + optionIndex + 1;
      } else {
        // insert after 'row' panel
        const insertPos = panelIndex + (rowPanels.length + 1) * optionIndex + 1;
        each(rowPanels, (rowPanel: PanelModel, i: number) => {
          setScopedVars(rowPanel, option);
          if (optionIndex > 0) {
            const cloneRowPanel = new PanelModel(rowPanel);
            this.updateRepeatedPanelIds(cloneRowPanel, true);
            // For exposed row additionally set proper Y grid position and add it to dashboard panels
            cloneRowPanel.gridPos.y += rowHeight * optionIndex;
            this.panels.splice(insertPos + i, 0, cloneRowPanel);
          }
        });
        rowCopy.panels = [];
        rowCopy.gridPos.y += rowHeight * optionIndex;
        yPos += rowHeight;
        panelBelowIndex = insertPos + rowPanels.length;
      }

      // Update gridPos for panels below if we inserted more than 1 repeated row panel
      if (selectedOptions.length > 1) {
        for (let i = panelBelowIndex; i < this.panels.length; i++) {
          this.panels[i].gridPos.y += yPos;
        }
      }
    }
  }

  updateRepeatedPanelIds(panel: PanelModel, repeatedByRow?: boolean) {
    panel.repeatPanelId = panel.id;
    panel.id = this.getNextPanelId();
    panel.key = `${panel.id}`;
    panel.repeatIteration = this.iteration;
    if (repeatedByRow) {
      panel.repeatedByRow = true;
    } else {
      panel.repeat = undefined;
    }
    return panel;
  }

  getSelectedVariableOptions(variable: any) {
    let selectedOptions: any[];
    if (isAllVariable(variable)) {
      selectedOptions = variable.options.slice(1, variable.options.length);
    } else {
      selectedOptions = filter(variable.options, { selected: true });
    }
    return selectedOptions;
  }

  getRowHeight(rowPanel: PanelModel): number {
    if (!rowPanel.panels || rowPanel.panels.length === 0) {
      return 0;
    }
    const rowYPos = rowPanel.gridPos.y;
    const positions = map(rowPanel.panels, 'gridPos');
    const maxPos = maxBy(positions, (pos: GridPos) => {
      return pos.y + pos.h;
    });
    return maxPos!.y + maxPos!.h - rowYPos;
  }

  removePanel(panel: PanelModel) {
    this.panels = this.panels.filter((item) => item !== panel);
    this.events.publish(new DashboardPanelsChangedEvent());
  }

  removeRow(row: PanelModel, removePanels: boolean) {
    const needToogle = (!removePanels && row.collapsed) || (removePanels && !row.collapsed);

    if (needToogle) {
      this.toggleRow(row);
    }

    this.removePanel(row);
  }

  expandRows() {
    for (let i = 0; i < this.panels.length; i++) {
      const panel = this.panels[i];

      if (panel.type !== 'row') {
        continue;
      }

      if (panel.collapsed) {
        this.toggleRow(panel);
      }
    }
  }

  collapseRows() {
    for (let i = 0; i < this.panels.length; i++) {
      const panel = this.panels[i];

      if (panel.type !== 'row') {
        continue;
      }

      if (!panel.collapsed) {
        this.toggleRow(panel);
      }
    }
  }

  isSubMenuVisible() {
    if (this.links.length > 0) {
      return true;
    }

    if (this.getVariables().find((variable) => variable.hide !== 2)) {
      return true;
    }

    if (this.annotations.list.find((annotation) => annotation.hide !== true)) {
      return true;
    }

    return false;
  }

  getPanelInfoById(panelId: number) {
    for (let i = 0; i < this.panels.length; i++) {
      if (this.panels[i].id === panelId) {
        return {
          panel: this.panels[i],
          index: i,
        };
      }
    }

    return null;
  }

  duplicatePanel(panel: PanelModel) {
    const newPanel = panel.getSaveModel();
    newPanel.id = this.getNextPanelId();

    delete newPanel.repeat;
    delete newPanel.repeatIteration;
    delete newPanel.repeatPanelId;
    delete newPanel.scopedVars;
    if (newPanel.alert) {
      delete newPanel.thresholds;
    }
    delete newPanel.alert;

    // does it fit to the right?
    if (panel.gridPos.x + panel.gridPos.w * 2 <= GRID_COLUMN_COUNT) {
      newPanel.gridPos.x += panel.gridPos.w;
    } else {
      // add below
      newPanel.gridPos.y += panel.gridPos.h;
    }

    this.addPanel(newPanel);
    return newPanel;
  }

  formatDate(date: DateTimeInput, format?: string) {
    return dateTimeFormat(date, {
      format,
      timeZone: this.getTimezone(),
    });
  }

  destroy() {
    this.events.removeAllListeners();
    for (const panel of this.panels) {
      panel.destroy();
    }
  }

  toggleRow(row: PanelModel) {
    const rowIndex = indexOf(this.panels, row);

    if (row.collapsed) {
      row.collapsed = false;
      const hasRepeat = some(row.panels as PanelModel[], (p: PanelModel) => p.repeat);

      if (row.panels.length > 0) {
        // Use first panel to figure out if it was moved or pushed
        const firstPanel = row.panels[0];
        const yDiff = firstPanel.gridPos.y - (row.gridPos.y + row.gridPos.h);

        // start inserting after row
        let insertPos = rowIndex + 1;
        // y max will represent the bottom y pos after all panels have been added
        // needed to know home much panels below should be pushed down
        let yMax = row.gridPos.y;

        for (const panel of row.panels) {
          // make sure y is adjusted (in case row moved while collapsed)
          // console.log('yDiff', yDiff);
          panel.gridPos.y -= yDiff;
          // insert after row
          this.panels.splice(insertPos, 0, new PanelModel(panel));
          // update insert post and y max
          insertPos += 1;
          yMax = Math.max(yMax, panel.gridPos.y + panel.gridPos.h);
        }

        const pushDownAmount = yMax - row.gridPos.y - 1;

        // push panels below down
        for (let panelIndex = insertPos; panelIndex < this.panels.length; panelIndex++) {
          this.panels[panelIndex].gridPos.y += pushDownAmount;
        }

        row.panels = [];

        if (hasRepeat) {
          this.processRowRepeats(row);
        }
      }

      // sort panels
      this.sortPanelsByGridPos();

      // emit change event
      this.events.publish(new DashboardPanelsChangedEvent());
      return;
    }

    const rowPanels = this.getRowPanels(rowIndex);

    // remove panels
    pull(this.panels, ...rowPanels);
    // save panel models inside row panel
    row.panels = map(rowPanels, (panel: PanelModel) => panel.getSaveModel());
    row.collapsed = true;

    // emit change event
    this.events.publish(new DashboardPanelsChangedEvent());
  }

  /**
   * Will return all panels after rowIndex until it encounters another row
   */
  getRowPanels(rowIndex: number): PanelModel[] {
    const rowPanels = [];

    for (let index = rowIndex + 1; index < this.panels.length; index++) {
      const panel = this.panels[index];

      // break when encountering another row
      if (panel.type === 'row') {
        break;
      }

      // this panel must belong to row
      rowPanels.push(panel);
    }

    return rowPanels;
  }

  /** @deprecated */
  on<T>(event: AppEvent<T>, callback: (payload?: T) => void) {
    console.log('DashboardModel.on is deprecated use events.subscribe');
    this.events.on(event, callback);
  }

  /** @deprecated */
  off<T>(event: AppEvent<T>, callback: (payload?: T) => void) {
    console.log('DashboardModel.off is deprecated');
    this.events.off(event, callback);
  }

  cycleGraphTooltip() {
    this.graphTooltip = (this.graphTooltip + 1) % 3;
  }

  sharedTooltipModeEnabled() {
    return this.graphTooltip > 0;
  }

  sharedCrosshairModeOnly() {
    return this.graphTooltip === 1;
  }

  getRelativeTime(date: DateTimeInput) {
    return dateTimeFormatTimeAgo(date, {
      timeZone: this.getTimezone(),
    });
  }

  isSnapshot() {
    return this.snapshot !== undefined;
  }

  getTimezone(): TimeZone {
    return (this.timezone ? this.timezone : contextSrv?.user?.timezone) as TimeZone;
  }

  private updateSchema(old: any) {
    const migrator = new DashboardMigrator(this);
    migrator.updateSchema(old);
  }

  resetOriginalTime() {
    this.originalTime = cloneDeep(this.time);
  }

  hasTimeChanged() {
    return !isEqual(this.time, this.originalTime);
  }

  resetOriginalVariables(initial = false) {
    if (initial) {
      this.originalTemplating = this.cloneVariablesFrom(this.templating.list);
      return;
    }

    this.originalTemplating = this.cloneVariablesFrom(this.getVariablesFromState());
  }

  hasVariableValuesChanged() {
    return this.hasVariablesChanged(this.originalTemplating, this.getVariablesFromState());
  }

  autoFitPanels(viewHeight: number, kioskMode?: UrlQueryValue) {
    const currentGridHeight = Math.max(
      ...this.panels.map((panel) => {
        return panel.gridPos.h + panel.gridPos.y;
      })
    );

    const navbarHeight = 55;
    const margin = 20;
    const submenuHeight = 50;

    let visibleHeight = viewHeight - navbarHeight - margin;

    // Remove submenu height if visible
    if (this.meta.submenuEnabled && !kioskMode) {
      visibleHeight -= submenuHeight;
    }

    // add back navbar height
    if (kioskMode && kioskMode !== KioskMode.TV) {
      visibleHeight += navbarHeight;
    }

    const visibleGridHeight = Math.floor(visibleHeight / (GRID_CELL_HEIGHT + GRID_CELL_VMARGIN));
    const scaleFactor = currentGridHeight / visibleGridHeight;

    this.panels.forEach((panel, i) => {
      panel.gridPos.y = Math.round(panel.gridPos.y / scaleFactor) || 1;
      panel.gridPos.h = Math.round(panel.gridPos.h / scaleFactor) || 1;
    });
  }

  templateVariableValueUpdated() {
    this.processRepeats();
    this.events.emit(CoreEvents.templateVariableValueUpdated);
  }

  getPanelByUrlId(panelUrlId: string) {
    const panelId = parseInt(panelUrlId ?? '0', 10);

    // First try to find it in a collapsed row and exand it
    for (const panel of this.panels) {
      if (panel.collapsed) {
        for (const rowPanel of panel.panels) {
          if (rowPanel.id === panelId) {
            this.toggleRow(panel);
            break;
          }
        }
      }
    }

    return this.getPanelById(panelId);
  }

  toggleLegendsForAll() {
    const panelsWithLegends = this.panels.filter((panel) => {
      return panel.legend !== undefined && panel.legend !== null;
    });

    // determine if more panels are displaying legends or not
    const onCount = panelsWithLegends.filter((panel) => panel.legend!.show).length;
    const offCount = panelsWithLegends.length - onCount;
    const panelLegendsOn = onCount >= offCount;

    for (const panel of panelsWithLegends) {
      panel.legend!.show = !panelLegendsOn;
      panel.render();
    }
  }

  getVariables = () => {
    return this.getVariablesFromState();
  };

  canAddAnnotations() {
    return this.meta.canEdit || this.meta.canMakeEditable;
  }

  shouldUpdateDashboardPanelFromJSON(updatedPanel: PanelModel, panel: PanelModel) {
    const shouldUpdateGridPositionLayout = !isEqual(updatedPanel?.gridPos, panel?.gridPos);
    if (shouldUpdateGridPositionLayout) {
      this.events.publish(new DashboardPanelsChangedEvent());
    }
  }

  private getPanelRepeatVariable(panel: PanelModel) {
    return this.getVariablesFromState().find((variable) => variable.name === panel.repeat);
  }

  private isSnapshotTruthy() {
    return this.snapshot;
  }

  private hasVariables() {
    return this.getVariablesFromState().length > 0;
  }

  private hasVariablesChanged(originalVariables: any[], currentVariables: any[]): boolean {
    if (originalVariables.length !== currentVariables.length) {
      return false;
    }

    const updated = map(currentVariables, (variable: any) => {
      return {
        name: variable.name,
        type: variable.type,
        current: cloneDeep(variable.current),
        filters: cloneDeep(variable.filters),
      };
    });

    return !isEqual(updated, originalVariables);
  }

  private cloneVariablesFrom(variables: any[]): any[] {
    return variables.map((variable) => {
      return {
        name: variable.name,
        type: variable.type,
        current: cloneDeep(variable.current),
        filters: cloneDeep(variable.filters),
      };
    });
  }
}
