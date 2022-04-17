const execa = require('execa');
const program = require('commander');
const resolveBin = require('resolve-as-bin');
const { resolve, sep } = require('path');

const cypress = (commandName, { updateScreenshots }) => {
  // Support running an unpublished dev build
  const dirname = __dirname.split(sep).pop();
  const projectPath = resolve(`${__dirname}${dirname === 'dist' ? '/..' : ''}`);

  // For plugins/extendConfig
  const CWD = `CWD=${process.cwd()}`;

  // For plugins/compareSnapshots
  const UPDATE_SCREENSHOTS = `UPDATE_SCREENSHOTS=${updateScreenshots ? 1 : 0}`;

  const cypressOptions = [commandName, '--env', `${CWD},${UPDATE_SCREENSHOTS}`, `--project=${projectPath}`];

  const execaOptions = {
    cwd: __dirname,
    stdio: 'inherit',
  };

  return execa(resolveBin('cypress'), cypressOptions, execaOptions)
    .then(() => {}) // no return value
    .catch((error) => {
      console.error(error.message);
      process.exitCode = 1;
    });
};

module.exports = () => {
  const updateOption = '-u, --update-screenshots';
  const updateDescription = 'update expected screenshots';

  program
    .command('open')
    .description('runs tests within the interactive GUI')
    .option(updateOption, updateDescription)
    .action((options) => cypress('open', options));

  program
    .command('run')
    .description('runs tests from the CLI without the GUI')
    .option(updateOption, updateDescription)
    .action((options) => cypress('run', options));

  program.parse(process.argv);
};
