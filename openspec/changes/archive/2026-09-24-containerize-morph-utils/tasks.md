## 1. URL resolver

- [x] 1.1 Add a failing `node --experimental-strip-types --test` for `publicUrl.ts`: first non-empty wins (runtime primary, runtime alias, built primary, built alias); blank runtime does not hide a built value; dev localhost fallback is used only when every candidate is blank and `dev` is true; production (`dev` false) stays empty
- [x] 1.2 Implement `publicUrl.ts` so that test passes, and add an `npm test` script that runs it
- [x] 1.3 Wire `auth.ts` and `config.ts` through the resolver, with localhost embed literals only behind `import.meta.env.DEV ? … : ''`. Load `/config.js` from `index.html` before the module, and add a no-op `public/config.js` for dev

## 2. Image

- [x] 2.1 Add `morph-utils/Dockerfile` (Node build with empty public URL `ARG`s, nginx runtime, `curl` `HEALTHCHECK` on `${PORT}` `/health`, no secret `ARG`), `.dockerignore`, nginx template (`/health`, `/ready`, `config.js` no-store, SPA fallback), and an entrypoint that JSON-escapes non-empty `VITE_*` values into `config.js` and substitutes only `${PORT}`
- [x] 2.2 Add `morph-utils/docker-compose.yml` that publishes `127.0.0.1:${MORPH_UTILS_PUBLISH_PORT:-3040}` to container port 3040 and passes the public URL variables through
- [x] 2.3 Add `morph-utils/deploy/check-container-contract.sh` covering health paths, no `$$`, no secret `ARG`, env files ignored, and the nine `VITE_*` names present in the Dockerfile, entrypoint, and README
- [x] 2.4 Document build, run, `PORT`, and the public URL variables in `morph-utils/README.md`

## 3. Verify

- [x] 3.1 Run the resolver test, the contract script, and `npm run build`. Confirm the production assets do not contain the localhost embed defaults
- [x] 3.2 Build and run the image with `VITE_MORPH_API_URL` set to a non-loopback origin. `GET /health` and `GET /ready` return 200. `config.js` contains that origin. The image history and build args contain no JWT, password, or API key
