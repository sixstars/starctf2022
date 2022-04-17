// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from 'react';
import IoAlert from 'react-icons/lib/io/alert';
import IoArrowRightA from 'react-icons/lib/io/arrow-right-a';
import IoNetwork from 'react-icons/lib/io/network';
import MdFileUpload from 'react-icons/lib/md/file-upload';
import { css } from '@emotion/css';
import cx from 'classnames';

import ReferencesButton from './ReferencesButton';
import TimelineRow from './TimelineRow';
import { formatDuration, ViewedBoundsFunctionType } from './utils';
import SpanTreeOffset from './SpanTreeOffset';
import SpanBar from './SpanBar';
import Ticks from './Ticks';

import { TNil } from '../types';
import { TraceSpan } from '../types/trace';
import { autoColor, createStyle, Theme, withTheme } from '../Theme';

const spanBarClassName = 'spanBar';
const spanBarLabelClassName = 'spanBarLabel';
const nameWrapperClassName = 'nameWrapper';
const nameWrapperMatchingFilterClassName = 'nameWrapperMatchingFilter';
const viewClassName = 'jaegerView';
const nameColumnClassName = 'nameColumn';

const getStyles = createStyle((theme: Theme) => {
  return {
    nameWrapper: css`
      label: nameWrapper;
      line-height: 27px;
      overflow: hidden;
      display: flex;
    `,
    nameWrapperMatchingFilter: css`
      label: nameWrapperMatchingFilter;
      background-color: ${autoColor(theme, '#fffce4')};
    `,
    nameColumn: css`
      label: nameColumn;
      position: relative;
      white-space: nowrap;
      z-index: 1;
      &:hover {
        z-index: 1;
      }
    `,
    endpointName: css`
      label: endpointName;
      color: ${autoColor(theme, '#808080')};
    `,
    view: css`
      label: view;
      position: relative;
    `,
    viewExpanded: css`
      label: viewExpanded;
      background: ${autoColor(theme, '#f8f8f8')};
      outline: 1px solid ${autoColor(theme, '#ddd')};
    `,
    viewExpandedAndMatchingFilter: css`
      label: viewExpandedAndMatchingFilter;
      background: ${autoColor(theme, '#fff3d7')};
      outline: 1px solid ${autoColor(theme, '#ddd')};
    `,
    row: css`
      label: row;
      &:hover .${spanBarClassName} {
        opacity: 1;
      }
      &:hover .${spanBarLabelClassName} {
        color: ${autoColor(theme, '#000')};
      }
      &:hover .${nameWrapperClassName} {
        background: #f8f8f8;
        background: linear-gradient(
          90deg,
          ${autoColor(theme, '#fafafa')},
          ${autoColor(theme, '#f8f8f8')} 75%,
          ${autoColor(theme, '#eee')}
        );
      }
      &:hover .${viewClassName} {
        background-color: ${autoColor(theme, '#f5f5f5')};
        outline: 1px solid ${autoColor(theme, '#ddd')};
      }
    `,
    rowClippingLeft: css`
      label: rowClippingLeft;
      & .${nameColumnClassName}::before {
        content: ' ';
        height: 100%;
        position: absolute;
        width: 6px;
        background-image: linear-gradient(
          to right,
          ${autoColor(theme, 'rgba(25, 25, 25, 0.25)')},
          ${autoColor(theme, 'rgba(32, 32, 32, 0)')}
        );
        left: 100%;
        z-index: -1;
      }
    `,
    rowClippingRight: css`
      label: rowClippingRight;
      & .${viewClassName}::before {
        content: ' ';
        height: 100%;
        position: absolute;
        width: 6px;
        background-image: linear-gradient(
          to left,
          ${autoColor(theme, 'rgba(25, 25, 25, 0.25)')},
          ${autoColor(theme, 'rgba(25, 25, 25, 0.25)')}
        );
        right: 0%;
        z-index: 1;
      }
    `,
    rowExpanded: css`
      label: rowExpanded;
      & .${spanBarClassName} {
        opacity: 1;
      }
      & .${spanBarLabelClassName} {
        color: ${autoColor(theme, '#000')};
      }
      & .${nameWrapperClassName}, &:hover .${nameWrapperClassName} {
        background: ${autoColor(theme, '#f0f0f0')};
        box-shadow: 0 1px 0 ${autoColor(theme, '#ddd')};
      }
      & .${nameWrapperMatchingFilterClassName} {
        background: ${autoColor(theme, '#fff3d7')};
      }
      &:hover .${viewClassName} {
        background: ${autoColor(theme, '#eee')};
      }
    `,
    rowMatchingFilter: css`
      label: rowMatchingFilter;
      background-color: ${autoColor(theme, '#fffce4')};
      &:hover .${nameWrapperClassName} {
        background: linear-gradient(
          90deg,
          ${autoColor(theme, '#fff5e1')},
          ${autoColor(theme, '#fff5e1')} 75%,
          ${autoColor(theme, '#ffe6c9')}
        );
      }
      &:hover .${viewClassName} {
        background-color: ${autoColor(theme, '#fff3d7')};
        outline: 1px solid ${autoColor(theme, '#ddd')};
      }
    `,

    rowExpandedAndMatchingFilter: css`
      label: rowExpandedAndMatchingFilter;
      &:hover .${viewClassName} {
        background: ${autoColor(theme, '#ffeccf')};
      }
    `,

    name: css`
      label: name;
      color: ${autoColor(theme, '#000')};
      cursor: pointer;
      flex: 1 1 auto;
      outline: none;
      overflow: hidden;
      padding-left: 4px;
      padding-right: 0.25em;
      position: relative;
      text-overflow: ellipsis;
      &::before {
        content: ' ';
        position: absolute;
        top: 4px;
        bottom: 4px;
        left: 0;
        border-left: 4px solid;
        border-left-color: inherit;
      }

      /* This is so the hit area of the span-name extends the rest of the width of the span-name column */
      &::after {
        background: transparent;
        bottom: 0;
        content: ' ';
        left: 0;
        position: absolute;
        top: 0;
        width: 1000px;
      }
      &:focus {
        text-decoration: none;
      }
      &:hover > small {
        color: ${autoColor(theme, '#000')};
      }
    `,
    nameDetailExpanded: css`
      label: nameDetailExpanded;
      &::before {
        bottom: 0;
      }
    `,
    svcName: css`
      label: svcName;
      padding: 0 0.25rem 0 0.5rem;
      font-size: 1.05em;
    `,
    svcNameChildrenCollapsed: css`
      label: svcNameChildrenCollapsed;
      font-weight: bold;
      font-style: italic;
    `,
    errorIcon: css`
      label: errorIcon;
      border-radius: 6.5px;
      color: ${autoColor(theme, '#fff')};
      font-size: 0.85em;
      margin-right: 0.25rem;
      padding: 1px;
    `,
    rpcColorMarker: css`
      label: rpcColorMarker;
      border-radius: 6.5px;
      display: inline-block;
      font-size: 0.85em;
      height: 1em;
      margin-right: 0.25rem;
      padding: 1px;
      width: 1em;
      vertical-align: middle;
    `,
    labelRight: css`
      label: labelRight;
      left: 100%;
    `,
    labelLeft: css`
      label: labelLeft;
      right: 100%;
    `,
  };
});

type SpanBarRowProps = {
  className?: string;
  theme: Theme;
  color: string;
  columnDivision: number;
  isChildrenExpanded: boolean;
  isDetailExpanded: boolean;
  isMatchingFilter: boolean;
  onDetailToggled: (spanID: string) => void;
  onChildrenToggled: (spanID: string) => void;
  numTicks: number;
  rpc?:
    | {
        viewStart: number;
        viewEnd: number;
        color: string;
        operationName: string;
        serviceName: string;
      }
    | TNil;
  noInstrumentedServer?:
    | {
        color: string;
        serviceName: string;
      }
    | TNil;
  showErrorIcon: boolean;
  getViewedBounds: ViewedBoundsFunctionType;
  traceStartTime: number;
  span: TraceSpan;
  focusSpan: (spanID: string) => void;
  hoverIndentGuideIds: Set<string>;
  addHoverIndentGuideId: (spanID: string) => void;
  removeHoverIndentGuideId: (spanID: string) => void;
  clippingLeft?: boolean;
  clippingRight?: boolean;
  createSpanLink?: (
    span: TraceSpan
  ) => { href: string; onClick?: (e: React.MouseEvent) => void; content: React.ReactNode };
};

/**
 * This was originally a stateless function, but changing to a PureComponent
 * reduced the render time of expanding a span row detail by ~50%. This is
 * even true in the case where the stateless function has the same prop types as
 * this class and arrow functions are created in the stateless function as
 * handlers to the onClick props. E.g. for now, the PureComponent is more
 * performance than the stateless function.
 */
export class UnthemedSpanBarRow extends React.PureComponent<SpanBarRowProps> {
  static displayName = 'UnthemedSpanBarRow';
  static defaultProps: Partial<SpanBarRowProps> = {
    className: '',
    rpc: null,
  };

  _detailToggle = () => {
    this.props.onDetailToggled(this.props.span.spanID);
  };

  _childrenToggle = () => {
    this.props.onChildrenToggled(this.props.span.spanID);
  };

  render() {
    const {
      className,
      color,
      columnDivision,
      isChildrenExpanded,
      isDetailExpanded,
      isMatchingFilter,
      numTicks,
      rpc,
      noInstrumentedServer,
      showErrorIcon,
      getViewedBounds,
      traceStartTime,
      span,
      focusSpan,
      hoverIndentGuideIds,
      addHoverIndentGuideId,
      removeHoverIndentGuideId,
      clippingLeft,
      clippingRight,
      theme,
      createSpanLink,
    } = this.props;
    const {
      duration,
      hasChildren: isParent,
      operationName,
      process: { serviceName },
    } = span;
    const label = formatDuration(duration);
    const viewBounds = getViewedBounds(span.startTime, span.startTime + span.duration);
    const viewStart = viewBounds.start;
    const viewEnd = viewBounds.end;
    const styles = getStyles(theme);

    const labelDetail = `${serviceName}::${operationName}`;
    let longLabel;
    let hintClassName;
    if (viewStart > 1 - viewEnd) {
      longLabel = `${labelDetail} | ${label}`;
      hintClassName = styles.labelLeft;
    } else {
      longLabel = `${label} | ${labelDetail}`;
      hintClassName = styles.labelRight;
    }

    return (
      <TimelineRow
        className={cx(
          styles.row,
          {
            [styles.rowExpanded]: isDetailExpanded,
            [styles.rowMatchingFilter]: isMatchingFilter,
            [styles.rowExpandedAndMatchingFilter]: isMatchingFilter && isDetailExpanded,
            [styles.rowClippingLeft]: clippingLeft,
            [styles.rowClippingRight]: clippingRight,
          },
          className
        )}
      >
        <TimelineRow.Cell className={cx(styles.nameColumn, nameColumnClassName)} width={columnDivision}>
          <div
            className={cx(styles.nameWrapper, nameWrapperClassName, {
              [styles.nameWrapperMatchingFilter]: isMatchingFilter,
              nameWrapperMatchingFilter: isMatchingFilter,
            })}
          >
            <SpanTreeOffset
              childrenVisible={isChildrenExpanded}
              span={span}
              onClick={isParent ? this._childrenToggle : undefined}
              hoverIndentGuideIds={hoverIndentGuideIds}
              addHoverIndentGuideId={addHoverIndentGuideId}
              removeHoverIndentGuideId={removeHoverIndentGuideId}
            />
            <a
              className={cx(styles.name, { [styles.nameDetailExpanded]: isDetailExpanded })}
              aria-checked={isDetailExpanded}
              title={labelDetail}
              onClick={this._detailToggle}
              role="switch"
              style={{ borderColor: color }}
              tabIndex={0}
            >
              <span
                className={cx(styles.svcName, {
                  [styles.svcNameChildrenCollapsed]: isParent && !isChildrenExpanded,
                })}
              >
                {showErrorIcon && (
                  <IoAlert
                    style={{
                      backgroundColor: span.errorIconColor
                        ? autoColor(theme, span.errorIconColor)
                        : autoColor(theme, '#db2828'),
                    }}
                    className={styles.errorIcon}
                  />
                )}
                {serviceName}{' '}
                {rpc && (
                  <span>
                    <IoArrowRightA /> <i className={styles.rpcColorMarker} style={{ background: rpc.color }} />
                    {rpc.serviceName}
                  </span>
                )}
                {noInstrumentedServer && (
                  <span>
                    <IoArrowRightA />{' '}
                    <i className={styles.rpcColorMarker} style={{ background: noInstrumentedServer.color }} />
                    {noInstrumentedServer.serviceName}
                  </span>
                )}
              </span>
              <small className={styles.endpointName}>{rpc ? rpc.operationName : operationName}</small>
              <small className={styles.endpointName}> | {label}</small>
            </a>
            {createSpanLink &&
              (() => {
                const link = createSpanLink(span);
                return (
                  <a
                    href={link.href}
                    // Needs to have target otherwise preventDefault would not work due to angularRouter.
                    target={'_blank'}
                    style={{ marginRight: '5px' }}
                    rel="noopener noreferrer"
                    onClick={
                      link.onClick
                        ? (event) => {
                            if (!(event.ctrlKey || event.metaKey || event.shiftKey) && link.onClick) {
                              event.preventDefault();
                              link.onClick(event);
                            }
                          }
                        : undefined
                    }
                  >
                    {link.content}
                  </a>
                );
              })()}

            {span.references && span.references.length > 1 && (
              <ReferencesButton
                references={span.references}
                tooltipText="Contains multiple references"
                focusSpan={focusSpan}
              >
                <IoNetwork />
              </ReferencesButton>
            )}
            {span.subsidiarilyReferencedBy && span.subsidiarilyReferencedBy.length > 0 && (
              <ReferencesButton
                references={span.subsidiarilyReferencedBy}
                tooltipText={`This span is referenced by ${
                  span.subsidiarilyReferencedBy.length === 1 ? 'another span' : 'multiple other spans'
                }`}
                focusSpan={focusSpan}
              >
                <MdFileUpload />
              </ReferencesButton>
            )}
          </div>
        </TimelineRow.Cell>
        <TimelineRow.Cell
          className={cx(styles.view, viewClassName, {
            [styles.viewExpanded]: isDetailExpanded,
            [styles.viewExpandedAndMatchingFilter]: isMatchingFilter && isDetailExpanded,
          })}
          data-test-id="span-view"
          style={{ cursor: 'pointer' }}
          width={1 - columnDivision}
          onClick={this._detailToggle}
        >
          <Ticks numTicks={numTicks} />
          <SpanBar
            rpc={rpc}
            viewStart={viewStart}
            viewEnd={viewEnd}
            theme={theme}
            getViewedBounds={getViewedBounds}
            color={color}
            shortLabel={label}
            longLabel={longLabel}
            traceStartTime={traceStartTime}
            span={span}
            labelClassName={`${spanBarLabelClassName} ${hintClassName}`}
            className={spanBarClassName}
          />
        </TimelineRow.Cell>
      </TimelineRow>
    );
  }
}

export default withTheme(UnthemedSpanBarRow);
