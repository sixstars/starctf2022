import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { within } from '@testing-library/dom';
import { OrgRole } from '@grafana/data';
import { selectors } from '@grafana/e2e-selectors';

import { Props, UserProfileEditPage } from './UserProfileEditPage';
import { initialUserState } from './state/reducers';
import { getNavModel } from '../../core/selectors/navModel';
import { backendSrv } from '../../core/services/backend_srv';
import { TeamPermissionLevel } from '../../types';

const defaultProps: Props = {
  ...initialUserState,
  user: {
    id: 1,
    name: 'Test User',
    email: 'test@test.com',
    login: 'test',
    isDisabled: false,
    isGrafanaAdmin: false,
    orgId: 0,
  },
  teams: [
    {
      id: 0,
      name: 'Team One',
      email: 'team.one@test.com',
      avatarUrl: '/avatar/07d881f402480a2a511a9a15b5fa82c0',
      memberCount: 2000,
      permission: TeamPermissionLevel.Admin,
    },
  ],
  orgs: [
    {
      name: 'Main',
      orgId: 0,
      role: OrgRole.Editor,
    },
    {
      name: 'Second',
      orgId: 1,
      role: OrgRole.Viewer,
    },
    {
      name: 'Third',
      orgId: 2,
      role: OrgRole.Admin,
    },
  ],
  sessions: [
    {
      id: 0,
      browser: 'Chrome',
      browserVersion: '90',
      clientIp: 'localhost',
      createdAt: '2021-01-01 04:00:00',
      device: 'Macbook Pro',
      isActive: true,
      os: 'Mac OS X',
      osVersion: '11',
      seenAt: new Date().toUTCString(),
    },
  ],
  navModel: getNavModel(
    {
      'profile-settings': {
        icon: 'sliders-v-alt',
        id: 'profile-settings',
        parentItem: {
          id: 'profile',
          text: 'Test User',
          img: '/avatar/46d229b033af06a191ff2267bca9ae56',
          url: '/profile',
        },
        text: 'Preferences',
        url: '/profile',
      },
    },
    'profile-settings'
  ),
  initUserProfilePage: jest.fn().mockResolvedValue(undefined),
  revokeUserSession: jest.fn().mockResolvedValue(undefined),
  changeUserOrg: jest.fn().mockResolvedValue(undefined),
  updateUserProfile: jest.fn().mockResolvedValue(undefined),
};

function getSelectors() {
  const dashboardSelect = () => screen.getByLabelText(/user preferences home dashboard drop down/i);
  const timepickerSelect = () => screen.getByLabelText(selectors.components.TimeZonePicker.container);
  const teamsTable = () => screen.getByRole('table', { name: /user teams table/i });
  const orgsTable = () => screen.getByRole('table', { name: /user organizations table/i });
  const sessionsTable = () => screen.getByRole('table', { name: /user sessions table/i });
  return {
    name: () => screen.getByRole('textbox', { name: /^name$/i }),
    email: () => screen.getByRole('textbox', { name: /email/i }),
    username: () => screen.getByRole('textbox', { name: /username/i }),
    saveProfile: () => screen.getByRole('button', { name: /edit user profile save button/i }),
    dashboardSelect,
    dashboardValue: () => within(dashboardSelect()).getByText(/default/i),
    timepickerSelect,
    timepickerValue: () => within(timepickerSelect()).getByText(/coordinated universal time/i),
    savePreferences: () => screen.getByRole('button', { name: /user preferences save button/i }),
    teamsTable,
    teamsRow: () => within(teamsTable()).getByRole('row', { name: /team one team.one@test\.com 2000/i }),
    orgsTable,
    orgsEditorRow: () => within(orgsTable()).getByRole('row', { name: /main editor current/i }),
    orgsViewerRow: () => within(orgsTable()).getByRole('row', { name: /second viewer select/i }),
    orgsAdminRow: () => within(orgsTable()).getByRole('row', { name: /third admin select/i }),
    sessionsTable,
    sessionsRow: () =>
      within(sessionsTable()).getByRole('row', {
        name: /now 2021-01-01 04:00:00 localhost chrome on mac os x 11/i,
      }),
  };
}

async function getTestContext(overrides: Partial<Props> = {}) {
  jest.clearAllMocks();
  const putSpy = jest.spyOn(backendSrv, 'put');
  const getSpy = jest
    .spyOn(backendSrv, 'get')
    .mockResolvedValue({ timezone: 'UTC', homeDashboardId: 0, theme: 'dark' });
  const searchSpy = jest.spyOn(backendSrv, 'search').mockResolvedValue([]);

  const props = { ...defaultProps, ...overrides };
  const { rerender } = render(<UserProfileEditPage {...props} />);

  await waitFor(() => expect(props.initUserProfilePage).toHaveBeenCalledTimes(1));

  return { rerender, putSpy, getSpy, searchSpy, props };
}

describe('UserProfileEditPage', () => {
  describe('when loading user', () => {
    it('should show loading placeholder', async () => {
      await getTestContext({ user: null });

      expect(screen.getByText(/loading \.\.\./i)).toBeInTheDocument();
    });
  });

  describe('when user has loaded', () => {
    it('should show edit profile form', async () => {
      await getTestContext();

      const { name, email, username, saveProfile } = getSelectors();
      expect(screen.getByText(/edit profile/i)).toBeInTheDocument();
      expect(name()).toBeInTheDocument();
      expect(name()).toHaveValue('Test User');
      expect(email()).toBeInTheDocument();
      expect(email()).toHaveValue('test@test.com');
      expect(username()).toBeInTheDocument();
      expect(username()).toHaveValue('test');
      expect(saveProfile()).toBeInTheDocument();
    });

    it('should show shared preferences', async () => {
      await getTestContext();

      const { dashboardSelect, dashboardValue, timepickerSelect, timepickerValue, savePreferences } = getSelectors();
      expect(screen.getByRole('group', { name: /preferences/i })).toBeInTheDocument();
      expect(screen.getByRole('radio', { name: /default/i })).toBeInTheDocument();
      expect(screen.getByRole('radio', { name: /dark/i })).toBeInTheDocument();
      expect(screen.getByRole('radio', { name: /light/i })).toBeInTheDocument();
      expect(dashboardSelect()).toBeInTheDocument();
      expect(dashboardValue()).toBeInTheDocument();
      expect(timepickerSelect()).toBeInTheDocument();
      expect(timepickerValue()).toBeInTheDocument();
      expect(savePreferences()).toBeInTheDocument();
    });

    describe('and teams are loading', () => {
      it('should show teams loading placeholder', async () => {
        await getTestContext({ teamsAreLoading: true });

        expect(screen.getByText(/loading teams\.\.\./i)).toBeInTheDocument();
      });
    });

    describe('and teams are loaded', () => {
      it('should show teams', async () => {
        await getTestContext();

        const { teamsTable, teamsRow } = getSelectors();
        expect(screen.getByRole('heading', { name: /teams/i })).toBeInTheDocument();
        expect(teamsTable()).toBeInTheDocument();
        expect(teamsRow()).toBeInTheDocument();
      });
    });

    describe('and organizations are loading', () => {
      it('should show teams loading placeholder', async () => {
        await getTestContext({ orgsAreLoading: true });

        expect(screen.getByText(/loading organizations\.\.\./i)).toBeInTheDocument();
      });
    });

    describe('and organizations are loaded', () => {
      it('should show organizations', async () => {
        await getTestContext();

        const { orgsTable, orgsEditorRow, orgsViewerRow, orgsAdminRow } = getSelectors();
        expect(screen.getByRole('heading', { name: /organizations/i })).toBeInTheDocument();
        expect(orgsTable()).toBeInTheDocument();
        expect(orgsEditorRow()).toBeInTheDocument();
        expect(orgsViewerRow()).toBeInTheDocument();
        expect(orgsAdminRow()).toBeInTheDocument();
      });
    });

    describe('and sessions are loading', () => {
      it('should show teams loading placeholder', async () => {
        await getTestContext({ sessionsAreLoading: true });

        expect(screen.getByText(/loading sessions\.\.\./i)).toBeInTheDocument();
      });
    });

    describe('and sessions are loaded', () => {
      it('should show sessions', async () => {
        await getTestContext();

        const { sessionsTable, sessionsRow } = getSelectors();
        expect(sessionsTable()).toBeInTheDocument();
        expect(sessionsRow()).toBeInTheDocument();
      });
    });

    describe('and user is edited and saved', () => {
      it('should call updateUserProfile', async () => {
        const { props } = await getTestContext();

        const { email, saveProfile } = getSelectors();
        userEvent.clear(email());
        await userEvent.type(email(), 'test@test.se');
        userEvent.click(saveProfile());

        await waitFor(() => expect(props.updateUserProfile).toHaveBeenCalledTimes(1));
        expect(props.updateUserProfile).toHaveBeenCalledWith({
          email: 'test@test.se',
          login: 'test',
          name: 'Test User',
        });
      });
    });

    describe('and organization is changed', () => {
      it('should call changeUserOrg', async () => {
        const { props } = await getTestContext();
        const orgsAdminSelectButton = () =>
          within(getSelectors().orgsAdminRow()).getByRole('button', {
            name: /switch to the organization named Third/i,
          });

        userEvent.click(orgsAdminSelectButton());

        await waitFor(() => expect(props.changeUserOrg).toHaveBeenCalledTimes(1));
        expect(props.changeUserOrg).toHaveBeenCalledWith({
          name: 'Third',
          orgId: 2,
          role: 'Admin',
        });
      });
    });

    describe('and session is revoked', () => {
      it('should call revokeUserSession', async () => {
        const { props } = await getTestContext();
        const sessionsRevokeButton = () =>
          within(getSelectors().sessionsRow()).getByRole('button', {
            name: /revoke user session/i,
          });

        userEvent.click(sessionsRevokeButton());

        await waitFor(() => expect(props.revokeUserSession).toHaveBeenCalledTimes(1));
        expect(props.revokeUserSession).toHaveBeenCalledWith(0);
      });
    });
  });
});
