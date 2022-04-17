import { importDashboard, Dashboard } from './importDashboard';
import { e2e } from '../index';

/**
 * Smoke test several dashboard json files from a test directory
 * and validate that all the panels in each import finish loading their queries
 * @param dirPath the relative path to a directory which contains json files representing dashboards,
 * for example if your dashboards live in `cypress/testDashboards` you can pass `/testDashboards`
 * @param queryTimeout a number of ms to wait for the imported dashboard to finish loading
 */
export const importDashboards = async (dirPath: string, queryTimeout?: number) => {
  e2e()
    .getJSONFilesFromDir(dirPath)
    .then((jsonFiles: Dashboard[]) => {
      jsonFiles.forEach((file) => {
        importDashboard(file, queryTimeout || 6000);
      });
    });
};
