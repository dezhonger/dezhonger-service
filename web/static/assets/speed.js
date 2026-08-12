const runButton = document.querySelector('#run-speed-test')
const message = document.querySelector('#speed-message')
const chart = document.querySelector('#speed-chart')
const t = (english, chinese) => window.DezhongerI18n?.t(english, chinese) || english

function setText(id, value) {
  document.querySelector(id).textContent = value
}

function milliseconds(value) {
  return Number.isFinite(value) ? `${value.toFixed(1)} ms` : '—'
}

function percentile(sorted, ratio) {
  return sorted[Math.min(sorted.length - 1, Math.ceil(sorted.length * ratio) - 1)]
}

function latencyGrade(value) {
  if (value < 80) return t('Excellent connection. Interactions should feel immediate.', '连接优秀，交互响应会很流畅。')
  if (value < 150) return t('Good connection for everyday use.', '连接良好，适合日常使用。')
  if (value < 300) return t('Moderate latency. Some actions may have a noticeable delay.', '延迟一般，部分操作可能有可感知等待。')
  return t('High latency. Check your local network or proxy route.', '延迟较高，请检查本地网络或代理线路。')
}

function drawChart(samples) {
  const ratio = window.devicePixelRatio || 1
  const width = Math.max(chart.clientWidth, 320)
  const height = Math.max(chart.clientHeight, 160)
  chart.width = Math.round(width * ratio)
  chart.height = Math.round(height * ratio)
  const context = chart.getContext('2d')
  context.scale(ratio, ratio)
  context.clearRect(0, 0, width, height)

  const padding = 18
  const max = Math.max(...samples, 1) * 1.18
  const step = (width - padding * 2) / Math.max(samples.length - 1, 1)
  context.strokeStyle = '#e5e5e5'
  context.lineWidth = 1
  for (let line = 0; line < 4; line += 1) {
    const y = padding + ((height - padding * 2) * line) / 3
    context.beginPath()
    context.moveTo(padding, y)
    context.lineTo(width - padding, y)
    context.stroke()
  }

  context.strokeStyle = '#10a37f'
  context.lineWidth = 2
  context.lineJoin = 'round'
  context.lineCap = 'round'
  context.beginPath()
  samples.forEach((sample, index) => {
    const x = padding + index * step
    const y = height - padding - (sample / max) * (height - padding * 2)
    if (index === 0) context.moveTo(x, y)
    else context.lineTo(x, y)
  })
  context.stroke()
}

async function runSpeedTest() {
  runButton.disabled = true
  runButton.textContent = t('Testing…', '测速中…')
  message.hidden = true
  const samples = []
  let failed = 0
  let region = 'Tencent Cloud · Hong Kong'

  for (let index = 0; index < 12; index += 1) {
    const started = performance.now()
    try {
      const response = await fetch(`/api/ping?sample=${Date.now()}-${index}`, {
        cache: 'no-store',
        credentials: 'same-origin',
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const payload = await response.json()
      region = payload.region || region
      samples.push(performance.now() - started)
    } catch {
      failed += 1
    }
    await new Promise(resolve => window.setTimeout(resolve, 90))
  }

  if (!samples.length) {
    message.textContent = t('All test requests failed. Refresh the page or check your network and try again.', '测速请求全部失败，请刷新页面或检查网络后重试。')
    message.dataset.tone = 'danger'
    message.hidden = false
    setText('#latency-grade', t('The server is temporarily unreachable.', '暂时无法连接服务器'))
  } else {
    const sorted = [...samples].sort((left, right) => left - right)
    const average = samples.reduce((sum, value) => sum + value, 0) / samples.length
    const differences = samples.slice(1).map((value, index) => Math.abs(value - samples[index]))
    const jitter = differences.length ? differences.reduce((sum, value) => sum + value, 0) / differences.length : 0
    document.querySelector('#latency-value').innerHTML = `${average.toFixed(1)}<span>ms</span>`
    setText('#latency-grade', latencyGrade(average))
    setText('#metric-min', milliseconds(sorted[0]))
    setText('#metric-median', milliseconds(percentile(sorted, .5)))
    setText('#metric-p95', milliseconds(percentile(sorted, .95)))
    setText('#metric-jitter', milliseconds(jitter))
    setText('#metric-success', `${samples.length}/12`)
    setText('#metric-region', region.replace('Tencent Cloud · ', ''))
    window.lastSpeedSamples = samples
    if (failed) {
      message.textContent = t(`${failed} requests failed; results use successful samples only.`, `${failed} 次请求失败，结果仅基于成功样本。`)
      message.dataset.tone = 'danger'
      message.hidden = false
    }
    drawChart(samples)
  }

  runButton.disabled = false
  runButton.textContent = t('Run again', '重新测速')
}

function renderNavigationTiming() {
  const entry = performance.getEntriesByType('navigation')[0]
  if (!entry) return

  const total = entry.loadEventEnd || entry.duration || performance.now()
  setText('#nav-dns', milliseconds(Math.max(0, entry.domainLookupEnd - entry.domainLookupStart)))
  setText('#nav-tcp', milliseconds(Math.max(0, entry.connectEnd - entry.connectStart)))
  const tls = entry.secureConnectionStart > 0 ? entry.connectEnd - entry.secureConnectionStart : 0
  setText('#nav-tls', milliseconds(Math.max(0, tls)))
  setText('#nav-ttfb', milliseconds(Math.max(0, entry.responseStart - entry.requestStart)))
  setText('#nav-total', milliseconds(total))
  setText('#navigation-total', t(`Total page time ${milliseconds(total)}`, `页面总耗时 ${milliseconds(total)}`))
}

runButton.addEventListener('click', runSpeedTest)
window.addEventListener('resize', () => {
  if (window.lastSpeedSamples) drawChart(window.lastSpeedSamples)
})
window.addEventListener('load', () => window.setTimeout(renderNavigationTiming, 0), { once: true })
runSpeedTest()
