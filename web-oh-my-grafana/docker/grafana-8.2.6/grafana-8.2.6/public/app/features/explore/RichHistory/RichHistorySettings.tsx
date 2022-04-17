import React from 'react';
import { css } from '@emotion/css';
import { stylesFactory, useTheme, Select, Button, Switch, Field } from '@grafana/ui';
import { GrafanaTheme, SelectableValue } from '@grafana/data';
import appEvents from 'app/core/app_events';
import { ShowConfirmModalEvent } from '../../../types/events';
import { dispatch } from 'app/store/store';
import { notifyApp } from 'app/core/actions';
import { createSuccessNotification } from 'app/core/copy/appNotification';
import { MAX_HISTORY_ITEMS } from '../../../core/utils/richHistory';

export interface RichHistorySettingsProps {
  retentionPeriod: number;
  starredTabAsFirstTab: boolean;
  activeDatasourceOnly: boolean;
  onChangeRetentionPeriod: (option: SelectableValue<number>) => void;
  toggleStarredTabAsFirstTab: () => void;
  toggleactiveDatasourceOnly: () => void;
  deleteRichHistory: () => void;
}

const getStyles = stylesFactory((theme: GrafanaTheme) => {
  return {
    container: css`
      font-size: ${theme.typography.size.sm};
      .space-between {
        margin-bottom: ${theme.spacing.lg};
      }
    `,
    input: css`
      max-width: 200px;
    `,
    switch: css`
      display: flex;
      align-items: center;
    `,
    label: css`
      margin-left: ${theme.spacing.md};
    `,
  };
});

const retentionPeriodOptions = [
  { value: 2, label: '2 days' },
  { value: 5, label: '5 days' },
  { value: 7, label: '1 week' },
  { value: 14, label: '2 weeks' },
];

export function RichHistorySettings(props: RichHistorySettingsProps) {
  const {
    retentionPeriod,
    starredTabAsFirstTab,
    activeDatasourceOnly,
    onChangeRetentionPeriod,
    toggleStarredTabAsFirstTab,
    toggleactiveDatasourceOnly,
    deleteRichHistory,
  } = props;
  const theme = useTheme();
  const styles = getStyles(theme);
  const selectedOption = retentionPeriodOptions.find((v) => v.value === retentionPeriod);

  const onDelete = () => {
    appEvents.publish(
      new ShowConfirmModalEvent({
        title: 'Delete',
        text: 'Are you sure you want to permanently delete your query history?',
        yesText: 'Delete',
        icon: 'trash-alt',
        onConfirm: () => {
          deleteRichHistory();
          dispatch(notifyApp(createSuccessNotification('Query history deleted')));
        },
      })
    );
  };

  return (
    <div className={styles.container}>
      <Field
        label="History time span"
        description={`Select the period of time for which Grafana will save your query history. Up to ${MAX_HISTORY_ITEMS} entries will be stored.`}
        className="space-between"
      >
        <div className={styles.input}>
          <Select
            menuShouldPortal
            value={selectedOption}
            options={retentionPeriodOptions}
            onChange={onChangeRetentionPeriod}
          ></Select>
        </div>
      </Field>
      <Field label="Default active tab" description=" " className="space-between">
        <div className={styles.switch}>
          <Switch value={starredTabAsFirstTab} onChange={toggleStarredTabAsFirstTab}></Switch>
          <div className={styles.label}>Change the default active tab from “Query history” to “Starred”</div>
        </div>
      </Field>
      <Field label="Data source behaviour" description=" " className="space-between">
        <div className={styles.switch}>
          <Switch value={activeDatasourceOnly} onChange={toggleactiveDatasourceOnly}></Switch>
          <div className={styles.label}>Only show queries for data source currently active in Explore</div>
        </div>
      </Field>
      <div
        className={css`
          font-weight: ${theme.typography.weight.bold};
        `}
      >
        Clear query history
      </div>
      <div
        className={css`
          margin-bottom: ${theme.spacing.sm};
        `}
      >
        Delete all of your query history, permanently.
      </div>
      <Button variant="destructive" onClick={onDelete}>
        Clear query history
      </Button>
    </div>
  );
}
