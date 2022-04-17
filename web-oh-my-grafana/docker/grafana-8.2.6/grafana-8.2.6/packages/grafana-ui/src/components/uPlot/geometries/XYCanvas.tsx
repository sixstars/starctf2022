import React, { useMemo } from 'react';
import { css } from '@emotion/css';

interface XYCanvasProps {
  top: number; // css pxls
  left: number; // css pxls
}

/**
 * Renders absolutely positioned element on top of the uPlot's plotting area (axes are not included!).
 * Useful when you want to render some overlay with canvas-independent elements on top of the plot.
 */
export const XYCanvas: React.FC<XYCanvasProps> = ({ children, left, top }) => {
  const className = useMemo(() => {
    return css`
      position: absolute;
      overflow: visible;
      left: ${left}px;
      top: ${top}px;
    `;
  }, [left, top]);

  return <div className={className}>{children}</div>;
};
