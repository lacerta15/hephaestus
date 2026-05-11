// Hephaestus reference dashboard — small, dependency-free.
// Talks to the API gateway at the configured base URL.

const API = (window.__HEPHAESTUS_API_BASE__ || 'http://localhost:3000') + '/api/v1';

// --- session state -----------------------------------------------------------
const session = {
  token: null,
  username: null,
  org: null,
  isBank: () =>
    ['BCAMSP', 'MandiriMSP', 'BRIMSP', 'BNIMSP', 'CIMBMSP'].includes(session.org),
  isBI: () => session.org === 'BIMSP',
};

// --- helpers -----------------------------------------------------------------
async function api(path, opts = {}) {
  const headers = { 'Content-Type': 'application/json', ...(opts.headers || {}) };
  if (session.token) headers['Authorization'] = `Bearer ${session.token}`;
  const res = await fetch(`${API}${path}`, { ...opts, headers });
  const text = await res.text();
  let body = text;
  try { body = JSON.parse(text); } catch (_) { /* keep as text */ }
  if (!res.ok) {
    const msg = (body && body.error) || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return body;
}

async function sha256Hex(str) {
  const buf = new TextEncoder().encode(str);
  const hash = await crypto.subtle.digest('SHA-256', buf);
  return Array.from(new Uint8Array(hash))
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}

function show(id) { document.getElementById(id).classList.remove('hidden'); }
function hide(id) { document.getElementById(id).classList.add('hidden'); }

function renderSession() {
  const el = document.getElementById('session');
  if (!session.token) {
    el.innerHTML = '<span class="not-logged">Not signed in</span>';
    return;
  }
  el.innerHTML = `<span class="who">${session.username}</span><span class="org">${session.org}</span>
    <button class="secondary" id="logout-btn" style="margin-left:1rem;padding:0.3rem 0.75rem;font-size:0.8rem;">Sign out</button>`;
  document.getElementById('logout-btn').onclick = logout;
}

function logout() {
  session.token = null;
  session.username = null;
  session.org = null;
  ['submit-section', 'validate-section', 'explore-section', 'audit-section'].forEach(hide);
  show('auth-section');
  renderSession();
}

function showSectionsByRole() {
  hide('auth-section');
  show('explore-section');
  show('audit-section');
  if (session.isBank()) show('submit-section');
  if (session.isBI()) show('validate-section');
}

// --- login -------------------------------------------------------------------
document.getElementById('login-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const username = document.getElementById('username').value.trim();
  const org = document.getElementById('org').value;
  try {
    const r = await api('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, org }),
    });
    session.token = r.token;
    session.username = r.username;
    session.org = r.org;
    renderSession();
    showSectionsByRole();
  } catch (err) {
    alert(`Login failed: ${err.message}`);
  }
});

// --- submit ------------------------------------------------------------------
document.getElementById('payloadFile')?.addEventListener('change', async (e) => {
  const f = e.target.files[0];
  if (!f) return;
  const text = await f.text();
  document.getElementById('payloadContent').value = text;
});

document.getElementById('submit-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const payload = document.getElementById('payloadContent').value;
  if (!payload) { alert('Please paste or upload a payload first.'); return; }
  const payloadHash = await sha256Hex(payload);
  const body = {
    bankCode: document.getElementById('bankCode').value.trim(),
    reportType: document.getElementById('reportType').value,
    reportingPeriod: document.getElementById('reportingPeriod').value.trim(),
    payloadHash,
    payloadURI: document.getElementById('payloadURI').value.trim(),
    schemaVersion: document.getElementById('schemaVersion').value.trim(),
  };
  const out = document.getElementById('submit-result');
  try {
    const r = await api('/reports', { method: 'POST', body: JSON.stringify(body) });
    out.classList.remove('error');
    out.textContent = JSON.stringify(r, null, 2);
  } catch (err) {
    out.classList.add('error');
    out.textContent = `Error: ${err.message}`;
  }
});

// --- validate ----------------------------------------------------------------
document.getElementById('validate-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const id = document.getElementById('validate-id').value.trim();
  const out = document.getElementById('validate-result');
  try {
    const r = await api(`/reports/${encodeURIComponent(id)}/validate`, {
      method: 'POST',
      body: JSON.stringify({
        markFinal: document.getElementById('validate-final').checked,
        notes: document.getElementById('validate-notes').value,
      }),
    });
    out.classList.remove('error');
    out.textContent = JSON.stringify(r, null, 2);
  } catch (err) {
    out.classList.add('error');
    out.textContent = `Error: ${err.message}`;
  }
});

// --- explore -----------------------------------------------------------------
document.getElementById('query-btn').addEventListener('click', async () => {
  const mode = document.getElementById('query-mode').value;
  const val = document.getElementById('query-input').value.trim();
  if (!val) return;
  const list = document.getElementById('reports-list');
  list.innerHTML = '<p class="subtitle">Loading...</p>';
  try {
    const r = await api(`/reports?${mode}=${encodeURIComponent(val)}`);
    if (!r.length) { list.innerHTML = '<p class="subtitle">No reports found.</p>'; return; }
    list.innerHTML = r.map(rep => `
      <div class="report-row">
        <div>
          <div><strong>${rep.reportId}</strong></div>
          <div class="subtitle">${rep.bankMSPID} · ${rep.reportType} · ${rep.reportingPeriod}</div>
        </div>
        <div><span class="status-pill status-${rep.status}">${rep.status}</span></div>
        <div class="subtitle">${rep.submittedBy || '—'}</div>
        <button class="secondary" data-id="${rep.reportId}">Audit</button>
      </div>
    `).join('');
    list.querySelectorAll('button[data-id]').forEach(btn => {
      btn.onclick = () => {
        document.getElementById('audit-id').value = btn.dataset.id;
        document.getElementById('audit-btn').click();
        document.getElementById('audit-section').scrollIntoView({ behavior: 'smooth' });
      };
    });
  } catch (err) {
    list.innerHTML = `<p class="result error">Error: ${err.message}</p>`;
  }
});

// --- audit -------------------------------------------------------------------
document.getElementById('audit-btn').addEventListener('click', async () => {
  const id = document.getElementById('audit-id').value.trim();
  if (!id) return;
  const trail = document.getElementById('audit-trail');
  trail.innerHTML = '<p class="subtitle">Loading audit trail...</p>';
  try {
    const r = await api(`/reports/${encodeURIComponent(id)}/audit`);
    if (!r.length) { trail.innerHTML = '<p class="subtitle">No history.</p>'; return; }
    trail.innerHTML = r.map((e, i) => `
      <div class="audit-entry">
        <div class="ts">tx ${e.txId} · ${new Date(e.timestamp).toISOString()}</div>
        <div><strong>v${i + 1}</strong>
          <span class="status-pill status-${e.report.status}">${e.report.status}</span>
          ${e.report.validatedBy ? '· validated by <em>' + e.report.validatedBy + '</em>' : ''}
        </div>
        <div class="subtitle">payload hash: ${e.report.payloadHash}</div>
      </div>
    `).join('');
  } catch (err) {
    trail.innerHTML = `<p class="result error">Error: ${err.message}</p>`;
  }
});
