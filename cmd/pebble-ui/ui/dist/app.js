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

    async function fetchStats() {
        try {
            const resp = await fetch('/api/stats');
            const data = await resp.json();
            dbPathEl.textContent = `DB: ${data.db_path}`;
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
            const resp = await fetch(`/api/keys?q=${encodeURIComponent(query)}&mode=${mode}&offset=${offset}&limit=${limit}`);
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
            const resp = await fetch(`/api/key/${encodeURIComponent(key)}`);
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

    searchInput.addEventListener('input', debounce(() => {
        currentPage = 0;
        fetchKeys(searchInput.value, 0);
    }, 300));

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
        currentPage = 0;
        fetchKeys(searchInput.value, 0);
    });

    refreshBtn.onclick = () => {
        fetchStats();
        fetchKeys(searchInput.value, currentPage * limit);
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

    function debounce(func, wait) {
        let timeout;
        return function(...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }

    fetchStats();
    fetchKeys();
});
