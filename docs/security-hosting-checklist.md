# Security and hosting checklist

Use this before running Morph anywhere other than a local `./start-all.sh` checkout. Local mode needs none of the overrides below.

`./start-all.sh` does not set `MORPH_ENV`. Unset, `development`, `dev`, `local`, and `test` are local mode: Morph starts with the development admin login from the [root README](../README.md) and logs a warning. Do not commit `.env`.

Set these on the host process environment. This repo does not ship a production secret, and the `deploy/` tree is not a complete hosting runbook.

## Before the process starts

- [ ] `MORPH_ENV=production` (`prod` is accepted). Any other value except the local names above refuses to start. The error names `MORPH_ENV` and does not print secret values.
- [ ] `MORPH_SECRETS_KEY` is 32 random bytes of standard base64 (`openssl rand -base64 32`) before any provider API key is stored. It is not derived from `JWT_SECRET`. Rotation uses `MORPH_SECRETS_KEY_PREVIOUS`. See `docs/agents/14-secrets-key.md`. Startup does not load this key yet. The follow-up passes the bool from `config.ParseMorphEnv()` into `secretbox.Resolve` after `config.ValidateStartup` and `NewTranSQL`.
- [ ] `JWT_SECRET` is a unique random string of at least 32 characters. Do not keep the development value from `.env.example`. When first enabling `MORPH_ENV=production`, set a **new** `JWT_SECRET` even if the current secret is already strong. That invalidates every existing session, including tokens issued for the old 876000-hour lifetime. Morph also rejects tokens with no `iat` or `exp`, and tokens whose `exp` minus `iat` is longer than 168 hours. Other services that validate Morph tokens with this same secret do not apply that window. Project (`morph-engi`) reads `JWT_SECRET` and must get the new value too; if the variable is unset there, Project falls back to its own development secret.
- [ ] `ADMIN_PASSWORD` is a unique password of at least 12 characters. `ADMIN_PASSWORD` is used before `BOOTSTRAP_ADMIN_PASSWORD`. Do not keep the development value from `.env.example`.
- [ ] If `plat_users` was created on an earlier run, set `MORPH_ROTATE_DEFAULT_ADMIN=1` for one production start. Morph replaces stored development passwords and keeps account ids. Unset the flag afterward. Leaving it set does not reset a password that no longer matches the development default. The flag does nothing unless `MORPH_ENV` is production.
- [ ] `JWT_EXPIRY_HOURS` is a whole number from 1 to 168, or unset (production default is 24 hours). Do not leave the local value `876000`.
- [ ] `MORPH_AI_API_KEY` is set for the model provider. Set `TRAN_OPENAI_API_KEY` as well when Content Maker embeddings are used.
- [ ] `ADMIN_EMAIL` and `ADMIN_USERNAME` are the operator you want, if not the local defaults.
- [ ] Confirm `.env` is not committed. Rotate any key that was shared or checked in.

Startup errors name the variable to set (`JWT_SECRET`, `ADMIN_PASSWORD`, `JWT_EXPIRY_HOURS`, `MORPH_ENV`, `MORPH_ROTATE_DEFAULT_ADMIN`) and do not print the secret or password.

Provider API keys stored inside the app are not encrypted at rest. That is a separate change.
