// NOTE: by default Component string selectors are set up to be aria-labels,
// however there are many cases where your component may not need an aria-label
// (a <button> with clear text, for example, does not need an aria-label as it's already labeled)
// but you still might need to select it for testing,
// in that case please add the attribute data-test-id={selector} in the component and
// prefix your selector string with 'data-test-id' so that when create the selectors we know to search for it on the right attribute
/**
 * Selectors grouped/defined in Components
 *
 * @alpha
 */
export const Components = {
  TimePicker: {
    openButton: 'data-testid TimePicker Open Button',
    fromField: 'TimePicker from field',
    toField: 'TimePicker to field',
    applyTimeRange: 'data-testid TimePicker submit button',
    calendar: 'TimePicker calendar',
    absoluteTimeRangeTitle: 'data-testid-absolute-time-range-narrow',
  },
  DataSource: {
    TestData: {
      QueryTab: {
        scenarioSelectContainer: 'Test Data Query scenario select container',
        scenarioSelect: 'Test Data Query scenario select',
        max: 'TestData max',
        min: 'TestData min',
        noise: 'TestData noise',
        seriesCount: 'TestData series count',
        spread: 'TestData spread',
        startValue: 'TestData start value',
      },
    },
    Jaeger: {
      traceIDInput: 'Trace ID',
    },
    Prometheus: {
      configPage: {
        exemplarsAddButton: 'Add exemplar config button',
        internalLinkSwitch: 'Internal link switch',
      },
      exemplarMarker: 'Exemplar marker',
    },
  },
  Menu: {
    MenuComponent: (title: string) => `${title} menu`,
    MenuGroup: (title: string) => `${title} menu group`,
    MenuItem: (title: string) => `${title} menu item`,
  },
  Panels: {
    Panel: {
      title: (title: string) => `data-testid Panel header ${title}`,
      headerItems: (item: string) => `Panel header item ${item}`,
      containerByTitle: (title: string) => `${title} panel`,
      headerCornerInfo: (mode: string) => `Panel header ${mode}`,
    },
    Visualization: {
      Graph: {
        VisualizationTab: {
          legendSection: 'Legend section',
        },
        Legend: {
          legendItemAlias: (name: string) => `gpl alias ${name}`,
          showLegendSwitch: 'gpl show legend',
        },
        xAxis: {
          labels: () => 'div.flot-x-axis > div.flot-tick-label',
        },
      },
      BarGauge: {
        value: 'Bar gauge value',
      },
      PieChart: {
        svgSlice: 'Pie Chart Slice',
      },
      Text: {
        container: () => '.markdown-html',
      },
      Table: {
        header: 'table header',
        footer: 'table-footer',
      },
    },
  },
  VizLegend: {
    seriesName: (name: string) => `VizLegend series ${name}`,
  },
  Drawer: {
    General: {
      title: (title: string) => `Drawer title ${title}`,
      expand: 'Drawer expand',
      contract: 'Drawer contract',
      close: 'Drawer close',
      rcContentWrapper: () => '.drawer-content-wrapper',
    },
  },
  PanelEditor: {
    General: {
      content: 'Panel editor content',
    },
    OptionsPane: {
      content: 'Panel editor option pane content',
      select: 'Panel editor option pane select',
      fieldLabel: (type: string) => `${type} field property editor`,
    },
    // not sure about the naming *DataPane*
    DataPane: {
      content: 'Panel editor data pane content',
    },
    applyButton: 'panel editor apply',
    toggleVizPicker: 'toggle-viz-picker',
    toggleVizOptions: 'toggle-viz-options',
    toggleTableView: 'toggle-table-view',
  },
  PanelInspector: {
    Data: {
      content: 'Panel inspector Data content',
    },
    Stats: {
      content: 'Panel inspector Stats content',
    },
    Json: {
      content: 'Panel inspector Json content',
    },
    Query: {
      content: 'Panel inspector Query content',
      refreshButton: 'Panel inspector Query refresh button',
      jsonObjectKeys: () => '.json-formatter-key',
    },
  },
  Tab: {
    title: (title: string) => `Tab ${title}`,
    active: () => '[class*="-activeTabStyle"]',
  },
  RefreshPicker: {
    runButton: 'RefreshPicker run button',
    intervalButton: 'RefreshPicker interval button',
  },
  QueryTab: {
    content: 'Query editor tab content',
    queryInspectorButton: 'Query inspector button',
    addQuery: 'Query editor add query button',
  },
  QueryEditorRows: {
    rows: 'Query editor row',
  },
  QueryEditorRow: {
    actionButton: (title: string) => `${title} query operation action`,
    title: (refId: string) => `Query editor row title ${refId}`,
  },
  AlertTab: {
    content: 'Alert editor tab content',
  },
  Alert: {
    alert: (severity: string) => `Alert ${severity}`,
  },
  TransformTab: {
    content: 'Transform editor tab content',
    newTransform: (name: string) => `New transform ${name}`,
    transformationEditor: (name: string) => `Transformation editor ${name}`,
    transformationEditorDebugger: (name: string) => `Transformation editor debugger ${name}`,
  },
  Transforms: {
    card: (name: string) => `New transform ${name}`,
    Reduce: {
      modeLabel: 'Transform mode label',
      calculationsLabel: 'Transform calculations label',
    },
    searchInput: 'search transformations',
  },
  PageToolbar: {
    container: () => '.page-toolbar',
    item: (tooltip: string) => `${tooltip}`,
  },
  QueryEditorToolbarItem: {
    button: (title: string) => `QueryEditor toolbar item button ${title}`,
  },
  BackButton: {
    backArrow: 'Go Back',
  },
  OptionsGroup: {
    group: (title?: string) => (title ? `Options group ${title}` : 'Options group'),
    toggle: (title?: string) => (title ? `Options group ${title} toggle` : 'Options group toggle'),
  },
  PluginVisualization: {
    item: (title: string) => `Plugin visualization item ${title}`,
    current: () => '[class*="-currentVisualizationItem"]',
  },
  Select: {
    option: 'Select option',
    input: () => 'input[id*="react-select-"]',
    singleValue: () => 'div[class*="-singleValue"]',
  },
  FieldConfigEditor: {
    content: 'Field config editor content',
  },
  OverridesConfigEditor: {
    content: 'Field overrides editor content',
  },
  FolderPicker: {
    container: 'Folder picker select container',
  },
  ReadonlyFolderPicker: {
    container: 'data-testid Readonly folder picker select container',
  },
  DataSourcePicker: {
    container: 'Data source picker select container',
    input: () => 'input[id="data-source-picker"]',
  },
  TimeZonePicker: {
    container: 'Time zone picker select container',
  },
  TraceViewer: {
    spanBar: () => '[data-test-id="SpanBar--wrapper"]',
  },
  QueryField: { container: 'Query field' },
  ValuePicker: {
    button: (name: string) => `Value picker button ${name}`,
    select: (name: string) => `Value picker select ${name}`,
  },
  Search: {
    section: 'Search section',
    items: 'Search items',
  },
  DashboardLinks: {
    container: 'data-testid Dashboard link container',
    dropDown: 'data-testid Dashboard link dropdown',
    link: 'data-testid Dashboard link',
  },
  LoadingIndicator: {
    icon: 'Loading indicator',
  },
  CallToActionCard: {
    button: (name: string) => `Call to action button ${name}`,
  },
  DataLinksContextMenu: {
    singleLink: 'Data link',
  },
  CodeEditor: {
    container: 'Code editor container',
  },
  DashboardImportPage: {
    textarea: 'data-testid-import-dashboard-textarea',
    submit: 'data-testid-load-dashboard',
  },
  ImportDashboardForm: {
    name: 'data-testid-import-dashboard-title',
    submit: 'data-testid-import-dashboard-submit',
  },
  PanelAlertTabContent: {
    content: 'Unified alert editor tab content',
  },
};
