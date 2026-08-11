const tableList = document.querySelector('#db-table-list')
const tableCount = document.querySelector('#db-table-count')
const emptyView = document.querySelector('#db-empty')
const contentView = document.querySelector('#db-content')
const output = document.querySelector('#db-view')
const message = document.querySelector('#db-message')
const pagination = document.querySelector('#db-pagination')
const dataTab = document.querySelector('#db-data-tab')
const structureTab = document.querySelector('#db-structure-tab')
const pageSize = 25
let tables = []
let selectedTable = ''
let tableData = null
let offset = 0
let mode = 'data'

async function request(path) {
  const response = await fetch(path, { credentials: 'same-origin', headers: { Accept: 'application/json' } })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    const error = new Error(payload?.message || `请求失败（HTTP ${response.status}）`)
    error.status = response.status
    error.code = payload?.code
    throw error
  }
  return payload
}

function showMessage(text, tone = 'neutral') {
  message.textContent = text
  message.dataset.tone = tone
  message.hidden = !text
}

function cellText(value) {
  if (value === null) return 'NULL'
  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  return String(value)
}

function appendCell(row, value, sensitive = false) {
  const cell = document.createElement('td')
  const content = document.createElement('pre')
  content.className = 'db-cell'
  if (value === null) content.classList.add('db-null')
  if (sensitive) content.classList.add('db-sensitive')
  content.textContent = cellText(value)
  cell.append(content)
  row.append(cell)
}

function renderData() {
  output.replaceChildren()
  pagination.hidden = false
  const wrap = document.createElement('div')
  wrap.className = 'db-table-wrap'
  const table = document.createElement('table')
  table.className = 'db-data-table'
  const head = document.createElement('thead')
  const headRow = document.createElement('tr')
  tableData.columns.forEach(column => {
    const cell = document.createElement('th')
    cell.textContent = column.name
    headRow.append(cell)
  })
  head.append(headRow)
  table.append(head)

  const body = document.createElement('tbody')
  if (!tableData.rows.length) {
    const row = document.createElement('tr')
    const cell = document.createElement('td')
    cell.colSpan = Math.max(tableData.columns.length, 1)
    cell.textContent = '当前页没有数据。'
    row.append(cell)
    body.append(row)
  } else {
    tableData.rows.forEach(record => {
      const row = document.createElement('tr')
      tableData.columns.forEach(column => appendCell(row, record[column.name], column.sensitive))
      body.append(row)
    })
  }
  table.append(body)
  wrap.append(table)
  output.append(wrap)

  const start = tableData.total ? tableData.offset + 1 : 0
  const end = Math.min(tableData.offset + tableData.rows.length, tableData.total)
  document.querySelector('#db-page-label').textContent = `${start}–${end} / ${tableData.total}`
  document.querySelector('#db-prev').disabled = tableData.offset === 0
  document.querySelector('#db-next').disabled = tableData.offset + tableData.limit >= tableData.total
}

function renderStructure() {
  output.replaceChildren()
  pagination.hidden = true
  const wrap = document.createElement('div')
  wrap.className = 'db-table-wrap'
  const table = document.createElement('table')
  table.className = 'db-data-table'
  const head = document.createElement('thead')
  const headRow = document.createElement('tr')
  ;['字段', '类型', '可空', '默认值', '保护'].forEach(label => {
    const cell = document.createElement('th')
    cell.textContent = label
    headRow.append(cell)
  })
  head.append(headRow)
  table.append(head)
  const body = document.createElement('tbody')
  tableData.columns.forEach(column => {
    const row = document.createElement('tr')
    const name = document.createElement('td')
    name.textContent = column.name
    if (column.primary_key) {
      const key = document.createElement('span')
      key.className = 'db-key'
      key.textContent = 'PK'
      name.append(key)
    }
    row.append(name)
    appendCell(row, column.data_type === 'ARRAY' ? column.udt_name : column.data_type)
    appendCell(row, column.nullable ? 'YES' : 'NO')
    appendCell(row, column.default)
    appendCell(row, column.sensitive ? '值已隐藏' : '—', column.sensitive)
    body.append(row)
  })
  table.append(body)
  wrap.append(table)
  output.append(wrap)
}

function renderCurrentView() {
  dataTab.setAttribute('aria-pressed', String(mode === 'data'))
  structureTab.setAttribute('aria-pressed', String(mode === 'structure'))
  if (!tableData) return
  if (mode === 'data') renderData()
  else renderStructure()
}

function renderTableList() {
  tableList.replaceChildren()
  tables.forEach(table => {
    const button = document.createElement('button')
    button.type = 'button'
    button.className = 'db-table-button'
    button.setAttribute('aria-current', String(table.name === selectedTable))
    const name = document.createElement('span')
    name.textContent = table.name
    const count = document.createElement('small')
    count.textContent = String(table.row_count)
    button.append(name, count)
    button.addEventListener('click', () => selectTable(table.name))
    tableList.append(button)
  })
}

async function loadTable() {
  if (!selectedTable) return
  showMessage('')
  output.textContent = '正在读取表…'
  try {
    tableData = await request(`/api/admin/database/tables/${encodeURIComponent(selectedTable)}?limit=${pageSize}&offset=${offset}`)
    document.querySelector('#db-title').textContent = tableData.table
    document.querySelector('#db-subtitle').textContent = `${tableData.columns.length} 个字段 · ${tableData.total} 行 · 只读`
    renderCurrentView()
  } catch (error) {
    output.textContent = ''
    showMessage(error.message, 'danger')
  }
}

async function selectTable(name) {
  selectedTable = name
  offset = 0
  tableData = null
  emptyView.hidden = true
  contentView.hidden = false
  renderTableList()
  await loadTable()
}

async function bootstrap() {
  try {
    const { user } = await request('/api/auth/me')
    if (user.must_change_password) return window.location.replace('/change-password/')
    if (user.role !== 'admin') return window.location.replace('/memo/')
    const payload = await request('/api/admin/database/tables')
    tables = payload.tables
    tableCount.textContent = `${tables.length} 张表`
    renderTableList()
    if (tables.length) await selectTable(tables[0].name)
  } catch (error) {
    if (error.status === 401) window.location.replace('/memo/')
    else showMessage(error.message, 'danger')
  }
}

dataTab.addEventListener('click', () => { mode = 'data'; renderCurrentView() })
structureTab.addEventListener('click', () => { mode = 'structure'; renderCurrentView() })
document.querySelector('#db-refresh').addEventListener('click', loadTable)
document.querySelector('#db-prev').addEventListener('click', async () => { offset = Math.max(0, offset - pageSize); await loadTable() })
document.querySelector('#db-next').addEventListener('click', async () => { offset += pageSize; await loadTable() })
bootstrap()
