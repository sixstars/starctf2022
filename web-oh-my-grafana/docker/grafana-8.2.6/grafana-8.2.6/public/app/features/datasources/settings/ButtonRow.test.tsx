import React from 'react';
import { shallow } from 'enzyme';
import ButtonRow, { Props } from './ButtonRow';

jest.mock('app/core/core', () => {
  return {
    contextSrv: {
      hasPermission: () => true,
    },
  };
});

const setup = (propOverrides?: object) => {
  const props: Props = {
    isReadOnly: true,
    onSubmit: jest.fn(),
    onDelete: jest.fn(),
    onTest: jest.fn(),
    exploreUrl: '/explore',
  };

  Object.assign(props, propOverrides);

  return shallow(<ButtonRow {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();

    expect(wrapper).toMatchSnapshot();
  });

  it('should render with buttons enabled', () => {
    const wrapper = setup({
      isReadOnly: false,
    });

    expect(wrapper).toMatchSnapshot();
  });
});
