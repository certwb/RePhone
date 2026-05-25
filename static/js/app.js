document.addEventListener('DOMContentLoaded', () => {
    // --- Theme Toggling ---
    const themeToggle = document.getElementById('themeToggle');
    const html = document.documentElement;
    
    // Check local storage or system preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
        html.setAttribute('data-theme', savedTheme);
    } else {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        html.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
    }

    themeToggle.addEventListener('click', () => {
        const currentTheme = html.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        html.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });

    // --- State ---
    let limit = 8;
    let lastId = 0;
    let hasMore = true;
    let isFetching = false;

    // --- DOM Elements ---
    const phonesGrid = document.getElementById('phonesGrid');
    const spinner = document.getElementById('spinner');
    const endMessage = document.getElementById('endOfFeed');
    
    // Filters
    const brandFilter = document.getElementById('brandFilter');
    const cityFilter = document.getElementById('cityFilter');
    const minPrice = document.getElementById('minPrice');
    const maxPrice = document.getElementById('maxPrice');

    // --- Formatting ---
    const formatPrice = (price) => {
        return new Intl.NumberFormat('ru-KZ', { style: 'currency', currency: 'KZT', maximumFractionDigits: 0 }).format(price);
    };

    // --- Fetch API ---
    const fetchPhones = async (reset = false) => {
        if (isFetching || (!hasMore && !reset)) return;
        isFetching = true;
        
        if (reset) {
            lastId = 0;
            hasMore = true;
            phonesGrid.innerHTML = Array(8).fill(`
                <div class="skeleton">
                    <div class="skeleton-img"></div>
                    <div class="skeleton-text" style="margin-top: 20px;"></div>
                    <div class="skeleton-text short"></div>
                    <div class="skeleton-text" style="margin-top: 30px;"></div>
                </div>
            `).join('');
            endMessage.classList.add('hidden');
        }

        spinner.classList.remove('hidden');

        try {
            const urlParams = new URLSearchParams(window.location.search);
            const brandId = urlParams.get('brand_id') || brandFilter?.value || '';
            const minP = urlParams.get('min_price') || minPrice.value || '';
            const maxP = urlParams.get('max_price') || maxPrice.value || '';
            const q = urlParams.get('q') || document.getElementById('searchInput')?.value || '';
            const condition = urlParams.get('condition') || document.getElementById('conditionFilter')?.value || '';
            const cityId = urlParams.get('city_id') || cityFilter?.value || '';

            const params = new URLSearchParams({
                limit,
                last_id: lastId,
                brand_id: brandId,
                city_id: cityId,
                min_price: minP,
                max_price: maxP,
                q: q,
                condition: condition
            });

            const res = await fetch(`/api/phones?${params.toString()}`);
            if (!res.ok) throw new Error('Failed to fetch');
            
            const data = await res.json() || [];
            
            if (reset) {
                phonesGrid.innerHTML = '';
            }
            
            if (data.length < limit) {
                hasMore = false;
                endMessage.classList.remove('hidden');
            }
            
            if (data.length > 0) {
                lastId = data[data.length - 1].id;
            }

            data.forEach(phone => {
                const imgUrl = (phone.images && phone.images.length > 0) ? phone.images[0].image_url : 'https://via.placeholder.com/300x400?text=No+Image';
                const card = document.createElement('article');
                card.className = 'phone-card glass-panel';
                card.innerHTML = `
                    <div class="phone-img-wrapper">
                        <img src="${imgUrl}" alt="${phone.title}" class="phone-img" loading="lazy">
                        <button class="favorite-btn ${phone.is_favorited ? 'active' : ''}" data-id="${phone.id}" aria-label="В избранное">
                            <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" stroke-width="2" fill="${phone.is_favorited ? 'currentColor' : 'none'}"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg>
                        </button>
                    </div>
                    <div class="phone-details">
                        <div style="display: flex; justify-content: space-between;">
                            <div class="phone-brand">${phone.brand_name || 'Бренд'} ${phone.city_name ? `• ${phone.city_name}` : ''}</div>
                            <div style="font-size: 0.75rem; color: var(--text-muted);"><svg viewBox="0 0 24 24" width="12" height="12" stroke="currentColor" stroke-width="2" fill="none" style="vertical-align: middle;"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg> ${phone.views_count || 0}</div>
                        </div>
                        <h3 class="phone-title">${phone.title}</h3>
                        <div class="phone-specs">
                            <span>${phone.storage} ГБ</span>
                            <span>АКБ: ${phone.battery_health}%</span>
                            <span>${phone.condition}</span>
                        </div>
                        <div class="phone-price">${formatPrice(phone.price)}</div>
                        <div style="display: flex; gap: 10px; margin-top: 15px;">
                            <button class="btn btn-primary message-seller-btn" data-userid="${phone.user_id}" data-phoneid="${phone.id}" data-phonetitle="${phone.title}" style="flex: 1; font-size: 0.9rem;">Написать</button>
                            <button class="btn btn-secondary show-phone-btn" data-id="${phone.id}" style="flex: 1; font-size: 0.9rem;">${phone.contact_phone || 'Номер'}</button>
                        </div>
                    </div>
                `;
                phonesGrid.appendChild(card);
            });
            
        } catch (error) {
            console.error('Error loading phones:', error);
        } finally {
            isFetching = false;
            spinner.classList.add('hidden');
        }
    };

    // --- Infinite Scroll (Intersection Observer) ---
    const observer = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting && hasMore && !isFetching) {
            fetchPhones();
        }
    }, { rootMargin: '100px' });
    
    observer.observe(spinner);

    // --- Favorites Logic (Event Delegation) ---
    phonesGrid.addEventListener('click', async (e) => {
        const btn = e.target.closest('.favorite-btn');
        if (!btn) return;
        
        e.stopPropagation();
        
        const isAuth = localStorage.getItem('isAuthenticated') === 'true';
        if (!isAuth) {
            authModalOverlay.classList.remove('hidden'); // Показать окно входа
            return;
        }

        const phoneId = btn.getAttribute('data-id');
        try {
            const res = await fetch('/api/favorites/toggle', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ phone_id: parseInt(phoneId) })
            });

            if (res.ok) {
                const data = await res.json();
                const svg = btn.querySelector('svg');
                if (data.status === 'added') {
                    btn.classList.add('active');
                    svg.setAttribute('fill', 'currentColor');
                } else {
                    btn.classList.remove('active');
                    svg.setAttribute('fill', 'none');
                }
            } else if (res.status === 401) {
                localStorage.removeItem('isAuthenticated');
                updateAuthUI();
                authModalOverlay.classList.remove('hidden');
            }
        } catch (error) {
            console.error('Error toggling favorite:', error);
        }
    });

    // --- Filter Handling ---
    const applyFiltersBtn = document.getElementById('applyFiltersBtn');
    const initFiltersFromURL = () => {
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('brand_id')) brandFilter.value = urlParams.get('brand_id');
        if (urlParams.has('min_price')) minPrice.value = urlParams.get('min_price');
        if (urlParams.has('max_price')) maxPrice.value = urlParams.get('max_price');
        const searchInput = document.getElementById('searchInput');
        if (searchInput && urlParams.has('q')) searchInput.value = urlParams.get('q');
        const conditionFilter = document.getElementById('conditionFilter');
        if (conditionFilter && urlParams.has('condition')) conditionFilter.value = urlParams.get('condition');
    };
    initFiltersFromURL();

    const applyFilters = () => {
        const urlParams = new URLSearchParams();
        if (brandFilter.value) urlParams.set('brand_id', brandFilter.value);
        if (minPrice.value) urlParams.set('min_price', minPrice.value);
        if (maxPrice.value) urlParams.set('max_price', maxPrice.value);
        const searchInput = document.getElementById('searchInput');
        if (searchInput && searchInput.value) urlParams.set('q', searchInput.value);
        const conditionFilter = document.getElementById('conditionFilter');
        if (conditionFilter && conditionFilter.value) urlParams.set('condition', conditionFilter.value);

        const newUrl = window.location.pathname + (urlParams.toString() ? '?' + urlParams.toString() : '');
        window.history.pushState({ path: newUrl }, '', newUrl);
        fetchPhones(true);
    };

    applyFiltersBtn.addEventListener('click', applyFilters);

    const brandChips = document.querySelectorAll('.brand-chip');
    brandChips.forEach(chip => {
        chip.addEventListener('click', (e) => {
            brandFilter.value = e.target.getAttribute('data-brand');
            applyFilters();
            // Scroll smoothly to filters
            document.querySelector('.filters-section').scrollIntoView({ behavior: 'smooth' });
        });
    });

    window.addEventListener('popstate', () => {
        initFiltersFromURL();
        fetchPhones(true);
    });

    // --- Modal Logic (MVP) ---
    const btnAddAd = document.getElementById('btnAddAd');
    const modalOverlay = document.getElementById('modalOverlay');
    const closeModal = document.getElementById('closeModal');
    const addPhoneForm = document.getElementById('addPhoneForm');
    const addPhoneMsg = document.getElementById('addPhoneMsg');
    
    btnAddAd.addEventListener('click', () => {
        addPhoneMsg.classList.add('hidden');
        addPhoneForm.reset();
        modalOverlay.classList.remove('hidden');
    });
    closeModal.addEventListener('click', () => modalOverlay.classList.add('hidden'));
    modalOverlay.addEventListener('click', (e) => {
        if (e.target === modalOverlay) modalOverlay.classList.add('hidden');
    });

    addPhoneForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        addPhoneMsg.classList.add('hidden');

        const submitBtn = addPhoneForm.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = 'Публикация...';

        const formData = new FormData();
        formData.append('title', document.getElementById('pTitle').value);
        formData.append('brand_id', document.getElementById('pBrand').value);
        formData.append('city_id', document.getElementById('pCity').value);
        formData.append('model', document.getElementById('pModel').value);
        formData.append('storage', document.getElementById('pStorage').value || '0');
        formData.append('battery_health', document.getElementById('pBattery').value || '0');
        formData.append('price', document.getElementById('pPrice').value);
        formData.append('condition', document.getElementById('pCondition').value);
        formData.append('description', document.getElementById('pDesc').value);
        formData.append('contact_phone', document.getElementById('pPhone').value);

        const files = document.getElementById('pImages').files;
        for (let i = 0; i < files.length; i++) {
            if (files[i].size > 2 * 1024 * 1024) {
                addPhoneMsg.textContent = 'Ошибка: Размер файла "' + files[i].name + '" превышает 2 МБ.';
                addPhoneMsg.style.color = '#ef4444';
                addPhoneMsg.classList.remove('hidden');
                return;
            }
            if (!files[i].type.startsWith('image/')) {
                addPhoneMsg.textContent = 'Ошибка: Файл "' + files[i].name + '" не является изображением.';
                addPhoneMsg.style.color = '#ef4444';
                addPhoneMsg.classList.remove('hidden');
                return;
            }
            formData.append('images', files[i]);
        }

        try {
            const res = await fetch('/api/phones', {
                method: 'POST',
                body: formData // Не указываем Content-Type, браузер сам выставит multipart/form-data и boundary
            });

            if (res.ok) {
                addPhoneMsg.textContent = 'Объявление успешно опубликовано!';
                addPhoneMsg.style.color = '#10b981'; // Green
                addPhoneMsg.classList.remove('hidden');
                setTimeout(() => {
                    modalOverlay.classList.add('hidden');
                    fetchPhones(true); // Перезагрузить ленту
                }, 1500);
            } else {
                const text = await res.text();
                addPhoneMsg.textContent = text || 'Ошибка при публикации';
                addPhoneMsg.style.color = '#ef4444'; // Red
                addPhoneMsg.classList.remove('hidden');
            }
        } catch (error) {
            addPhoneMsg.textContent = 'Сетевая ошибка';
            addPhoneMsg.style.color = '#ef4444';
            addPhoneMsg.classList.remove('hidden');
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Опубликовать';
        }
    });

    // --- Auth Logic ---
    const btnLogin = document.getElementById('btnLogin');
    const btnLogout = document.getElementById('btnLogout');
    const btnProfile = document.getElementById('btnProfile');
    const authModalOverlay = document.getElementById('authModalOverlay');
    const closeAuthModal = document.getElementById('closeAuthModal');
    const authForm = document.getElementById('authForm');
    const authSwitchLink = document.getElementById('authSwitchLink');
    const authTitle = document.getElementById('authTitle');
    const authSubmitBtn = document.getElementById('authSubmitBtn');
    const authSwitchText = document.getElementById('authSwitchText');
    const authError = document.getElementById('authError');
    const authEmail = document.getElementById('authEmail');
    const authPassword = document.getElementById('authPassword');
    
    let isLoginMode = true;

    // Check auth state from localStorage
    const updateAuthUI = () => {
        const isAuth = localStorage.getItem('isAuthenticated') === 'true';
        if (isAuth) {
            btnLogin.classList.add('hidden');
            btnLogout.classList.remove('hidden');
            btnAddAd.classList.remove('hidden');
            btnProfile.classList.remove('hidden');
        } else {
            btnLogin.classList.remove('hidden');
            btnLogout.classList.add('hidden');
            btnAddAd.classList.add('hidden');
            btnProfile.classList.add('hidden');
        }
    };
    updateAuthUI();

    btnLogin.addEventListener('click', () => authModalOverlay.classList.remove('hidden'));
    closeAuthModal.addEventListener('click', () => authModalOverlay.classList.add('hidden'));
    authModalOverlay.addEventListener('click', (e) => {
        if (e.target === authModalOverlay) authModalOverlay.classList.add('hidden');
    });

    authSwitchLink.addEventListener('click', (e) => {
        e.preventDefault();
        isLoginMode = !isLoginMode;
        authTitle.textContent = isLoginMode ? 'Вход' : 'Регистрация';
        authSubmitBtn.textContent = isLoginMode ? 'Войти' : 'Зарегистрироваться';
        authSwitchText.textContent = isLoginMode ? 'Нет аккаунта?' : 'Уже есть аккаунт?';
        authSwitchLink.textContent = isLoginMode ? 'Зарегистрироваться' : 'Войти';
        authError.classList.add('hidden');
    });

    authForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        authError.classList.add('hidden');
        const endpoint = isLoginMode ? '/api/auth/login' : '/api/auth/register';
        
        try {
            const res = await fetch(endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email: authEmail.value,
                    password: authPassword.value
                })
            });
            
            if (res.ok) {
                localStorage.setItem('isAuthenticated', 'true');
                updateAuthUI();
                authModalOverlay.classList.add('hidden');
                authEmail.value = '';
                authPassword.value = '';
                // Connect WebSocket now that user is logged in
                connectWebSocket();
                // Optional: fetchPhones(true) to refresh specific UI parts
            } else {
                const text = await res.text();
                authError.textContent = text || 'Произошла ошибка';
                authError.classList.remove('hidden');
            }
        } catch (error) {
            authError.textContent = 'Сетевая ошибка';
            authError.classList.remove('hidden');
        }
    });

    btnLogout.addEventListener('click', async () => {
        try {
            await fetch('/api/auth/logout', { method: 'POST' });
            localStorage.removeItem('isAuthenticated');
            updateAuthUI();
            if (ws) {
                if (wsReconnectTimeout) clearTimeout(wsReconnectTimeout);
                ws.onclose = null; // Prevent reconnect logic
                ws.close();
                ws = null;
            }
        } catch (error) {
            console.error(error);
        }
    });

    // --- Profile Logic ---
    const mainView = document.getElementById('mainView');
    const profileView = document.getElementById('profileView');
    const profileGrid = document.getElementById('profileGrid');
    
    // Empty states and tabs
    const emptyStateMyAds = document.getElementById('emptyStateMyAds');
    const emptyStateFavorites = document.getElementById('emptyStateFavorites');
    const emptyStateMessages = document.getElementById('emptyStateMessages');
    const settingsTabContent = document.getElementById('settingsTabContent');
    const chatTabContent = document.getElementById('chatTabContent');
    const chatList = document.getElementById('chatList');
    const chatMessages = document.getElementById('chatMessages');
    const chatHeader = document.getElementById('chatHeader');
    const chatActiveUser = document.getElementById('chatActiveUser');
    const chatActivePhone = document.getElementById('chatActivePhone');
    const chatForm = document.getElementById('chatForm');
    const chatInput = document.getElementById('chatInput');
    const tabBtns = document.querySelectorAll('.profile-tabs .tab-btn');
    const logo = document.querySelector('.logo');

    const transitionView = (callback) => {
        if (!document.startViewTransition) {
            callback();
            return;
        }
        document.startViewTransition(() => {
            callback();
        });
    };

    const showMainView = () => {
        transitionView(() => {
            profileView.classList.add('hidden');
            mainView.classList.remove('hidden');
            if (phonesGrid.children.length === 0) fetchPhones(true);
        });
    };

    document.getElementById('btnEmptyAddAd')?.addEventListener('click', () => document.getElementById('btnAddAd').click());
    document.getElementById('btnEmptyCatalog')?.addEventListener('click', showMainView);

    const hideAllProfileContent = () => {
        profileGrid.classList.add('hidden');
        emptyStateMyAds.classList.add('hidden');
        emptyStateFavorites.classList.add('hidden');
        emptyStateMessages.classList.add('hidden');
        settingsTabContent.classList.add('hidden');
        chatTabContent.classList.add('hidden');
    };

    const loadProfileData = async (type) => {
        hideAllProfileContent();
        profileGrid.innerHTML = Array(4).fill(`
            <div class="skeleton">
                <div class="skeleton-img"></div>
                <div class="skeleton-text" style="margin-top: 20px;"></div>
                <div class="skeleton-text short"></div>
                <div class="skeleton-text" style="margin-top: 30px;"></div>
            </div>
        `).join('');
        profileGrid.classList.remove('hidden');
        
        try {
            const endpoint = type === 'ads' ? '/api/user/phones' : '/api/user/favorites';
            const res = await fetch(endpoint);
            if (!res.ok) throw new Error('Failed to fetch');
            const data = await res.json();
            
            profileGrid.innerHTML = '';
            
            if (!data || data.length === 0) {
                profileGrid.classList.add('hidden');
                if (type === 'ads') emptyStateMyAds.classList.remove('hidden');
                if (type === 'favorites') emptyStateFavorites.classList.remove('hidden');
                return;
            }

            data.forEach(phone => {
                const card = document.createElement('div');
                card.className = 'phone-card glass-panel';
                
                const imgUrl = phone.images && phone.images.length > 0 ? phone.images[0].image_url : 'https://placehold.co/400x300?text=No+Image';
                
                let actionBtn = '';
                if (type === 'ads') {
                    actionBtn = `
                        <button class="delete-btn" data-id="${phone.id}" aria-label="Снять с публикации" style="position: absolute; top: 12px; right: 12px; background: rgba(239, 68, 68, 0.9); border: none; border-radius: 50%; width: 36px; height: 36px; color: white; cursor: pointer; z-index: 2;">
                            <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                        </button>
                    `;
                } else {
                    actionBtn = `
                        <button class="favorite-btn active" data-id="${phone.id}" aria-label="В избранное">
                            <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" stroke-width="2" fill="currentColor"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg>
                        </button>
                    `;
                }

                card.innerHTML = `
                    <div class="phone-img-wrapper">
                        <img src="${imgUrl}" alt="${phone.title}" class="phone-img" loading="lazy">
                        ${actionBtn}
                    </div>
                    <div class="phone-details">
                        <div style="display: flex; justify-content: space-between;">
                            <div class="phone-brand">${phone.brand_name || 'Бренд'} ${phone.city_name ? `• ${phone.city_name}` : ''}</div>
                            <div style="font-size: 0.75rem; color: var(--text-muted);"><svg viewBox="0 0 24 24" width="12" height="12" stroke="currentColor" stroke-width="2" fill="none" style="vertical-align: middle;"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg> ${phone.views_count || 0}</div>
                        </div>
                        <h3 class="phone-title">${phone.title}</h3>
                        <div class="phone-specs">
                            <span>${phone.storage} ГБ</span>
                            <span>АКБ: ${phone.battery_health}%</span>
                            <span>${phone.condition}</span>
                        </div>
                        <div class="phone-price">${formatPrice(phone.price)}</div>
                        ${type === 'ads' ? 
                            `<div style="display: flex; gap: 10px; margin-top: 15px;">
                                <button class="btn btn-primary full-width bump-btn" data-id="${phone.id}" style="flex: 1;">Поднять</button>
                                <button class="btn btn-secondary full-width show-phone-btn" data-id="${phone.id}" style="flex: 1;">Номер</button>
                             </div>
                             ${phone.promo_slug 
                                ? `<button class="btn full-width copy-promo-btn" data-slug="${phone.promo_slug}" style="margin-top: 10px; background: rgba(99, 102, 241, 0.1); color: #818cf8; border: 1px solid rgba(99, 102, 241, 0.3);">🔗 Скопировать ссылку</button>`
                                : `<button class="btn btn-primary full-width promo-btn" data-id="${phone.id}" style="margin-top: 10px; background: linear-gradient(135deg, #a855f7, #6366f1); border: none;">✨ Прокачать продажу</button>`
                             }` 
                            : 
                            `<button class="btn btn-secondary full-width show-phone-btn" data-id="${phone.id}" style="margin-top: 15px;">${phone.contact_phone || 'Показать номер'}</button>`
                        }
                    </div>
                `;
                profileGrid.appendChild(card);
            });
        } catch (error) {
            profileGrid.innerHTML = '<p>Ошибка загрузки</p>';
        }
    };

    profileGrid.addEventListener('click', async (e) => {
        const bumpBtn = e.target.closest('.bump-btn');
        if (bumpBtn) {
            const phoneId = bumpBtn.getAttribute('data-id');
            const originalText = bumpBtn.textContent;
            bumpBtn.textContent = '...';
            bumpBtn.disabled = true;

            try {
                const res = await fetch(`/api/phones/${phoneId}/bump`, { method: 'POST' });
                const data = await res.json();
                
                if (res.ok) {
                    alert(data.message || 'Успешно поднято');
                    loadProfileData('ads'); // Refresh list
                } else {
                    alert(data.message || 'Ошибка. Возможно, еще не прошло 24 часа.');
                }
            } catch (error) {
                alert('Ошибка сети');
            } finally {
                bumpBtn.textContent = originalText;
                bumpBtn.disabled = false;
            }
        }

        const promoBtn = e.target.closest('.promo-btn');
        if (promoBtn) {
            const phoneId = promoBtn.getAttribute('data-id');
            const originalText = promoBtn.textContent;
            promoBtn.textContent = '✨ Создаем магию...';
            promoBtn.disabled = true;

            try {
                const res = await fetch(`/api/phones/${phoneId}/promo`, { method: 'POST' });
                const data = await res.json();
                
                if (res.ok) {
                    alert(data.message);
                    loadProfileData('ads'); // Refresh list to show the new link button
                } else {
                    alert(data.message || 'Ошибка генерации');
                }
            } catch (error) {
                alert('Ошибка сети');
            } finally {
                promoBtn.textContent = originalText;
                promoBtn.disabled = false;
            }
        }

        const copyPromoBtn = e.target.closest('.copy-promo-btn');
        if (copyPromoBtn) {
            const slug = copyPromoBtn.getAttribute('data-slug');
            const url = window.location.origin + '/promo/' + slug;
            navigator.clipboard.writeText(url).then(() => {
                const originalText = copyPromoBtn.textContent;
                copyPromoBtn.textContent = '✅ Скопировано!';
                setTimeout(() => copyPromoBtn.textContent = originalText, 2000);
            }).catch(err => {
                alert('Не удалось скопировать ссылку. Ссылка: ' + url);
            });
        }
    });

    btnProfile.addEventListener('click', () => {
        mainView.classList.add('hidden');
        profileView.classList.remove('hidden');
        tabBtns.forEach(b => b.classList.remove('active'));
        document.querySelector('.tab-btn[data-tab="my-ads"]').classList.add('active');
        loadProfileData('ads');
    });

    logo.addEventListener('click', (e) => {
        e.preventDefault();
        showMainView();
    });

    let currentChatUserId = null;
    let currentChatPhoneId = null;

    const loadChats = async () => {
        hideAllProfileContent();
        chatTabContent.classList.remove('hidden');
        chatList.innerHTML = Array(5).fill(`
            <div class="chat-item" style="pointer-events: none;">
                <div class="skeleton-text" style="height: 15px; margin: 0 0 5px 0;"></div>
                <div class="skeleton-text short" style="height: 12px; margin: 0 0 5px 0;"></div>
                <div class="skeleton-text" style="height: 12px; margin: 0; width: 80%;"></div>
            </div>
        `).join('');
        try {
            const res = await fetch('/api/user/chats');
            if (!res.ok) throw new Error('Failed');
            const data = await res.json();
            
            chatList.innerHTML = '';
            if (!data || data.length === 0) {
                chatTabContent.classList.add('hidden');
                emptyStateMessages.classList.remove('hidden');
                return;
            }

            data.forEach(chat => {
                const item = document.createElement('div');
                item.className = 'chat-item';
                if (currentChatUserId === chat.other_user_id && currentChatPhoneId === chat.phone_id) {
                    item.classList.add('active');
                }
                const timeStr = new Date(chat.last_message_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
                const unreadBadge = chat.unread_count > 0 ? `<span style="background:var(--primary-color);color:white;border-radius:10px;padding:2px 6px;font-size:0.7rem;margin-left:5px">${chat.unread_count}</span>` : '';
                
                item.innerHTML = `
                    <div class="chat-item-header">
                        <h4>${chat.other_user_email.split('@')[0]} ${unreadBadge}</h4>
                        <span class="chat-item-time">${timeStr}</span>
                    </div>
                    <div class="chat-item-phone">${chat.phone_title}</div>
                    <div class="chat-item-last">${chat.last_message}</div>
                `;
                item.addEventListener('click', () => {
                    document.querySelectorAll('.chat-item').forEach(i => i.classList.remove('active'));
                    item.classList.add('active');
                    openChat(chat.other_user_id, chat.other_user_email, chat.phone_id, chat.phone_title);
                });
                chatList.appendChild(item);
            });
        } catch(e) {
            chatList.innerHTML = '<p>Ошибка</p>';
        }
    };

    const openChat = async (userId, userEmail, phoneId, phoneTitle) => {
        currentChatUserId = userId;
        currentChatPhoneId = phoneId;
        
        chatTabContent.classList.remove('hidden');
        emptyStateMessages.classList.add('hidden');

        chatHeader.classList.remove('hidden');
        chatForm.classList.remove('hidden');
        
        // Fetch user info to show their rating and name properly
        try {
            const uRes = await fetch(`/api/user/info?id=${userId}`);
            if (uRes.ok) {
                const uData = await uRes.json();
                chatActiveUser.innerHTML = `${uData.name || userEmail.split('@')[0]} <span style="font-size: 0.8rem; color: #f59e0b;">★ ${uData.average_rating.toFixed(1)} (${uData.review_count})</span>`;
            } else {
                chatActiveUser.textContent = userEmail.split('@')[0];
            }
        } catch(e) {
            chatActiveUser.textContent = userEmail.split('@')[0];
        }

        chatActivePhone.textContent = 'По объявлению: ' + phoneTitle;
        const btnLeaveReview = document.getElementById('btnLeaveReview');
        if (btnLeaveReview) {
            btnLeaveReview.style.display = 'block';
            btnLeaveReview.onclick = () => openReviewModal(userId);
        }
        
        chatMessages.innerHTML = Array(3).fill(`
            <div class="chat-bubble received skeleton" style="height: 40px; margin-bottom: 10px; width: 60%;"></div>
            <div class="chat-bubble sent skeleton" style="height: 40px; margin-bottom: 10px; width: 50%; align-self: flex-end;"></div>
        `).join('');
        
        try {
            const res = await fetch(`/api/user/messages?other_user_id=${userId}&phone_id=${phoneId}`);
            if (!res.ok) throw new Error('Failed');
            const data = await res.json();
            
            chatMessages.innerHTML = '';
            
            data.forEach(msg => {
                const isSent = msg.sender_id !== userId;
                const bubble = document.createElement('div');
                bubble.className = `chat-bubble ${isSent ? 'sent' : 'received'}`;
                const timeStr = new Date(msg.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
                bubble.innerHTML = `
                    <div>${msg.content}</div>
                    <div class="chat-bubble-time">${timeStr}</div>
                `;
                chatMessages.appendChild(bubble);
            });
            chatMessages.scrollTop = chatMessages.scrollHeight;
            
        } catch(e) {
            chatMessages.innerHTML = '<p>Ошибка загрузки</p>';
        }
    };

    chatForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const content = chatInput.value.trim();
        if (!content || !currentChatUserId || !currentChatPhoneId) return;
        
        chatInput.disabled = true;
        try {
            const res = await fetch('/api/user/messages/send', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    receiver_id: currentChatUserId,
                    phone_id: currentChatPhoneId,
                    content: content
                })
            });
            if (res.ok) {
                chatInput.value = '';
                // Reload chat
                openChat(currentChatUserId, chatActiveUser.textContent, currentChatPhoneId, chatActivePhone.textContent.replace('По объявлению: ', ''));
                // Reload list to update last message
                // (Implementation could be optimized, but this works)
            }
        } catch(e) {}
        chatInput.disabled = false;
        chatInput.focus();
    });

    tabBtns.forEach(btn => {
        btn.addEventListener('click', (e) => {
            transitionView(() => {
                tabBtns.forEach(b => b.classList.remove('active'));
                const targetBtn = e.target;
                targetBtn.classList.add('active');
                
                const tab = targetBtn.getAttribute('data-tab');
                if (tab === 'my-ads') loadProfileData('ads');
                else if (tab === 'favorites') loadProfileData('favorites');
                else if (tab === 'messages') loadChats();
                else if (tab === 'settings') {
                    hideAllProfileContent();
                    settingsTabContent.classList.remove('hidden');
                }
            });
        });
    });


    // --- Profile Editing & Settings Logic ---
    const profileName = document.getElementById('profileName');
    const profileAvatarDisplay = document.getElementById('profileAvatarDisplay');
    const profileAvatarFallback = document.getElementById('profileAvatarFallback');
    
    const loadUserProfile = async () => {
        try {
            const res = await fetch('/api/user/profile');
            if (res.ok) {
                const data = await res.json();
                profileName.textContent = data.name || 'Пользователь';
                window.currentUserCityId = data.city_id;
                const createdDate = new Date(data.created_at).toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
                document.querySelector('.user-meta').innerHTML = `На RePhone с ${createdDate} • <span>★ 5.0 (Отличный продавец)</span>`;
                
                if (data.avatar_url) {
                    profileAvatarDisplay.style.backgroundImage = `url(${data.avatar_url})`;
                    profileAvatarFallback.style.display = 'none';
                } else {
                    profileAvatarDisplay.style.backgroundImage = 'none';
                    profileAvatarFallback.style.display = 'block';
                }
            }
        } catch (e) {}
    };

    document.getElementById('btnProfile').addEventListener('click', () => {
        transitionView(() => {
            mainView.classList.add('hidden');
            profileView.classList.remove('hidden');
            loadUserProfile();
            document.querySelector('.tab-btn[data-tab="my-ads"]').click();
        });
    });

    // Edit Profile Modal
    const editProfileModal = document.getElementById('editProfileModal');
    const btnEditProfile = document.getElementById('btnEditProfile');
    const closeEditProfileModal = document.getElementById('closeEditProfileModal');
    const editProfileForm = document.getElementById('editProfileForm');
    const editProfileMsg = document.getElementById('editProfileMsg');

    btnEditProfile.addEventListener('click', () => {
        document.getElementById('editProfileName').value = profileName.textContent !== 'Пользователь' ? profileName.textContent : '';
        if (window.currentUserCityId) {
            document.getElementById('editProfileCity').value = window.currentUserCityId;
        }
        editProfileMsg.classList.add('hidden');
        editProfileModal.classList.remove('hidden');
    });

    closeEditProfileModal.addEventListener('click', () => editProfileModal.classList.add('hidden'));

    editProfileForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData();
        formData.append('name', document.getElementById('editProfileName').value);
        formData.append('city_id', document.getElementById('editProfileCity').value);
        const avatarFile = document.getElementById('editProfileAvatar').files[0];
        if (avatarFile) {
            if (avatarFile.size > 2 * 1024 * 1024) {
                editProfileMsg.textContent = 'Аватарка слишком большая (макс 2 МБ)';
                editProfileMsg.style.color = '#ef4444';
                editProfileMsg.classList.remove('hidden');
                return;
            }
            formData.append('avatar', avatarFile);
        }

        try {
            const res = await fetch('/api/user/profile', { method: 'PUT', body: formData });
            if (res.ok) {
                editProfileMsg.textContent = 'Профиль обновлен!';
                editProfileMsg.style.color = '#10b981';
                editProfileMsg.classList.remove('hidden');
                loadUserProfile(); // refresh UI
                setTimeout(() => editProfileModal.classList.add('hidden'), 1000);
            } else {
                editProfileMsg.textContent = 'Ошибка обновления';
                editProfileMsg.style.color = '#ef4444';
                editProfileMsg.classList.remove('hidden');
            }
        } catch(e) {
            editProfileMsg.textContent = 'Сетевая ошибка';
            editProfileMsg.style.color = '#ef4444';
            editProfileMsg.classList.remove('hidden');
        }
    });

    // Change Password
    const changePasswordForm = document.getElementById('changePasswordForm');
    const passwordMsg = document.getElementById('passwordMsg');
    changePasswordForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const oldPassword = document.getElementById('oldPassword').value;
        const newPassword = document.getElementById('newPassword').value;

        try {
            const res = await fetch('/api/user/password', {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({old_password: oldPassword, new_password: newPassword})
            });
            if (res.ok) {
                passwordMsg.textContent = 'Пароль успешно изменен';
                passwordMsg.style.color = '#10b981';
                passwordMsg.classList.remove('hidden');
                changePasswordForm.reset();
            } else {
                passwordMsg.textContent = 'Неверный старый пароль или ошибка';
                passwordMsg.style.color = '#ef4444';
                passwordMsg.classList.remove('hidden');
            }
        } catch(e) {}
    });

    // Logout All
    const btnLogoutAll = document.getElementById('btnLogoutAll');
    const logoutAllMsg = document.getElementById('logoutAllMsg');
    btnLogoutAll.addEventListener('click', async () => {
        if (!confirm('Вы уверены, что хотите выйти со всех устройств?')) return;
        try {
            const res = await fetch('/api/user/logout-all', { method: 'POST' });
            if (res.ok) {
                logoutAllMsg.textContent = 'Сессии завершены. Выход...';
                logoutAllMsg.style.color = '#10b981';
                logoutAllMsg.classList.remove('hidden');
                setTimeout(() => {
                    localStorage.removeItem('isAuthenticated');
                    window.location.reload();
                }, 1500);
            }
        } catch(e) {}
    });

    profileGrid.addEventListener('click', async (e) => {
        const delBtn = e.target.closest('.delete-btn');
        if (delBtn) {
            e.stopPropagation();
            if (confirm('Снять объявление с публикации?')) {
                const phoneId = delBtn.getAttribute('data-id');
                try {
                    const res = await fetch('/api/phones/' + phoneId, { method: 'DELETE' });
                    if (res.ok) loadProfileData('ads');
                } catch (error) { console.error(error); }
            }
            return;
        }

        const favBtn = e.target.closest('.favorite-btn');
        if (favBtn) {
            e.stopPropagation();
            const phoneId = favBtn.getAttribute('data-id');
            try {
                const res = await fetch('/api/favorites/toggle', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ phone_id: parseInt(phoneId) })
                });
                if (res.ok) loadProfileData('favorites');
            } catch (error) { console.error(error); }
            return;
        }

        const showPhoneBtn = e.target.closest('.show-phone-btn');
        if (showPhoneBtn) {
            e.stopPropagation();
            if (showPhoneBtn.textContent.includes('***')) {
                const phoneId = showPhoneBtn.getAttribute('data-id');
                showPhoneBtn.textContent = 'Загрузка...';
                try {
                    const res = await fetch('/api/phones/' + phoneId + '/phone');
                    if (res.ok) {
                        const data = await res.json();
                        showPhoneBtn.textContent = data.contact_phone;
                        showPhoneBtn.classList.replace('btn-secondary', 'btn-primary');
                        showPhoneBtn.style.background = 'var(--bg-color)';
                        showPhoneBtn.style.color = 'var(--text-primary)';
                        showPhoneBtn.style.border = '1px solid var(--primary-color)';
                    }
                } catch (error) { showPhoneBtn.textContent = 'Ошибка'; }
            }
            return;
        }
    });

    phonesGrid.addEventListener('click', async (e) => {
        const messageBtn = e.target.closest('.message-seller-btn');
        if (messageBtn) {
            e.stopPropagation();
            const isAuth = localStorage.getItem('isAuthenticated') === 'true';
            if (!isAuth) {
                authModalOverlay.classList.remove('hidden');
                return;
            }
            
            const sellerId = parseInt(messageBtn.getAttribute('data-userid'));
            const phoneId = parseInt(messageBtn.getAttribute('data-phoneid'));
            const phoneTitle = messageBtn.getAttribute('data-phonetitle');
            
            // Открываем профиль, вкладку сообщений и сразу этот чат
            mainView.classList.add('hidden');
            profileView.classList.remove('hidden');
            tabBtns.forEach(b => b.classList.remove('active'));
            document.querySelector('.tab-btn[data-tab="messages"]').classList.add('active');
            
            loadChats().then(() => {
                openChat(sellerId, `Продавец (ID: ${sellerId})`, phoneId, phoneTitle);
            });
            return;
        }

        const showPhoneBtn = e.target.closest('.show-phone-btn');
        if (showPhoneBtn) {
            e.stopPropagation();
            if (showPhoneBtn.textContent.includes('***') || showPhoneBtn.textContent === 'Показать номер') {
                const phoneId = showPhoneBtn.getAttribute('data-id');
                showPhoneBtn.textContent = 'Загрузка...';
                try {
                    const res = await fetch('/api/phones/' + phoneId + '/phone');
                    if (res.ok) {
                        const data = await res.json();
                        showPhoneBtn.textContent = data.contact_phone;
                        showPhoneBtn.classList.replace('btn-secondary', 'btn-primary');
                        showPhoneBtn.style.background = 'var(--bg-color)';
                        showPhoneBtn.style.color = 'var(--text-primary)';
                        showPhoneBtn.style.border = '1px solid var(--primary-color)';
                    }
                } catch (error) { showPhoneBtn.textContent = 'Ошибка'; }
            }
            return;
        }
    });

    // --- Infinite Scroll (IntersectionObserver) ---
    const observerOptions = {
        root: null,
        rootMargin: '100px',
        threshold: 0.1
    };

    const scrollObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting && hasMore && !isFetching) {
                fetchPhones();
            }
        });
    }, observerOptions);

    const scrollSentinel = document.getElementById('scrollSentinel');
    if (scrollSentinel) {
        scrollObserver.observe(scrollSentinel);
    }

    // --- Cities API ---
    const loadCities = async () => {
        try {
            const res = await fetch('/api/cities');
            if (res.ok) {
                const cities = await res.json();
                const pCity = document.getElementById('pCity');
                const editCity = document.getElementById('editProfileCity');
                
                let optionsHtml = '';
                cities.forEach(c => {
                    optionsHtml += `<option value="${c.id}">${c.name}</option>`;
                });
                
                if (cityFilter) cityFilter.innerHTML += optionsHtml;
                if (pCity) pCity.innerHTML += optionsHtml;
                if (editCity) editCity.innerHTML += optionsHtml;
            }
        } catch(e) {}
    };
    
    cityFilter?.addEventListener('change', () => fetchPhones(true));
    loadCities();

    // --- WebSocket Logic ---
    let ws = null;
    let wsReconnectTimeout = null;

    const connectWebSocket = () => {
        // Only connect if user is logged in (we can check if btnProfile is visible)
        if (btnProfile.classList.contains('hidden')) return;

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(`${protocol}//${window.location.host}/api/ws`);

        ws.onopen = () => {
            console.log('WebSocket connected');
            if (wsReconnectTimeout) clearTimeout(wsReconnectTimeout);
        };

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                if (msg.type === 'new_message') {
                    handleNewMessage(msg.payload);
                }
            } catch(e) {
                console.error('Error parsing WS message', e);
            }
        };

        ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting in 5s...');
            wsReconnectTimeout = setTimeout(connectWebSocket, 5000);
        };
    };

    const handleNewMessage = (payload) => {
        // If we are currently chatting with this user and phone:
        if (currentChatUserId === payload.sender_id && currentChatPhoneId === payload.phone_id) {
            const bubble = document.createElement('div');
            bubble.className = 'chat-bubble received';
            const timeStr = new Date(payload.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
            bubble.innerHTML = `
                <div>${payload.content}</div>
                <div class="chat-bubble-time">${timeStr}</div>
            `;
            chatMessages.appendChild(bubble);
            chatMessages.scrollTop = chatMessages.scrollHeight;
        } else {
            // Show a tiny notification dot or refresh chats list if chats tab is open
            if (!chatTabContent.classList.contains('hidden')) {
                loadChats();
            }
        }
    };

    // Try connecting on load
    connectWebSocket();

    // --- Reviews Logic ---
    let targetReviewUserId = null;
    const reviewModalOverlay = document.getElementById('reviewModalOverlay');
    const closeReviewModal = document.getElementById('closeReviewModal');
    const reviewForm = document.getElementById('reviewForm');
    const reviewMsg = document.getElementById('reviewMsg');
    const ratingSpans = document.querySelectorAll('.rating-selector span');
    const reviewRatingInput = document.getElementById('reviewRating');

    window.openReviewModal = (userId) => {
        targetReviewUserId = userId;
        reviewModalOverlay.classList.remove('hidden');
        reviewForm.reset();
        reviewRatingInput.value = '0';
        ratingSpans.forEach(s => s.classList.remove('selected', 'hovered'));
        reviewMsg.classList.add('hidden');
    };

    closeReviewModal?.addEventListener('click', () => reviewModalOverlay.classList.add('hidden'));
    reviewModalOverlay?.addEventListener('click', (e) => {
        if (e.target === reviewModalOverlay) reviewModalOverlay.classList.add('hidden');
    });

    // Rating star interactions
    ratingSpans.forEach(span => {
        span.addEventListener('mouseover', (e) => {
            const val = parseInt(e.target.getAttribute('data-value'));
            ratingSpans.forEach(s => {
                if (parseInt(s.getAttribute('data-value')) <= val) {
                    s.classList.add('hovered');
                } else {
                    s.classList.remove('hovered');
                }
            });
        });
        
        span.addEventListener('mouseout', () => {
            ratingSpans.forEach(s => s.classList.remove('hovered'));
        });

        span.addEventListener('click', (e) => {
            const val = parseInt(e.target.getAttribute('data-value'));
            reviewRatingInput.value = val;
            ratingSpans.forEach(s => {
                if (parseInt(s.getAttribute('data-value')) <= val) {
                    s.classList.add('selected');
                } else {
                    s.classList.remove('selected');
                }
            });
        });
    });

    reviewForm?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const rating = parseInt(reviewRatingInput.value);
        if (!rating) {
            reviewMsg.textContent = 'Пожалуйста, выберите оценку';
            reviewMsg.style.color = '#ef4444';
            reviewMsg.classList.remove('hidden');
            return;
        }

        const comment = document.getElementById('reviewComment').value.trim();

        try {
            const res = await fetch('/api/reviews', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    seller_id: parseInt(targetReviewUserId),
                    rating: rating,
                    comment: comment
                })
            });

            if (res.ok) {
                reviewMsg.textContent = 'Отзыв успешно отправлен!';
                reviewMsg.style.color = '#10b981';
                reviewMsg.classList.remove('hidden');
                setTimeout(() => {
                    reviewModalOverlay.classList.add('hidden');
                    // Refresh chat info
                    if (currentChatUserId === targetReviewUserId) {
                        // Re-fetch info to update rating
                        fetch(`/api/user/info?id=${targetReviewUserId}`).then(r => r.json()).then(uData => {
                            chatActiveUser.innerHTML = `${uData.name || 'Пользователь'} <span style="font-size: 0.8rem; color: #f59e0b;">★ ${uData.average_rating.toFixed(1)} (${uData.review_count})</span>`;
                        });
                    }
                }, 1500);
            } else {
                const txt = await res.text();
                reviewMsg.textContent = txt || 'Ошибка отправки отзыва';
                reviewMsg.style.color = '#ef4444';
                reviewMsg.classList.remove('hidden');
            }
        } catch (error) {
            reviewMsg.textContent = 'Сетевая ошибка';
            reviewMsg.style.color = '#ef4444';
            reviewMsg.classList.remove('hidden');
        }
    });

});
