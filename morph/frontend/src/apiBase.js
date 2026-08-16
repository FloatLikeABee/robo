/**
 * API origin for axios/fetch. Default '' = same origin as the React app.
 * In CRA dev, `package.json` "proxy" forwards /api to the Go server, so CORS is avoided.
 * Set REACT_APP_API_URL when the UI is served from a different host than the API.
 */
export const API_BASE_URL = process.env.REACT_APP_API_URL || '';
