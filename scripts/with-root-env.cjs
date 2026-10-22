#!/usr/bin/env node
/**
 * Load repo-root `.env` (next to start-all.sh) into process.env without overriding
 * already-set variables, then spawn the remaining argv. Nested `.env` files are ignored.
 */
const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

function findRepoRoot(start) {
  let dir = path.resolve(start);
  for (;;) {
    if (fs.existsSync(path.join(dir, 'start-all.sh'))) {
      return dir;
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      return null;
    }
    dir = parent;
  }
}

function loadEnvFile(filePath) {
  if (!fs.existsSync(filePath)) {
    return;
  }
  const text = fs.readFileSync(filePath, 'utf8');
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim();
    if (!line || line.startsWith('#')) {
      continue;
    }
    const eq = line.indexOf('=');
    if (eq <= 0) {
      continue;
    }
    const key = line.slice(0, eq).trim();
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) {
      continue;
    }
    if (process.env[key] !== undefined) {
      continue;
    }
    let val = line.slice(eq + 1).trim();
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    }
    process.env[key] = val;
  }
}

function aliasLegacyEnv(env) {
  for (const [key, val] of Object.entries(env)) {
    let newKey = '';
    if (key.startsWith('REACT_APP_MORPH_')) {
      newKey = 'REACT_APP_DORPH_' + key.slice('REACT_APP_MORPH_'.length);
    } else if (key.startsWith('VITE_MORPH_')) {
      newKey = 'VITE_DORPH_' + key.slice('VITE_MORPH_'.length);
    } else if (key.startsWith('MORPH_')) {
      newKey = 'DORPH_' + key.slice('MORPH_'.length);
    }
    if (!newKey) continue;
    if (env[newKey] !== undefined && String(env[newKey]).trim() !== '') continue;
    env[newKey] = val;
  }
}

function selfCheck() {
  const prefer = { DORPH_AI_API_KEY: 'new', MORPH_AI_API_KEY: 'old' };
  aliasLegacyEnv(prefer);
  if (prefer.DORPH_AI_API_KEY !== 'new') {
    throw new Error('DORPH_ must win when both are set');
  }
  const fallback = { MORPH_AI_API_KEY: 'old' };
  aliasLegacyEnv(fallback);
  if (fallback.DORPH_AI_API_KEY !== 'old') {
    throw new Error('MORPH_ must copy into DORPH_ when the new name is empty');
  }
  const vite = { VITE_MORPH_AI_URL: 'http://old' };
  aliasLegacyEnv(vite);
  if (vite.VITE_DORPH_AI_URL !== 'http://old') {
    throw new Error('VITE_MORPH_ must copy into VITE_DORPH_');
  }
  const cra = { REACT_APP_MORPH_UTILS_URL: 'http://old' };
  aliasLegacyEnv(cra);
  if (cra.REACT_APP_DORPH_UTILS_URL !== 'http://old') {
    throw new Error('REACT_APP_MORPH_ must copy into REACT_APP_DORPH_');
  }
}

const args = process.argv.slice(2);
if (args[0] === '--self-check') {
  selfCheck();
  process.exit(0);
}

const root = findRepoRoot(__dirname);
if (root) {
  loadEnvFile(path.join(root, '.env'));
  aliasLegacyEnv(process.env);
}

if (!args.length) {
  console.error('usage: with-root-env.cjs <command> [args...]');
  process.exit(1);
}

const child = spawn(args[0], args.slice(1), {
  stdio: 'inherit',
  env: process.env,
});
child.on('exit', (code, signal) => {
  if (signal) {
    process.exit(1);
  }
  process.exit(code ?? 1);
});
