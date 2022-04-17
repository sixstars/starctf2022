import React from 'react';
import { FieldConfigEditorProps, SliderFieldConfigSettings } from '@grafana/data';
import { Slider } from '../Slider/Slider';

export const SliderValueEditor: React.FC<FieldConfigEditorProps<number, SliderFieldConfigSettings>> = ({
  value,
  onChange,
  item,
}) => {
  const { settings } = item;
  const initialValue = typeof value === 'number' ? value : typeof value === 'string' ? +value : 0;

  return (
    <Slider
      value={initialValue}
      min={settings?.min || 0}
      max={settings?.max || 100}
      step={settings?.step}
      onChange={onChange}
    />
  );
};
