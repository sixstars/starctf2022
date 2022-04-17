import { useSpinner } from '../utils/useSpinner';
import { testPlugin } from './plugin/tests';
import { Task, TaskRunner } from './task';
import rimrafCallback from 'rimraf';
import { resolve as resolvePath } from 'path';
import { promisify } from 'util';
import globby from 'globby';
import execa from 'execa';
import { constants as fsConstants, promises as fs } from 'fs';
import { CLIEngine } from 'eslint';
import { bundlePlugin as bundleFn, PluginBundleOptions } from './plugin/bundle';

const { access, copyFile } = fs;
const { COPYFILE_EXCL } = fsConstants;
const rimraf = promisify(rimrafCallback);

interface PluginBuildOptions {
  coverage: boolean;
  maxJestWorkers?: string;
  preserveConsole?: boolean;
  skipTest?: boolean;
  skipLint?: boolean;
}

interface Fixable {
  fix?: boolean;
}

const bundlePlugin = (options: PluginBundleOptions) => useSpinner('Compiling...', () => bundleFn(options));

// @ts-ignore
const clean = () => useSpinner('Cleaning', () => rimraf(`${process.cwd()}/dist`));

const copyIfNonExistent = (srcPath: string, destPath: string) =>
  copyFile(srcPath, destPath, COPYFILE_EXCL)
    .then(() => console.log(`Created: ${destPath}`))
    .catch((error) => {
      if (error.code !== 'EEXIST') {
        throw error;
      }
    });

export const prepare = () =>
  useSpinner('Preparing', () =>
    Promise.all([
      // Remove local dependencies for @grafana/data/node_modules
      // See: https://github.com/grafana/grafana/issues/26748
      rimraf(resolvePath(__dirname, 'node_modules/@grafana/data/node_modules')),

      // Copy only if local tsconfig does not exist.  Otherwise this will work, but have odd behavior
      copyIfNonExistent(
        resolvePath(__dirname, '../../config/tsconfig.plugin.local.json'),
        resolvePath(process.cwd(), 'tsconfig.json')
      ),
      // Copy only if local prettierrc does not exist.  Otherwise this will work, but have odd behavior
      copyIfNonExistent(
        resolvePath(__dirname, '../../config/prettier.plugin.rc.js'),
        resolvePath(process.cwd(), '.prettierrc.js')
      ),
    ])
  );

export const versions = async () => {
  try {
    const nodeVersion = await execa('node', ['--version']);
    console.log(`Using Node.js ${nodeVersion.stdout}`);

    const toolkitVersion = await execa('grafana-toolkit', ['--version']);
    console.log(`Using @grafana/toolkit ${toolkitVersion.stdout}`);
  } catch (err) {
    console.log(`Error reading versions`, err);
  }
};

// @ts-ignore
const typecheckPlugin = () => useSpinner('Typechecking', () => execa('tsc', ['--noEmit']));

const getTypescriptSources = () => globby(resolvePath(process.cwd(), 'src/**/*.+(ts|tsx)'));

// @ts-ignore
const getStylesSources = () => globby(resolvePath(process.cwd(), 'src/**/*.+(scss|css)'));

export const lintPlugin = ({ fix }: Fixable = {}) =>
  useSpinner('Linting', async () => {
    try {
      // Show a warning if the tslint file exists
      await access(resolvePath(process.cwd(), 'tslint.json'));
      console.log('\n');
      console.log('--------------------------------------------------------------');
      console.log('NOTE: @grafana/toolkit has migrated to use eslint');
      console.log('Update your configs to use .eslintrc rather than tslint.json');
      console.log('--------------------------------------------------------------');
    } catch {
      // OK: tslint does not exist
    }

    // @todo should remove this because the config file could be in a parent dir or within package.json
    const configFile = await globby(resolvePath(process.cwd(), '.eslintrc?(.cjs|.js|.json|.yaml|.yml)')).then(
      (filePaths) => {
        if (filePaths.length > 0) {
          return filePaths[0];
        } else {
          return resolvePath(__dirname, '../../config/eslint.plugin.json');
        }
      }
    );

    const cli = new CLIEngine({
      configFile,
      fix,
    });

    const report = cli.executeOnFiles(await getTypescriptSources());

    if (fix) {
      CLIEngine.outputFixes(report);
    }

    const { errorCount, results, warningCount } = report;
    const formatter = cli.getFormatter();

    if (errorCount > 0 || warningCount > 0) {
      console.log('\n');
      console.log(formatter(results));
      console.log('\n');
    }

    if (errorCount > 0) {
      throw new Error(`${errorCount} linting errors found in ${results.length} files`);
    }
  });

export const pluginBuildRunner: TaskRunner<PluginBuildOptions> = async ({
  coverage,
  maxJestWorkers,
  preserveConsole,
  skipTest,
  skipLint,
}) => {
  await versions();
  await prepare();
  if (!skipLint) {
    await lintPlugin({ fix: false });
  }
  if (!skipTest) {
    await testPlugin({ updateSnapshot: false, coverage, maxWorkers: maxJestWorkers, watch: false });
  }
  await bundlePlugin({ watch: false, production: true, preserveConsole });
};

export const pluginBuildTask = new Task<PluginBuildOptions>('Build plugin', pluginBuildRunner);
