const form = document.querySelector('#password-form')
const submit = document.querySelector('#password-submit')
const message = document.querySelector('#password-message')
const accountCopy = document.querySelector('#account-copy')
const t = (english, chinese) => window.DezhongerI18n?.t(english, chinese) || english
const localize = (text) => window.DezhongerI18n?.translateString(text) || text

function showMessage(text, tone = 'neutral') {
  message.textContent = localize(text)
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
    const error = new Error(payload?.message || t(`Request failed (HTTP ${response.status})`, `请求失败（HTTP ${response.status}）`))
    error.code = payload?.code
    error.status = response.status
    throw error
  }
  return payload
}

async function bootstrap() {
  try {
    const { user } = await request('/api/auth/me')
    accountCopy.textContent = user.must_change_password
      ? t(`${user.username}, you signed in with a temporary password. Set a new password first.`, `${user.username}，这是临时密码登录，请先设置自己的新密码。`)
      : t(`${user.username}, enter your current password to set a new one.`, `${user.username}，输入当前密码后即可更新密码。`)
  } catch (error) {
    if (error.status === 401) window.location.replace('/memo/')
    else showMessage(t('The API is temporarily unavailable. Refresh later.', '暂时无法连接 API，请稍后刷新。'), 'danger')
  }
}

form.addEventListener('submit', async (event) => {
  event.preventDefault()
  const currentPassword = document.querySelector('#current-password').value
  const newPassword = document.querySelector('#new-password').value
  const confirmPassword = document.querySelector('#confirm-password').value
  if (newPassword !== confirmPassword) {
    showMessage(t('The new passwords do not match.', '两次输入的新密码不一致。'), 'danger')
    return
  }
  if (newPassword === currentPassword) {
    showMessage(t('The new password must differ from the current password.', '新密码不能与当前密码相同。'), 'danger')
    return
  }

  submit.disabled = true
  showMessage(t('Saving…', '正在保存…'))
  try {
    await request('/api/account/password', {
      method: 'PUT',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    })
    form.reset()
    showMessage(t('Password updated and the session was signed out safely. Returning to sign-in…', '密码已更新，登录状态已安全退出。正在返回登录页…'), 'success')
    window.setTimeout(() => window.location.replace('/memo/'), 1200)
  } catch (error) {
    showMessage(error.code === 'invalid_credentials' ? t('The current password is incorrect.', '当前密码不正确。') : error.message, 'danger')
  } finally {
    submit.disabled = false
  }
})

bootstrap()
