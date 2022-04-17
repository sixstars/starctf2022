import React, { useState, useMemo, useEffect, useRef } from 'react';
import { applyFieldOverrides, FieldConfigSource, getTimeZone, PanelData, PanelPlugin } from '@grafana/data';
import { PanelRendererProps } from '@grafana/runtime';
import { appEvents } from 'app/core/core';
import { useAsync } from 'react-use';
import { getPanelOptionsWithDefaults, OptionDefaults } from '../dashboard/state/getPanelOptionsWithDefaults';
import { importPanelPlugin } from '../plugins/plugin_loader';
import { useTheme2 } from '@grafana/ui';
const defaultFieldConfig = { defaults: {}, overrides: [] };
export function PanelRenderer<P extends object = any, F extends object = any>(props: PanelRendererProps<P, F>) {
  const {
    pluginId,
    data,
    timeZone = getTimeZone(),
    options = {},
    width,
    height,
    title,
    onOptionsChange = () => {},
    onChangeTimeRange = () => {},
    fieldConfig: externalFieldConfig = defaultFieldConfig,
  } = props;

  const [localFieldConfig, setFieldConfig] = useState(externalFieldConfig);
  const { value: plugin, error, loading } = useAsync(() => importPanelPlugin(pluginId), [pluginId]);
  const optionsWithDefaults = useOptionDefaults(plugin, options, localFieldConfig);
  const dataWithOverrides = useFieldOverrides(plugin, optionsWithDefaults, data, timeZone);

  useEffect(() => {
    setFieldConfig((lfc) => ({ ...lfc, ...externalFieldConfig }));
  }, [externalFieldConfig]);

  if (error) {
    return <div>Failed to load plugin: {error.message}</div>;
  }

  if (pluginIsLoading(loading, plugin, pluginId)) {
    return <div>Loading plugin panel...</div>;
  }

  if (!plugin || !plugin.panel) {
    return <div>Seems like the plugin you are trying to load does not have a panel component.</div>;
  }

  if (!dataWithOverrides) {
    return <div>No panel data</div>;
  }

  const PanelComponent = plugin.panel;

  return (
    <PanelComponent
      id={1}
      data={dataWithOverrides}
      title={title}
      timeRange={dataWithOverrides.timeRange}
      timeZone={timeZone}
      options={optionsWithDefaults!.options}
      fieldConfig={localFieldConfig}
      transparent={false}
      width={width}
      height={height}
      renderCounter={0}
      replaceVariables={(str: string) => str}
      onOptionsChange={onOptionsChange}
      onFieldConfigChange={setFieldConfig}
      onChangeTimeRange={onChangeTimeRange}
      eventBus={appEvents}
    />
  );
}

function useOptionDefaults<P extends object = any, F extends object = any>(
  plugin: PanelPlugin | undefined,
  options: P,
  fieldConfig: FieldConfigSource<F>
): OptionDefaults | undefined {
  return useMemo(() => {
    if (!plugin) {
      return;
    }

    return getPanelOptionsWithDefaults({
      plugin,
      currentOptions: options,
      currentFieldConfig: fieldConfig,
      isAfterPluginChange: false,
    });
  }, [plugin, fieldConfig, options]);
}

function useFieldOverrides(
  plugin: PanelPlugin | undefined,
  defaultOptions: OptionDefaults | undefined,
  data: PanelData | undefined,
  timeZone: string
): PanelData | undefined {
  const fieldConfig = defaultOptions?.fieldConfig;
  const series = data?.series;
  const fieldConfigRegistry = plugin?.fieldConfigRegistry;
  const theme = useTheme2();
  const structureRev = useRef(0);

  return useMemo(() => {
    if (!fieldConfigRegistry || !fieldConfig || !data) {
      return;
    }
    structureRev.current = structureRev.current + 1;

    return {
      ...data,
      series: applyFieldOverrides({
        data: series,
        fieldConfig,
        fieldConfigRegistry,
        replaceVariables: (str: string) => str,
        theme,
        timeZone,
      }),
      structureRev: structureRev.current,
    };
  }, [fieldConfigRegistry, fieldConfig, data, series, timeZone, theme]);
}

function pluginIsLoading(loading: boolean, plugin: PanelPlugin<any, any> | undefined, pluginId: string) {
  return loading || plugin?.meta.id !== pluginId;
}
