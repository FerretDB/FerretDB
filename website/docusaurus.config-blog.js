// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

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

  scripts: [{src: 'https://plausible.io/js/script.js', defer: true, "data-domain": "blog.ferretdb.io"}],

  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexBlog: true, // Index blog posts in search engine
        indexDocs: false, // Docs plugin is disabled, docs search needs to be disabled too
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
              } = params;

              return blogPosts.slice(0, 10).map(post => ({
                title: post.metadata.title,
                link: post.metadata.frontMatter.link,
                image: `${config.url}${post.metadata.frontMatter.image}`,
                date: post.metadata.date,
                description: post.metadata.description,
              }));
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
            to: '/',
            label: 'Blog',
            position: 'left'
          },
          {
            href: 'https://docs.ferretdb.io/',
            position: 'left',
            label: 'Documentation',
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
                href: 'https://ferretdb.io/',
                label: 'Go to FerretDB.io',
                position: 'right',
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
        additionalLanguages: ['go', 'json5', 'sql'],
      },
    }),
};

module.exports = config;
