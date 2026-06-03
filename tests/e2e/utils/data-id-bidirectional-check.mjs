import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const BASE = path.dirname(fileURLToPath(import.meta.url))

const vueDirs = [
  path.join(BASE, '../../../typescript/admin-web/src/views/config/'),
  path.join(BASE, '../../../typescript/admin-web/src/views/i18n/'),
]
const specDir = path.join(BASE, '../specs')

function extractIds(dirs, ext) {
  const ids = new Set()
  for (const d of dirs) {
    if (!fs.existsSync(d)) continue
    for (const f of fs.readdirSync(d).filter(x => x.endsWith(ext))) {
      const c = fs.readFileSync(path.join(d, f), 'utf-8')
      for (const m of c.matchAll(/data-id="([^"]+)"/g)) ids.add(m[1])
    }
  }
  return ids
}

const vueIds = extractIds(vueDirs, '.vue')
let specIds = new Set()
if (fs.existsSync(specDir)) {
  specIds = extractIds([specDir], '.spec.ts')
}

console.log(`\ndata-id 双向校验报告`)
console.log(`  Vue 声明: ${vueIds.size}`)
console.log(`  Spec 引用: ${specIds.size}`)

const missingInSpec = [...vueIds].filter(id => !specIds.has(id))
if (missingInSpec.length > 0) {
  console.warn(`\n⚠️  Vue 有 ${missingInSpec.length} 个 data-id 未被 Spec 引用（装饰性/动态绑定，允许）:`)
  missingInSpec.forEach(id => console.warn(`   ${id}`))
}

const missingInVue = [...specIds].filter(id => !vueIds.has(id))
if (missingInVue.length > 0) {
  console.error(`\n❌ Spec 引用了 ${missingInVue.length} 个 data-id 在 Vue 中不存在（HARD GATE）:`)
  missingInVue.forEach(id => console.error(`   ${id}`))
  process.exit(1)
}

if (missingInSpec.length === 0 && missingInVue.length === 0) {
  console.log(`\n✅ 双向校验通过: Vue=${vueIds.size}, Spec=${specIds.size}, 100% 命中\n\n`)
}
