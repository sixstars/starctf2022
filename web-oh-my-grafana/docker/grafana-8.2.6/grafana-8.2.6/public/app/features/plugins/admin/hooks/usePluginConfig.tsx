import { useAsync } from 'react-use';
import { CatalogPlugin } from '../types';
import { loadPlugin } from '../../PluginPage';

export const usePluginConfig = (plugin?: CatalogPlugin) => {
  return useAsync(async () => {
    if (!plugin) {
      return null;
    }

    if (plugin.isInstalled && !plugin.isDisabled) {
      return loadPlugin(plugin.id);
    }
    return null;
  }, [plugin?.id, plugin?.isInstalled, plugin?.isDisabled]);
};
