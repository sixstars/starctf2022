import React, { useState, useCallback } from 'react';
import { Meta, Story } from '@storybook/react';
import { withCenteredStory, withHorizontallyCenteredStory } from '../../utils/storybook/withCenteredStory';
import { InlineField, Switch, InlineSwitch } from '@grafana/ui';
import mdx from './Switch.mdx';
import { InlineFieldRow } from '../Forms/InlineFieldRow';
import { Field } from '../Forms/Field';

export default {
  title: 'Forms/Switch',
  component: Switch,
  decorators: [withCenteredStory, withHorizontallyCenteredStory],
  parameters: {
    docs: {
      page: mdx,
    },
  },
  args: {
    disabled: false,
    value: false,
    transparent: false,
  },
} as Meta;

export const Controlled: Story = (args) => {
  return (
    <div>
      <div style={{ marginBottom: '32px' }}>
        <Field label="Normal switch" description="For horizontal forms">
          <Switch value={args.value} disabled={args.disabled} transparent={args.transparent} />
        </Field>
      </div>
      <div style={{ marginBottom: '32px' }}>
        <InlineFieldRow>
          <InlineField label="My switch">
            <InlineSwitch value={args.value} disabled={args.disabled} transparent={args.transparent} />
          </InlineField>
        </InlineFieldRow>
      </div>
      <div style={{ marginBottom: '32px' }}>
        <div>just inline switch with show label</div>
        <span>
          <InlineSwitch
            label="Raw data"
            showLabel={true}
            value={args.value}
            disabled={args.disabled}
            transparent={args.transparent}
          />
        </span>
      </div>
    </div>
  );
};

export const Uncontrolled: Story = (args) => {
  const [checked, setChecked] = useState(args.value);
  const onChange = useCallback((e) => setChecked(e.currentTarget.checked), [setChecked]);
  return <Switch value={checked} disabled={args.disabled} transparent={args.transparent} onChange={onChange} />;
};
