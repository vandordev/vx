import config from './config.mjs';

const apiReference = {
  label: 'API Reference',
  items: [
            { label: 'app', link: '/api/app' },
            { label: 'config', link: '/api/config' },
            { label: 'domain', link: '/api/domain' },
            { label: 'errors', link: '/api/errors' },
            { label: 'gen', link: '/api/gen' },
            { label: 'input', link: '/api/input' },
            { label: 'package', link: '/api/package' },
            { label: 'project', link: '/api/project' },
            { label: 'resolve', link: '/api/resolve' },
            { label: 'ui', link: '/api/ui' },
            { label: 'utils', link: '/api/utils' },
            { label: 'view', link: '/api/view' },
            { label: 'vpkg', link: '/api/vpkg' },
            { label: 'workflow', link: '/api/workflow' },
    {
      label: 'Adapters',
      items: [
              { label: 'clipboard', link: '/api/adapters/clipboard' },
              { label: 'editor', link: '/api/adapters/editor' },
              { label: 'icon', link: '/api/adapters/icon' },
              { label: 'shell', link: '/api/adapters/shell' },
              { label: 'tty', link: '/api/adapters/tty' },
      ],
    },
  ],
};

const sidebar = [
  {
    label: 'vx',
    link: '/',
  },
  {
    label: 'Install',
    link: '/install',
  },
  {
    label: 'Commands',
    items: [
      { label: 'vx', link: '/commands/vx' },
            { label: 'completion', link: '/commands/completion' },
            { label: 'config', link: '/commands/config' },
            { label: 'config init', link: '/commands/config-init' },
            { label: 'gen', link: '/commands/gen' },
            { label: 'view', link: '/commands/view' },
    ],
  },
  {
    label: 'Configuration',
    link: '/configuration',
  },
];

// Add API Reference in non-production environments only
const isProduction = process.env.NODE_ENV === 'production';
if (!isProduction) {
  sidebar.push(apiReference);
}

sidebar.push({ label: 'Contributing', link: '/contributing' });
export default sidebar;
