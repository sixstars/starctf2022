import { GrafanaTheme2, createTheme } from '@grafana/data';
//@ts-ignore
import { create } from '@storybook/theming/create';
import '../src/components/Icon/iconBundle';

const createStorybookTheme = (theme: GrafanaTheme2) => {
  return create({
    colorPrimary: theme.colors.primary.main,
    colorSecondary: theme.colors.error.main,

    // UI
    appBg: theme.colors.background.canvas,
    appContentBg: theme.colors.background.primary,
    appBorderColor: theme.colors.border.medium,
    appBorderRadius: theme.shape.borderRadius(1),

    // Typography
    fontBase: theme.typography.fontFamily,
    fontCode: theme.typography.fontFamilyMonospace,

    // Text colors
    textColor: theme.colors.text.primary,
    textInverseColor: theme.colors.background.primary,

    // Toolbar default and active colors
    barTextColor: theme.colors.text.primary,
    barSelectedColor: theme.colors.emphasize(theme.colors.primary.text),
    barBg: theme.colors.background.primary,

    // Form colors
    inputBg: theme.components.input.background,
    inputBorder: theme.components.input.borderColor,
    inputTextColor: theme.components.input.text,
    inputBorderRadius: theme.shape.borderRadius(1),

    brandTitle: 'Grafana UI',
    brandUrl: './',
    brandImage: 'public/img/grafana_icon.svg',
  });
};

const GrafanaLight = createStorybookTheme(createTheme({ colors: { mode: 'light' } }));
const GrafanaDark = createStorybookTheme(createTheme({ colors: { mode: 'dark' } }));

export { GrafanaLight, GrafanaDark };
