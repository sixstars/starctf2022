import React, { ComponentType, PureComponent } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { ClickOutsideWrapper } from '@grafana/ui';
import { LoadingState } from '@grafana/data';

import { StoreState } from 'app/types';
import { VariableInput } from '../shared/VariableInput';
import { commitChangesToVariable, filterOrSearchOptions, navigateOptions, openOptions } from './actions';
import { OptionsPickerState, toggleAllOptions, toggleOption } from './reducer';
import { VariableOption, VariableWithMultiSupport, VariableWithOptions } from '../../types';
import { VariableOptions } from '../shared/VariableOptions';
import { isMulti } from '../../guard';
import { VariablePickerProps } from '../types';
import { formatVariableLabel } from '../../shared/formatVariable';
import { toVariableIdentifier } from '../../state/types';
import { getVariableQueryRunner } from '../../query/VariableQueryRunner';
import { VariableLink } from '../shared/VariableLink';

export const optionPickerFactory = <Model extends VariableWithOptions | VariableWithMultiSupport>(): ComponentType<
  VariablePickerProps<Model>
> => {
  const mapDispatchToProps = {
    openOptions,
    commitChangesToVariable,
    filterOrSearchOptions,
    toggleAllOptions,
    toggleOption,
    navigateOptions,
  };

  const mapStateToProps = (state: StoreState) => ({
    picker: state.templating.optionsPicker,
  });

  const connector = connect(mapStateToProps, mapDispatchToProps);

  interface OwnProps extends VariablePickerProps<Model> {}

  type Props = OwnProps & ConnectedProps<typeof connector>;

  class OptionsPickerUnconnected extends PureComponent<Props> {
    onShowOptions = () =>
      this.props.openOptions(toVariableIdentifier(this.props.variable), this.props.onVariableChange);
    onHideOptions = () => this.props.commitChangesToVariable(this.props.onVariableChange);

    onToggleOption = (option: VariableOption, clearOthers: boolean) => {
      const toggleFunc =
        isMulti(this.props.variable) && this.props.variable.multi
          ? this.onToggleMultiValueVariable
          : this.onToggleSingleValueVariable;
      toggleFunc(option, clearOthers);
    };

    onToggleSingleValueVariable = (option: VariableOption, clearOthers: boolean) => {
      this.props.toggleOption({ option, clearOthers, forceSelect: false });
      this.onHideOptions();
    };

    onToggleMultiValueVariable = (option: VariableOption, clearOthers: boolean) => {
      this.props.toggleOption({ option, clearOthers, forceSelect: false });
    };

    render() {
      const { variable, picker } = this.props;
      const showOptions = picker.id === variable.id;

      return (
        <div className="variable-link-wrapper">
          {showOptions ? this.renderOptions(picker) : this.renderLink(variable)}
        </div>
      );
    }

    renderLink(variable: VariableWithOptions) {
      const linkText = formatVariableLabel(variable);
      const loading = variable.state === LoadingState.Loading;

      return (
        <VariableLink
          id={variable.id}
          text={linkText}
          onClick={this.onShowOptions}
          loading={loading}
          onCancel={this.onCancel}
        />
      );
    }

    onCancel = () => {
      getVariableQueryRunner().cancelRequest(toVariableIdentifier(this.props.variable));
    };

    renderOptions(picker: OptionsPickerState) {
      const { id } = this.props.variable;
      return (
        <ClickOutsideWrapper onClick={this.onHideOptions}>
          <VariableInput
            id={id}
            value={picker.queryValue}
            onChange={this.props.filterOrSearchOptions}
            onNavigate={this.props.navigateOptions}
            aria-expanded={true}
            aria-controls={`options-${id}`}
          />
          <VariableOptions
            values={picker.options}
            onToggle={this.onToggleOption}
            onToggleAll={this.props.toggleAllOptions}
            highlightIndex={picker.highlightIndex}
            multi={picker.multi}
            selectedValues={picker.selectedValues}
            id={`options-${id}`}
          />
        </ClickOutsideWrapper>
      );
    }
  }

  const OptionsPicker = connector(OptionsPickerUnconnected);
  OptionsPicker.displayName = 'OptionsPicker';

  return OptionsPicker;
};
