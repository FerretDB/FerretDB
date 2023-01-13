// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'FerretDB Documentation',
  tagline: 'A truly Open Source MongoDB alternative',

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

  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexBlog: false,
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
      navbar: {
        logo: {
          alt: 'FerretDB Logo',
          src: 'img/logo_dark.png',
          srcDark:'img/logo_light.png'
        },
        items: [
          {
            type: 'doc',
            docId: 'intro',
            position: 'left',
            label: 'Documentation',
          },
          {
            href: 'https://blog.ferertdb.io/',
            label: 'Blog',
            position: 'left'
          },
          {
            href: 'https://github.com/FerretDB/',
            label: 'GitHub',
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
                to: '/intro/',
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
