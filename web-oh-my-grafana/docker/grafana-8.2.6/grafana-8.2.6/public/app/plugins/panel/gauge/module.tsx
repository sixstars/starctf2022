import { PanelPlugin } from '@grafana/data';
import { GaugePanel } from './GaugePanel';
import { GaugeOptions } from './types';
import { addOrientationOption, addStandardDataReduceOptions } from '../stat/types';
import { gaugePanelMigrationHandler, gaugePanelChangedHandler } from './GaugeMigrations';
import { commonOptionsBuilder } from '@grafana/ui';

export const plugin = new PanelPlugin<GaugeOptions>(GaugePanel)
  .useFieldConfig()
  .setPanelOptions((builder) => {
    addStandardDataReduceOptions(builder);
    addOrientationOption(builder);

    builder
      .addBooleanSwitch({
        path: 'showThresholdLabels',
        name: 'Show threshold labels',
        description: 'Render the threshold values around the gauge bar',
        defaultValue: false,
      })
      .addBooleanSwitch({
        path: 'showThresholdMarkers',
        name: 'Show threshold markers',
        description: 'Renders the thresholds as an outer bar',
        defaultValue: true,
      });

    commonOptionsBuilder.addTextSizeOptions(builder);
  })
  .setPanelChangeHandler(gaugePanelChangedHandler)
  .setMigrationHandler(gaugePanelMigrationHandler);
