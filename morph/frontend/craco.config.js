/**
 * CRA (react-scripts) warns when a dependency's .map file points at unpublished sources
 * (e.g. dompurify via jspdf). Those maps are optional; suppress the noise.
 */
module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      const ignore = [/Failed to parse source map/];
      if (Array.isArray(webpackConfig.ignoreWarnings)) {
        webpackConfig.ignoreWarnings.push(...ignore);
      } else {
        webpackConfig.ignoreWarnings = ignore;
      }
      return webpackConfig;
    },
  },
};
