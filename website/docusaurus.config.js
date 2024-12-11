// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import {themes} from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'FerretDB',
  tagline: 'A truly Open Source MongoDB alternative',

  url: 'https://docs.ferretdb.io',
  baseUrl: '/',

  favicon: 'img/favicon.ico',
  trailingSlash: true,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',
  onBrokenAnchors: 'throw',

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
        lunr:{
          tokenizerSeparator: /[\s\-\$]+/,
        }
      },
    ],
    'plugin-image-zoom'
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
      image: 'img/logo-dark.jpg',
      navbar: {
        logo: {
          alt: 'FerretDB Logo',
          src: 'img/logo-dark.jpg',
          srcDark: 'img/logo-light.png'
        },
        items: [
          {
            to: '/',
            label: 'Documentation',
            position: 'right',
            type: 'docsVersionDropdown'
          },
          {
            href: 'https://blog.ferretdb.io/',
            label: 'Blog',
            position: 'right'
          },
          {
            href: 'https://github.com/FerretDB/',
            label: 'GitHub',
            position: 'right',
          },
          {
            href: 'https://www.ferretdb.com/',
            label: 'FerretDB.com',
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
                href: 'https://www.ferretdb.com/',
                label: 'FerretDB.com',
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
        theme: themes.github,
        darkTheme: themes.dracula,
        additionalLanguages: ['go', 'sql', 'json', 'json5', 'systemd'],
      },
      mermaid: {
        theme: {light: 'default', dark: 'dark'},
      },
    }),
  markdown: {
    mermaid: true,
  },
  themes: ['@docusaurus/theme-mermaid'],
};

module.exports = config;
