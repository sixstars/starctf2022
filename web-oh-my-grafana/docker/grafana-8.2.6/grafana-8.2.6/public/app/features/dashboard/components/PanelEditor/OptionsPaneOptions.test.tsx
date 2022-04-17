import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import {
  FieldConfigSource,
  LoadingState,
  PanelData,
  standardEditorsRegistry,
  standardFieldConfigEditorRegistry,
} from '@grafana/data';

import { selectors } from '@grafana/e2e-selectors';
import { OptionsPaneOptions } from './OptionsPaneOptions';
import { DashboardModel, PanelModel } from '../../state';
import { Provider } from 'react-redux';
import configureMockStore from 'redux-mock-store';
import { getPanelPlugin } from 'app/features/plugins/__mocks__/pluginMocks';
import { getStandardFieldConfigs, getStandardOptionEditors } from '@grafana/ui';

standardEditorsRegistry.setInit(getStandardOptionEditors);
standardFieldConfigEditorRegistry.setInit(getStandardFieldConfigs);

const mockStore = configureMockStore<any, any>();
const OptionsPaneSelector = selectors.components.PanelEditor.OptionsPane;

class OptionsPaneOptionsTestScenario {
  onFieldConfigsChange = jest.fn();
  onPanelOptionsChanged = jest.fn();
  onPanelConfigChange = jest.fn();

  panelData: PanelData = {
    series: [],
    state: LoadingState.Done,
    timeRange: {} as any,
  };

  plugin = getPanelPlugin({
    id: 'TestPanel',
  }).useFieldConfig({
    standardOptions: {},
    useCustomConfig: (b) => {
      b.addBooleanSwitch({
        name: 'CustomBool',
        path: 'CustomBool',
      })
        .addBooleanSwitch({
          name: 'HiddenFromDef',
          path: 'HiddenFromDef',
          hideFromDefaults: true,
        })
        .addTextInput({
          name: 'TextPropWithCategory',
          path: 'TextPropWithCategory',
          settings: {
            placeholder: 'CustomTextPropPlaceholder',
          },
          category: ['Axis'],
        });
    },
  });

  panel = new PanelModel({
    title: 'Test title',
    type: this.plugin.meta.id,
    fieldConfig: {
      defaults: {
        max: 100,
        thresholds: {
          mode: 'absolute',
          steps: [
            { value: -Infinity, color: 'green' },
            { value: 100, color: 'green' },
          ],
        },
      },
      overrides: [],
    },
    options: {},
  });

  dashboard = new DashboardModel({});
  store = mockStore({
    dashboard: { panels: [] },
    templating: {
      variables: {},
    },
  });

  render() {
    render(
      <Provider store={this.store}>
        <OptionsPaneOptions
          data={this.panelData}
          plugin={this.plugin}
          panel={this.panel}
          dashboard={this.dashboard}
          onFieldConfigsChange={this.onFieldConfigsChange}
          onPanelConfigChange={this.onPanelConfigChange}
          onPanelOptionsChanged={this.onPanelOptionsChanged}
        />
      </Provider>
    );
  }
}

describe('OptionsPaneOptions', () => {
  it('should render panel frame options', async () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    expect(screen.getByLabelText(OptionsPaneSelector.fieldLabel('Panel options Title'))).toBeInTheDocument();
  });

  it('should render all categories', async () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    expect(screen.getByRole('heading', { name: /Panel options/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Standard options/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Thresholds/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /TestPanel/ })).toBeInTheDocument();
  });

  it('should render custom  options', () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    expect(screen.getByLabelText(OptionsPaneSelector.fieldLabel('TestPanel CustomBool'))).toBeInTheDocument();
  });

  it('should not render options that are marked as hidden from defaults', () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    expect(screen.queryByLabelText(OptionsPaneSelector.fieldLabel('TestPanel HiddenFromDef'))).not.toBeInTheDocument();
  });

  it('should create categories for field options with category', () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    expect(screen.getByRole('heading', { name: /Axis/ })).toBeInTheDocument();
  });

  it('should not render categories with hidden fields only', () => {
    const scenario = new OptionsPaneOptionsTestScenario();

    scenario.plugin = getPanelPlugin({
      id: 'TestPanel',
    }).useFieldConfig({
      standardOptions: {},
      useCustomConfig: (b) => {
        b.addBooleanSwitch({
          name: 'CustomBool',
          path: 'CustomBool',
          hideFromDefaults: true,
          category: ['Axis'],
        });
      },
    });

    scenario.render();
    expect(screen.queryByRole('heading', { name: /Axis/ })).not.toBeInTheDocument();
  });

  it('should call onPanelConfigChange when updating title', () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    const input = screen.getByDisplayValue(scenario.panel.title);
    fireEvent.change(input, { target: { value: 'New' } });
    fireEvent.blur(input);

    expect(scenario.onPanelConfigChange).toHaveBeenCalledWith('title', 'New');
  });

  it('should call onFieldConfigsChange when updating field config', () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    const input = screen.getByPlaceholderText('CustomTextPropPlaceholder');
    fireEvent.change(input, { target: { value: 'New' } });
    fireEvent.blur(input);

    const newFieldConfig: FieldConfigSource = scenario.panel.fieldConfig;
    newFieldConfig.defaults.custom = { TextPropWithCategory: 'New' };

    expect(scenario.onFieldConfigsChange).toHaveBeenCalledWith(newFieldConfig);
  });

  it('should only render hits when search query specified', async () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.render();

    const input = screen.getByPlaceholderText('Search options');
    fireEvent.change(input, { target: { value: 'TextPropWithCategory' } });
    fireEvent.blur(input);

    expect(screen.queryByLabelText(OptionsPaneSelector.fieldLabel('Panel options Title'))).not.toBeInTheDocument();
    expect(screen.getByLabelText(OptionsPaneSelector.fieldLabel('Axis TextPropWithCategory'))).toBeInTheDocument();
  });

  it('should not render field override options non data panel', async () => {
    const scenario = new OptionsPaneOptionsTestScenario();
    scenario.plugin = getPanelPlugin({
      id: 'TestPanel',
    });

    scenario.render();

    expect(
      screen.queryByLabelText(selectors.components.ValuePicker.button('Add field override'))
    ).not.toBeInTheDocument();
  });

  it('should allow standard properties extension', async () => {
    const scenario = new OptionsPaneOptionsTestScenario();

    scenario.plugin = getPanelPlugin({
      id: 'TestPanel',
    }).useFieldConfig({
      useCustomConfig: (b) => {
        b.addBooleanSwitch({
          name: 'CustomThresholdOption',
          path: 'CustomThresholdOption',
          category: ['Thresholds'],
        });
      },
    });

    scenario.render();

    const thresholdsSection = screen.getByLabelText(selectors.components.OptionsGroup.group('Thresholds'));
    expect(
      within(thresholdsSection).getByLabelText(OptionsPaneSelector.fieldLabel('Thresholds CustomThresholdOption'))
    ).toBeInTheDocument();
  });
});
