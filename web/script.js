const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => Array.from(document.querySelectorAll(sel));

let token = localStorage.getItem('token') || '';
let currentUser = null;
let charts = {};
let currencies = [];
let accounts = [];
let categories = [];
let exchangeRates = [];
let currentTab = 'overview';

// Ensure API base path is set: dev uses explicit localhost backend (including file://), prod uses relative path
(function(){
    const isFile = location.protocol === 'file:';
    const isLocalhost = ['localhost', '127.0.0.1', '::1'].includes(location.hostname);
    window.API_BASE = (isFile || isLocalhost)
        ? 'http://localhost:8080/api/v1'
        : '/api/v1';
})();

class NotificationSystem {
    static show(message, type = 'info', duration = 5000) {
        const container = $('#notifications');
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        
        const icons = {
            success: 'fas fa-check-circle',
            error: 'fas fa-exclamation-triangle',
            info: 'fas fa-info-circle'
        };
        
        notification.innerHTML = `
            <i class="${icons[type]}"></i>
            <div class="notification-content">
                <div class="notification-title">${this.getTitle(type)}</div>
                <div class="notification-message">${message}</div>
            </div>
            <button class="notification-close" onclick="this.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        `;
        
        container.appendChild(notification);
        
        setTimeout(() => {
            if (notification.parentElement) {
                notification.style.animation = 'slideInRight 0.3s ease reverse';
                setTimeout(() => notification.remove(), 300);
            }
        }, duration);
    }
    
    static getTitle(type) {
        const titles = {
            success: 'Успех',
            error: 'Ошибка',
            info: 'Информация'
        };
        return titles[type];
    }
}
 
// Settings handlers are implemented in the main FormHandlers class below.

// --- Currency Converter ---
class CurrencyConverter {
    static init() {
        this.bindEvents();
        this.populateCurrencySelects();
        this.setDefaultCurrencies();
        this.performAutoConversion();
    }

    static bindEvents() {
        const amountEl = document.getElementById('convert-amount');
        const fromEl = document.getElementById('convert-from');
        const toEl = document.getElementById('convert-to');
        const swapEl = document.getElementById('convert-swap');
        const quickBtn = document.getElementById('quick-convert');
        if (!amountEl || !fromEl || !toEl) return;

        amountEl.addEventListener('input', () => this.performAutoConversion());
        fromEl.addEventListener('change', () => { this.updateCurrencySymbol('from'); this.performAutoConversion(); });
        toEl.addEventListener('change', () => { this.updateCurrencySymbol('to'); this.performAutoConversion(); });

        if (swapEl) swapEl.addEventListener('click', () => this.swapCurrencies());
        if (quickBtn) quickBtn.addEventListener('click', () => this.performConversion());

        this.debouncedConversion = debounce(() => this.performConversion(), 500);
    }

    static populateCurrencySelects() {
        const fromSelect = document.getElementById('convert-from');
        const toSelect = document.getElementById('convert-to');
        if (!fromSelect || !toSelect || !Array.isArray(currencies)) return;

        fromSelect.innerHTML = '<option value="">Выберите валюту</option>';
        toSelect.innerHTML = '<option value="">Выберите валюту</option>';

        currencies.forEach(currency => {
            const optionA = document.createElement('option');
            optionA.value = currency.id;
            optionA.textContent = `${currency.code} - ${currency.name}`;
            optionA.setAttribute('data-symbol', currency.symbol || CurrencyUI.getSymbol(currency.code));
            const optionB = optionA.cloneNode(true);
            fromSelect.appendChild(optionA);
            toSelect.appendChild(optionB);
        });
    }

    static setDefaultCurrencies() {
        const usd = currencies.find(c => c.code === 'USD');
        const rub = currencies.find(c => c.code === 'RUB');
        if (usd) { document.getElementById('convert-from').value = String(usd.id); this.updateCurrencySymbol('from'); }
        if (rub) { document.getElementById('convert-to').value = String(rub.id); this.updateCurrencySymbol('to'); }
    }

    static updateCurrencySymbol(type) {
        const select = document.getElementById(`convert-${type}`);
        const symbolEl = document.getElementById(`${type}-symbol`);
        if (!select || !symbolEl) return;
        const opt = select.options[select.selectedIndex];
        if (opt && opt.value) {
            const cur = currencies.find(c => c.id === parseInt(opt.value));
            const sym = cur ? (cur.symbol || UIManager.getCurrencySymbol(cur.code)) : UIManager.getCurrencySymbol((opt.textContent || '').split(' - ')[0]);
            symbolEl.textContent = sym;
        } else {
            symbolEl.textContent = CurrencyUI.getSymbol('USD');
        }
    }

    static async performAutoConversion() {
        const amount = document.getElementById('convert-amount');
        const fromSel = document.getElementById('convert-from');
        const toSel = document.getElementById('convert-to');
        if (!amount || !fromSel || !toSel) return;
        const val = parseFloat(amount.value);
        if (!val || val <= 0 || !fromSel.value || !toSel.value) {
            this.clearResult();
            return;
        }
        this.setLoadingState(true);
        this.debouncedConversion();
    }

    static async performConversion() {
        const amount = parseFloat(document.getElementById('convert-amount').value);
        const fromCurrencyId = parseInt(document.getElementById('convert-from').value);
        const toCurrencyId = parseInt(document.getElementById('convert-to').value);

        if (!amount || amount <= 0 || !fromCurrencyId || !toCurrencyId) {
            this.clearResult();
            return;
        }

        try {
            this.setLoadingState(true);
            const result = await ApiClient.request('/exchange/convert-simple', {
                method: 'POST',
                body: JSON.stringify({ from_currency_id: fromCurrencyId, to_currency_id: toCurrencyId, amount })
            });
            this.displayConversionResult(result, amount);
        } catch (error) {
            this.displayConversionError(error.message);
        } finally {
            this.setLoadingState(false);
        }
    }

    static displayConversionResult(result, originalAmount) {
        const fromCurrency = currencies.find(c => c.id === result.from_currency_id);
        const toCurrency = currencies.find(c => c.id === result.to_currency_id);
        if (!fromCurrency || !toCurrency) { this.displayConversionError('Ошибка отображения результата'); return; }
        const fromSymbol = CurrencyUI.getSymbol(fromCurrency.code);
        const toSymbol = CurrencyUI.getSymbol(toCurrency.code);
        const fmt = (n, d=2) => new Intl.NumberFormat('ru-RU', { minimumFractionDigits: d, maximumFractionDigits: d }).format(parseFloat(n));

        const resultInput = document.getElementById('convert-result');
        if (resultInput) {
            resultInput.value = fmt(result.converted_amount);
            resultInput.classList.add('updated');
            setTimeout(() => resultInput.classList.remove('updated'), 1000);
        }
        const rateEl = document.getElementById('exchange-rate-value');
        if (rateEl) rateEl.textContent = `1 ${fromSymbol} = ${fmt(result.exchange_rate, 4)} ${toSymbol}`;
        const details = document.getElementById('conversion-details');
        if (details) details.style.display = 'block';
    }

    static displayConversionError(message) {
        const res = document.getElementById('convert-result');
        if (res) res.value = 'Ошибка';
        const rate = document.getElementById('exchange-rate-value');
        if (rate) rate.textContent = 'Не удалось получить курс';
        const details = document.getElementById('conversion-details');
        if (details) details.style.display = 'block';
        NotificationSystem.show(message || 'Ошибка конвертации', 'error');
    }

    static clearResult() {
        const res = document.getElementById('convert-result');
        if (res) res.value = '';
        const rate = document.getElementById('exchange-rate-value');
        if (rate) rate.textContent = '-';
        const details = document.getElementById('conversion-details');
        if (details) details.style.display = 'none';
    }

    static swapCurrencies() {
        const fromSelect = document.getElementById('convert-from');
        const toSelect = document.getElementById('convert-to');
        if (!fromSelect || !toSelect) return;
        const temp = fromSelect.value;
        fromSelect.value = toSelect.value;
        toSelect.value = temp;
        this.updateCurrencySymbol('from');
        this.updateCurrencySymbol('to');
        this.performAutoConversion();
    }

    static setLoadingState(loading) {
        const card = document.querySelector('.converter-card');
        const swap = document.getElementById('convert-swap');
        if (!card) return;
        if (loading) {
            card.classList.add('loading');
            if (swap) { swap.style.opacity = '0.7'; swap.style.pointerEvents = 'none'; }
        } else {
            card.classList.remove('loading');
            if (swap) { swap.style.opacity = '1'; swap.style.pointerEvents = 'auto'; }
        }
    }
}

class CurrencyConversion {
    static getRate(fromCode, toCode) {
        if (!fromCode || !toCode || fromCode === toCode) return 1;
        const from = currencies.find(c => c.code === fromCode);
        const to = currencies.find(c => c.code === toCode);
        if (!from || !to) return null;
        const direct = exchangeRates.find(r => r.base_currency_id === from.id && r.target_currency_id === to.id);
        if (direct) return parseFloat(direct.rate);
        const reverse = exchangeRates.find(r => r.base_currency_id === to.id && r.target_currency_id === from.id);
        if (reverse) return 1 / parseFloat(reverse.rate);
        const usd = currencies.find(c => c.code === 'USD');
        if (!usd) return null;
        const toUSD = exchangeRates.find(r => r.base_currency_id === from.id && r.target_currency_id === usd.id)
            || (exchangeRates.find(r => r.base_currency_id === usd.id && r.target_currency_id === from.id) ? { rate: 1 / parseFloat(exchangeRates.find(r => r.base_currency_id === usd.id && r.target_currency_id === from.id).rate) } : null);
        const fromUSD = exchangeRates.find(r => r.base_currency_id === usd.id && r.target_currency_id === to.id)
            || (exchangeRates.find(r => r.base_currency_id === to.id && r.target_currency_id === usd.id) ? { rate: 1 / parseFloat(exchangeRates.find(r => r.base_currency_id === to.id && r.target_currency_id === usd.id).rate) } : null);
        if (toUSD && fromUSD) return parseFloat(toUSD.rate) * parseFloat(fromUSD.rate);
        return null;
    }

    static getEquivalents(amount, fromCode, codes) {
        const uniq = Array.from(new Set(codes));
        return uniq.map(code => {
            const rate = this.getRate(fromCode, code) || 0;
            const converted = (parseFloat(amount) * rate).toFixed(2);
            return { code, amount: converted };
        });
    }
}

class CurrencySelection {
    static async ensureOrPrompt() {
        try {
            const profile = await ApiClient.request('/user/profile');
            currentUser = profile;
        } catch (_) {}
        const needSelect = !currentUser || !currentUser.default_currency_id;
        if (!needSelect) return;
        const allowed = ['USD','RUB','TJS'];
        const available = currencies.filter(c => allowed.includes(c.code));
        if (available.length === 0) return;
        const modal = document.createElement('div');
        modal.id = 'currency-select-modal';
        modal.className = 'modal-overlay active';
        modal.innerHTML = `
            <div class="modal-content currency-modal">
                <div class="modal-header">
                    <h3>Выберите основную валюту</h3>
                    <p>Эта валюта будет использоваться для отчетов и расчетов по умолчанию</p>
                </div>
                <div class="modal-body">
                    <div class="currency-grid">
                        ${available.map(c => {
                            const icons = { USD: 'fa-dollar-sign', RUB: 'fa-ruble-sign', TJS: 'fa-coins', EUR: 'fa-euro-sign', GBP: 'fa-sterling-sign', JPY: 'fa-yen-sign' };
                            const icon = icons[c.code] || 'fa-money-bill-wave';
                            const symbol = CurrencyUI.getSymbol(c.code);
                            return `
                                <div class="currency-card" data-id="${c.id}" data-code="${c.code}">
                                    <div class="currency-card__header">
                                        <div class="currency-icon"><i class="fas ${icon}"></i></div>
                                        <div class="currency-badge">${c.code}</div>
                                    </div>
                                    <div class="currency-card__body">
                                        <h4 class="currency-name">${c.name}</h4>
                                        <div class="currency-symbol">${symbol}</div>
                                    </div>
                                    <div class="currency-card__footer">
                                        <span class="select-hint">Выбрать</span>
                                    </div>
                                </div>
                            `;
                        }).join('')}
                    </div>
                </div>
                <div class="modal-footer">
                    <p class="help-text">Вы всегда можете изменить валюту в настройках профиля</p>
                </div>
            </div>`;
        document.body.appendChild(modal);
        modal.querySelectorAll('.currency-card').forEach(card => {
            card.addEventListener('click', async () => {
                const id = parseInt(card.getAttribute('data-id'));
                const code = card.getAttribute('data-code');
                card.classList.add('selected', 'loading');
                const hint = card.querySelector('.select-hint');
                if (hint) hint.textContent = 'Установка';
                try {
                    await ApiClient.request('/user/default-currency', { method: 'PUT', body: JSON.stringify({ currency_id: id }) });
                    card.classList.remove('loading');
                    card.classList.add('success');
                    if (hint) hint.innerHTML = '<i class="fas fa-check"></i> Установлена';
                    await new Promise(r => setTimeout(r, 800));
                    NotificationSystem.show(`Основная валюта установлена: ${code}`, 'success');
                    try {
                        const profile = await ApiClient.request('/user/profile');
                        currentUser = profile;
                        try { appState.setState({ user: profile }); } catch(_) {}
                    } catch (_e) {}
                    await Promise.all([
                        DataManager.loadAccountsData(),
                        DataManager.loadOverviewData(),
                        DataManager.loadAnalytics()
                    ]);
                    try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
                } catch (e) {
                    card.classList.remove('loading', 'selected');
                    if (hint) hint.textContent = 'Выбрать';
                    NotificationSystem.show(e.message || 'Ошибка установки валюты', 'error');
                } finally {
                    const content = modal.querySelector('.currency-modal');
                    if (content) {
                        content.style.opacity = '0';
                        content.style.transform = 'scale(0.9)';
                    } else {
                        modal.style.opacity = '0';
                        modal.style.transform = 'scale(0.9)';
                    }
                    setTimeout(() => {
                        modal.remove();
                        try { localStorage.removeItem('needs_currency_select'); } catch (_) {}
                    }, 300);
                }
            });
        });
    }
}

class CurrencyUI {
    static getSymbol(code) {
        if (!code) return '$';
        const found = currencies && currencies.find(c => c.code === code);
        if (found && found.symbol) return found.symbol;
        const fallback = { 
            USD: '$',
            EUR: '€',
            RUB: '₽',
            TJS: 'SM',
            GBP: '£',
            JPY: '¥',
            CNY: '¥',
            KZT: '₸',
            INR: '₹',
            UZS: "so'm",
            KGS: '⃀',
            AZN: '₼',
            BYN: 'Br',
            UAH: '₴'
        };
        return fallback[code] || '$';
    }
}

class LoadingManager {
    static show() {
        $('#loading-overlay').classList.add('active');
        // skeleton start: add loading shimmer to key containers while data loads
        ['#transactions', '#recent-transactions', '#categories-container', '#accounts-container', '#exchange-rates-container'].forEach(sel => {
            const el = document.querySelector(sel);
            if (el) el.classList.add('skeleton-loading');
        });
    }
    
    static hide() {
        $('#loading-overlay').classList.remove('active');
        // skeleton end: remove shimmer
        ['#transactions', '#recent-transactions', '#categories-container', '#accounts-container', '#exchange-rates-container'].forEach(sel => {
            const el = document.querySelector(sel);
            if (el) el.classList.remove('skeleton-loading');
        });
    }
}

// --- Security & Validation ---
class SecurityManager {
    static sanitizeHTML(str) {
        const div = document.createElement('div');
        div.textContent = String(str == null ? '' : str);
        return div.innerHTML;
    }
    static validateEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    }
    static validatePassword(password) {
        const minLength = 8;
        const hasUpperCase = /[A-Z]/.test(password);
        const hasLowerCase = /[a-z]/.test(password);
        const hasNumbers = /\d/.test(password);
        return (password || '').length >= minLength && hasUpperCase && hasLowerCase && hasNumbers;
    }
    static validateTransaction(data) {
        if (data.amount <= 0) throw new Error('Сумма должна быть положительной');
        if (!Number.isFinite(data.amount)) throw new Error('Некорректная сумма');
        if (!['income', 'expense'].includes(data.type)) throw new Error('Некорректный тип транзакции');
        return true;
    }
}

// --- State Management ---
class StateManager {
    constructor() {
        this.state = {
            user: null,
            transactions: [],
            accounts: [],
            categories: [],
            currencies: [],
            exchangeRates: [],
            ui: { currentTab: 'overview', loading: false }
        };
        this.listeners = [];
    }
    setState(newState) {
        this.state = { ...this.state, ...newState };
        this.notifyListeners();
    }
    subscribe(listener) { this.listeners.push(listener); }
    notifyListeners() { this.listeners.forEach(l => { try { l(this.state); } catch(_) {} }); }
}

const appState = new StateManager();

// (moved) appState subscription is defined after debounce to avoid TDZ issues

// --- Cache & Debounce ---
class CacheService {
    constructor() { this.cache = new Map(); this.TTL = 5 * 60 * 1000; }
    set(key, data) { this.cache.set(key, { data, timestamp: Date.now() }); }
    get(key) {
        const item = this.cache.get(key);
        if (!item) return null;
        if (Date.now() - item.timestamp > this.TTL) { this.cache.delete(key); return null; }
        return item.data;
    }
    clear() { this.cache.clear(); }
}
const apiCache = new CacheService();

const debounce = (fn, delay) => {
    let timeout;
    return (...args) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => fn(...args), delay);
    };
};

// --- Error Handling & Retry ---

// Track currency-related changes to trigger UI refresh (placed after debounce)
let _lastDefaultCurrencyId = null;
let _lastCurrenciesLen = 0;
const _refreshCurrencyDisplayDebounced = debounce(() => {
    try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
}, 80);

// keep legacy globals in sync for backward compatibility and auto-refresh on currency changes
appState.subscribe((s) => {
    try { currentUser = s.user; } catch(_) {}
    try { accounts = s.accounts; } catch(_) {}
    try { categories = s.categories; } catch(_) {}
    try { currencies = s.currencies; } catch(_) {}
    try { exchangeRates = s.exchangeRates; } catch(_) {}
    try { currentTab = s.ui && s.ui.currentTab ? s.ui.currentTab : currentTab; } catch(_) {}

    // Detect changes and refresh currency display
    try {
        const newId = s && s.user ? s.user.default_currency_id : null;
        const curLen = Array.isArray(s && s.currencies) ? s.currencies.length : 0;
        if (newId !== _lastDefaultCurrencyId) {
            _lastDefaultCurrencyId = newId;
            _refreshCurrencyDisplayDebounced();
        }
        if (curLen !== _lastCurrenciesLen) {
            _lastCurrenciesLen = curLen;
            _refreshCurrencyDisplayDebounced();
        }
    } catch(_) {}
});
class ErrorHandler {
    static async withRetry(fn, maxRetries = 3, delay = 1000) {
        for (let i = 0; i < maxRetries; i++) {
            try { return await fn(); }
            catch (error) {
                if (i === maxRetries - 1) throw error;
                await new Promise(r => setTimeout(r, delay * (i + 1)));
            }
        }
    }
    static isNetworkError(error) {
        return (error && error.message && (error.message.includes('Network') || error.message.includes('fetch')));
    }
}

class ApiClient {
    static async request(path, opts = {}) {
        const headers = Object.assign({ 'Content-Type': 'application/json' }, opts.headers || {});
        if (token) headers['Authorization'] = `Bearer ${token}`;

        const method = (opts.method || 'GET').toUpperCase();
        const key = method === 'GET' ? `GET:${path}` : null;

        // cache: only GET without explicit cache bust
        if (key) {
            const cached = apiCache.get(key);
            if (cached) return cached;
        }

        LoadingManager.show();

        const doFetch = async () => {
            const res = await fetch(`${window.API_BASE}${path}`, { ...opts, headers });
            const isJson = (res.headers.get('content-type') || '').includes('application/json');
            const body = isJson ? await res.json() : await res.text();
            if (!res.ok) throw new Error(body?.error || body?.message || res.statusText);
            if (key) apiCache.set(key, body);
            else {
                // invalidate cache on mutations
                apiCache.clear();
            }
            return body;
        };

        try {
            // retry on transient network failures
            return await ErrorHandler.withRetry(doFetch, 3, 700);
        } catch (error) {
            if (error && (error.name === 'TypeError' || ErrorHandler.isNetworkError(error))) {
                throw new Error('Ошибка сети: Не удалось подключиться к серверу');
            }
            throw error;
        } finally {
            LoadingManager.hide();
        }
    }

    static async health() {
        try {
            const res = await fetch(`${window.API_BASE}/health`);
            return await res.json();
        } catch (e) { return null; }
    }
}

class UIManager {
    static setAuthenticated(user) {
        $('#auth-forms').classList.add('hidden');
        $('#auth-info').classList.remove('hidden');
        const authSection = document.getElementById('auth');
        if (authSection) authSection.classList.add('hidden');
        // Fill pretty name and avatar
        const email = user.email || '';
        const rawName = user.username || user.name || (email.includes('@') ? email.split('@')[0] : 'Пользователь');
        const prettyName = rawName
            .replace(/[_\-.]+/g, ' ')
            .trim()
            .replace(/\s+/g, ' ')
            .replace(/\b\w/g, ch => ch.toUpperCase());
        const avatarLetter = (prettyName[0] || 'U').toUpperCase();
        const nameEl = document.getElementById('user-name');
        const mailEl = document.getElementById('user-email');
        const avatarEl = document.getElementById('user-avatar');
        if (nameEl) nameEl.textContent = prettyName;
        if (mailEl) mailEl.textContent = email;
        if (avatarEl) avatarEl.textContent = avatarLetter;
        $('#app').classList.remove('hidden');
        currentUser = user;
        try { appState.setState({ user }); } catch(_) {}
        
        this.updateGreeting();
        this.updateCurrentDate();
        this.animateDashboardEntrance();
        this.initTabNavigation();
        try { this.refreshCurrencyDisplay(); } catch(_) {}
    }
    
    static setAnonymous() {
        $('#auth-forms').classList.remove('hidden');
        $('#auth-info').classList.add('hidden');
        const authSection = document.getElementById('auth');
        if (authSection) authSection.classList.remove('hidden');
        $('#app').classList.add('hidden');
        currentUser = null;
        try { appState.setState({ user: null }); } catch(_) {}
        const nameEl = document.getElementById('user-name');
        const mailEl = document.getElementById('user-email');
        const avatarEl = document.getElementById('user-avatar');
        if (nameEl) nameEl.textContent = 'User';
        if (mailEl) mailEl.textContent = '';
        if (avatarEl) avatarEl.textContent = 'U';
    }
    
    static updateGreeting() {
        const hour = new Date().getHours();
        let greeting = 'Добро пожаловать';
        
        if (hour < 12) greeting = 'Доброе утро';
        else if (hour < 18) greeting = 'Добрый день';
        else greeting = 'Добрый вечер';
        
        $('#greeting-text').textContent = greeting;
    }
    
    static updateCurrentDate() {
        const now = new Date();
        const options = { 
            weekday: 'long', 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
        };
        $('#current-date').textContent = now.toLocaleDateString('ru-RU', options);
    }
    
    static animateDashboardEntrance() {
        const elements = $$('.metric-card, .content-section');
        elements.forEach((el, index) => {
            el.style.opacity = '0';
            el.style.transform = 'translateY(30px)';
            
            setTimeout(() => {
                el.style.transition = 'all 0.6s ease';
                el.style.opacity = '1';
                el.style.transform = 'translateY(0)';
            }, index * 100);
        });
    }
    
    static initTabNavigation() {
        $$('.nav-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                const tabName = tab.getAttribute('data-tab');
                this.switchTab(tabName);
            });
        });
    }
    
    static switchTab(tabName) {
        $$('.nav-tab').forEach(tab => {
            tab.classList.toggle('active', tab.getAttribute('data-tab') === tabName);
        });
        
        $$('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `${tabName}-tab`);
        });
        
        currentTab = tabName;
        // sync state
        try { appState.setState({ ui: { ...appState.state.ui, currentTab: tabName } }); } catch(_) {}
        
        this.loadTabData(tabName);
    }
    
    static async loadTabData(tabName) {
        switch (tabName) {
            case 'overview':
                await DataManager.loadOverviewData();
                try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
                break;
            case 'transactions':
                await DataManager.loadTransactions();
                break;
            case 'analytics':
                await DataManager.loadAnalytics();
                break;
            case 'accounts':
                await DataManager.loadAccountsData();
                break;
            case 'categories':
                await DataManager.loadCategoriesData();
                break;
            case 'exchange':
                await DataManager.loadExchangeData();
                try { if (document.getElementById('convert-from')) CurrencyConverter.init(); } catch(_) {}
                break;
        }
    }
    
    static updateFinancialMetrics(summary) {
        if (summary) {
            $('#total-income').textContent = this.formatCurrency(summary.total_income);
            $('#total-expense').textContent = this.formatCurrency(summary.total_expense);
            $('#net-amount').textContent = this.formatCurrency(summary.net_amount);
            
            let savingsRate = 0;
            if (summary.total_income > 0) {
                savingsRate = Math.max(0, Math.min(100, (summary.net_amount / summary.total_income) * 100));
                savingsRate = parseFloat(savingsRate.toFixed(1));
            }
            $('#savings-rate').textContent = `${savingsRate}%`;
            
            const gauge = $('#savings-gauge');
            gauge.style.width = `${Math.min(savingsRate, 100)}%`;
            
            this.updateMetricColors('#net-amount', summary.net_amount);
        }
    }
    
    static formatCurrency(amount) {
        let currencyCode = 'USD';
        let currencySymbol = '$';
        try {
            if (currentUser && currentUser.default_currency_id && currencies && currencies.length) {
                const defaultCurrency = currencies.find(c => c.id === currentUser.default_currency_id);
                if (defaultCurrency) {
                    currencyCode = defaultCurrency.code;
                    currencySymbol = defaultCurrency.symbol || CurrencyUI.getSymbol(defaultCurrency.code);
                }
            }
        } catch (_) {}
        const formatted = new Intl.NumberFormat('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(amount || 0);
        return `${formatted} ${currencySymbol}`;
    }

    static getCurrencySymbol(code) {
        return CurrencyUI.getSymbol(code);
    }

    static refreshCurrencyDisplay() {
        if (currentUser && currentUser.default_currency_id) {
            const def = currencies.find(c => c.id === currentUser.default_currency_id);
            if (def) {
                const summary = appState && appState.state && appState.state.summary;
                if (summary) {
                    UIManager.updateFinancialMetrics(summary);
                }
                if (typeof CurrencyConverter !== 'undefined') {
                    try { CurrencyConverter.updateCurrencySymbol('from'); } catch(_) {}
                    try { CurrencyConverter.updateCurrencySymbol('to'); } catch(_) {}
                }
                console.log('Валюта по умолчанию обновлена:', def.code, def.symbol);
            }
        }
    }
    
    static updateMetricColors(selector, value) {
        const element = $(selector);
        element.classList.remove('positive', 'negative', 'neutral');
        
        if (value > 0) {
            element.classList.add('positive');
        } else if (value < 0) {
            element.classList.add('negative');
        } else {
            element.classList.add('neutral');
        }
    }
}

class AuthManager {
    static async login(email, password) {
        try {
            const data = await ApiClient.request('/login', {
                method: 'POST',
                body: JSON.stringify({ email, password })
            });
            
            token = data.token;
            localStorage.setItem('token', token);
            try { localStorage.setItem('user', JSON.stringify(data.user || {})); } catch(_) {}
            UIManager.setAuthenticated(data.user);
            NotificationSystem.show('Вход выполнен успешно!', 'success');
            
            await DataManager.loadInitialData();
            try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
            const needs = localStorage.getItem('needs_currency_select');
            if (needs === '1') {
                await CurrencySelection.ensureOrPrompt();
            }
        } catch (error) {
            throw error;
        }
    }
    
    static async register(username, email, password) {
        try {
            await ApiClient.request('/register', {
                method: 'POST',
                body: JSON.stringify({ username, email, password })
            });
            
            NotificationSystem.show('Аккаунт успешно создан!', 'success');
            try { localStorage.setItem('needs_currency_select', '1'); } catch (_) {}
            await this.login(email, password);
        } catch (error) {
            throw error;
        }
    }
    
    static async logout() {
        try {
            await ApiClient.request('/logout', { method: 'POST' });
        } catch (error) {
            console.warn('Logout error:', error);
        } finally {
            token = '';
            localStorage.removeItem('token');
            try { localStorage.removeItem('user'); } catch(_) {}
            UIManager.setAnonymous();
            NotificationSystem.show('Вы успешно вышли из системы', 'info');
        }
    }
}

class DataManager {
    static async loadInitialData() {
        await Promise.all([
            this.loadCurrencies(),
            this.loadAccounts(),
            this.loadCategories(),
            this.loadOverviewData()
        ]);
    }
    
    static async loadOverviewData() {
        await Promise.all([
            this.loadDefaultAccountTransactions(),
            this.loadExchangeRatesOverview(),
            this.loadDefaultAccountSummary(),
            this.loadUserBalances()
        ]);
    }
    
    static async loadAccountsData() {
        await Promise.all([
            this.loadAccounts(),
            this.loadUserBalances(),
            this.loadExchangeRates()
        ]);
    }
    
    static async loadCategoriesData() {
        await this.loadCategories();
    }
    
    static async loadExchangeData() {
        await Promise.all([
            this.loadExchangeRates(),
            this.loadUserBalances(),
            this.loadQuickRates()
        ]);
    }
    
    static async loadAnalytics() {
        await Promise.all([
            this.loadSummary(),
            this.loadCategoryBreakdown(),
            this.loadMonthlySummary(),
            this.renderCharts()
        ]);
    }
    
    static async loadCurrencies() {
        try {
            const response = await ApiClient.request('/currencies');
            currencies = response;
            try { appState.setState({ currencies: response }); } catch(_) {}
            this.populateCurrencySelects();
            try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
        } catch (error) {
            console.error('Ошибка загрузки валют:', error);
            NotificationSystem.show('Не удалось загрузить список валют', 'error');
        }
    }
    
    static async loadAccounts() {
        try {
            const response = await ApiClient.request('/accounts');
            accounts = response;
            try { appState.setState({ accounts: response }); } catch(_) {}
            this.renderAccountSelect();
            this.updateAccountsDisplay();
        } catch (error) {
            console.error('Ошибка загрузки счетов:', error);
            NotificationSystem.show('Не удалось загрузить счета', 'error');
        }
    }
    
    static async loadCategories() {
        try {
            const response = await ApiClient.request('/categories');
            categories = response;
            try { appState.setState({ categories: response }); } catch(_) {}
            this.renderCategories();
            this.populateCategorySelect();
        } catch (error) {
            console.error('Ошибка загрузки категорий:', error);
            NotificationSystem.show('Не удалось загрузить категории', 'error');
        }
    }
    
    static async loadRecentTransactions() {
        try {
            const transactions = await ApiClient.request('/transactions?limit=5');
            this.renderRecentTransactions(transactions);
        } catch (error) {
            console.error('Ошибка загрузки recent transactions:', error);
        }
    }
    
    static async loadDefaultAccountTransactions() {
        try {
            const transactions = await ApiClient.request('/transactions/default?limit=5');
            this.renderRecentTransactions(transactions);
        } catch (error) {
            console.error('Ошибка загрузки default account transactions:', error);
        }
    }
    
    static async loadTransactions(filters = {}) {
        try {
            let url = '/transactions';
            const params = new URLSearchParams();
            
            if (filters.startDate && filters.endDate) {
                // sanitize dates
                const start = String(filters.startDate).slice(0, 10);
                const end = String(filters.endDate).slice(0, 10);
                params.append('start', start);
                params.append('end', end);
            }
            
            if (filters.types && filters.types.length > 0) {
                const allowed = ['income','expense'];
                filters.types.filter(t => allowed.includes(t)).forEach(type => params.append('type', type));
            }
            
            if (params.toString()) {
                url += `?${params.toString()}`;
            }
            
            const transactions = await ApiClient.request(url);
            this.renderTransactions(transactions);
            try { appState.setState({ transactions }); } catch(_) {}
        } catch (error) {
            NotificationSystem.show('Не удалось загрузить транзакции', 'error');
        }
    }
    
    static async loadSummary() {
        try {
            const today = new Date();
            const start = new Date(today.getFullYear(), today.getMonth(), 1);
            const startStr = start.toISOString().slice(0, 10);
            const endStr = today.toISOString().slice(0, 10);
            const summary = await ApiClient.request(`/transactions/summary?start=${startStr}&end=${endStr}`);
            UIManager.updateFinancialMetrics(summary);
            try { appState.setState({ summary }); } catch(_) {}
        } catch (error) {
            console.error('Ошибка загрузки summary:', error);
        }
    }
    
    static async loadDefaultAccountSummary() {
        try {
            const today = new Date();
            const start = new Date(today.getFullYear(), today.getMonth(), 1);
            const startStr = start.toISOString().slice(0, 10);
            const endStr = today.toISOString().slice(0, 10);
            const summary = await ApiClient.request(`/transactions/default/summary?start=${startStr}&end=${endStr}`);
            UIManager.updateFinancialMetrics(summary);
            try { appState.setState({ summary }); } catch(_) {}
        } catch (error) {
            console.error('Ошибка загрузки default account summary:', error);
        }
    }
    
    static async loadCategoryBreakdown() {
        try {
            const today = new Date();
            const start = new Date(today.getFullYear(), today.getMonth(), 1);
            const startStr = start.toISOString().slice(0, 10);
            const endStr = today.toISOString().slice(0, 10);
            const breakdown = await ApiClient.request(`/transactions/by-category?start=${startStr}&end=${endStr}`);
            this.renderCategoryBreakdown(breakdown);
        } catch (error) {
            console.error('Ошибка загрузки breakdown:', error);
        }
    }
    
    static async loadMonthlySummary() {
        try {
            const year = new Date().getFullYear();
            const monthly = await ApiClient.request(`/transactions/monthly-summary?year=${year}`);
        } catch (error) {
            console.error('Ошибка загрузки monthly summary:', error);
        }
    }
    
    static async loadExchangeRates() {
        try {
            const response = await ApiClient.request('/exchange/rates');
            exchangeRates = response;
            try { appState.setState({ exchangeRates: response }); } catch(_) {}
            this.updateExchangeRatesDisplay();
        } catch (error) {
            console.error('Ошибка загрузки курсов:', error);
            NotificationSystem.show('Не удалось загрузить курсы валют', 'error');
        }
    }
    
    static async loadExchangeRatesOverview() {
        try {
            const rates = await ApiClient.request('/exchange/rates?limit=3');
            this.updateExchangeRatesOverview(rates);
        } catch (error) {
            console.error('Ошибка загрузки курсов для overview:', error);
        }
    }
    
    static async loadUserBalances() {
        try {
            const balances = await ApiClient.request('/exchange/balances');
            this.updateBalancesDisplay(balances);
        } catch (error) {
            console.error('Ошибка загрузки балансов:', error);
        }
    }
    
    static async loadQuickRates() {
        try {
            const rates = await ApiClient.request('/exchange/rates?limit=4');
            this.updateQuickRates(rates);
        } catch (error) {
            console.error('Ошибка загрузки quick rates:', error);
        }
    }
    
    static populateCurrencySelects() {
        const accountCurrencySelect = $('#account-currency');
        const convertFromSelect = $('#convert-from');
        const convertToSelect = $('#convert-to');
        const list = Array.isArray(currencies) ? currencies : [];

        if (accountCurrencySelect) {
            accountCurrencySelect.innerHTML = '<option value="">Выберите валюту</option>';
        }
        if (convertFromSelect) {
            convertFromSelect.innerHTML = '<option value="">Из валюты</option>';
        }
        if (convertToSelect) {
            convertToSelect.innerHTML = '<option value="">В валюту</option>';
        }

        list.forEach(currency => {
            if (accountCurrencySelect) {
                const accountOption = document.createElement('option');
                accountOption.value = currency.id;
                accountOption.textContent = `${currency.name} (${currency.code})`;
                accountCurrencySelect.appendChild(accountOption);
            }
            if (convertFromSelect && convertToSelect) {
                const convertOption1 = document.createElement('option');
                convertOption1.value = currency.id;
                convertOption1.textContent = `${currency.code} - ${currency.name}`;
                convertFromSelect.appendChild(convertOption1.cloneNode(true));
                convertToSelect.appendChild(convertOption1.cloneNode(true));
            }
        });
    }
    
    static renderAccountSelect() {
        const select = $('#tx-account');
        if (!select) return;
        select.innerHTML = '<option value="">Выберите счет</option>';
        const list = Array.isArray(accounts) ? accounts : [];
        list.forEach(account => {
            const currency = currencies && Array.isArray(currencies) ? currencies.find(c => c.id === account.currency_id) : null;
            const option = document.createElement('option');
            option.value = account.id;
            option.textContent = `${account.name} (${currency ? currency.code : 'N/A'}) - ${account.balance} ${currency ? currency.symbol : ''}`;
            select.appendChild(option);
        });
    }
    
    static populateCategorySelect() {
        const select = $('#tx-category');
        if (!select) return;
        select.innerHTML = '<option value="">Выберите категорию</option>';
        const list = Array.isArray(categories) ? categories : [];
        list.forEach(category => {
            const option = document.createElement('option');
            option.value = category.id;
            const icon = category.type === 'income' ? '📈' : '📉';
            option.textContent = `${icon} ${category.name} (${category.type === 'income' ? 'доход' : 'расход'})`;
            select.appendChild(option);
        });
    }
    
    static renderRecentTransactions(transactions) {
        const container = $('#recent-transactions');
        
        if (!transactions || transactions.length === 0) {
            container.innerHTML = `
                <div class="empty-state compact">
                    <i class="fas fa-receipt"></i>
                    <p>Нет недавних транзакций</p>
                </div>
            `;
            return;
        }
        
        container.innerHTML = transactions.map(transaction => {
            const date = new Date(transaction.date).toLocaleDateString('ru-RU');
            const amount = parseFloat(transaction.amount).toFixed(2);
            const currency = currencies.find(c => c.id === transaction.currency_id);
            const sym = currency ? (currency.symbol || UIManager.getCurrencySymbol(currency.code)) : UIManager.getCurrencySymbol('USD');
            const category = categories.find(cat => cat.id === transaction.category_id);
            
            return `
                <div class="transaction-item compact transaction-${transaction.type}">
                    <div class="transaction-icon">
                        <i class="fas ${transaction.type === 'income' ? 'fa-arrow-up' : 'fa-arrow-down'}"></i>
                    </div>
                    <div class="transaction-details">
                        <div class="transaction-title">${SecurityManager.sanitizeHTML(transaction.description || 'Без описания')}</div>
                        <div class="transaction-meta">
                            <span>${date}</span>
                            <span>${SecurityManager.sanitizeHTML(category ? category.name : 'Без категории')}</span>
                        </div>
                    </div>
                    <div class="transaction-amount amount-${transaction.type}">
                        ${transaction.type === 'income' ? '+' : '-'}${amount} ${sym}
                    </div>
                </div>
            `;
        }).join('');
    }

    static renderTransactions(transactions) {
        const container = $('#transactions');
        
        if (!transactions || transactions.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">
                        <i class="fas fa-receipt"></i>
                    </div>
                    <h4>Транзакций не найдено</h4>
                    <p>Попробуйте изменить фильтры или добавьте новую транзакцию</p>
                </div>
            `;
            return;
        }
        
        container.innerHTML = transactions.map(transaction => {
            const date = new Date(transaction.date).toLocaleDateString('ru-RU');
            const amount = parseFloat(transaction.amount).toFixed(2);
            const currency = currencies.find(c => c.id === transaction.currency_id);
            const sym = currency ? (currency.symbol || UIManager.getCurrencySymbol(currency.code)) : UIManager.getCurrencySymbol('USD');
            const account = accounts.find(a => a.id === transaction.account_id);
            const category = categories.find(cat => cat.id === transaction.category_id);
            
            return `
                <div class="transaction-item transaction-${transaction.type}">
                    <div class="transaction-icon">
                        <i class="fas ${transaction.type === 'income' ? 'fa-arrow-up' : 'fa-arrow-down'}"></i>
                    </div>
                    <div class="transaction-details">
                        <div class="transaction-title">${SecurityManager.sanitizeHTML(transaction.description || 'Без описания')}</div>
                        <div class="transaction-meta">
                            <span>${date}</span>
                            <span>${SecurityManager.sanitizeHTML(account ? account.name : 'Неизвестный счет')}</span>
                            <span>${SecurityManager.sanitizeHTML(category ? category.name : 'Без категории')}</span>
                        </div>
                    </div>
                    <div class="transaction-amount amount-${transaction.type}">
                        ${transaction.type === 'income' ? '+' : '-'}${amount} ${sym}
                    </div>
                </div>
            `;
        }).join('');
    }
    
    static renderCategories() {
        const container = $('#categories-container');
        
        if (!categories || categories.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">
                        <i class="fas fa-tags"></i>
                    </div>
                    <h4>Категорий пока нет</h4>
                    <p>Создайте первую категорию для организации транзакций</p>
                </div>
            `;
            return;
        }

        container.innerHTML = categories.map(category => `
            <div class="category-card ${category.type}">
                <div class="category-header">
                    <div class="category-info">
                        <h4>${category.name}</h4>
                        <div class="category-type ${category.type}">
                            <i class="fas ${category.type === 'income' ? 'fa-arrow-up' : 'fa-arrow-down'}"></i>
                            ${category.type === 'income' ? 'Доход' : 'Расход'}
                        </div>
                    </div>
                </div>
                ${category.description ? `
                    <div class="category-desc">${category.description}</div>
                ` : ''}
                <div class="category-stats">
                    <div class="category-stat">
                        <span class="stat-label">Создана:</span>
                        <span class="stat-value">${new Date(category.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                </div>
                <div class="category-actions">
                    <button class="btn-danger btn-sm" onclick="deleteCategory(${category.id})">
                        <i class="fas fa-trash"></i>
                        Удалить
                    </button>
                </div>
            </div>
        `).join('');
    }
    
    static renderCategoryBreakdown(breakdown) {
        const container = $('#by-category');
        
        if (!breakdown || breakdown.length === 0) {
            container.innerHTML = '<div class="empty-state compact">Нет данных для отображения</div>';
            return;
        }
        
        container.innerHTML = breakdown.map(item => `
            <div class="breakdown-item">
                <div class="breakdown-info">
                    <div class="breakdown-name">${item.category_name}</div>
                    <div class="category-type type-${item.type}">${item.type === 'income' ? 'доход' : 'расход'}</div>
                </div>
                <div class="breakdown-stats">
                    <div class="breakdown-amount">${UIManager.formatCurrency(item.total_amount)}</div>
                    <div class="breakdown-percentage">${parseFloat(item.percentage).toFixed(1)}%</div>
                </div>
            </div>
        `).join('');
    }
    
    static updateAccountsDisplay() {
        const containerEl = $('#accounts-container');
        if (!containerEl) return;
        
        if (!accounts || accounts.length === 0) {
            containerEl.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-wallet"></i>
                    <h4>У вас пока нет счетов</h4>
                    <p>Создайте свой первый счет для начала работы</p>
                    <button class="btn-primary" onclick="showCreateAccountModal()">
                        <i class="fas fa-plus"></i> Создать счет
                    </button>
                </div>
            `;
            return;
        }
        
        containerEl.innerHTML = accounts.map(account => {
            const currency = currencies.find(c => c.id === account.currency_id);
            const currencyCode = currency ? currency.code : 'USD';
            const currencySymbol = CurrencyUI.getSymbol(currencyCode);
            const equivalents = CurrencyConversion.getEquivalents(account.balance, currencyCode, ['USD','RUB','TJS']);
            
            return `
                <div class="account-card ${account.is_default ? 'default' : ''}" onclick="showAccountDetails(${account.id})">
                    <div class="account-header">
                        <div class="account-info">
                            <h4>${account.name || 'Основной счет'}</h4>
                            <span class="currency-code">${currencyCode}</span>
                        </div>
                        <div class="account-actions">
                            ${!account.is_default ? `
                                <button class="btn-secondary btn-sm" onclick="event.stopPropagation(); setDefaultAccount(${account.id})">
                                    <i class="fas fa-star"></i>
                                    Основной
                                </button>
                            ` : `
                                <span class="default-badge">
                                    <i class="fas fa-star"></i>
                                    Основной счет
                                </span>
                            `}
                        </div>
                    </div>
                    <div class="account-balance">
                        <span class="balance-amount">${parseFloat(account.balance).toFixed(2)}</span>
                        <span class="balance-currency">${currencySymbol}</span>
                    </div>
                    <div class="account-equivalents">
                        ${equivalents.filter(e => e.code !== currencyCode).map(e => `
                            <div class="equivalent-line">≈ ${e.amount} ${e.code}</div>
                        `).join('')}
                    </div>
                    <div class="account-meta">
                        <span>Обновлено: ${new Date(account.updated_at || Date.now()).toLocaleDateString('ru-RU')}</span>
                    </div>
                </div>
            `;
        }).join('');
    }
    
    static updateExchangeRatesOverview(rates) {
        const container = $('#exchange-rates-overview');
        if (!container || !rates) return;
        
        container.innerHTML = rates.slice(0, 3).map(rate => {
            const baseCurrency = currencies.find(c => c.id === rate.base_currency_id);
            const targetCurrency = currencies.find(c => c.id === rate.target_currency_id);
            
            if (!baseCurrency || !targetCurrency) return '';
            
            const change = Math.random() > 0.5 ? Math.random() * 2 : -Math.random() * 2;
            const changeClass = change > 0 ? 'positive' : change < 0 ? 'negative' : 'neutral';
            
            return `
                <div class="exchange-rate-item">
                    <span class="currency-pair">${baseCurrency.code}/${targetCurrency.code}</span>
                    <span class="rate-value">${parseFloat(rate.rate).toFixed(4)}</span>
                    <span class="rate-change ${changeClass}">
                        <i class="fas fa-arrow-${change > 0 ? 'up' : 'down'}"></i>
                        ${Math.abs(change).toFixed(2)}%
                    </span>
                </div>
            `;
        }).join('');
    }
    
    static updateExchangeRatesDisplay() {
        const container = $('#exchange-rates-container');
        if (!container || !exchangeRates) return;
        
        if (exchangeRates.length === 0) {
            container.innerHTML = `
                <div class="empty-state compact">
                    <i class="fas fa-exchange-alt"></i>
                    <p>Курсы валют не загружены</p>
                </div>
            `;
            return;
        }
        
        container.innerHTML = exchangeRates.map(rate => {
            const baseCurrency = currencies.find(c => c.id === rate.base_currency_id);
            const targetCurrency = currencies.find(c => c.id === rate.target_currency_id);
            
            if (!baseCurrency || !targetCurrency) return '';
            
            return `
                <div class="exchange-rate-card">
                    <div class="rate-pair">
                        <span class="base-currency">${baseCurrency.code}</span>
                        <i class="fas fa-arrow-right"></i>
                        <span class="target-currency">${targetCurrency.code}</span>
                    </div>
                    <div class="rate-value">${parseFloat(rate.rate).toFixed(6)}</div>
                    <div class="rate-updated">Обновлено: ${new Date(rate.last_updated || Date.now()).toLocaleString()}</div>
                </div>
            `;
        }).join('');
    }
    
    static updateBalancesDisplay(balances) {
        const containers = ['#balances-container', '#exchange-balances-container'];
        
        containers.forEach(containerId => {
            const container = $(containerId);
            if (!container) return;
            
            if (!balances || balances.length === 0) {
                container.innerHTML = '<div class="empty-state compact">Нет данных о балансах</div>';
                return;
            }
            
            container.innerHTML = balances.map(balance => {
                const code = balance.currency_code;
                const amount = parseFloat(balance.balance).toFixed(2);
                const usdApprox = balance.balance_in_usd ? parseFloat(balance.balance_in_usd).toFixed(2) : null;
                const usdSymbol = UIManager.getCurrencySymbol('USD');
                return `
                    <div class="balance-card">
                        <div class="balance-info">
                            <h4>${code}</h4>
                            <div class="balance-amount">${amount}</div>
                        </div>
                        ${usdApprox ? `
                            <div class="balance-usd">
                                ≈ ${usdSymbol}${usdApprox} USD
                            </div>
                        ` : ''}
                    </div>
                `;
            }).join('');
        });
        
        // Обновляем общий баланс на главной странице
        if (balances && balances.length > 0) {
            const totalUSD = balances.reduce((sum, balance) => {
                return sum + (balance.balance_in_usd ? parseFloat(balance.balance_in_usd) : 0);
            }, 0);
            
            const netAmountElement = $('#net-amount');
            if (netAmountElement) {
                netAmountElement.textContent = UIManager.formatCurrency(totalUSD);
                UIManager.updateMetricColors('#net-amount', totalUSD);
            } else {
            }
        } else {
        }
    }
    
    static updateQuickRates(rates) {
        const container = $('#quick-rates');
        if (!container || !rates) return;
        
        container.innerHTML = rates.map(rate => {
            const baseCurrency = currencies.find(c => c.id === rate.base_currency_id);
            const targetCurrency = currencies.find(c => c.id === rate.target_currency_id);
            
            if (!baseCurrency || !targetCurrency) return '';
            
            return `
                <div class="quick-rate-item">
                    <span class="rate-pair">${baseCurrency.code}/${targetCurrency.code}</span>
                    <span class="rate-value">${parseFloat(rate.rate).toFixed(4)}</span>
                </div>
            `;
        }).join('');
    }
    
    static renderCharts() {
        this.renderCategoryChart();
        this.renderMonthlyChart();
        this.renderExchangeRatesChart();
    }
    
    static async renderCategoryChart() {
        const ctx = document.getElementById('chart-by-category');
        if (!ctx) return;
        
        try {
            const today = new Date();
            const start = new Date(today.getFullYear(), today.getMonth(), 1);
            const startStr = start.toISOString().slice(0, 10);
            const endStr = today.toISOString().slice(0, 10);
            const breakdown = await ApiClient.request(`/transactions/by-category?start=${startStr}&end=${endStr}`);
            
            if (charts.category) {
                charts.category.destroy();
            }
            
            if (!breakdown || breakdown.length === 0) {
                ctx.parentElement.innerHTML = '<div class="empty-state compact">Нет данных для графика</div>';
                return;
            }
            
            const labels = breakdown.map(x => x.category_name);
            const data = breakdown.map(x => parseFloat(x.total_amount));
            const backgroundColors = breakdown.map((_, i) => {
                const hue = (i * 137.5) % 360;
                return `hsl(${hue}, 70%, 60%)`;
            });
            
            charts.category = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels,
                    datasets: [{
                        data,
                        backgroundColor: backgroundColors,
                        borderWidth: 2,
                        borderColor: 'var(--bg-primary)'
                    }]
                },
                options: {
                    responsive: true,
                    plugins: {
                        legend: {
                            position: 'right',
                            labels: {
                                color: 'var(--text-primary)',
                                font: { size: 12 }
                            }
                        }
                    },
                    animation: {
                        animateScale: true,
                        animateRotate: true
                    }
                }
            });
        } catch (error) {
            console.error('Category chart error:', error);
            ctx.parentElement.innerHTML = '<div class="empty-state compact">Ошибка загрузки графика</div>';
        }
    }
    
    static async renderMonthlyChart() {
        const ctx = document.getElementById('chart-monthly');
        if (!ctx) return;
        
        try {
            const year = new Date().getFullYear();
            const monthly = await ApiClient.request(`/transactions/monthly-summary?year=${year}`);
            
            if (charts.monthly) {
                charts.monthly.destroy();
            }
            
            if (!monthly || monthly.length === 0) {
                ctx.parentElement.innerHTML = '<div class="empty-state compact">Нет данных для графика</div>';
                return;
            }
            
            const monthNames = ['Янв', 'Фев', 'Мар', 'Апр', 'Май', 'Июн', 'Июл', 'Авг', 'Сен', 'Окт', 'Ноя', 'Дек'];
            const labels = monthly.map(m => monthNames[parseInt(m.month) - 1]);
            const income = monthly.map(m => parseFloat(m.total_income) || 0);
            const expense = monthly.map(m => parseFloat(m.total_expense) || 0);
            
            charts.monthly = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels,
                    datasets: [
                        {
                            label: 'Доходы',
                            data: income,
                            backgroundColor: 'rgba(6, 214, 160, 0.7)',
                            borderColor: 'rgba(6, 214, 160, 1)',
                            borderWidth: 1
                        },
                        {
                            label: 'Расходы',
                            data: expense,
                            backgroundColor: 'rgba(239, 68, 68, 0.7)',
                            borderColor: 'rgba(239, 68, 68, 1)',
                            borderWidth: 1
                        }
                    ]
                },
                options: {
                    responsive: true,
                    scales: {
                        x: {
                            stacked: false,
                            ticks: { color: 'var(--text-secondary)' },
                            grid: { color: 'var(--border)' }
                        },
                        y: {
                            beginAtZero: true,
                            ticks: { 
                                color: 'var(--text-secondary)',
                                callback: function(value) {
                                    try {
                                        let cur = 'USD';
                                        if (currentUser && currentUser.default_currency_id && currencies && currencies.length) {
                                            const c = currencies.find(x => x.id === currentUser.default_currency_id);
                                            if (c && c.code) cur = c.code;
                                        }
                                        const sym = (typeof UIManager !== 'undefined' && UIManager.getCurrencySymbol) ? UIManager.getCurrencySymbol(cur) : '$';
                                        return sym + value;
                                    } catch (_) { 
                                        const fallback = (typeof UIManager !== 'undefined' && UIManager.getCurrencySymbol) ? UIManager.getCurrencySymbol('USD') : '$';
                                        return fallback + value; 
                                    }
                                }
                            },
                            grid: { color: 'var(--border)' }
                        }
                    },
                    plugins: {
                        legend: {
                            labels: { color: 'var(--text-primary)' }
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Monthly chart error:', error);
            ctx.parentElement.innerHTML = '<div class="empty-state compact">Ошибка загрузки графика</div>';
        }
    }
    
    static async renderExchangeRatesChart() {
        const ctx = document.getElementById('chart-exchange-rates');
        if (!ctx) return;
        
        try {
            const days = 30;
            const labels = Array.from({length: days}, (_, i) => {
                const date = new Date();
                date.setDate(date.getDate() - (days - i - 1));
                return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
            });
            
            const usdToRub = Array.from({length: days}, () => 80 + Math.random() * 10);
            const eurToRub = Array.from({length: days}, () => 90 + Math.random() * 10);
            
            if (charts.exchange) {
                charts.exchange.destroy();
            }
            
            charts.exchange = new Chart(ctx, {
                type: 'line',
                data: {
                    labels,
                    datasets: [
                        {
                            label: 'USD/RUB',
                            data: usdToRub,
                            borderColor: 'rgba(99, 102, 241, 1)',
                            backgroundColor: 'rgba(99, 102, 241, 0.1)',
                            tension: 0.4,
                            fill: false
                        },
                        {
                            label: 'EUR/RUB',
                            data: eurToRub,
                            borderColor: 'rgba(6, 214, 160, 1)',
                            backgroundColor: 'rgba(6, 214, 160, 0.1)',
                            tension: 0.4,
                            fill: false
                        }
                    ]
                },
                options: {
                    responsive: true,
                    interaction: {
                        intersect: false,
                        mode: 'index'
                    },
                    scales: {
                        x: {
                            ticks: { color: 'var(--text-secondary)' },
                            grid: { color: 'var(--border)' }
                        },
                        y: {
                            ticks: { color: 'var(--text-secondary)' },
                            grid: { color: 'var(--border)' }
                        }
                    },
                    plugins: {
                        legend: {
                            labels: { color: 'var(--text-primary)' }
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Exchange rates chart error:', error);
            ctx.parentElement.innerHTML = '<div class="empty-state compact">Ошибка загрузки графика</div>';
        }
    }
}

class FormHandlers {
    static init() {
        // Floating labels: apply has-value class and keep it in sync
        const fields = $$('.input-group input, .input-group select, .input-group textarea');
        const apply = (el) => {
            if (!el) return;
            const has = !!(el.value && String(el.value).trim().length > 0);
            el.classList.toggle('has-value', has);
        };
        fields.forEach(el => {
            apply(el);
            el.addEventListener('input', () => apply(el));
            el.addEventListener('change', () => apply(el));
            el.addEventListener('blur', () => apply(el));
        });

        this.initAuthForms();
        this.initDataForms();
        this.initUIInteractions();
        this.initDateDefaults();
    }
    
    static initAuthForms() {
        $('#login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = $('#login-email').value;
            const password = $('#login-password').value;
            
            try {
                if (!SecurityManager.validateEmail(email)) {
                    NotificationSystem.show('Введите корректный email', 'error');
                    return;
                }
                if (!SecurityManager.validatePassword(password)) {
                    NotificationSystem.show('Пароль должен быть не менее 8 символов и содержать цифры, строчные и заглавные буквы', 'error');
                    return;
                }
                await AuthManager.login(email, password);
            } catch (error) {
                NotificationSystem.show(error.message, 'error');
            }
        });
        
        $('#register-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = $('#reg-username').value;
            const email = $('#reg-email').value;
            const password = $('#reg-password').value;
            
            try {
                if (!SecurityManager.validateEmail(email)) {
                    NotificationSystem.show('Введите корректный email', 'error');
                    return;
                }
                if (!SecurityManager.validatePassword(password)) {
                    NotificationSystem.show('Пароль должен быть не менее 8 символов и содержать цифры, строчные и заглавные буквы', 'error');
                    return;
                }
                await AuthManager.register(username, email, password);
            } catch (error) {
                NotificationSystem.show(error.message, 'error');
            }
        });

        $('#logout-btn').addEventListener('click', () => {
            AuthManager.logout();
        });

        $$('.auth-tabs .tab-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const tabName = btn.getAttribute('data-tab');
                this.switchAuthTab(tabName);
            });
        });

        const loginToggle = $('#login-pass-toggle');
        if (loginToggle) {
            loginToggle.addEventListener('click', () => {
                this.togglePassword('login-password', 'login-pass-toggle');
            });
        }
        const regToggle = $('#reg-pass-toggle');
        if (regToggle) {
            regToggle.addEventListener('click', () => {
                this.togglePassword('reg-password', 'reg-pass-toggle');
            });
        }

        const regPass = $('#reg-password');
        if (regPass) {
            regPass.addEventListener('input', () => this.updatePasswordStrength(regPass.value));
            this.updatePasswordStrength(regPass.value || '');
        }
    }

    static togglePassword(inputId, btnId) {
        const input = document.getElementById(inputId);
        const btn = document.getElementById(btnId);
        if (!input || !btn) return;
        const icon = btn.querySelector('i');
        if (input.type === 'password') {
            input.type = 'text';
            if (icon) { icon.classList.remove('fa-eye'); icon.classList.add('fa-eye-slash'); }
            btn.setAttribute('aria-label', 'Скрыть пароль');
        } else {
            input.type = 'password';
            if (icon) { icon.classList.remove('fa-eye-slash'); icon.classList.add('fa-eye'); }
            btn.setAttribute('aria-label', 'Показать пароль');
        }
    }

    static updatePasswordStrength(value) {
        const bar = document.getElementById('password-strength-bar');
        const text = document.getElementById('password-strength-text');
        if (!bar || !text) return;
        let score = 0;
        if (value.length >= 8) score++;
        if (/[0-9]/.test(value)) score++;
        if (/[a-z]/.test(value) && /[A-Z]/.test(value)) score++;
        if (/[^A-Za-z0-9]/.test(value)) score++;
        const pct = [0, 25, 50, 75, 100][score];
        bar.style.width = pct + '%';
        bar.className = '';
        bar.classList.add('strength-fill');
        bar.classList.toggle('weak', score <= 1);
        bar.classList.toggle('medium', score === 2);
        bar.classList.toggle('good', score === 3);
        bar.classList.toggle('strong', score >= 4);
        const labels = ['Очень слабый', 'Слабый', 'Средний', 'Хороший', 'Сильный'];
        text.textContent = 'Надёжность пароля: ' + labels[score];
    }
    
    // Settings: open modal
    static openSettingsModal() {
        try {
            this.populateSettingsCurrencies();
            const nameInput = document.getElementById('settings-username');
            if (nameInput) nameInput.value = (currentUser && (currentUser.username || currentUser.name)) || '';
            const modal = document.getElementById('settings-modal');
            if (modal) modal.classList.remove('hidden');
        } catch (_) {}
    }

    // Settings: fill default currency list
    static populateSettingsCurrencies() {
        const sel = document.getElementById('settings-default-currency');
        if (!sel) return;
        sel.innerHTML = '<option value="">Валюта по умолчанию</option>';
        const list = Array.isArray(currencies) ? currencies : [];
        list.forEach(c => {
            const opt = document.createElement('option');
            opt.value = String(c.id);
            opt.textContent = `${c.name} (${c.code})`;
            sel.appendChild(opt);
        });
        if (currentUser && currentUser.default_currency_id) sel.value = String(currentUser.default_currency_id);
    }

    // Settings: submit handler
    static async handleSettingsSubmit(e) {
        e.preventDefault();
        try {
            const unameEl = document.getElementById('settings-username');
            const curSel = document.getElementById('settings-default-currency');
            const oldPass = document.getElementById('old-password').value;
            const newPass = document.getElementById('new-password').value;
            const confirmPass = document.getElementById('confirm-password').value;

            const updates = [];

            const newUsername = unameEl ? unameEl.value.trim() : '';
            if (newUsername && currentUser && newUsername !== (currentUser.username || '')) {
                updates.push(ApiClient.request('/user/profile', {
                    method: 'PUT',
                    body: JSON.stringify({ username: newUsername })
                }));
            }

            const newCurrencyId = curSel && curSel.value ? parseInt(curSel.value, 10) : null;
            if (newCurrencyId && currentUser && newCurrencyId !== currentUser.default_currency_id) {
                updates.push(ApiClient.request('/user/default-currency', {
                    method: 'PUT',
                    body: JSON.stringify({ currency_id: newCurrencyId })
                }));
            }

            const wantPasswordChange = newPass || confirmPass || oldPass;
            if (wantPasswordChange) {
                if (!oldPass) { NotificationSystem.show('Введите текущий пароль', 'error'); return; }
                if (!newPass || newPass.length < 8 || !SecurityManager.validatePassword(newPass)) { NotificationSystem.show('Новый пароль должен быть надёжным', 'error'); return; }
                if (newPass !== confirmPass) { NotificationSystem.show('Пароли не совпадают', 'error'); return; }
                updates.push(ApiClient.request('/user/password', {
                    method: 'PUT',
                    body: JSON.stringify({ old_password: oldPass, new_password: newPass })
                }));
            }

            if (updates.length === 0) { NotificationSystem.show('Нет изменений', 'info'); return; }

            await Promise.all(updates);
            try {
                const profile = await ApiClient.request('/user/profile');
                currentUser = profile;
                try { localStorage.setItem('user', JSON.stringify(profile || {})); } catch(_) {}
                UIManager.setAuthenticated(profile);
                try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
            } catch (_) {}

            NotificationSystem.show('Настройки сохранены', 'success');
            closeModal('settings-modal');
            ['old-password','new-password','confirm-password'].forEach(id => { const el = document.getElementById(id); if (el) el.value=''; });
        } catch (error) {
            NotificationSystem.show(error.message || 'Ошибка сохранения настроек', 'error');
        }
    }
    
    static initDataForms() {
        const txForm = $('#tx-form');
        if (txForm) txForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            try {
                const accountId = $('#tx-account').value;
                const categoryId = $('#tx-category').value;
                
                if (!accountId) {
                    NotificationSystem.show('Выберите счет для транзакции', 'error');
                    return;
                }
                
                if (!categoryId) {
                    NotificationSystem.show('Выберите категорию для транзакции', 'error');
                    return;
                }

                const body = {
                    category_id: parseInt(categoryId, 10),
                    account_id: parseInt(accountId, 10),
                    amount: parseFloat($('#tx-amount').value),
                    description: $('#tx-desc').value,
                    date: $('#tx-date').value,
                    type: $('#tx-type').value
                };
                // validate and sanitize
                SecurityManager.validateTransaction(body);
                body.description = SecurityManager.sanitizeHTML(body.description || '');
                
                // optimistic UI hint
                const listEl = document.querySelector('#transactions');
                if (listEl) listEl.classList.add('optimistic-update');
                await ApiClient.request('/transactions', {
                    method: 'POST',
                    body: JSON.stringify(body)
                });
                
                $('#tx-amount').value = '';
                $('#tx-desc').value = '';
                $('#tx-form').classList.add('hidden');
                
                NotificationSystem.show('Транзакция успешно добавлена!', 'success');
                
                await Promise.all([
                    UIManager.loadTabData(currentTab),
                    DataManager.loadOverviewData()
                ]);
            } catch (error) {
                NotificationSystem.show(error.message, 'error');
            } finally {
                const listEl = document.querySelector('#transactions');
                if (listEl) listEl.classList.remove('optimistic-update');
            }
        });

        const categoryForm = $('#category-form');
        if (categoryForm) categoryForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            try {
                const body = {
                    name: $('#cat-name').value.trim(),
                    type: $('#cat-type').value,
                    description: $('#cat-desc').value.trim() || ''
                };
                
                if (!body.name) {
                    NotificationSystem.show('Введите название категории', 'error');
                    return;
                }
                
                await ApiClient.request('/categories', {
                    method: 'POST',
                    body: JSON.stringify(body)
                });
                
                // Очистка формы
                $('#cat-name').value = '';
                $('#cat-desc').value = '';
                $('#cat-type').value = 'expense';
                
                // Сброс состояния полей
                $$('#category-form input').forEach(input => {
                    input.classList.remove('has-value');
                });
                
                NotificationSystem.show('Категория успешно создана!', 'success');
                
                await DataManager.loadCategories();
                await DataManager.populateCategorySelect();
            } catch (error) {
                NotificationSystem.show(error.message, 'error');
            }
        });

        // Converter form is optional on some screens; guard to avoid init errors
        const converterForm = document.getElementById('converter-form');
        if (converterForm) {
            converterForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                await this.performConversion();
            });
        }

        const swapBtn = $('#convert-swap');
        if (swapBtn) {
            swapBtn.addEventListener('click', async () => {
                const fromSel = $('#convert-from');
                const toSel = $('#convert-to');
                const tmp = fromSel.value;
                fromSel.value = toSel.value;
                toSel.value = tmp;
                await this.performConversion(true);
            });
        }

        ['#convert-from', '#convert-to', '#convert-amount'].forEach(id => {
            const el = $(id);
            if (el) {
                el.addEventListener('change', () => this.performConversion(true));
                el.addEventListener('input', () => this.performConversion(true));
            }
        });
    }

    static async performConversion(silent = false) {
        try {
            const fromCurrencyId = $('#convert-from').value;
            const toCurrencyId = $('#convert-to').value;
            const amountVal = $('#convert-amount').value;
            const amount = parseFloat(amountVal || '0');
            if (!fromCurrencyId || !toCurrencyId) {
                if (!silent) NotificationSystem.show('Выберите валюты для конвертации', 'error');
                return;
            }
            if (!amount || amount <= 0) {
                if (!silent) NotificationSystem.show('Введите положительное значение суммы', 'error');
                return;
            }
            const body = {
                from_currency_id: parseInt(fromCurrencyId, 10),
                to_currency_id: parseInt(toCurrencyId, 10),
                amount: amount
            };
            const result = await ApiClient.request('/exchange/convert-simple', {
                method: 'POST',
                body: JSON.stringify(body)
            });
            this.displayConversionResult(result, amount);
        } catch (error) {
            if (!silent) NotificationSystem.show(error.message, 'error');
        }
    }
    
    static displayConversionResult(result, originalAmount) {
        const container = $('#converter-result');
        const fromCurrency = currencies.find(c => c.id === result.from_currency_id);
        const toCurrency = currencies.find(c => c.id === result.to_currency_id);
        
        if (!fromCurrency || !toCurrency) {
            container.innerHTML = '<div class="error">Ошибка отображения результата</div>';
            return;
        }
        
        const symFrom = CurrencyUI.getSymbol(fromCurrency.code);
        const symTo = CurrencyUI.getSymbol(toCurrency.code);
        const fmt = (n, d=2) => new Intl.NumberFormat('ru-RU', { minimumFractionDigits: d, maximumFractionDigits: d }).format(parseFloat(n));
        container.innerHTML = `
            <div class="conversion-result">
                ${fmt(originalAmount)} ${symFrom} (${fromCurrency.code}) = ${fmt(result.converted_amount)} ${symTo} (${toCurrency.code})
            </div>
            <div class="conversion-rate">
                Курс: 1 ${symFrom} (${fromCurrency.code}) = ${fmt(result.exchange_rate, 6)} ${symTo} (${toCurrency.code})
            </div>
        `;
    }
    
    static initUIInteractions() {
        // Settings open
        const settingsBtn = document.getElementById('settings-btn');
        if (settingsBtn) {
            settingsBtn.addEventListener('click', () => FormHandlers.openSettingsModal());
        }
        const settingsForm = document.getElementById('settings-form');
        if (settingsForm) {
            settingsForm.addEventListener('submit', (e) => FormHandlers.handleSettingsSubmit(e));
        }

        const addTxBtn = $('#add-transaction-btn');
        if (addTxBtn) {
            addTxBtn.addEventListener('click', () => {
                const form = $('#tx-form');
                if (form) form.classList.toggle('hidden');
            });
        }

        const filterToggle = $('#filter-toggle');
        if (filterToggle) {
            filterToggle.addEventListener('click', () => {
                const panel = $('#filters-panel');
                if (panel) panel.classList.toggle('active');
            });
        }

        const quickAdd = $('#quick-add');
        if (quickAdd) {
            quickAdd.addEventListener('click', () => {
                UIManager.switchTab('transactions');
                setTimeout(() => {
                    const form = $('#tx-form');
                    if (form) form.classList.remove('hidden');
                }, 300);
            });
        }

        const exportBtn = $('#export-report');
        if (exportBtn) {
            exportBtn.addEventListener('click', () => {
                NotificationSystem.show('Функция экспорта в разработке', 'info');
            });
        }

        const analyticsPeriod = $('#analytics-period');
        if (analyticsPeriod) {
            analyticsPeriod.addEventListener('change', async () => {
                await DataManager.renderCharts();
            });
        }

        const monthlyYear = $('#monthly-year');
        if (monthlyYear) {
            monthlyYear.addEventListener('change', async () => {
                await DataManager.renderCharts();
            });
        }

        const exchangePeriod = $('#exchange-period');
        if (exchangePeriod) {
            exchangePeriod.addEventListener('change', async () => {
                await DataManager.renderExchangeRatesChart();
            });
        }

        // Filters: debounced apply on change/input
        const startEl = document.getElementById('period-start');
        const endEl = document.getElementById('period-end');
        const typeEls = Array.from(document.querySelectorAll("#filters-panel input[name='type']"));
        if (startEl && endEl) {
            const handler = () => { if (window.applyTransactionFiltersDebounced) window.applyTransactionFiltersDebounced(); };
            startEl.addEventListener('change', handler);
            endEl.addEventListener('change', handler);
        }
        if (typeEls.length) {
            const handler = () => { if (window.applyTransactionFiltersDebounced) window.applyTransactionFiltersDebounced(); };
            typeEls.forEach(el => el.addEventListener('change', handler));
        }
    }
    
    static initDateDefaults() {
        const today = new Date().toISOString().split('T')[0];
        const firstDayOfMonth = new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0];
        
        $('#tx-date').value = today;
        $('#period-start').value = firstDayOfMonth;
        $('#period-end').value = today;
    }
    
    static switchAuthTab(tabName) {
        $$('.auth-tabs .tab-btn').forEach(btn => {
            btn.classList.toggle('active', btn.getAttribute('data-tab') === tabName);
        });
        
        $$('.auth-forms-container .tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `${tabName}-tab`);
        });

        // Sync auth hero scenes
        const loginHero = document.getElementById('login-hero');
        const registerHero = document.getElementById('register-hero');
        if (loginHero && registerHero) {
            const isLogin = tabName === 'login';
            loginHero.classList.toggle('active', isLogin);
            registerHero.classList.toggle('active', !isLogin);
            loginHero.setAttribute('aria-hidden', String(!isLogin));
            registerHero.setAttribute('aria-hidden', String(isLogin));
        }
    }
}

window.switchTab = (tabName) => UIManager.switchTab(tabName);

window.updateExchangeRates = async () => {
    try {
        await ApiClient.request('/exchange/rates/update', { method: 'POST' });
        NotificationSystem.show('Курсы валют обновлены!', 'success');
        await DataManager.loadExchangeRates();
        await DataManager.loadExchangeRatesOverview();
        await DataManager.loadQuickRates();
    } catch (error) {
        NotificationSystem.show('Ошибка обновления курсов', 'error');
    }
};

window.showCreateAccountModal = () => {
    const modal = $('#create-account-modal');
    if (modal) {
        modal.classList.remove('hidden');
        modal.classList.add('active');
    } else {
        console.error('Modal not found!');
    }
};

window.closeModal = (modalId) => {
    const modal = $(`#${modalId}`);
    if (modal) {
        modal.classList.remove('active');
        modal.classList.add('hidden');
    }
};

window.createAccount = async () => {
    const name = $('#account-name').value;
    const currencyId = $('#account-currency').value;
    const initialBalance = $('#initial-balance').value || 0;
    
    if (!name || !currencyId) {
        NotificationSystem.show('Заполните все обязательные поля', 'error');
        return;
    }

    try {
        const body = {
            name: name,
            currency_id: parseInt(currencyId, 10),
            initial_balance: parseFloat(initialBalance)
        };
        
        await ApiClient.request('/accounts', {
            method: 'POST',
            body: JSON.stringify(body)
        });

        NotificationSystem.show('Счет успешно создан!', 'success');
        closeModal('create-account-modal');
        
        $('#account-name').value = '';
        $('#account-currency').value = '';
        $('#initial-balance').value = '0';

        await DataManager.loadAccounts();
    } catch (error) {
        console.error('Error creating account:', error);
        NotificationSystem.show('Ошибка при создании счета: ' + error.message, 'error');
    }
};

window.setDefaultAccount = async (accountId) => {
    try {
        await ApiClient.request(`/accounts/${accountId}/default`, {
            method: 'PUT'
        });
        NotificationSystem.show('Счет установлен как основной!', 'success');
        await DataManager.loadAccounts();
    } catch (error) {
        NotificationSystem.show('Ошибка установки счета по умолчанию', 'error');
    }
};

window.showAccountDetails = async (accountId) => {
    try {
        const account = await ApiClient.request(`/accounts/${accountId}`);
        const currency = currencies.find(c => c.id === account.currency_id);
        
        const content = `
            <div class="account-details-content">
                <div class="detail-item">
                    <span class="detail-label">Название счета:</span>
                    <span class="detail-value">${account.name}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Валюта:</span>
                    <span class="detail-value">${currency ? `${currency.name} (${currency.code})` : 'Неизвестно'}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Баланс:</span>
                    <span class="detail-value">${parseFloat(account.balance).toFixed(2)} ${currency ? currency.symbol : ''}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Основной счет:</span>
                    <span class="detail-value">${account.is_default ? 'Да' : 'Нет'}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Создан:</span>
                    <span class="detail-value">${new Date(account.created_at).toLocaleDateString('ru-RU')}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Обновлен:</span>
                    <span class="detail-value">${new Date(account.updated_at).toLocaleDateString('ru-RU')}</span>
                </div>
            </div>
        `;
        
        $('#account-details-content').innerHTML = content;
        $('#account-details-modal').classList.add('active');
    } catch (error) {
        NotificationSystem.show('Ошибка загрузки деталей счета', 'error');
    }
};

window.deleteCategory = async (categoryId) => {
    if (!confirm('Вы уверены, что хотите удалить эту категорию?')) {
        return;
    }
    
    try {
        await ApiClient.request(`/categories/${categoryId}`, {
            method: 'DELETE'
        });
        
        NotificationSystem.show('Категория успешно удалена!', 'success');
        await DataManager.loadCategories();
    } catch (error) {
        NotificationSystem.show('Ошибка удаления категории', 'error');
    }
};

window.applyTransactionFilters = () => {
    const startDate = $('#period-start').value;
    const endDate = $('#period-end').value;
    const typeCheckboxes = $$('input[name="type"]:checked');
    const types = Array.from(typeCheckboxes).map(cb => cb.value);
    
    const filters = {};
    if (startDate && endDate) {
        filters.startDate = startDate;
        filters.endDate = endDate;
    }
    if (types.length > 0) {
        filters.types = types;
    }
    
    DataManager.loadTransactions(filters);
    $('#filters-panel').classList.remove('active');
};

// Debounced filters application to reduce API chatter during input changes
window.applyTransactionFiltersDebounced = debounce(() => {
    try { window.applyTransactionFilters(); } catch (_) {}
}, 300);

async function initApp() {
    try {
        const h = await ApiClient.health();
        if (h && h.status === 'ok') {
            // Backend health OK
        } else {
            console.warn('Backend health check failed');
        }
    } catch (_) {}
    FormHandlers.init();
    
    if (token) {
        // 1) Оптимистично показать UI, чтобы не выбрасывало на форму при перезагрузке
        let cachedUser = null;
        try { cachedUser = JSON.parse(localStorage.getItem('user') || 'null'); } catch(_) {}
        try { UIManager.setAuthenticated(cachedUser || (currentUser || { email: '' })); } catch(_) {}
        // 2) Обновить профиль с сервера и перезагрузить данные
        try {
            const profile = await ApiClient.request('/user/profile');
            try { localStorage.setItem('user', JSON.stringify(profile || {})); } catch(_) {}
            UIManager.setAuthenticated(profile);
            await DataManager.loadInitialData();
            try { UIManager.refreshCurrencyDisplay(); } catch(_) {}
        } catch (error) {
            console.error('Auth check failed:', error);
            const msg = (error && error.message) ? String(error.message) : '';
            const unauthorized = /unauthorized|401/i.test(msg);
            if (unauthorized) {
                UIManager.setAnonymous();
                NotificationSystem.show('Сессия истекла. Пожалуйста, войдите снова.', 'error');
            } else {
                // Остаёмся в режиме сессии и подгружаем данные по возможности
                try { await DataManager.loadInitialData(); } catch(_) {}
                NotificationSystem.show('Нет связи с сервером. Работаем в офлайн-режиме.', 'info');
            }
        }
    } else {
        UIManager.setAnonymous();
        try {
            await DataManager.loadCurrencies();
            await DataManager.loadExchangeRatesOverview();
            await DataManager.loadQuickRates();
            await DataManager.loadExchangeRates();
        } catch (e) {
            console.warn('Public data load failed:', e);
        }
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initApp);
} else {
    initApp();
}

window.debugPing = async () => {
    const h = await ApiClient.health();
    console.log('Health:', h);
    return h;
};