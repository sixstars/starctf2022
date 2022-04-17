import React, { FC, useCallback, useMemo, useState } from 'react';
import { Field, GrafanaTheme, SelectableValue } from '@grafana/data';
import { css, cx } from '@emotion/css';

import { TableStyles } from './styles';
import { stylesFactory, useStyles, useTheme2 } from '../../themes';
import { Button, ClickOutsideWrapper, HorizontalGroup, IconButton, Label, VerticalGroup } from '..';
import { FilterList } from './FilterList';
import { calculateUniqueFieldValues, getFilteredOptions, valuesToOptions } from './utils';

interface Props {
  column: any;
  tableStyles: TableStyles;
  onClose: () => void;
  field?: Field;
}

export const FilterPopup: FC<Props> = ({ column: { preFilteredRows, filterValue, setFilter }, onClose, field }) => {
  const theme = useTheme2();
  const uniqueValues = useMemo(() => calculateUniqueFieldValues(preFilteredRows, field), [preFilteredRows, field]);
  const options = useMemo(() => valuesToOptions(uniqueValues), [uniqueValues]);
  const filteredOptions = useMemo(() => getFilteredOptions(options, filterValue), [options, filterValue]);
  const [values, setValues] = useState<SelectableValue[]>(filteredOptions);
  const [matchCase, setMatchCase] = useState(false);

  const onCancel = useCallback((event?: React.MouseEvent) => onClose(), [onClose]);

  const onFilter = useCallback(
    (event: React.MouseEvent) => {
      const filtered = values.length ? values : undefined;

      setFilter(filtered);
      onClose();
    },
    [setFilter, values, onClose]
  );

  const onClearFilter = useCallback(
    (event: React.MouseEvent) => {
      setFilter(undefined);
      onClose();
    },
    [setFilter, onClose]
  );

  const clearFilterVisible = useMemo(() => filterValue !== undefined, [filterValue]);
  const styles = useStyles(getStyles);

  return (
    <ClickOutsideWrapper onClick={onCancel} useCapture={true}>
      <div className={cx(styles.filterContainer)} onClick={stopPropagation}>
        <VerticalGroup spacing="lg">
          <VerticalGroup spacing="xs">
            <HorizontalGroup justify="space-between" align="center">
              <Label className={styles.label}>Filter by values:</Label>
              <IconButton
                name="text-fields"
                tooltip="Match case"
                style={{ color: matchCase ? theme.colors.text.link : theme.colors.text.disabled }}
                onClick={() => {
                  setMatchCase((s) => !s);
                }}
              />
            </HorizontalGroup>
            <div className={cx(styles.listDivider)} />
            <FilterList onChange={setValues} values={values} options={options} caseSensitive={matchCase} />
          </VerticalGroup>
          <HorizontalGroup spacing="lg">
            <HorizontalGroup>
              <Button size="sm" onClick={onFilter}>
                Ok
              </Button>
              <Button size="sm" variant="secondary" onClick={onCancel}>
                Cancel
              </Button>
            </HorizontalGroup>
            {clearFilterVisible && (
              <HorizontalGroup>
                <Button fill="text" size="sm" onClick={onClearFilter}>
                  Clear filter
                </Button>
              </HorizontalGroup>
            )}
          </HorizontalGroup>
        </VerticalGroup>
      </div>
    </ClickOutsideWrapper>
  );
};

const getStyles = stylesFactory((theme: GrafanaTheme) => ({
  filterContainer: css`
    label: filterContainer;
    width: 100%;
    min-width: 250px;
    height: 100%;
    max-height: 400px;
    background-color: ${theme.colors.bg1};
    border: ${theme.border.width.sm} solid ${theme.colors.border2};
    padding: ${theme.spacing.md};
    margin: ${theme.spacing.sm} 0;
    box-shadow: 0px 0px 20px ${theme.colors.dropdownShadow};
    border-radius: ${theme.spacing.xs};
  `,
  listDivider: css`
    label: listDivider;
    width: 100%;
    border-top: ${theme.border.width.sm} solid ${theme.colors.border2};
    padding: ${theme.spacing.xs} ${theme.spacing.md};
  `,
  label: css`
    margin-bottom: 0;
  `,
}));

const stopPropagation = (event: React.MouseEvent) => {
  event.stopPropagation();
};
