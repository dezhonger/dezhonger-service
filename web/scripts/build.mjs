import { cp, mkdir, rm } from 'node:fs/promises'
import { resolve } from 'node:path'

import { build } from 'esbuild'

const root = resolve(import.meta.dirname, '..')
const output = resolve(root, 'dist')

await rm(output, { force: true, recursive: true })
await mkdir(output, { recursive: true })
await cp(resolve(root, 'static'), output, { recursive: true })

await build({
  entryPoints: [resolve(root, 'src/memo-app.mjs')],
  bundle: true,
  minify: true,
  format: 'esm',
  target: 'es2020',
  outfile: resolve(output, 'assets/memo/memo-app.js'),
})
