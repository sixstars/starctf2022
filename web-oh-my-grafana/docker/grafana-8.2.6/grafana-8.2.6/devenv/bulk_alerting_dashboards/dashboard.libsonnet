{
  alertingDashboard(dashboardCounter, datasourceCounter):: {
    title: "alerting-title-" + dashboardCounter,
    editable: true,
    gnetId: null,
    graphTooltip: 0,
    id: null,
    links: [],
    panels: [
      {
        alert: {
          conditions: [
            {
              evaluator: {
                params: [
                  65
                ],
                type: "gt"
              },
              operator: {
                type: "and"
              },
              query: {
                params: [
                  "A",
                  "5m",
                  "now"
                ]
              },
              reducer: {
                params: [],
                type: "avg"
              },
              type: "query"
            }
          ],
          executionErrorState: "alerting",
          frequency: "24h",
          handler: 1,
          name: "bulk alerting " + dashboardCounter,
          noDataState: "no_data",
          notifications: []
        },
        aliasColors: {},
        bars: false,
        dashLength: 10,
        dashes: false,
        datasource: "gfdev-bulkalerting-" + datasourceCounter,
        fill: 1,
        gridPos: {
          h: 9,
          w: 12,
          x: 0,
          y: 0
        },
        id: 1,
        legend: {
          avg: false,
          current: false,
          max: false,
          min: false,
          show: true,
          total: false,
          values: false
        },
        lines: true,
        linewidth: 1,
        nullPointMode: "null",
        percentage: false,
        pointradius: 5,
        points: false,
        renderer: "flot",
        seriesOverrides: [],
        spaceLength: 10,
        stack: false,
        steppedLine: false,
        targets: [
          {
            expr: "go_goroutines",
            format: "time_series",
            intervalFactor: 1,
            refId: "A"
          }
        ],
        thresholds: [
          {
            colorMode: "critical",
            fill: true,
            line: true,
            op: "gt",
            value: 50
          }
        ],
        timeFrom: null,
        timeShift: null,
        title: "Panel Title",
        tooltip: {
          shared: true,
          sort: 0,
          value_type: "individual"
        },
        type: "graph",
        xaxis: {
          buckets: null,
          mode: "time",
          name: null,
          show: true,
          values: []
        },
        yaxes: [
          {
            format: "short",
            label: null,
            logBase: 1,
            max: null,
            min: null,
            show: true
          },
          {
            format: "short",
            label: null,
            logBase: 1,
            max: null,
            min: null,
            show: true
          }
        ]
      }
    ],
    schemaVersion: 16,
    style: "dark",
    tags: [],
    templating: {
      list: []
    },
    time: {
      from: "now-6h",
      to: "now"
    },
    timepicker: {
      refresh_intervals: [
        "5s",
        "10s",
        "30s",
        "1m",
        "5m",
        "15m",
        "30m",
        "1h",
        "2h",
        "1d"
      ],
      time_options: [
        "5m",
        "15m",
        "1h",
        "6h",
        "12h",
        "24h",
        "2d",
        "7d",
        "30d"
      ]
    },
    timezone: "",
    uid: null,
    version: 0
  },
}

