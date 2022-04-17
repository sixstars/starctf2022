import React from 'react';
import { FieldConfig, LinkModel } from '@grafana/data';
import { selectors } from '@grafana/e2e-selectors';
import { css } from '@emotion/css';
import { WithContextMenu } from '../ContextMenu/WithContextMenu';
import { linkModelToContextMenuItems } from '../../utils/dataLinks';
import { MenuGroup, MenuItemsGroup } from '../Menu/MenuGroup';
import { MenuItem } from '../Menu/MenuItem';

interface DataLinksContextMenuProps {
  children: (props: DataLinksContextMenuApi) => JSX.Element;
  links: () => LinkModel[];
  config: FieldConfig;
}

export interface DataLinksContextMenuApi {
  openMenu?: React.MouseEventHandler<HTMLOrSVGElement>;
  targetClassName?: string;
}

export const DataLinksContextMenu: React.FC<DataLinksContextMenuProps> = ({ children, links, config }) => {
  const linksCounter = config.links!.length;
  const itemsGroup: MenuItemsGroup[] = [{ items: linkModelToContextMenuItems(links), label: 'Data links' }];
  const renderMenuGroupItems = () => {
    return itemsGroup.map((group, index) => (
      <MenuGroup key={`${group.label}${index}`} label={group.label}>
        {(group.items || []).map((item) => (
          <MenuItem
            key={item.label}
            url={item.url}
            label={item.label}
            target={item.target}
            icon={item.icon}
            active={item.active}
            onClick={item.onClick}
          />
        ))}
      </MenuGroup>
    ));
  };

  // Use this class name (exposed via render prop) to add context menu indicator to the click target of the visualization
  const targetClassName = css`
    cursor: context-menu;
  `;

  if (linksCounter > 1) {
    return (
      <WithContextMenu renderMenuItems={renderMenuGroupItems}>
        {({ openMenu }) => {
          return children({ openMenu, targetClassName });
        }}
      </WithContextMenu>
    );
  } else {
    const linkModel = links()[0];
    return (
      <a
        href={linkModel.href}
        onClick={linkModel.onClick}
        target={linkModel.target}
        title={linkModel.title}
        style={{ display: 'flex' }}
        aria-label={selectors.components.DataLinksContextMenu.singleLink}
      >
        {children({})}
      </a>
    );
  }
};
