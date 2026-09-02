# 10: Settings Page

**What to build:** Full-page Settings with sidebar navigation for 12 sections. Includes customizable keyboard shortcuts and Auth sub-page for saved credentials.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Settings Page renders when activeView is 'settings'
- [ ] Sidebar navigation shows 12 sections: General, Appearance, Editor, Network, Proxy, TLS, Environments, Auth, Storage, Keyboard Shortcuts, Notifications, Advanced
- [ ] Active section is highlighted in sidebar
- [ ] Section content renders in main area

### General
- [ ] App name, version, update check

### Appearance
- [ ] Theme selector (atlas-light, atlas-dark, system)
- [ ] Font size, font family

### Editor
- [ ] Default method, default body type, auto-save

### Network
- [ ] Timeout, follow redirects, verify SSL

### Proxy
- [ ] Global proxy settings: HTTP Proxy, HTTPS Proxy, SOCKS, Authentication

### TLS
- [ ] Global TLS settings: Verification, Custom CA, Client Certificates, TLS Version

### Environments
- [ ] Default environment, variable interpolation settings

### Auth
- [ ] Saved credentials list
- [ ] Add/Edit/Delete credentials
- [ ] OAuth clients list
- [ ] Add/Edit/Delete OAuth clients

### Storage
- [ ] History retention (30d, 90d, 1yr, Forever)
- [ ] Storage location

### Keyboard Shortcuts
- [ ] List of actions with current shortcut
- [ ] Edit shortcut per action
- [ ] Reset to defaults button

### Notifications
- [ ] Enable/disable background task notifications

### Advanced
- [ ] Debug mode, logging level

- [ ] All components use theme tokens
