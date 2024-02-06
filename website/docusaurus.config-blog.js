// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import {themes} from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'FerretDB Blog',
  tagline: 'A truly Open Source MongoDB alternative',

  url: 'https://blog.ferretdb.io',
  baseUrl: '/',

  favicon: 'img/favicon.ico',
  trailingSlash: true,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  stylesheets: [
    {            href: "https://unpkg.com/@antonz/codapi@0.13.0/dist/snippet.css"},
  ],

  scripts: [
    {src: 'https://plausible.io/js/script.js', defer: true, "data-domain": "blog.ferretdb.io"},
    {src: "https://unpkg.com/@antonz/codapi@0.13.0/dist/snippet.js", defer: true},
    {src: '/codapi/init.js', defer: true},
  ],

  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexBlog: true, // Index blog posts in search engine
        indexDocs: false, // Docs plugin is disabled, docs search needs to be disabled too
        lunr:{
          tokenizerSeparator: /[\s\-\$]+/,
        }
      },
    ],
  ],

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: false,
        blog: {
          routeBasePath: '/',
          blogTitle: 'FerretDB Blog',
          showReadingTime: true,
          authorsMapPath: 'authors.yml',
          postsPerPage: 8,

          blogSidebarTitle: 'All posts',
          blogSidebarCount: 'ALL',
          feedOptions: {
            type: 'all',
            title: 'FerretDB Blog',
            description: 'A truly Open Source MongoDB alternative',
            copyright: `Copyright © ${new Date().getFullYear()} FerretDB Inc.`,

            // override to add images; see https://github.com/facebook/docusaurus/discussions/8321#discussioncomment-7016367
            createFeedItems: async (params) => {
              const {
                blogPosts,
                defaultCreateFeedItems,
                siteConfig,
                outDir
              } = params;

              const allFeedItems = await defaultCreateFeedItems({
                blogPosts: blogPosts.slice(0, 10),
                siteConfig: siteConfig,
                outDir: outDir
              });

              return allFeedItems.map((item, index) => ({
                ...item,
                image: `${config.url}${blogPosts[index].metadata.frontMatter.image}`,
              }))

            },
          },
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
            href: 'https://docs.ferretdb.io/',
            position: 'right',
            label: 'Documentation',
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
                href: 'https://docs.ferretdb.io/',
                label: 'Documentation',
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
                to: '/',
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
        additionalLanguages: ['go', 'sql', 'json', 'json5'],
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
