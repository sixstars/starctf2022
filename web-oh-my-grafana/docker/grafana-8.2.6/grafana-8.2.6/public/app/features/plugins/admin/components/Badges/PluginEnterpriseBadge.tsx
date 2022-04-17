import React from 'react';
import { Badge, Button, HorizontalGroup, PluginSignatureBadge, useStyles2 } from '@grafana/ui';
import { CatalogPlugin } from '../../types';
import { getBadgeColor } from './sharedStyles';
import { config } from '@grafana/runtime';

type Props = { plugin: CatalogPlugin };

export function PluginEnterpriseBadge({ plugin }: Props): React.ReactElement {
  const customBadgeStyles = useStyles2(getBadgeColor);
  const onClick = (ev: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    ev.preventDefault();
    window.open(
      `https://grafana.com/grafana/plugins/${plugin.id}?utm_source=grafana_catalog_learn_more`,
      '_blank',
      'noopener,noreferrer'
    );
  };

  if (config.licenseInfo?.hasValidLicense) {
    return <Badge text="Enterprise" color="blue" />;
  }

  return (
    <HorizontalGroup>
      <PluginSignatureBadge status={plugin.signature} />
      <Badge icon="lock" aria-label="lock icon" text="Enterprise" color="blue" className={customBadgeStyles} />
      <Button size="sm" fill="text" icon="external-link-alt" onClick={onClick}>
        Learn more
      </Button>
    </HorizontalGroup>
  );
}
