import React, { FC, useState } from 'react';
import { useDispatch } from 'react-redux';
import { useLocation } from 'react-router-dom';
import { css } from '@emotion/css';
import { AppEvents, GrafanaTheme2, urlUtil } from '@grafana/data';
import { config } from '@grafana/runtime';
import { Button, ConfirmModal, ClipboardButton, HorizontalGroup, LinkButton, useStyles2 } from '@grafana/ui';
import { contextSrv } from 'app/core/services/context_srv';
import { appEvents } from 'app/core/core';
import { useIsRuleEditable } from '../../hooks/useIsRuleEditable';
import { Annotation } from '../../utils/constants';
import { getRulesSourceName, isCloudRulesSource } from '../../utils/datasource';
import { createExploreLink, createViewLink } from '../../utils/misc';
import * as ruleId from '../../utils/rule-id';
import { deleteRuleAction } from '../../state/actions';
import { CombinedRule, RulesSource } from 'app/types/unified-alerting';

interface Props {
  rule: CombinedRule;
  rulesSource: RulesSource;
}

export const RuleDetailsActionButtons: FC<Props> = ({ rule, rulesSource }) => {
  const dispatch = useDispatch();
  const location = useLocation();
  const style = useStyles2(getStyles);
  const { namespace, group, rulerRule } = rule;
  const [ruleToDelete, setRuleToDelete] = useState<CombinedRule>();

  const leftButtons: JSX.Element[] = [];
  const rightButtons: JSX.Element[] = [];

  const { isEditable } = useIsRuleEditable(getRulesSourceName(rulesSource), rulerRule);
  const returnTo = location.pathname + location.search;
  const isViewMode = inViewMode(location.pathname);

  const deleteRule = () => {
    if (ruleToDelete && ruleToDelete.rulerRule) {
      const identifier = ruleId.fromRulerRule(
        getRulesSourceName(ruleToDelete.namespace.rulesSource),
        ruleToDelete.namespace.name,
        ruleToDelete.group.name,
        ruleToDelete.rulerRule
      );

      dispatch(deleteRuleAction(identifier, { navigateTo: isViewMode ? '/alerting/list' : undefined }));
      setRuleToDelete(undefined);
    }
  };

  const buildShareUrl = () => {
    if (isCloudRulesSource(rulesSource)) {
      const ruleUrl = `${encodeURIComponent(rulesSource.name)}/${encodeURIComponent(rule.name)}`;
      return `${config.appUrl}${config.appSubUrl}/alerting/${ruleUrl}/find`;
    }

    return window.location.href.split('?')[0];
  };

  // explore does not support grafana rule queries atm
  if (isCloudRulesSource(rulesSource) && contextSrv.isEditor) {
    leftButtons.push(
      <LinkButton
        className={style.button}
        size="xs"
        key="explore"
        variant="primary"
        icon="chart-line"
        target="__blank"
        href={createExploreLink(rulesSource.name, rule.query)}
      >
        See graph
      </LinkButton>
    );
  }
  if (rule.annotations[Annotation.runbookURL]) {
    leftButtons.push(
      <LinkButton
        className={style.button}
        size="xs"
        key="runbook"
        variant="primary"
        icon="book"
        target="__blank"
        href={rule.annotations[Annotation.runbookURL]}
      >
        View runbook
      </LinkButton>
    );
  }
  if (rule.annotations[Annotation.dashboardUID]) {
    const dashboardUID = rule.annotations[Annotation.dashboardUID];
    if (dashboardUID) {
      leftButtons.push(
        <LinkButton
          className={style.button}
          size="xs"
          key="dashboard"
          variant="primary"
          icon="apps"
          target="__blank"
          href={`d/${encodeURIComponent(dashboardUID)}`}
        >
          Go to dashboard
        </LinkButton>
      );
      const panelId = rule.annotations[Annotation.panelID];
      if (panelId) {
        leftButtons.push(
          <LinkButton
            className={style.button}
            size="xs"
            key="dashboard"
            variant="primary"
            icon="apps"
            target="__blank"
            href={`d/${encodeURIComponent(dashboardUID)}?viewPanel=${encodeURIComponent(panelId)}`}
          >
            Go to panel
          </LinkButton>
        );
      }
    }
  }

  if (!isViewMode) {
    rightButtons.push(
      <LinkButton
        className={style.button}
        size="xs"
        key="view"
        variant="secondary"
        icon="eye"
        href={createViewLink(rulesSource, rule, returnTo)}
      >
        View
      </LinkButton>
    );
  }

  if (isEditable && rulerRule) {
    const sourceName = getRulesSourceName(rulesSource);
    const identifier = ruleId.fromRulerRule(sourceName, namespace.name, group.name, rulerRule);

    const editURL = urlUtil.renderUrl(
      `${config.appSubUrl}/alerting/${encodeURIComponent(ruleId.stringifyIdentifier(identifier))}/edit`,
      {
        returnTo,
      }
    );

    if (isViewMode) {
      rightButtons.push(
        <ClipboardButton
          onClipboardCopy={() => {
            appEvents.emit(AppEvents.alertSuccess, ['URL copied!']);
          }}
          onClipboardError={(e) => {
            appEvents.emit(AppEvents.alertError, ['Error while copying URL', e.text]);
          }}
          className={style.button}
          size="sm"
          getText={buildShareUrl}
        >
          Copy link to rule
        </ClipboardButton>
      );
    }

    rightButtons.push(
      <LinkButton className={style.button} size="xs" key="edit" variant="secondary" icon="pen" href={editURL}>
        Edit
      </LinkButton>,
      <Button
        className={style.button}
        size="xs"
        type="button"
        key="delete"
        variant="secondary"
        icon="trash-alt"
        onClick={() => setRuleToDelete(rule)}
      >
        Delete
      </Button>
    );
  }
  if (leftButtons.length || rightButtons.length) {
    return (
      <>
        <div className={style.wrapper}>
          <HorizontalGroup width="auto">{leftButtons.length ? leftButtons : <div />}</HorizontalGroup>
          <HorizontalGroup width="auto">{rightButtons.length ? rightButtons : <div />}</HorizontalGroup>
        </div>
        {!!ruleToDelete && (
          <ConfirmModal
            isOpen={true}
            title="Delete rule"
            body="Deleting this rule will permanently remove it from your alert rule list. Are you sure you want to delete this rule?"
            confirmText="Yes, delete"
            icon="exclamation-triangle"
            onConfirm={deleteRule}
            onDismiss={() => setRuleToDelete(undefined)}
          />
        )}
      </>
    );
  }

  return null;
};

function inViewMode(pathname: string): boolean {
  return pathname.endsWith('/view');
}

export const getStyles = (theme: GrafanaTheme2) => ({
  wrapper: css`
    padding: ${theme.spacing(2)} 0;
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    flex-wrap: wrap;
    border-bottom: solid 1px ${theme.colors.border.medium};
  `,
  button: css`
    height: 24px;
    margin-top: ${theme.spacing(1)};
    font-size: ${theme.typography.size.sm};
  `,
});
