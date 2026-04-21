(function () {
  'use strict';

  const $ = (sel) => document.querySelector(sel);
  const $$ = (sel) => document.querySelectorAll(sel);

  // ── State ──
  const state = {
    token: localStorage.getItem('mtv_token'),
    user: null,
  };

  // ── API Client ──
  const api = {
    base: '/api',
    async request(method, path, body) {
      const headers = { 'Content-Type': 'application/json' };
      if (state.token) headers['Authorization'] = `Bearer ${state.token}`;
      const res = await fetch(this.base + path, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });
      const data = await res.json();
      if (data.code !== 0) throw new Error(data.message || 'request failed');
      return data.data;
    },
    get(p) { return this.request('GET', p); },
    post(p, b) { return this.request('POST', p, b); },
    put(p, b) { return this.request('PUT', p, b); },
    del(p, b) { return this.request('DELETE', p, b); },
  };

  // ── Toast ──
  function toast(msg, type = 'info') {
    const el = document.createElement('div');
    el.className = `toast toast-${type}`;
    el.textContent = msg;
    $('#toast-container').appendChild(el);
    setTimeout(() => el.remove(), 3000);
  }

  // ── Modal ──
  function openModal(title, bodyHTML) {
    $('#modal-title').textContent = title;
    $('#modal-body').innerHTML = bodyHTML;
    $('#modal-overlay').classList.remove('hidden');
  }
  function closeModal() {
    $('#modal-overlay').classList.add('hidden');
  }

  // ── Auth ──
  function parseToken(token) {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return { id: payload.user_id, username: payload.username, role: payload.role };
    } catch { return null; }
  }

  function login(token, user) {
    state.token = token;
    state.user = user;
    localStorage.setItem('mtv_token', token);
    showMain();
  }

  function logout() {
    state.token = null;
    state.user = null;
    localStorage.removeItem('mtv_token');
    showLogin();
  }

  function showLogin() {
    $('#login-page').classList.remove('hidden');
    $('#main-layout').classList.add('hidden');
    $('#login-username').focus();
  }

  function showMain() {
    state.user = parseToken(state.token);
    if (!state.user) { logout(); return; }
    $('#login-page').classList.add('hidden');
    $('#main-layout').classList.remove('hidden');
    $('#sidebar-username').textContent = state.user.username;
    const roleEl = $('#sidebar-role');
    roleEl.textContent = state.user.role;
    roleEl.className = 'badge badge-' + state.user.role;

    if (state.user.role !== 'admin') {
      $$('.admin-only').forEach(el => el.classList.add('hidden'));
    } else {
      $$('.admin-only').forEach(el => el.classList.remove('hidden'));
    }

    route();
  }

  // ── Router ──
  const routes = {
    '#/': renderDashboard,
    '#/users': renderUsers,
    '#/invites': renderInvites,
    '#/sources': renderSources,
    '#/account': renderAccount,
    '#/docs': renderDocs,
  };

  function route() {
    const hash = location.hash || '#/';
    const render = routes[hash] || renderDashboard;
    $$('.nav-item').forEach(el => el.classList.toggle('active', el.getAttribute('href') === hash));
    render($('#content'));
  }

  // ── Pages ──

  // Dashboard
  async function renderDashboard(el) {
    el.innerHTML = '<div class="page-header"><h2>仪表盘</h2></div><div class="stats-grid" id="stats-grid">加载中...</div>';
    try {
      const s = await api.get('/admin/stats');
      $('#stats-grid').innerHTML = `
        <div class="stat-card"><div class="stat-label">总用户</div><div class="stat-value">${s.total_users}</div></div>
        <div class="stat-card"><div class="stat-label">活跃用户</div><div class="stat-value">${s.active_users}</div></div>
        <div class="stat-card"><div class="stat-label">采集源总数</div><div class="stat-value">${s.total_sources}</div></div>
        <div class="stat-card"><div class="stat-label">今日 API 调用</div><div class="stat-value">${s.api_calls_today}</div></div>
        <div class="stat-card"><div class="stat-label">7 日 API 调用</div><div class="stat-value">${s.api_calls_7days}</div></div>
      `;
    } catch (e) {
      $('#stats-grid').innerHTML = `<div class="empty">加载失败：${esc(e.message)}</div>`;
    }
  }

  // Users
  let usersPage = 1;
  async function renderUsers(el) {
    el.innerHTML = '<div class="page-header"><h2>用户管理</h2></div><div id="users-table">加载中...</div>';
    await loadUsers();
  }

  async function loadUsers() {
    try {
      const d = await api.get(`/admin/users?page=${usersPage}&page_size=20`);
      const users = d.users || [];
      const total = d.total || 0;
      const pages = Math.ceil(total / 20);

      let rows = users.map(u => `
        <tr>
          <td>${u.id}</td>
          <td>${esc(u.username)}</td>
          <td><span class="badge badge-${u.role}">${u.role}</span></td>
          <td>${u.banned ? '<span class="badge badge-danger">已封禁</span>' : '<span class="badge badge-success">正常</span>'}</td>
          <td>${u.api_key_cipher ? '<span class="badge badge-success">已配置</span>' : '<span class="badge badge-warning">未配置</span>'}</td>
          <td>${formatDate(u.created_at)}</td>
          <td>
            <div class="btn-group">
              ${u.role !== 'admin' ? `<button class="btn btn-sm ${u.banned ? 'btn-success' : 'btn-warning'}" onclick="App.toggleBan(${u.id},${!u.banned})">${u.banned ? '解封' : '封禁'}</button>` : ''}
              ${u.role !== 'admin' ? `<button class="btn btn-sm btn-danger" onclick="App.deleteUser(${u.id})">删除</button>` : ''}
            </div>
          </td>
        </tr>
      `).join('');

      if (!rows) rows = '<tr><td colspan="7" class="empty">暂无用户</td></tr>';

      $('#users-table').innerHTML = `
        <div class="table-wrap">
          <table>
            <thead><tr><th>ID</th><th>用户名</th><th>角色</th><th>状态</th><th>API Key</th><th>注册时间</th><th>操作</th></tr></thead>
            <tbody>${rows}</tbody>
          </table>
          <div class="pagination">
            <span>共 ${total} 条</span>
            <div class="btn-group">
              <button class="btn btn-sm btn-outline" ${usersPage <= 1 ? 'disabled' : ''} onclick="App.usersPageTo(${usersPage - 1})">上一页</button>
              <span style="padding:0 8px;line-height:30px">${usersPage} / ${pages || 1}</span>
              <button class="btn btn-sm btn-outline" ${usersPage >= pages ? 'disabled' : ''} onclick="App.usersPageTo(${usersPage + 1})">下一页</button>
            </div>
          </div>
        </div>`;
    } catch (e) {
      $('#users-table').innerHTML = `<div class="empty">加载失败：${esc(e.message)}</div>`;
    }
  }

  // Invites
  let invitesPage = 1;
  async function renderInvites(el) {
    el.innerHTML = `
      <div class="page-header">
        <h2>邀请码管理</h2>
        <button class="btn btn-primary" onclick="App.showGenInvites()">生成邀请码</button>
      </div>
      <div id="invites-table">加载中...</div>`;
    await loadInvites();
  }

  async function loadInvites() {
    try {
      const d = await api.get(`/admin/invites?page=${invitesPage}&page_size=20`);
      const invites = d.invites || [];
      const total = d.total || 0;
      const pages = Math.ceil(total / 20);

      let rows = invites.map(inv => {
        const used = inv.used_by != null;
        const expired = inv.expires_at && new Date(inv.expires_at) < new Date();
        let status = '<span class="badge badge-success">可用</span>';
        if (used) status = '<span class="badge badge-warning">已使用</span>';
        else if (expired) status = '<span class="badge badge-danger">已过期</span>';
        return `
          <tr>
            <td><code>${esc(inv.code)}</code></td>
            <td>${status}</td>
            <td>${inv.used_by || '-'}</td>
            <td>${inv.expires_at ? formatDate(inv.expires_at) : '永不'}</td>
            <td>${formatDate(inv.created_at)}</td>
            <td>
              <div class="btn-group">
                ${!used ? `<button class="btn btn-sm btn-outline" onclick="App.copyText('${esc(inv.code)}')">复制</button>` : ''}
                <button class="btn btn-sm btn-danger" onclick="App.deleteInvite('${esc(inv.code)}')">删除</button>
              </div>
            </td>
          </tr>`;
      }).join('');

      if (!rows) rows = '<tr><td colspan="6" class="empty">暂无邀请码</td></tr>';

      $('#invites-table').innerHTML = `
        <div class="table-wrap">
          <table>
            <thead><tr><th>邀请码</th><th>状态</th><th>使用者</th><th>过期时间</th><th>创建时间</th><th>操作</th></tr></thead>
            <tbody>${rows}</tbody>
          </table>
          <div class="pagination">
            <span>共 ${total} 条</span>
            <div class="btn-group">
              <button class="btn btn-sm btn-outline" ${invitesPage <= 1 ? 'disabled' : ''} onclick="App.invitesPageTo(${invitesPage - 1})">上一页</button>
              <span style="padding:0 8px;line-height:30px">${invitesPage} / ${pages || 1}</span>
              <button class="btn btn-sm btn-outline" ${invitesPage >= pages ? 'disabled' : ''} onclick="App.invitesPageTo(${invitesPage + 1})">下一页</button>
            </div>
          </div>
        </div>`;
    } catch (e) {
      $('#invites-table').innerHTML = `<div class="empty">加载失败：${esc(e.message)}</div>`;
    }
  }

  // Global Sources
  async function renderSources(el) {
    el.innerHTML = `
      <div class="page-header">
        <h2>全局采集源</h2>
        <div class="btn-group">
          <button class="btn btn-outline" onclick="App.showSortSources()">排序</button>
          <button class="btn btn-primary" onclick="App.showAddSource()">添加源</button>
        </div>
      </div>
      <div id="sources-table">加载中...</div>`;
    await loadSources();
  }

  async function loadSources() {
    try {
      const sources = await api.get('/admin/sources') || [];
      window._globalSources = sources;

      let rows = sources.map(s => `
        <tr>
          <td><code>${esc(s.key)}</code></td>
          <td>${esc(s.name)}</td>
          <td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${esc(s.api_url)}">${esc(s.api_url)}</td>
          <td>${s.detail_url ? esc(s.detail_url) : '-'}</td>
          <td>${s.disabled ? '<span class="badge badge-danger">禁用</span>' : '<span class="badge badge-success">启用</span>'}</td>
          <td>${s.sort_order}</td>
          <td>
            <div class="btn-group">
              <button class="btn btn-sm btn-outline" onclick="App.showEditSource('${esc(s.key)}')">编辑</button>
              <button class="btn btn-sm btn-danger" onclick="App.deleteSource('${esc(s.key)}')">删除</button>
            </div>
          </td>
        </tr>`).join('');

      if (!rows) rows = '<tr><td colspan="7" class="empty">暂无采集源</td></tr>';

      $('#sources-table').innerHTML = `
        <div class="table-wrap">
          <table>
            <thead><tr><th>Key</th><th>名称</th><th>API URL</th><th>Detail URL</th><th>状态</th><th>排序</th><th>操作</th></tr></thead>
            <tbody>${rows}</tbody>
          </table>
        </div>`;
    } catch (e) {
      $('#sources-table').innerHTML = `<div class="empty">加载失败：${esc(e.message)}</div>`;
    }
  }

  // Account
  async function renderAccount(el) {
    el.innerHTML = `
      <div class="page-header"><h2>我的账号</h2></div>
      <div class="card">
        <h3>账号信息</h3>
        <p><strong>用户名：</strong>${esc(state.user.username)}</p>
        <p style="margin-top:8px"><strong>角色：</strong><span class="badge badge-${state.user.role}">${state.user.role}</span></p>
      </div>
      <div class="card">
        <h3>API Key</h3>
        <p style="margin-bottom:16px;color:var(--text-light)">API Key 用于外部接口调用认证。生成后请妥善保管，仅显示一次。</p>
        <div id="apikey-area">
          <div class="btn-group">
            <button class="btn btn-primary" onclick="App.generateApiKey()">生成 API Key</button>
            <button class="btn btn-danger" onclick="App.revokeApiKey()">撤销 API Key</button>
          </div>
        </div>
      </div>`;
  }

  // API Docs
  function renderDocs(el) {
    const baseURL = location.origin;
    el.innerHTML = `
      <div class="page-header"><h2>API 对接文档</h2></div>

      <div class="doc-toc">
        <h4>目录</h4>
        <ul>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-overview')">概述</a></li>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-auth')">认证方式</a></li>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-search')">搜索接口</a></li>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-source')">源管理</a></li>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-account')">账号接口</a></li>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-admin')">管理接口</a></li>
          <li><a href="#/docs" onclick="App.scrollToDoc('sec-errors')">错误码</a></li>
        </ul>
      </div>

      <!-- 概述 -->
      <div class="doc-section" id="sec-overview">
        <h3>概述</h3>
        <div class="card">
          <p style="margin-bottom:12px">MoonTV Server 是一个多源聚合影视搜索 API，支持多租户隔离。</p>
          <p class="doc-label">Base URL</p>
          <div class="doc-code">${esc(baseURL)}/api</div>
          <p class="doc-label">统一响应格式</p>
          <div class="doc-code">{
  "code": 0,        // 0 = 成功，非 0 = 失败
  "data": { ... },  // 业务数据
  "message": "ok"   // 描述信息
}</div>
          <p class="doc-label">认证方式</p>
          <table class="doc-params">
            <tr><td><span class="doc-auth-badge apikey">API Key</span></td><td>外部接口调用，通过 <code>X-API-Key</code> 请求头或 <code>?apikey=</code> 查询参数传递</td></tr>
            <tr><td><span class="doc-auth-badge jwt">JWT</span></td><td>管理面板操作，通过 <code>Authorization: Bearer &lt;token&gt;</code> 请求头传递</td></tr>
            <tr><td><span class="doc-auth-badge admin">Admin</span></td><td>需要 JWT 认证 + 管理员角色</td></tr>
            <tr><td><span class="doc-auth-badge public">Public</span></td><td>无需认证</td></tr>
          </table>
        </div>
      </div>

      <!-- 认证 -->
      <div class="doc-section" id="sec-auth">
        <h3>认证方式</h3>

        ${docEndpoint('POST', '/api/auth/login', '用户登录', 'public', {
          body: [
            ['username', 'string', '是', '用户名'],
            ['password', 'string', '是', '密码'],
          ],
          response: `{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": 1,
    "username": "nineone",
    "role": "admin"
  },
  "message": "ok"
}`,
          curl: `curl -X POST ${esc(baseURL)}/api/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{"username":"your_name","password":"your_pass"}'`
        })}

        ${docEndpoint('POST', '/api/auth/register', '用户注册（需要邀请码）', 'public', {
          body: [
            ['username', 'string', '是', '用户名（3-32 位）'],
            ['password', 'string', '是', '密码（≥6 位）'],
            ['invite_code', 'string', '是', '邀请码'],
          ],
          response: `{
  "code": 0,
  "data": {
    "user_id": 2,
    "username": "testuser"
  },
  "message": "ok"
}`,
          note: '注册成功后自动继承全局采集源配置。'
        })}
      </div>

      <!-- 搜索接口 -->
      <div class="doc-section" id="sec-search">
        <h3>搜索接口</h3>

        ${docEndpoint('GET', '/api/search', '多源聚合搜索', 'apikey', {
          query: [
            ['q', 'string', '是', '搜索关键词'],
            ['page', 'int', '否', '页码，默认 1'],
            ['yellow_filter', 'bool', '否', '黄色内容过滤，默认 true'],
          ],
          response: `{
  "code": 0,
  "data": [
    {
      "source": "feifan",
      "source_name": "非凡",
      "page_count": 5,
      "results": [
        {
          "id": "74250",
          "title": "情感价值",
          "poster": "https://example.com/cover.jpg",
          "episodes": [
            "https://example.com/ep1.m3u8",
            "https://example.com/ep2.m3u8"
          ],
          "episodes_titles": ["第1集", "第2集"],
          "source": "feifan",
          "source_name": "非凡",
          "class": "剧情,喜剧",
          "year": "2025",
          "desc": "剧情简介...",
          "type_name": "国产剧"
        }
      ]
    }
  ],
  "message": "ok"
}`,
          curl: `curl "${esc(baseURL)}/api/search?q=情感价值&page=1" \\
  -H "X-API-Key: mtv_your_api_key"`,
          note: '并发查询用户已启用的所有采集源，结果按源分组返回。'
        })}

        ${docEndpoint('GET', '/api/search/sse', 'SSE 流式搜索', 'apikey', {
          query: [
            ['q', 'string', '是', '搜索关键词'],
            ['yellow_filter', 'bool', '否', '黄色内容过滤，默认 true'],
          ],
          response: `event: message
data: {"source":"feifan","source_name":"非凡","page_count":5,"results":[...]}

event: message
data: {"source":"jisu","source_name":"极速","page_count":3,"results":[...]}

event: done
data: {}`,
          curl: `curl -N "${esc(baseURL)}/api/search/sse?q=情感价值" \\
  -H "X-API-Key: mtv_your_api_key"`,
          note: '使用 Server-Sent Events 协议，每个源返回结果时立即推送一条 data 事件，全部完成后发送 done 事件。适合前端实时展示搜索进度。'
        })}

        ${docEndpoint('GET', '/api/detail', '获取视频详情', 'apikey', {
          query: [
            ['source', 'string', '是', '源标识（如 feifan）'],
            ['id', 'string', '是', '视频 ID（搜索结果中的 id 字段）'],
          ],
          response: `{
  "code": 0,
  "data": {
    "id": "74250",
    "title": "情感价值",
    "poster": "https://example.com/cover.jpg",
    "episodes": ["https://example.com/ep1.m3u8"],
    "episodes_titles": ["HD中字"],
    "source": "feifan",
    "source_name": "非凡",
    "class": "剧情,喜剧",
    "year": "2025",
    "desc": "完整剧情简介...",
    "type_name": "喜剧片"
  },
  "message": "ok"
}`,
          curl: `curl "${esc(baseURL)}/api/detail?source=feifan&id=74250" \\
  -H "X-API-Key: mtv_your_api_key"`
        })}

        ${docEndpoint('GET', '/api/suggest', '搜索建议', 'apikey', {
          query: [
            ['q', 'string', '是', '搜索关键词'],
          ],
          response: `{
  "code": 0,
  "data": ["情感价值", "情感的禁区", "情感导师"],
  "message": "ok"
}`,
          curl: `curl "${esc(baseURL)}/api/suggest?q=情感" \\
  -H "X-API-Key: mtv_your_api_key"`,
          note: '从前 3 个源中检索，返回最多 10 条不重复的影片标题。'
        })}
      </div>

      <!-- 源管理 -->
      <div class="doc-section" id="sec-source">
        <h3>源管理（租户级）</h3>
        <div class="doc-note">以下接口管理当前用户的个人采集源配置，不影响其他用户。</div>

        ${docEndpoint('GET', '/api/sources', '获取我的源列表', 'apikey', {
          response: `{
  "code": 0,
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "key": "feifan",
      "name": "非凡",
      "api_url": "https://api.ffzyapi.com/api.php/provide/vod/",
      "detail_url": "",
      "disabled": false,
      "sort_order": 0,
      "created_at": "2026-04-21T12:00:00+08:00"
    }
  ],
  "message": "ok"
}`,
          curl: `curl "${esc(baseURL)}/api/sources" \\
  -H "X-API-Key: mtv_your_api_key"`
        })}

        ${docEndpoint('POST', '/api/sources', '添加采集源', 'apikey', {
          body: [
            ['key', 'string', '是', '源唯一标识'],
            ['name', 'string', '是', '显示名称'],
            ['api_url', 'string', '是', '采集 API 地址'],
            ['detail_url', 'string', '否', '详情页地址（用于 HTML 解析模式）'],
          ],
          curl: `curl -X POST "${esc(baseURL)}/api/sources" \\
  -H "X-API-Key: mtv_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"key":"mysrc","name":"我的源","api_url":"https://example.com/api.php/provide/vod/"}'`
        })}

        ${docEndpoint('PUT', '/api/sources/:key', '更新采集源', 'apikey', {
          params: [['key', 'string', '是', '源标识']],
          body: [
            ['name', 'string', '否', '显示名称'],
            ['api_url', 'string', '否', 'API 地址'],
            ['detail_url', 'string', '否', '详情页地址'],
            ['disabled', 'bool', '否', '是否禁用'],
          ],
          curl: `curl -X PUT "${esc(baseURL)}/api/sources/mysrc" \\
  -H "X-API-Key: mtv_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"disabled":true}'`
        })}

        ${docEndpoint('DELETE', '/api/sources/:key', '删除采集源', 'apikey', {
          params: [['key', 'string', '是', '源标识']],
          curl: `curl -X DELETE "${esc(baseURL)}/api/sources/mysrc" \\
  -H "X-API-Key: mtv_your_api_key"`
        })}

        ${docEndpoint('PUT', '/api/sources/sort', '调整源排序', 'apikey', {
          body: [['keys', 'string[]', '是', '按顺序排列的源 key 数组']],
          curl: `curl -X PUT "${esc(baseURL)}/api/sources/sort" \\
  -H "X-API-Key: mtv_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"keys":["feifan","jisu","hongniu"]}'`
        })}
      </div>

      <!-- 账号接口 -->
      <div class="doc-section" id="sec-account">
        <h3>账号接口</h3>

        ${docEndpoint('POST', '/api/user/apikey', '生成 API Key', 'jwt', {
          response: `{
  "code": 0,
  "data": {
    "api_key": "mtv_2MeDeis1RM3ATrVopseeuwT2IO0KFt8XdK2NpungRBW0D5AISJfidcl81325cGmoPtnL2rZNYzej"
  },
  "message": "ok"
}`,
          curl: `curl -X POST "${esc(baseURL)}/api/user/apikey" \\
  -H "Authorization: Bearer your_jwt_token"`,
          note: 'API Key 仅在生成时返回一次，请妥善保存。如已存在 Key 需先撤销再重新生成。',
          noteWarn: true
        })}

        ${docEndpoint('DELETE', '/api/user/apikey', '撤销 API Key', 'jwt', {
          response: `{
  "code": 0,
  "data": { "message": "api key revoked" },
  "message": "ok"
}`,
          curl: `curl -X DELETE "${esc(baseURL)}/api/user/apikey" \\
  -H "Authorization: Bearer your_jwt_token"`,
          note: '撤销后使用该 Key 的所有请求将立即失效。'
        })}
      </div>

      <!-- 管理接口 -->
      <div class="doc-section" id="sec-admin">
        <h3>管理接口</h3>
        <div class="doc-note">以下接口需要管理员权限（JWT + admin 角色）。</div>

        ${docEndpoint('GET', '/api/admin/stats', '系统统计', 'admin', {
          response: `{
  "code": 0,
  "data": {
    "total_users": 10,
    "active_users": 8,
    "total_sources": 120,
    "api_calls_today": 1500,
    "api_calls_7days": 8700
  },
  "message": "ok"
}`
        })}

        ${docEndpoint('GET', '/api/admin/users', '用户列表', 'admin', {
          query: [
            ['page', 'int', '否', '页码，默认 1'],
            ['page_size', 'int', '否', '每页条数，默认 20，最大 100'],
          ],
          response: `{
  "code": 0,
  "data": {
    "users": [{ "id":1, "username":"nineone", "role":"admin", "banned":false, ... }],
    "total": 10,
    "page": 1
  },
  "message": "ok"
}`
        })}

        ${docEndpoint('PUT', '/api/admin/users/:id/ban', '封禁/解封用户', 'admin', {
          params: [['id', 'int', '是', '用户 ID']],
          body: [['banned', 'bool', '是', 'true=封禁, false=解封']],
        })}

        ${docEndpoint('DELETE', '/api/admin/users/:id', '删除用户', 'admin', {
          params: [['id', 'int', '是', '用户 ID']],
          note: '删除用户会同时删除其所有采集源配置。不可删除自己。'
        })}

        ${docEndpoint('POST', '/api/admin/invites', '生成邀请码', 'admin', {
          body: [
            ['count', 'int', '是', '生成数量（1-50）'],
            ['expire_days', 'int', '否', '有效天数（0=永不过期）'],
          ],
          response: `{
  "code": 0,
  "data": [
    { "id":1, "code":"58024529225cd8f0", "expires_at":"2026-04-28T...", ... }
  ],
  "message": "ok"
}`
        })}

        ${docEndpoint('GET', '/api/admin/invites', '邀请码列表', 'admin', {
          query: [
            ['page', 'int', '否', '页码'],
            ['page_size', 'int', '否', '每页条数'],
          ]
        })}

        ${docEndpoint('DELETE', '/api/admin/invites/:code', '删除邀请码', 'admin', {
          params: [['code', 'string', '是', '邀请码']],
        })}

        ${docEndpoint('GET', '/api/admin/sources', '全局源列表', 'admin', {})}
        ${docEndpoint('POST', '/api/admin/sources', '添加全局源', 'admin', {
          body: [
            ['key', 'string', '是', '源唯一标识'],
            ['name', 'string', '是', '显示名称'],
            ['api_url', 'string', '是', '采集 API 地址'],
            ['detail_url', 'string', '否', '详情页地址'],
          ],
          note: '新用户注册时会自动继承当时的所有全局源。'
        })}
        ${docEndpoint('PUT', '/api/admin/sources/:key', '更新全局源', 'admin', {
          params: [['key', 'string', '是', '源标识']],
          body: [
            ['name', 'string', '否', ''],
            ['api_url', 'string', '否', ''],
            ['detail_url', 'string', '否', ''],
            ['disabled', 'bool', '否', ''],
          ]
        })}
        ${docEndpoint('DELETE', '/api/admin/sources/:key', '删除全局源', 'admin', {
          params: [['key', 'string', '是', '源标识']]
        })}
        ${docEndpoint('PUT', '/api/admin/sources/sort', '全局源排序', 'admin', {
          body: [['keys', 'string[]', '是', '按顺序排列的源 key 数组']]
        })}
      </div>

      <!-- 错误码 -->
      <div class="doc-section" id="sec-errors">
        <h3>错误码</h3>
        <div class="card" style="padding:0;overflow:hidden">
          <table class="doc-params">
            <thead><tr><th>错误码</th><th>HTTP 状态码</th><th>说明</th></tr></thead>
            <tbody>
              <tr><td><code>0</code></td><td>200</td><td>成功</td></tr>
              <tr><td><code>40001</code></td><td>401</td><td>未认证 / Token 缺失</td></tr>
              <tr><td><code>40002</code></td><td>401</td><td>Token 无效或过期</td></tr>
              <tr><td><code>40003</code></td><td>403</td><td>账号已被封禁</td></tr>
              <tr><td><code>40004</code></td><td>401</td><td>缺少 API Key</td></tr>
              <tr><td><code>40005</code></td><td>401</td><td>API Key 无效或已撤销</td></tr>
              <tr><td><code>40006</code></td><td>403</td><td>无权限（需管理员）</td></tr>
              <tr><td><code>40101</code></td><td>400</td><td>请求参数错误</td></tr>
              <tr><td><code>40102</code></td><td>400</td><td>参数校验失败</td></tr>
              <tr><td><code>40103</code></td><td>409</td><td>数据重复</td></tr>
              <tr><td><code>40104</code></td><td>400</td><td>邀请码无效/已使用/已过期</td></tr>
              <tr><td><code>40401</code></td><td>404</td><td>资源不存在</td></tr>
              <tr><td><code>50001</code></td><td>500</td><td>服务器内部错误</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    `;

    // bind toggle
    el.querySelectorAll('.doc-endpoint-header').forEach(header => {
      header.addEventListener('click', () => {
        header.parentElement.classList.toggle('open');
      });
    });
  }

  function docEndpoint(method, path, desc, auth, opts = {}) {
    const m = method.toLowerCase();
    const authBadge = {
      public: '<span class="doc-auth-badge public">Public</span>',
      apikey: '<span class="doc-auth-badge apikey">API Key</span>',
      jwt: '<span class="doc-auth-badge jwt">JWT</span>',
      admin: '<span class="doc-auth-badge admin">Admin</span>',
    }[auth] || '';

    let body = '';

    if (opts.params && opts.params.length) {
      body += '<p class="doc-label">路径参数</p>';
      body += paramTable(opts.params);
    }
    if (opts.query && opts.query.length) {
      body += '<p class="doc-label">查询参数</p>';
      body += paramTable(opts.query);
    }
    if (opts.body && opts.body.length) {
      body += '<p class="doc-label">请求体 (JSON)</p>';
      body += paramTable(opts.body);
    }
    if (opts.response) {
      body += '<p class="doc-label">响应示例</p>';
      body += `<div class="doc-code">${esc(opts.response)}</div>`;
    }
    if (opts.curl) {
      body += '<p class="doc-label">cURL 示例</p>';
      body += `<div class="doc-code">${esc(opts.curl)}<button class="copy-btn" onclick="event.stopPropagation();App.copyText(this.previousSibling.textContent.trim())">复制</button></div>`;
    }
    if (opts.note) {
      body += `<div class="doc-note${opts.noteWarn ? ' warn' : ''}">${esc(opts.note)}</div>`;
    }

    return `
      <div class="doc-endpoint">
        <div class="doc-endpoint-header">
          <span class="doc-method ${m}">${method}</span>
          <span class="doc-path">${esc(path)}</span>
          ${authBadge}
          <span class="doc-desc">${esc(desc)}</span>
          <span class="doc-chevron">&#9654;</span>
        </div>
        <div class="doc-endpoint-body">${body}</div>
      </div>`;
  }

  function paramTable(rows) {
    let html = '<table class="doc-params"><thead><tr><th>参数</th><th>类型</th><th>必填</th><th>说明</th></tr></thead><tbody>';
    for (const [name, type, req, desc] of rows) {
      const reqBadge = req === '是' ? '<span class="required">必填</span>' : '<span class="optional">可选</span>';
      html += `<tr><td><code>${esc(name)}</code></td><td>${esc(type)}</td><td>${reqBadge}</td><td>${esc(desc)}</td></tr>`;
    }
    html += '</tbody></table>';
    return html;
  }

  // ── Actions (exposed globally) ──
  window.App = {
    // Users
    async toggleBan(id, banned) {
      try {
        await api.put(`/admin/users/${id}/ban`, { banned });
        toast(banned ? '已封禁' : '已解封', 'success');
        await loadUsers();
      } catch (e) { toast(e.message, 'error'); }
    },
    async deleteUser(id) {
      if (!confirm('确认删除该用户？此操作不可恢复。')) return;
      try {
        await api.del(`/admin/users/${id}`);
        toast('已删除', 'success');
        await loadUsers();
      } catch (e) { toast(e.message, 'error'); }
    },
    usersPageTo(p) { usersPage = p; loadUsers(); },

    // Invites
    showGenInvites() {
      openModal('生成邀请码', `
        <form id="gen-invite-form">
          <div class="form-group">
            <label>数量</label>
            <input type="number" id="invite-count" value="5" min="1" max="50" required>
          </div>
          <div class="form-group">
            <label>有效天数（0 = 永不过期）</label>
            <input type="number" id="invite-days" value="7" min="0">
          </div>
          <button type="submit" class="btn btn-primary btn-block">生成</button>
        </form>
      `);
      $('#gen-invite-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
          const codes = await api.post('/admin/invites', {
            count: parseInt($('#invite-count').value),
            expire_days: parseInt($('#invite-days').value),
          });
          closeModal();
          toast(`已生成 ${codes.length} 个邀请码`, 'success');
          await loadInvites();
        } catch (err) { toast(err.message, 'error'); }
      };
    },
    async deleteInvite(code) {
      if (!confirm('确认删除该邀请码？')) return;
      try {
        await api.del(`/admin/invites/${code}`);
        toast('已删除', 'success');
        await loadInvites();
      } catch (e) { toast(e.message, 'error'); }
    },
    invitesPageTo(p) { invitesPage = p; loadInvites(); },

    // Sources
    showAddSource() {
      openModal('添加采集源', `
        <form id="add-source-form">
          <div class="form-group"><label>Key（唯一标识）</label><input type="text" id="src-key" required></div>
          <div class="form-group"><label>名称</label><input type="text" id="src-name" required></div>
          <div class="form-group"><label>API URL</label><input type="url" id="src-apiurl" required></div>
          <div class="form-group"><label>Detail URL（可选）</label><input type="url" id="src-detailurl"></div>
          <button type="submit" class="btn btn-primary btn-block">添加</button>
        </form>
      `);
      $('#add-source-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
          await api.post('/admin/sources', {
            key: $('#src-key').value.trim(),
            name: $('#src-name').value.trim(),
            api_url: $('#src-apiurl').value.trim(),
            detail_url: $('#src-detailurl').value.trim() || undefined,
          });
          closeModal();
          toast('添加成功', 'success');
          await loadSources();
        } catch (err) { toast(err.message, 'error'); }
      };
    },
    showEditSource(key) {
      const src = (window._globalSources || []).find(s => s.key === key);
      if (!src) return;
      openModal('编辑采集源', `
        <form id="edit-source-form">
          <div class="form-group"><label>Key</label><input type="text" value="${esc(src.key)}" disabled></div>
          <div class="form-group"><label>名称</label><input type="text" id="edit-name" value="${esc(src.name)}" required></div>
          <div class="form-group"><label>API URL</label><input type="url" id="edit-apiurl" value="${esc(src.api_url)}" required></div>
          <div class="form-group"><label>Detail URL</label><input type="url" id="edit-detailurl" value="${esc(src.detail_url || '')}"></div>
          <div class="form-group">
            <label><input type="checkbox" id="edit-disabled" ${src.disabled ? 'checked' : ''}> 禁用</label>
          </div>
          <button type="submit" class="btn btn-primary btn-block">保存</button>
        </form>
      `);
      $('#edit-source-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
          await api.put(`/admin/sources/${key}`, {
            name: $('#edit-name').value.trim(),
            api_url: $('#edit-apiurl').value.trim(),
            detail_url: $('#edit-detailurl').value.trim() || '',
            disabled: $('#edit-disabled').checked,
          });
          closeModal();
          toast('保存成功', 'success');
          await loadSources();
        } catch (err) { toast(err.message, 'error'); }
      };
    },
    async deleteSource(key) {
      if (!confirm(`确认删除采集源 "${key}"？`)) return;
      try {
        await api.del(`/admin/sources/${key}`);
        toast('已删除', 'success');
        await loadSources();
      } catch (e) { toast(e.message, 'error'); }
    },
    showSortSources() {
      const sources = window._globalSources || [];
      if (!sources.length) { toast('暂无采集源', 'info'); return; }
      let items = sources.map((s, i) => `
        <li class="sort-item" data-key="${esc(s.key)}">
          <span class="sort-name">${esc(s.name)} <code style="font-size:12px;color:var(--text-light)">${esc(s.key)}</code></span>
          <div class="sort-btns">
            <button onclick="App.sortMove(${i},-1)">&uarr;</button>
            <button onclick="App.sortMove(${i},1)">&darr;</button>
          </div>
        </li>
      `).join('');
      openModal('排序采集源', `
        <ul class="sort-list" id="sort-list">${items}</ul>
        <button class="btn btn-primary btn-block" style="margin-top:16px" onclick="App.saveSortSources()">保存排序</button>
      `);
    },
    sortMove(idx, dir) {
      const list = $('#sort-list');
      const items = Array.from(list.children);
      const newIdx = idx + dir;
      if (newIdx < 0 || newIdx >= items.length) return;
      if (dir === -1) list.insertBefore(items[idx], items[newIdx]);
      else list.insertBefore(items[newIdx], items[idx]);
      // Re-bind indices
      Array.from(list.children).forEach((li, i) => {
        const btns = li.querySelectorAll('.sort-btns button');
        btns[0].setAttribute('onclick', `App.sortMove(${i},-1)`);
        btns[1].setAttribute('onclick', `App.sortMove(${i},1)`);
      });
    },
    async saveSortSources() {
      const keys = Array.from($('#sort-list').children).map(li => li.dataset.key);
      try {
        await api.put('/admin/sources/sort', { keys });
        closeModal();
        toast('排序已保存', 'success');
        await loadSources();
      } catch (e) { toast(e.message, 'error'); }
    },

    // API Key
    async generateApiKey() {
      try {
        const d = await api.post('/user/apikey');
        $('#apikey-area').innerHTML = `
          <div class="apikey-display">
            ${esc(d.api_key)}
            <button class="copy-btn" onclick="App.copyText('${esc(d.api_key)}')">复制</button>
          </div>
          <p style="color:var(--warning);font-size:13px;margin-top:8px">请立即复制保存，此 Key 仅显示一次。</p>
        `;
        toast('API Key 已生成', 'success');
      } catch (e) { toast(e.message, 'error'); }
    },
    async revokeApiKey() {
      if (!confirm('确认撤销 API Key？撤销后当前 Key 将立即失效。')) return;
      try {
        await api.del('/user/apikey');
        toast('API Key 已撤销', 'success');
        renderAccount($('#content'));
      } catch (e) { toast(e.message, 'error'); }
    },

    // Util
    copyText(text) {
      navigator.clipboard.writeText(text).then(() => toast('已复制', 'success')).catch(() => toast('复制失败', 'error'));
    },
    scrollToDoc(id) {
      setTimeout(() => {
        const el = document.getElementById(id);
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 50);
    },
  };

  // ── Helpers ──
  function esc(s) {
    if (s == null) return '';
    const d = document.createElement('div');
    d.textContent = String(s);
    return d.innerHTML;
  }

  function formatDate(s) {
    if (!s) return '-';
    const d = new Date(s);
    return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  }

  // ── Init ──
  document.addEventListener('DOMContentLoaded', () => {
    // Login form
    $('#login-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const username = $('#login-username').value.trim();
      const password = $('#login-password').value;
      try {
        const data = await api.post('/auth/login', { username, password });
        login(data.token, { id: data.user_id, username: data.username, role: data.role });
        toast('登录成功', 'success');
      } catch (err) {
        toast(err.message || '登录失败', 'error');
      }
    });

    // Logout
    $('#logout-btn').addEventListener('click', logout);

    // Modal close
    $('#modal-close').addEventListener('click', closeModal);
    $('#modal-overlay').addEventListener('click', (e) => {
      if (e.target === $('#modal-overlay')) closeModal();
    });

    // Hash routing
    window.addEventListener('hashchange', () => {
      if (state.token) route();
    });

    // Initial render
    if (state.token) {
      showMain();
    } else {
      showLogin();
    }
  });
})();
