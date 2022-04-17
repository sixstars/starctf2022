import React, { useState } from 'react';
import { ContextMenu } from '../ContextMenu/ContextMenu';

interface WithContextMenuProps {
  /** Menu item trigger that accepts openMenu prop */
  children: (props: { openMenu: React.MouseEventHandler<HTMLElement> }) => JSX.Element;
  /** A function that returns an array of menu items */
  renderMenuItems: () => React.ReactNode;
}

export const WithContextMenu: React.FC<WithContextMenuProps> = ({ children, renderMenuItems }) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [menuPosition, setMenuPosition] = useState({ x: 0, y: 0 });
  return (
    <>
      {children({
        openMenu: (e) => {
          setIsMenuOpen(true);
          setMenuPosition({
            x: e.pageX,
            y: e.pageY,
          });
        },
      })}

      {isMenuOpen && (
        <ContextMenu
          onClose={() => setIsMenuOpen(false)}
          x={menuPosition.x}
          y={menuPosition.y}
          renderMenuItems={renderMenuItems}
        />
      )}
    </>
  );
};
