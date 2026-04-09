document.addEventListener('DOMContentLoaded', () => {
    let currentKeys = [];
    let currentKey = '';
    let currentValue = null;
    let currentPage = 0;
    const limit = 50;
    let totalKeys = 0;

    const keyList = document.getElementById('key-list');
    const valueDisplay = document.getElementById('value-display');
    const currentKeyEl = document.getElementById('current-key');
    const valueSizeEl = document.getElementById('value-size');
    const dbPathEl = document.getElementById('db-path-display');
    const keyCountEl = document.getElementById('key-count-display');
    const searchInput = document.getElementById('search-input');
    const searchMode = document.getElementById('search-mode');
    const refreshBtn = document.getElementById('refresh-btn');
    const prevBtn = document.getElementById('prev-btn');
    const nextBtn = document.getElementById('next-btn');
    const pageInfo = document.getElementById('page-info');
    const noSelection = document.getElementById('no-selection');
    const valueViewer = document.getElementById('value-viewer');
    const tabBtns = document.querySelectorAll('.tab-btn');
    const sidebar = document.getElementById('sidebar');
    const resizeHandle = document.getElementById('resize-handle');

    let activeTab = 'raw';
    let substringWarningShown = false;

    // Resolve API base relative to where the page is served from.
    // This ensures the UI works when mounted at a sub-path (e.g. /pebble-ui/).
    const scriptEl = document.querySelector('script[src$="app.js"]');
    const basePath = new URL('.', scriptEl.src).pathname.replace(/\/$/, '');

    async function fetchStats() {
        try {
            const resp = await fetch(basePath + '/api/stats');
            const data = await resp.json();
            if (data.db_path) {
                dbPathEl.textContent = `DB: ${data.db_path}`;
            } else {
                dbPathEl.style.display = 'none';
            }
            keyCountEl.textContent = `Keys: ${data.total_keys}`;
            totalKeys = data.total_keys;
            updatePagination();
        } catch (e) {
            console.error('Failed to fetch stats', e);
        }
    }

    async function fetchKeys(query = '', offset = 0) {
        try {
            const mode = searchMode.value;
            const resp = await fetch(`${basePath}/api/keys?q=${encodeURIComponent(query)}&mode=${mode}&offset=${offset}&limit=${limit}`);
            const data = await resp.json();
            currentKeys = data.keys || [];
            totalKeys = data.total;
            renderKeyList();
            updatePagination();
        } catch (e) {
            console.error('Failed to fetch keys', e);
        }
    }

    async function fetchValue(key) {
        try {
            const resp = await fetch(`${basePath}/api/key/${encodeURIComponent(key)}`);
            if (!resp.ok) throw new Error('Not found');
            const data = await resp.json();
            currentKey = data.key;
            currentValue = data;
            
            noSelection.classList.add('hidden');
            valueViewer.classList.remove('hidden');
            renderValue();
        } catch (e) {
            console.error('Failed to fetch value', e);
        }
    }

    function renderKeyList() {
        keyList.innerHTML = '';
        currentKeys.forEach(key => {
            const li = document.createElement('li');
            li.textContent = key;
            if (key === currentKey) li.classList.add('selected');
            li.onclick = () => {
                document.querySelectorAll('#key-list li').forEach(el => el.classList.remove('selected'));
                li.classList.add('selected');
                fetchValue(key);
            };
            keyList.appendChild(li);
        });
    }

    function renderValue() {
        currentKeyEl.textContent = currentKey;
        valueSizeEl.textContent = `${currentValue.size} bytes`;
        
        if (activeTab === 'raw') {
            valueDisplay.textContent = currentValue.value;
        } else if (activeTab === 'hex') {
            valueDisplay.textContent = currentValue.value_hex.match(/.{1,2}/g).join(' ');
        } else if (activeTab === 'json') {
            try {
                const parsed = JSON.parse(currentValue.value);
                valueDisplay.textContent = JSON.stringify(parsed, null, 2);
            } catch (e) {
                valueDisplay.textContent = 'Invalid JSON';
            }
        }
    }

    function updatePagination() {
        const totalPages = Math.ceil(totalKeys / limit);
        pageInfo.textContent = `Page ${currentPage + 1} of ${totalPages || 1}`;
        prevBtn.disabled = currentPage === 0;
        nextBtn.disabled = (currentPage + 1) >= totalPages;
    }

    searchMode.addEventListener('change', () => {
        if (searchMode.value === 'substring') {
            searchMode.classList.add('warning');
            if (!substringWarningShown) {
                substringWarningShown = true;
                alert('⚠️ Contains search scans ALL keys in memory.\nThis may be slow for large databases.');
            }
        } else {
            searchMode.classList.remove('warning');
        }
    });

    refreshBtn.onclick = () => {
        currentPage = 0;
        fetchStats();
        fetchKeys(searchInput.value, 0);
    };

    prevBtn.onclick = () => {
        if (currentPage > 0) {
            currentPage--;
            fetchKeys(searchInput.value, currentPage * limit);
        }
    };

    nextBtn.onclick = () => {
        currentPage++;
        fetchKeys(searchInput.value, currentPage * limit);
    };

    // Sidebar resize
    let isResizing = false;
    resizeHandle.addEventListener('mousedown', (e) => {
        isResizing = true;
        resizeHandle.classList.add('dragging');
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';
    });

    document.addEventListener('mousemove', (e) => {
        if (!isResizing) return;
        const newWidth = e.clientX;
        if (newWidth >= 150 && newWidth <= window.innerWidth * 0.7) {
            sidebar.style.width = newWidth + 'px';
        }
    });

    document.addEventListener('mouseup', () => {
        if (isResizing) {
            isResizing = false;
            resizeHandle.classList.remove('dragging');
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        }
    });

    tabBtns.forEach(btn => {
        btn.onclick = () => {
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            activeTab = btn.dataset.tab;
            if (currentValue) renderValue();
        };
    });

    document.getElementById('copy-btn').onclick = () => {
        navigator.clipboard.writeText(valueDisplay.textContent);
    };

    async function fetchConfig() {
        try {
            const resp = await fetch(basePath + '/api/config');
            const data = await resp.json();
            if (data.substring_search) {
                document.querySelector('#search-mode option[value="substring"]').classList.remove('hidden');
            }
        } catch (e) {
            console.error('Failed to fetch config', e);
        }
    }

    fetchConfig();
    fetchStats();
    fetchKeys();
});
