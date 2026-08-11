const list = document.querySelector('#user-list')
const message = document.querySelector('#admin-message')
const createForm = document.querySelector('#create-user-form')
const createSubmit = document.querySelector('#create-user-submit')
let currentUser = null

function showMessage(text, tone = 'neutral') {
  message.textContent = text
  message.dataset.tone = tone
  message.hidden = !text
}

async function request(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    credentials: 'same-origin',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json', ...(options.headers || {}) },
  })
  const payload = response.status === 204 ? null : await response.json().catch(() => null)
  if (!response.ok) {
    const error = new Error(payload?.message || `请求失败（HTTP ${response.status}）`)
    error.code = payload?.code
    error.status = response.status
    throw error
  }
  return payload
}

function field(labelText, control) {
  const label = document.createElement('label')
  label.className = 'field'
  label.append(labelText, control)
  return label
}

function renderUsers(users) {
  list.replaceChildren()
  if (!users.length) {
    list.textContent = '暂无用户。'
    return
  }

  for (const user of users) {
    const row = document.createElement('form')
    row.className = 'user-row'
    row.dataset.userId = user.id

    const identity = document.createElement('div')
    identity.className = 'user-name'
    const name = document.createElement('strong')
    name.textContent = user.username
    const meta = document.createElement('small')
    meta.textContent = user.must_change_password ? '等待修改临时密码' : '密码已由用户设置'
    identity.append(name, meta)

    const role = document.createElement('select')
    role.innerHTML = '<option value="user">普通用户</option><option value="admin">管理员</option>'
    role.value = user.role

    const activeLabel = document.createElement('label')
    activeLabel.className = 'checkbox-field'
    const active = document.createElement('input')
    active.type = 'checkbox'
    active.checked = user.is_active
    activeLabel.append(active, '账号启用')

    const password = document.createElement('input')
    password.type = 'password'
    password.minLength = 12
    password.maxLength = 72
    password.placeholder = '留空则不重置'
    password.autocomplete = 'new-password'

    const save = document.createElement('button')
    save.className = 'button primary'
    save.type = 'submit'
    save.textContent = '保存'

    if (user.id === currentUser.id) {
      role.disabled = true
      active.disabled = true
      meta.textContent += ' · 当前账号'
    }

    row.append(identity, field('角色', role), activeLabel, field('新临时密码', password), save)
    row.addEventListener('submit', async (event) => {
      event.preventDefault()
      save.disabled = true
      const body = { role: role.value, is_active: active.checked }
      if (password.value) body.password = password.value
      try {
        await request(`/api/admin/users/${encodeURIComponent(user.id)}`, {
          method: 'PATCH',
          body: JSON.stringify(body),
        })
        password.value = ''
        showMessage(`已更新用户 ${user.username}。`, 'success')
        await loadUsers()
      } catch (error) {
        showMessage(error.message, 'danger')
      } finally {
        save.disabled = false
      }
    })
    list.append(row)
  }
}

async function loadUsers() {
  list.textContent = '正在加载…'
  try {
    const { users } = await request('/api/admin/users')
    renderUsers(users)
  } catch (error) {
    list.textContent = ''
    showMessage(error.message, 'danger')
  }
}

createForm.addEventListener('submit', async (event) => {
  event.preventDefault()
  createSubmit.disabled = true
  try {
    await request('/api/admin/users', {
      method: 'POST',
      body: JSON.stringify({
        username: document.querySelector('#new-username').value,
        password: document.querySelector('#new-user-password').value,
        role: document.querySelector('#new-user-role').value,
      }),
    })
    createForm.reset()
    showMessage('用户已创建。', 'success')
    await loadUsers()
  } catch (error) {
    showMessage(error.code === 'username_exists' ? '该用户名已经存在。' : error.message, 'danger')
  } finally {
    createSubmit.disabled = false
  }
})

document.querySelector('#refresh-users').addEventListener('click', loadUsers)

async function bootstrap() {
  try {
    const { user } = await request('/api/auth/me')
    if (user.must_change_password) {
      window.location.replace('/change-password/')
      return
    }
    if (user.role !== 'admin') {
      window.location.replace('/memo/')
      return
    }
    currentUser = user
    await loadUsers()
  } catch (error) {
    if (error.status === 401) window.location.replace('/memo/')
    else showMessage('无法初始化管理页面。', 'danger')
  }
}

bootstrap()
