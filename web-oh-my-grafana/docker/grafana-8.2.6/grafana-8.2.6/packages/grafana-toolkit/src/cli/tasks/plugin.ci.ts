import { Task, TaskRunner } from './task';
import { pluginBuildRunner } from './plugin.build';
import { getPluginJson } from '../../config/utils/pluginValidation';
import { getPluginId } from '../../config/utils/getPluginId';

// @ts-ignore
import execa = require('execa');
import path = require('path');
import fs from 'fs-extra';
import { getPackageDetails, getGrafanaVersions, readGitLog } from '../../plugins/utils';
import { buildManifest, signManifest, saveManifest } from '../../plugins/manifest';
import {
  getJobFolder,
  writeJobStats,
  getCiFolder,
  getPluginBuildInfo,
  getPullRequestNumber,
  getCircleDownloadBaseURL,
} from '../../plugins/env';
import { agregateWorkflowInfo, agregateCoverageInfo, agregateTestInfo } from '../../plugins/workflow';
import { PluginPackageDetails, PluginBuildReport } from '../../plugins/types';
import rimrafCallback from 'rimraf';
import { promisify } from 'util';
const rimraf = promisify(rimrafCallback);

export interface PluginCIOptions {
  finish?: boolean;
  upload?: boolean;
  signatureType?: string;
  rootUrls?: string[];
  maxJestWorkers?: string;
}

/**
 * 1. BUILD
 *
 *  when platform exists it is building backend, otherwise frontend
 *
 *  Each build writes data:
 *   ~/ci/jobs/build_xxx/
 *
 *  Anything that should be put into the final zip file should be put in:
 *   ~/ci/jobs/build_xxx/dist
 *
 * @deprecated -- this task was written with a specific circle-ci build in mind.  That system
 * has been replaced with Drone, and this is no longer the best practice.  Any new work
 * should be defined in the grafana build pipeline tool or drone configs directly.
 */
const buildPluginRunner: TaskRunner<PluginCIOptions> = async ({ finish, maxJestWorkers }) => {
  const start = Date.now();

  if (finish) {
    const workDir = getJobFolder();
    await rimraf(workDir);
    fs.mkdirSync(workDir);

    // Move local folders to the scoped job folder
    for (const name of ['dist', 'coverage']) {
      const dir = path.resolve(process.cwd(), name);
      if (fs.existsSync(dir)) {
        fs.moveSync(dir, path.resolve(workDir, name));
      }
    }
    writeJobStats(start, workDir);
  } else {
    // Do regular build process with coverage
    await pluginBuildRunner({ coverage: true, maxJestWorkers });
  }
};

export const ciBuildPluginTask = new Task<PluginCIOptions>('Build Plugin', buildPluginRunner);

/**
 * 2. Package
 *
 *  Take everything from `~/ci/job/{any}/dist` and
 *  1. merge it into: `~/ci/dist`
 *  2. zip it into packages in `~/ci/packages`
 *  3. prepare grafana environment in: `~/ci/grafana-test-env`
 *
 *
 * @deprecated -- this task was written with a specific circle-ci build in mind.  That system
 * has been replaced with Drone, and this is no longer the best practice.  Any new work
 * should be defined in the grafana build pipeline tool or drone configs directly.
 */
const packagePluginRunner: TaskRunner<PluginCIOptions> = async ({ signatureType, rootUrls }) => {
  const start = Date.now();
  const ciDir = getCiFolder();
  const packagesDir = path.resolve(ciDir, 'packages');
  const distDir = path.resolve(ciDir, 'dist');
  const docsDir = path.resolve(ciDir, 'docs');
  const jobsDir = path.resolve(ciDir, 'jobs');

  fs.exists(jobsDir, (jobsDirExists) => {
    if (!jobsDirExists) {
      throw new Error('You must run plugin:ci-build prior to running plugin:ci-package');
    }
  });

  const grafanaEnvDir = path.resolve(ciDir, 'grafana-test-env');
  await execa('rimraf', [packagesDir, distDir, grafanaEnvDir]);
  fs.mkdirSync(packagesDir);
  fs.mkdirSync(distDir);

  // Updating the dist dir to have a pluginId named directory in it
  // The zip needs to contain the plugin code wrapped in directory with a pluginId name
  const distContentDir = path.resolve(distDir, getPluginId());
  fs.mkdirSync(grafanaEnvDir);

  console.log('Build Dist Folder');

  // 1. Check for a local 'dist' folder
  const d = path.resolve(process.cwd(), 'dist');
  if (fs.existsSync(d)) {
    await execa('cp', ['-rn', d + '/.', distContentDir]);
  }

  // 2. Look for any 'dist' folders under ci/job/XXX/dist
  const dirs = fs.readdirSync(path.resolve(ciDir, 'jobs'));
  for (const j of dirs) {
    const contents = path.resolve(ciDir, 'jobs', j, 'dist');
    if (fs.existsSync(contents)) {
      try {
        await execa('cp', ['-rn', contents + '/.', distContentDir]);
      } catch (er) {
        throw new Error('Duplicate files found in dist folders');
      }
    }
  }

  console.log('Save the source info in plugin.json');
  const pluginJsonFile = path.resolve(distContentDir, 'plugin.json');
  const pluginInfo = getPluginJson(pluginJsonFile);
  pluginInfo.info.build = await getPluginBuildInfo();
  fs.writeFileSync(pluginJsonFile, JSON.stringify(pluginInfo, null, 2), { encoding: 'utf-8' });

  // Write a MANIFEST.txt file in the dist folder
  try {
    const manifest = await buildManifest(distContentDir);
    if (signatureType) {
      manifest.signatureType = signatureType;
    }
    if (rootUrls) {
      manifest.rootUrls = rootUrls;
    }
    const signedManifest = await signManifest(manifest);
    await saveManifest(distContentDir, signedManifest);
  } catch (err) {
    console.warn(`Error signing manifest: ${distContentDir}`, err);
  }

  console.log('Building ZIP');
  let zipName = pluginInfo.id + '-' + pluginInfo.info.version + '.zip';
  let zipFile = path.resolve(packagesDir, zipName);
  await execa('zip', ['-r', zipFile, '.'], { cwd: distDir });

  const zipStats = fs.statSync(zipFile);
  if (zipStats.size < 100) {
    throw new Error('Invalid zip file: ' + zipFile);
  }

  // Make a copy so it is easy for report to read
  await execa('cp', [pluginJsonFile, distDir]);

  const info: PluginPackageDetails = {
    plugin: await getPackageDetails(zipFile, distDir),
  };

  console.log('Setup Grafana Environment');
  let p = path.resolve(grafanaEnvDir, 'plugins', pluginInfo.id);
  fs.mkdirSync(p, { recursive: true });
  await execa('unzip', [zipFile, '-d', p]);

  // If docs exist, zip them into packages
  if (fs.existsSync(docsDir)) {
    console.log('Creating documentation zip');
    zipName = pluginInfo.id + '-' + pluginInfo.info.version + '-docs.zip';
    zipFile = path.resolve(packagesDir, zipName);
    await execa('zip', ['-r', zipFile, '.'], { cwd: docsDir });

    info.docs = await getPackageDetails(zipFile, docsDir);
  }

  p = path.resolve(packagesDir, 'info.json');
  fs.writeFileSync(p, JSON.stringify(info, null, 2), { encoding: 'utf-8' });

  // Write the custom settings
  p = path.resolve(grafanaEnvDir, 'custom.ini');
  const customIniBody =
    `# Autogenerated by @grafana/toolkit \n` +
    `[paths] \n` +
    `plugins = ${path.resolve(grafanaEnvDir, 'plugins')}\n` +
    `\n`; // empty line
  fs.writeFileSync(p, customIniBody, { encoding: 'utf-8' });

  writeJobStats(start, getJobFolder());
};

export const ciPackagePluginTask = new Task<PluginCIOptions>('Bundle Plugin', packagePluginRunner);

/**
 * 4. Report
 *
 *  Create a report from all the previous steps
 *
 * @deprecated -- this task was written with a specific circle-ci build in mind.  That system
 * has been replaced with Drone, and this is no longer the best practice.  Any new work
 * should be defined in the grafana build pipeline tool or drone configs directly.
 */
const pluginReportRunner: TaskRunner<PluginCIOptions> = async ({ upload }) => {
  const ciDir = path.resolve(process.cwd(), 'ci');
  const packageDir = path.resolve(ciDir, 'packages');
  const packageInfo = require(path.resolve(packageDir, 'info.json')) as PluginPackageDetails;

  const pluginJsonFile = path.resolve(ciDir, 'dist', 'plugin.json');
  console.log('Load info from: ' + pluginJsonFile);

  const pluginMeta = getPluginJson(pluginJsonFile);
  const report: PluginBuildReport = {
    plugin: pluginMeta,
    packages: packageInfo,
    workflow: agregateWorkflowInfo(),
    coverage: agregateCoverageInfo(),
    tests: agregateTestInfo(),
    artifactsBaseURL: await getCircleDownloadBaseURL(),
    grafanaVersion: getGrafanaVersions(),
    git: await readGitLog(),
  };
  const pr = getPullRequestNumber();
  if (pr) {
    report.pullRequest = pr;
  }

  // Save the report to disk
  const file = path.resolve(ciDir, 'report.json');
  fs.writeFileSync(file, JSON.stringify(report, null, 2), { encoding: 'utf-8' });

  const GRAFANA_API_KEY = process.env.GRAFANA_API_KEY;
  if (!GRAFANA_API_KEY) {
    console.log('Enter a GRAFANA_API_KEY to upload the plugin report');
    return;
  }
  const url = `https://grafana.com/api/plugins/${report.plugin.id}/ci`;

  console.log('Sending report to:', url);
  const axios = require('axios');
  const info = await axios.post(url, report, {
    headers: { Authorization: 'Bearer ' + GRAFANA_API_KEY },
  });
  if (info.status === 200) {
    console.log('OK: ', info.data);
  } else {
    console.warn('Error: ', info);
  }
};

export const ciPluginReportTask = new Task<PluginCIOptions>('Generate Plugin Report', pluginReportRunner);
