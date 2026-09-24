/**
 * CRA + mermaid@11.
 *
 * mermaid's package exports point webpack at dist/mermaid.core.mjs. Those
 * chunks are strict ESM and import dayjs (and dayjs/plugin/*) as bare
 * specifiers. Webpack's fullySpecified resolver then reports
 * "Can't resolve 'dayjs'" from the chunk directory. Turn fullySpecified off
 * for mermaid's own .mjs files so those imports resolve through node_modules.
 * Do not alias dayjs to an absolute path: CRA's ModuleScopePlugin treats that
 * as an import from outside src/ and fails the build.
 *
 * The same package ships source maps whose "sources" point at unpublished
 * src/*.ts. source-map-loader turns each missing file into
 * "Failed to parse source map ... ENOENT". Exclude mermaid from that loader.
 * Other third-party maps (dompurify via jspdf) stay warnings and are ignored
 * by message, not by turning source maps or CI checks off.
 */
const MERMAID_DIR = /[\\/]node_modules[\\/]mermaid[\\/]/;

module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      const ignore = [/Failed to parse source map/];
      if (Array.isArray(webpackConfig.ignoreWarnings)) {
        webpackConfig.ignoreWarnings.push(...ignore);
      } else {
        webpackConfig.ignoreWarnings = ignore;
      }

      for (const rule of webpackConfig.module.rules) {
        if (!rule || typeof rule.loader !== 'string') continue;
        if (!rule.loader.includes('source-map-loader')) continue;
        if (!rule.exclude) {
          rule.exclude = MERMAID_DIR;
        } else if (Array.isArray(rule.exclude)) {
          rule.exclude.push(MERMAID_DIR);
        } else {
          rule.exclude = [rule.exclude, MERMAID_DIR];
        }
      }

      webpackConfig.module.rules.push({
        test: /\.mjs$/,
        include: MERMAID_DIR,
        resolve: {
          fullySpecified: false,
        },
      });

      return webpackConfig;
    },
  },
};
