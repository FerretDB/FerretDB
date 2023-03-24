// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'FerretDB Documentation',
  tagline: 'Open Source, MongoDB-compatible document database',

  url: 'https://docs.ferretdb.io',
  baseUrl: '/',

  favicon: 'img/favicon.ico',
  trailingSlash: true,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  scripts: [{src: 'https://plausible.io/js/script.js', defer: true, "data-domain": "docs.ferretdb.io"}],

  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexBlog: true, // Index blog posts in search engine
        indexDocs: true, // Blog plugin is disabled, blog search needs to be disabled too
      },
    ],
  ],

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/FerretDB/FerretDB/tree/main/website',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: 'img/logo_dark.jpg',
      navbar: {
        logo: {
          alt: 'FerretDB Logo',
          src: 'img/logo_dark.jpg',
          srcDark:'img/logo_light.png'
        },
        items: [
          {
            to: '/',
            label: 'Documentation',
            position: 'left'
          },
          {
            href: 'https://blog.ferretdb.io/',
            label: 'Blog',
            position: 'left'
          },
          {
            href: 'https://github.com/FerretDB/',
            label: 'GitHub',
            position: 'right',
          },
          {
            href: 'https://ferretdb.io/',
            label: 'Go to FerretDB.io',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'FerretDB Docs',
            items: [
              {
                label: 'Documentation',
                to: '/',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'GitHub Discussions',
                href: 'https://github.com/FerretDB/FerretDB/discussions/',
              },
              {
                label: 'Slack',
                href: 'https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A',
              },
              {
                label: 'Twitter',
                href: 'https://twitter.com/ferret_db',
              },
              {
                label: 'Mastodon',
                href: 'https://techhub.social/@ferretdb',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                href: 'https://ferretdb.io/',
                label: 'Go to FerretDB.io',
                position: 'right',
              },
              {
                label: 'Blog',
                to: 'https://blog.ferretdb.io/',
              },
              {
                label: 'GitHub',
                href: 'https://github.com/FerretDB/',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} FerretDB Inc. Built with Docusaurus.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
};

module.exports = config;
