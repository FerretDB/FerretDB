// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import { themes } from "prism-react-renderer";

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "FerretDB",
  tagline: "A truly Open Source MongoDB alternative",

  url: "https://docs.ferretdb.io",
  baseUrl: "/",

  favicon: "img/favicon.ico",
  trailingSlash: true,

  onBrokenAnchors: "throw",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "throw",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  scripts: [{ src: "https://plausible.io/js/script.js", defer: true, "data-domain": "docs.ferretdb.io" }],

  plugins: [
    [
      // https://docusaurus.io/docs/api/plugins/@docusaurus/plugin-client-redirects
      // Note that it does not work in development (`task docs-dev`).
      require.resolve("@docusaurus/plugin-client-redirects"),
      {
        redirects: [
          { to: "/migration/diff", from: "/diff" },
          { to: "/reference", from: ["/reference/supported_commands", "/reference/supported-commands"] },
          { to: "/installation", from: "/quickstart" },
        ],

        createRedirects(existingPath) {
          if (existingPath.startsWith("/installation/ferretdb")) {
            return [
              // old blog posts
              // for example: /quickstart-guide/docker/ -> /installation/ferretdb/docker/
              existingPath.replace("/installation/ferretdb", "/quickstart-guide"),
              existingPath.replace("/installation/ferretdb", "/quickstart_guide"),
            ];
          }

          return undefined;
        },
      },
    ],
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexBlog: true, // Index blog posts in search engine
        indexDocs: true, // Blog plugin is disabled, blog search needs to be disabled too
        lunr: {
          tokenizerSeparator: /[\s\-\$]+/,
        },
      },
    ],
    "plugin-image-zoom",
  ],

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: "/",
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/FerretDB/FerretDB/tree/main/website",

          // https://docusaurus.io/docs/versioning#configuring-versioning-behavior
          // https://docusaurus.io/docs/api/plugins/@docusaurus/plugin-content-docs#configuration
          lastVersion: "current",
          versions: {
            current: {
              label: "v2.0 RC",
              banner: "none",
            },
            "v1.24": {
              label: "v1.24",
              path: "v1.24",
              banner: "none",
            },
          },
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: "img/logo-dark.jpg",
      navbar: {
        logo: {
          alt: "FerretDB Logo",
          src: "img/logo-dark.jpg",
          srcDark: "img/logo-light.png",
        },
        items: [
          {
            to: "/",
            label: "Documentation",
            position: "right",
            type: "docsVersionDropdown",
          },
          {
            href: "https://blog.ferretdb.io/",
            label: "Blog",
            position: "right",
          },
          {
            href: "https://github.com/FerretDB/",
            label: "GitHub",
            position: "right",
          },
          {
            href: "https://www.ferretdb.com/",
            label: "FerretDB.com",
            position: "right",
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: "FerretDB Docs",
            items: [
              {
                label: "Documentation",
                to: "/",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "GitHub Discussions",
                href: "https://github.com/FerretDB/FerretDB/discussions/",
              },
              {
                label: "Slack",
                href: "https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A",
              },
              {
                label: "X (Twitter)",
                href: "https://x.com/ferret_db",
              },
              {
                label: "Mastodon",
                href: "https://techhub.social/@ferretdb",
              },
            ],
          },
          {
            title: "More",
            items: [
              {
                href: "https://www.ferretdb.com/",
                label: "FerretDB.com",
                position: "right",
              },
              {
                label: "Blog",
                to: "https://blog.ferretdb.io/",
              },
              {
                label: "GitHub",
                href: "https://github.com/FerretDB/",
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} FerretDB Inc. Built with Docusaurus.`,
      },
      prism: {
        theme: themes.github,
        darkTheme: themes.dracula,
        additionalLanguages: ["go", "sql", "json", "json5", "systemd"],
      },
      mermaid: {
        theme: { light: "default", dark: "dark" },
      },
    }),
  markdown: {
    mermaid: true,
  },
  themes: ["@docusaurus/theme-mermaid"],
};

module.exports = config;
