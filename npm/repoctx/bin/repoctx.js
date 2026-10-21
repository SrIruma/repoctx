#!/usr/bin/env node
'use strict';
const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const name = `@sriruma/repoctx-${process.platform}-${process.arch}`;
let dir;
try {
  dir = path.dirname(require.resolve(`${name}/package.json`));
} catch (e) {
  console.error(`repoctx: platform package ${name} is not installed.`);
  console.error('Reinstall @sriruma/repoctx, or install the matching platform package manually.');
  process.exit(1);
}
const binName = process.platform === 'win32' ? 'repoctx.exe' : 'repoctx';
const binPath = path.join(dir, 'bin', binName);
if (!fs.existsSync(binPath)) {
  console.error(`repoctx: binary not found at ${binPath}`);
  process.exit(1);
}
const r = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
process.exit(r.status === null ? 1 : r.status);
