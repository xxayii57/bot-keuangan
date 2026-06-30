/**
 * Undangan Intim - SaaS Wedding Invitation Builder
 * app.js - Client-Side Engine & Bridge Module
 */

// Global State
const appState = {
    isLoggedIn: false,
    userRole: 'client', // 'client' or 'reseller'
    clientSlug: 'aria-bima',
    invitationData: {
        brideName: "Aria",
        groomName: "Bima",
        brideFullName: "Aria Lestari, S.Kom",
        groomFullName: "Bima Adi Wijaya, S.T",
        brideParents: "Putri bungsu dari Bapak Hartono & Ibu Sri",
        groomParents: "Putra sulung dari Bapak Bambang & Ibu Retno",
        eventDateISO: "2026-08-18",
        akadTime: "08:00 - 10:00 WIB",
        resepsiTime: "11:00 - 14:00 WIB",
        venueName: "Pendopo Agung Terracotta",
        venueAddress: "Jl. Lavender No. 24, Kotagede, Yogyakarta",
        mapsUrl: "https://maps.google.com/?q=-7.820251,110.398634",
        musicUrl: "https://assets.mixkit.co/music/preview/mixkit-beautiful-dream-acoustic-guitar-and-piano-1002.mp3",
        videoUrl: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
        bankName: "BCA",
        bankNumber: "8021984221",
        bankHolder: "Bima Adi Wijaya",
        walletName: "Gopay",
        walletNumber: "081234567890",
        showVideo: true,
        showGallery: true,
        enableRsvp: true,
        gallery: [
            "https://images.unsplash.com/photo-1519741497674-611481863552?q=80&w=300&auto=format&fit=crop",
            "https://images.unsplash.com/photo-1583939003579-730e3918a45a?q=80&w=300&auto=format&fit=crop",
            "https://images.unsplash.com/photo-1511285560929-80b456fea0bc?q=80&w=300&auto=format&fit=crop"
        ]
    },
    gifts: [
        { id: 1, name: "Blender Philips HR-2115", bookedBy: "Siti Rahma", status: "Booked" },
        { id: 2, name: "Microwave Sharp R-728", bookedBy: "", status: "Available" },
        { id: 3, name: "Set Piring Keramik Bohemian (6 pcs)", bookedBy: "", status: "Available" }
    ],
    resellerData: {
        affiliateId: "IND-88219",
        commission: 0,
        points: 0,
        clients: [
            { slug: "aria-bima", couple: "Aria & Bima", package: "Platinum", status: "Active" },
            { slug: "clara-dani", couple: "Clara & Dani", package: "Gold", status: "Active" },
            { slug: "elisa-fajar", couple: "Elisa & Fajar", package: "Silver", status: "Draft" }
        ]
    },
    contacts: [],
    waMessageTemplate: "Halo {nama},\n\nKabar bahagia! Kami mengundang Anda untuk hadir di momen spesial pernikahan kami.\n\nDetail Acara:\n🤵 Bima & Aria 👰\n📅 Tanggal: {tanggal}\n📍 Tempat: {tempat}\n\nBuka link undangan digital kami untuk RSVP dan melihat galeri foto:\n{link}\n\nSuatu kehormatan bagi kami bila Anda berkenan hadir.\nTerima kasih 🌸"
};

// Check if running inside Android Native Webview
const isAndroid = typeof AndroidBridge !== 'undefined';

// Core App Initializer
document.addEventListener("DOMContentLoaded", () => {
    initNavigation();
    initLogin();
    initAccordions();
    initForms();
    initMediaManager();
    initContactsShare();
    initGiftRegistry();
    initKadoManager();
    initResellerPanel();
    loadThemesDynamic();
    
    // Auto-scroll inputs above soft keyboard on focus (Optimized behavior: 'auto' to avoid lag)
    setTimeout(() => {
        document.querySelectorAll('input, textarea').forEach(input => {
            input.addEventListener('focus', (e) => {
                setTimeout(() => {
                    e.target.scrollIntoView({ behavior: 'auto', block: 'center' });
                }, 100);
            });
        });
    }, 1000);
    
    // Draw initial chart on login (or when state becomes active)
    drawAnalyticsChart();
    
    // Check saved local sessions for auto-login
    setTimeout(checkAutoLogin, 500);
});

// Toast Helper
function showToast(message, type = "success") {
    const container = document.getElementById("toast-container");
    const toast = document.createElement("div");
    toast.className = `p-4 rounded-2xl shadow-lg border text-xs font-semibold flex items-center space-x-2.5 transition-all duration-300 opacity-0 translate-y-2 pointer-events-auto bg-alabaster ${
        type === "success" ? "text-sage border-sage/20 shadow-sage/5" : "text-terracotta border-terracotta/20 shadow-terracotta/5"
    }`;
    
    const icon = type === "success" ? "fa-solid fa-circle-check" : "fa-solid fa-triangle-exclamation";
    toast.innerHTML = `<i class="${icon} text-base"></i><span>${message}</span>`;
    
    container.appendChild(toast);
    
    // Animate In
    setTimeout(() => {
        toast.classList.remove("opacity-0", "translate-y-2");
        toast.classList.add("opacity-100", "translate-y-0");
    }, 50);
    
    // Remove Toast after 3.5s
    setTimeout(() => {
        toast.classList.remove("opacity-100", "translate-y-0");
        toast.classList.add("opacity-0", "translate-y-[-10px]");
        setTimeout(() => toast.remove(), 300);
    }, 3500);
}

// ----------------------------------------------------
// NAVIGATION SYSTEM
// ----------------------------------------------------
function initNavigation() {
    const tabs = document.querySelectorAll(".nav-tab, .nav-action-btn");
    const views = ["view-welcome", "view-order-form", "view-payment", "view-activation-pending", "view-login", "view-dashboard", "view-editor-form", "view-media-manager", "view-contacts-share", "view-gift-registry", "view-reseller", "view-preview", "view-profile"];
    
    tabs.forEach(tab => {
        tab.addEventListener("click", (e) => {
            const target = tab.getAttribute("data-target");
            if (!appState.isLoggedIn && !['view-login', 'view-welcome', 'view-order-form', 'view-payment', 'view-activation-pending'].includes(target)) {
                showToast("Anda harus masuk terlebih dahulu", "error");
                return;
            }
            switchView(target);
        });
    });

    // Back to dashboard triggers
    document.querySelectorAll(".btn-back-dash").forEach(btn => {
        btn.addEventListener("click", () => switchView("view-dashboard"));
    });

    // Logout
    document.getElementById("btn-logout").addEventListener("click", () => {
        handleLogout();
    });

    // Cloud Sync Trigger
    document.getElementById("btn-sync").addEventListener("click", () => {
        syncDataWithServer();
    });
}

function switchView(targetId) {
    const views = ["view-welcome", "view-order-form", "view-payment", "view-activation-pending", "view-login", "view-dashboard", "view-editor-form", "view-media-manager", "view-contacts-share", "view-gift-registry", "view-reseller", "view-preview", "view-profile"];
    views.forEach(viewId => {
        const el = document.getElementById(viewId);
        if (!el) return;
        if (viewId === targetId) {
            el.classList.remove("hidden");
            el.classList.add("flex");
        } else {
            el.classList.remove("flex");
            el.classList.add("hidden");
        }
    });

    // Update bottom navigation selected state
    if (appState.isLoggedIn) {
        document.getElementById("global-nav").classList.remove("hidden");
        const navTabs = document.querySelectorAll(".nav-tab");
        navTabs.forEach(tab => {
            const target = tab.getAttribute("data-target");
            if (target === targetId) {
                tab.classList.remove("text-dustyrose");
                tab.classList.add("text-terracotta");
            } else {
                tab.classList.remove("text-terracotta");
                tab.classList.add("text-dustyrose");
            }
        });
    } else {
        document.getElementById("global-nav").classList.add("hidden");
    }

    // Specially handle scanner backdrop transparency
    if (targetId === "view-qr-scanner") {
        document.body.style.backgroundColor = "transparent";
    } else {
        document.body.style.backgroundColor = "#F5EFEB";
    }

    // Auto load contacts when entering contacts tab
    if (targetId === "view-contacts-share") {
        if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.readPhoneContacts === 'function') {
            AndroidBridge.readPhoneContacts();
        }
    }

    // Refresh charts & fetch real stats if dashboard selected
    if (targetId === "view-dashboard") {
        updateDashboardStats();
        setTimeout(drawAnalyticsChart, 50);
    }
}

// ----------------------------------------------------
// LOGIN CONTROLLER
// ----------------------------------------------------
function initLogin() {
    const tabClient = document.getElementById("tab-login-client");
    const tabReseller = document.getElementById("tab-login-reseller");
    const fieldSlug = document.getElementById("field-slug");
    const fieldEmail = document.getElementById("field-email");
    const loginIcon = document.getElementById("login-icon");
    const loginTitle = document.getElementById("login-title");
    const loginSubtitle = document.getElementById("login-subtitle");

    tabClient.addEventListener("click", () => {
        appState.userRole = 'client';
        tabClient.className = "flex-1 py-2.5 rounded-full text-sm font-semibold text-alabaster bg-terracotta transition-all duration-300 focus:outline-none z-10";
        tabReseller.className = "flex-1 py-2.5 rounded-full text-sm font-semibold text-dustyrose hover:text-chocolate transition-all duration-300 focus:outline-none z-10";
        fieldSlug.classList.remove("hidden");
        fieldEmail.classList.add("hidden");
        loginIcon.className = "fa-solid fa-circle-user text-terracotta";
        loginTitle.textContent = "Login Akun Pengantin";
        loginSubtitle.textContent = "Masukkan slug undangan & sandi Anda";
    });

    tabReseller.addEventListener("click", () => {
        appState.userRole = 'reseller';
        tabReseller.className = "flex-1 py-2.5 rounded-full text-sm font-semibold text-alabaster bg-terracotta transition-all duration-300 focus:outline-none z-10";
        tabClient.className = "flex-1 py-2.5 rounded-full text-sm font-semibold text-dustyrose hover:text-chocolate transition-all duration-300 focus:outline-none z-10";
        fieldSlug.classList.add("hidden");
        fieldEmail.classList.remove("hidden");
        loginIcon.className = "fa-solid fa-users-viewfinder text-terracotta";
        loginTitle.textContent = "Login Mitra / Reseller";
        loginSubtitle.textContent = "Masukkan email mitra & sandi Anda";
    });

    document.getElementById("form-login").addEventListener("submit", (e) => {
        e.preventDefault();
        
        const password = document.getElementById("login-password").value;
        
        if (appState.userRole === 'client') {
            const slug = document.getElementById("login-slug").value.trim().toLowerCase();
            const password = document.getElementById("login-password").value;
            if (!slug) {
                showToast("Slug tidak boleh kosong", "error");
                return;
            }
            
            const submitBtn = document.getElementById("btn-login-submit");
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<i class="fa-solid fa-spinner animate-spin"></i><span>Memproses...</span>';
            
            fetch("https://intim.my.id/server-api.php?action=login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ slug, password })
            })
            .then(res => {
                if (!res.ok) {
                    return res.json().then(err => { throw new Error(err.message || "Gagal masuk") });
                }
                return res.json();
            })
            .then(res => {
                appState.clientSlug = slug;
                appState.isLoggedIn = true;
                appState.invitationData = res.data;
                
                // Clear any trial countdowns & overlays
                if (trialInterval) clearInterval(trialInterval);
                const trialCard = document.getElementById("trial-alert-card");
                if (trialCard) trialCard.classList.add("hidden");
                const lockOverlay = document.getElementById("dashboard-locked-overlay");
                if (lockOverlay) lockOverlay.classList.add("hidden");
                const saveBtn = document.getElementById("btn-save-editor");
                if (saveBtn) saveBtn.disabled = false;

                if (res.data.status === 'trial') {
                    startTrialCountdown(res.data.trial_expires_at);
                }
                
                // Set data to form inputs
                document.getElementById("groom-name").value = res.data.groomName || "";
                document.getElementById("groom-fullname").value = res.data.groomFullName || "";
                document.getElementById("groom-parents").value = res.data.groomParents || "";
                document.getElementById("groom-instagram").value = res.data.groomInstagram || "";
                document.getElementById("bride-name").value = res.data.brideName || "";
                document.getElementById("bride-fullname").value = res.data.brideFullName || "";
                document.getElementById("bride-parents").value = res.data.brideParents || "";
                document.getElementById("bride-instagram").value = res.data.brideInstagram || "";
                document.getElementById("event-date").value = res.data.eventDateISO || "";
                document.getElementById("akad-time").value = res.data.akadTime || "";
                document.getElementById("resepsi-time").value = res.data.resepsiTime || "";
                document.getElementById("venue-name").value = res.data.venueName || "";
                document.getElementById("venue-address").value = res.data.venueAddress || "";
                document.getElementById("maps-url").value = res.data.mapsUrl || "";
                document.getElementById("bank-name").value = res.data.bankName || "";
                document.getElementById("bank-holder").value = res.data.bankHolder || "";
                document.getElementById("bank-number").value = res.data.bankNumber || "";
                document.getElementById("wallet-name").value = res.data.walletName || "";
                document.getElementById("wallet-number").value = res.data.walletNumber || "";
                
                // Populate Kustomisasi Teks (with initial fallback values from website theme defaults)
                document.getElementById("custom-title").value = res.data.customTitle || "The Wedding Of";
                document.getElementById("custom-quote").value = res.data.customQuote || "Dan di antara tanda-tanda (kebesaran)-Nya ialah Dia menciptakan pasangan-pasangan untukmu dari jenismu sendiri, agar kamu cenderung dan merasa tenteram kepadanya, dan Dia menjadikan di antaramu rasa kasih dan sayang. Sungguh, pada yang demikian itu benar-benar terdapat tanda-tanda (kebesaran Allah) bagi kaum yang berpikir.";
                document.getElementById("custom-quote-src").value = res.data.customQuoteSrc || "Ar-Rum: 21";
                document.getElementById("custom-intro").value = res.data.customIntro || "Maha Suci Allah yang telah menciptakan makhluk-Nya berpasang-pasangan. Ya Allah, perkenankanlah kami merangkai cinta kasih yang Kau ridhai dalam ikatan pernikahan suci kami...";
                document.getElementById("custom-outro").value = res.data.customOutro || "Merupakan suatu kehormatan dan kebahagiaan bagi kami apabila Bapak/Ibu/Saudara/i berkenan hadir dan memberikan doa restu kepada kedua mempelai. Atas kehadiran dan doa restunya kami ucapkan terima kasih.";
                
                // Populate display order & Adab walimah
                document.getElementById("couple-display-order").value = res.data.coupleDisplayOrder || "groom_first";
                document.getElementById("toggle-show-adab").checked = res.data.showAdabWalimah === true;
                
                // Update dynamic theme banner details
                const currentActiveTheme = res.data.order_theme || "v9";
                document.getElementById("active-theme-name").textContent = "Tema Elegan (" + currentActiveTheme.toUpperCase() + ")";
                
                // Populate Kado Digital
                const isKadoEnabled = res.data.enableKado === true;
                document.getElementById("toggle-enable-kado").checked = isKadoEnabled;
                document.getElementById("kado-recipient-name").value = res.data.kadoRecipientName || "";
                document.getElementById("kado-recipient-address").value = res.data.kadoRecipientAddress || "";
                document.getElementById("kado-recipient-phone").value = res.data.kadoRecipientPhone || "";
                
                if (isKadoEnabled) {
                    document.getElementById("kado-details-fields").classList.remove("hidden");
                } else {
                    document.getElementById("kado-details-fields").classList.add("hidden");
                }
                
                // Render Love Story nodes
                renderStoryNodes(res.data.story || []);
                
                // Handle Profile View population
                document.getElementById("profile-name").textContent = res.data.order_name || res.data.brideName || "Klien Undangan";
                document.getElementById("profile-email").textContent = res.data.order_email || "klien@gmail.com";
                const userStatus = res.data.status || "active";
                document.getElementById("profile-quota").textContent = userStatus === "trial" ? "0" : "1";
                
                // Setup trial lockers in editor views
                // Centralized trial locking rules
                applyTrialLockRules(userStatus);
                
                // Dynamic profile avatar initial
                const firstLetter = (res.data.order_name || "K").charAt(0).toUpperCase();
                const avatarEl = document.getElementById("profile-avatar-container");
                if (avatarEl) {
                    avatarEl.innerHTML = `<span class="text-terracotta font-black text-2xl">${firstLetter}</span>`;
                }

                // Populate background music preset
                const musicUrlVal = res.data.musicUrl || "https://assets.mixkit.co/music/preview/mixkit-beautiful-dream-acoustic-guitar-and-piano-1002.mp3";
                const presetSelect = document.getElementById("media-music-preset");
                const customInput = document.getElementById("media-music-custom-url");
                const wrapCustom = document.getElementById("wrapper-custom-music");
                
                if (presetSelect) {
                    const presets = [
                        "https://assets.mixkit.co/music/preview/mixkit-beautiful-dream-acoustic-guitar-and-piano-1002.mp3",
                        "https://assets.mixkit.co/music/preview/mixkit-love-story-piano-solo-1005.mp3",
                        "https://assets.mixkit.co/music/preview/mixkit-wedding-day-happy-acoustic-guitar-1012.mp3"
                    ];
                    if (presets.includes(musicUrlVal)) {
                        presetSelect.value = musicUrlVal;
                        if (customInput) customInput.value = "";
                        if (wrapCustom) wrapCustom.classList.add("hidden");
                    } else {
                        presetSelect.value = "custom";
                        if (customInput) customInput.value = musicUrlVal;
                        if (wrapCustom) wrapCustom.classList.remove("hidden");
                    }
                }
                
                document.getElementById("dash-couple-title").textContent = `${res.data.brideName} & ${res.data.groomName} Wedding`;
                document.getElementById("dash-link").textContent = `https://intim.my.id/${slug}`;
                document.getElementById("nav-tab-reseller").classList.add("hidden");
                
                // Store session to localStorage
                localStorage.setItem("session_slug", slug);
                localStorage.setItem("session_password", password);
                localStorage.setItem("session_role", "client");

                showToast("Login Berhasil! Selamat datang di panel Undangan Intim.", "success");
                document.getElementById("btn-logout").classList.remove("hidden");
                switchView("view-dashboard");
            })
            .catch(err => {
                showToast(err.message, "error");
            })
            .finally(() => {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '<i class="fa-solid fa-circle-user"></i><span>Masuk Sekarang</span>';
            });
            
        } else {
            const email = document.getElementById("login-email").value.trim();
            if (!email) {
                showToast("Email tidak boleh kosong", "error");
                return;
            }
            
            const submitBtn = document.getElementById("btn-login-submit");
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<i class="fa-solid fa-spinner animate-spin"></i><span>Memproses...</span>';
            
            fetch("https://intim.my.id/server-api.php?action=reseller-login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email: email })
            })
            .then(res => {
                if (!res.ok) {
                    throw new Error("Gagal memproses login reseller");
                }
                return res.json();
            })
            .then(res => {
                if (res.status !== "success" || !res.data) {
                    throw new Error(res.message || "Gagal memproses data reseller");
                }
                
                appState.isLoggedIn = true;
                appState.userRole = 'reseller';
                appState.resellerData = res.data;
                
                // Refresh reseller UI displays
                document.getElementById("reseller-affiliate-id").textContent = `ID: ${res.data.affiliateId}`;
                const commission = res.data.commission || 0;
                const points = res.data.points || 0;
                document.getElementById("reseller-commission").textContent = `Rp ${commission.toLocaleString()}`;
                document.getElementById("reseller-points").textContent = `${points} PTS`;
                
                // Initialize default reseller-custom-price and ref link
                const customPriceInput = document.getElementById("reseller-custom-price");
                if (customPriceInput) {
                    customPriceInput.value = 35000;
                }
                const refLinkEl = document.getElementById("reseller-ref-link");
                if (refLinkEl) {
                    refLinkEl.textContent = `https://intim.my.id/register?ref=${res.data.affiliateId}&price=35000`;
                }

                // Render dynamic client list
                renderResellerClients();

                // Show reseller tab
                document.getElementById("nav-tab-reseller").classList.remove("hidden");
                
                // Store reseller session to localStorage
                localStorage.setItem("session_slug", email);
                localStorage.setItem("session_role", "reseller");

                showToast("Login Mitra Berhasil! Mengalihkan ke panel Reseller.", "success");
                document.getElementById("btn-logout").classList.remove("hidden");
                switchView("view-reseller");
            })
            .catch(err => {
                showToast(err.message || "Email atau jaringan error", "error");
            })
            .finally(() => {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '<i class="fa-solid fa-circle-user"></i><span>Masuk Sekarang</span>';
            });
        }
    });
}

// ----------------------------------------------------
// ACCORDIONS SYSTEM (EDITOR)
// ----------------------------------------------------
function initAccordions() {
    const accordions = document.querySelectorAll(".accordion-header");
    accordions.forEach(header => {
        header.addEventListener("click", () => {
            const targetId = header.getAttribute("data-target");
            const content = document.getElementById(targetId);
            const icon = header.querySelector("i.fa-chevron-down") || header.querySelector("i.fa-solid");
            
            // Toggle visibility
            if (content) {
                if (content.classList.contains("hidden")) {
                    content.classList.remove("hidden");
                    content.classList.add("animate-slide-up");
                    if (icon) icon.style.transform = "rotate(180deg)";
                } else {
                    content.classList.add("hidden");
                    if (icon) icon.style.transform = "rotate(0deg)";
                }
            }
        });
    });
}

// ----------------------------------------------------
// FORMS INPUT & BINDING CONTROLLER (EDITOR)
// ----------------------------------------------------
function initForms() {
    // Fill default values from state if logged in
    const data = appState.invitationData;
    if (data) {
        document.getElementById("groom-name").value = data.groomName || "";
        document.getElementById("groom-fullname").value = data.groomFullName || "";
        document.getElementById("groom-parents").value = data.groomParents || "";
        document.getElementById("groom-instagram").value = "bima_adi"; // sample
        
        document.getElementById("bride-name").value = data.brideName || "";
        document.getElementById("bride-fullname").value = data.brideFullName || "";
        document.getElementById("bride-parents").value = data.brideParents || "";
        document.getElementById("bride-instagram").value = "aria_lestari"; // sample

        document.getElementById("event-date").value = data.eventDateISO || "";
        document.getElementById("akad-time").value = data.akadTime || "";
        document.getElementById("resepsi-time").value = data.resepsiTime || "";
        document.getElementById("venue-name").value = data.venueName || "";
        document.getElementById("venue-address").value = data.venueAddress || "";
        document.getElementById("maps-url").value = data.mapsUrl || "";

        document.getElementById("bank-name").value = data.bankName || "";
        document.getElementById("bank-number").value = data.bankNumber || "";
        document.getElementById("bank-holder").value = data.bankHolder || "";
        document.getElementById("wallet-name").value = data.walletName || "";
        document.getElementById("wallet-number").value = data.walletNumber || "";
    }

    // Handle Quick Action Publish/Draft status dropdown
    const statusSelect = document.getElementById("select-invitation-status");
    if (statusSelect) {
        statusSelect.addEventListener("change", (e) => {
            const val = e.target.value;
            const badge = document.getElementById("dash-badge-status");
            if (!badge) return;
            if (val === "Active") {
                badge.className = "px-3 py-1 bg-sage/20 text-sage text-xs font-bold rounded-full uppercase tracking-wider flex items-center";
                badge.innerHTML = `<i class="fa-solid fa-circle text-[8px] mr-1.5 animate-pulse"></i>Aktif / Published`;
                showToast("Undangan dipublikasi secara Online!", "success");
            } else if (val === "Draft") {
                badge.className = "px-3 py-1 bg-dustyrose/20 text-dustyrose text-xs font-bold rounded-full uppercase tracking-wider flex items-center";
                badge.innerHTML = `<i class="fa-solid fa-circle text-[8px] mr-1.5"></i>Draft / Offline`;
                showToast("Undangan disembunyikan (Draft)", "success");
            } else {
                badge.className = "px-3 py-1 bg-terracotta/20 text-terracotta text-xs font-bold rounded-full uppercase tracking-wider flex items-center";
                badge.innerHTML = `<i class="fa-solid fa-circle text-[8px] mr-1.5"></i>Expired`;
                showToast("Undangan diset kadaluarsa", "error");
            }
        });
    }

    // Save Invitation Editor Changes
    const btnSaveEditor = document.getElementById("btn-save-editor");
    if (btnSaveEditor) {
        btnSaveEditor.addEventListener("click", () => {
            const data = appState.invitationData;
            if (!data) {
                showToast("Data undangan belum dimuat!", "error");
                return;
            }
            // Collect UI values & update global state
            data.groomName = document.getElementById("groom-name").value;
            data.groomFullName = document.getElementById("groom-fullname").value;
            data.groomParents = document.getElementById("groom-parents").value;
            data.brideName = document.getElementById("bride-name").value;
            data.brideFullName = document.getElementById("bride-fullname").value;
            data.brideParents = document.getElementById("bride-parents").value;
            
            data.eventDateISO = document.getElementById("event-date").value;
            data.akadTime = document.getElementById("akad-time").value;
            data.resepsiTime = document.getElementById("resepsi-time").value;
            data.venueName = document.getElementById("venue-name").value;
            data.venueAddress = document.getElementById("venue-address").value;
            data.mapsUrl = document.getElementById("maps-url").value;

            data.bankName = document.getElementById("bank-name").value;
            data.bankNumber = document.getElementById("bank-number").value;
            data.bankHolder = document.getElementById("bank-holder").value;
            data.walletName = document.getElementById("wallet-name").value;
            data.walletNumber = document.getElementById("wallet-number").value;

            // Custom texts
            data.customTitle = document.getElementById("custom-title").value;
            data.customQuote = document.getElementById("custom-quote").value;
            data.customQuoteSrc = document.getElementById("custom-quote-src").value;
            data.customIntro = document.getElementById("custom-intro").value;
            data.customOutro = document.getElementById("custom-outro").value;

            // Kado digital
            data.enableKado = document.getElementById("toggle-enable-kado").checked;
            data.kadoRecipientName = document.getElementById("kado-recipient-name").value;
            data.kadoRecipientAddress = document.getElementById("kado-recipient-address").value;
            data.kadoRecipientPhone = document.getElementById("kado-recipient-phone").value;

            // Story nodes
            data.story = getStoryNodesData();

            // Dynamic Titles
            const coupleTitle = document.getElementById("dash-couple-title");
            if (coupleTitle) coupleTitle.textContent = `${data.brideName} & ${data.groomName} Wedding`;
            const dashLink = document.getElementById("dash-link");
            if (dashLink) dashLink.textContent = `https://intim.my.id/${appState.clientSlug}`;

            // Native notification triggers (alarm calculation H-7, H-3, H-1)
            scheduleNativeAlarms(data.eventDateISO);

            showToast("Perubahan data pengantin disimpan!", "success");
            switchView("view-dashboard");
        });
    }
}

// ----------------------------------------------------
// A. IMAGE PIPELINE - CANVAS COMPRESSION (WEBP)
// ----------------------------------------------------
function compressImage(file, maxWidth = 1080, quality = 0.75) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = (event) => {
            const img = new Image();
            img.src = event.target.result;
            img.onload = () => {
                const canvas = document.createElement("canvas");
                let width = img.width;
                let height = img.height;

                // Scale down proportionally
                if (width > maxWidth) {
                    height = Math.round((height * maxWidth) / width);
                    width = maxWidth;
                }

                canvas.width = width;
                canvas.height = height;

                const ctx = canvas.getContext("2d");
                ctx.drawImage(img, 0, 0, width, height);

                // Convert to compressed JPEG data URL
                const compressedBase64 = canvas.toDataURL("image/jpeg", quality);
                resolve(compressedBase64);
            };
            img.onerror = (err) => reject(err);
        };
        reader.onerror = (err) => reject(err);
    });
}

// ----------------------------------------------------
// MEDIA & GALLERY MODULE
// ----------------------------------------------------
function initMediaManager() {
    const toggleShowVideo = document.getElementById("toggle-show-video");
    const wrapperVideoUrl = document.getElementById("wrapper-video-url");
    const mediaVideoUrl = document.getElementById("media-video-url");

    const toggleShowGallery = document.getElementById("toggle-show-gallery");
    const wrapperPhotoGallery = document.getElementById("wrapper-photo-gallery");

    const galleryInput = document.getElementById("gallery-input");

    const mediaMusicPreset = document.getElementById("media-music-preset");
    const mediaMusicCustomUrl = document.getElementById("media-music-custom-url");
    const wrapperCustomMusic = document.getElementById("wrapper-custom-music");

    if (mediaMusicPreset) {
        mediaMusicPreset.addEventListener("change", (e) => {
            const isTrial = appState.invitationData && appState.invitationData.status === "trial";
            if (e.target.value === "custom") {
                if (isTrial) {
                    showToast("Fitur URL Lagu Kustom hanya tersedia untuk pengguna Premium!", "error");
                    mediaMusicPreset.value = "https://assets.mixkit.co/music/preview/mixkit-beautiful-dream-acoustic-guitar-and-piano-1002.mp3";
                    if (wrapperCustomMusic) wrapperCustomMusic.classList.add("hidden");
                    triggerUpgradeModal();
                } else {
                    if (wrapperCustomMusic) wrapperCustomMusic.classList.remove("hidden");
                }
            } else {
                if (wrapperCustomMusic) wrapperCustomMusic.classList.add("hidden");
            }
        });
    }

    // Bind Toggles
    if (toggleShowVideo) {
        toggleShowVideo.addEventListener("change", (e) => {
            if (e.target.checked) {
                if (wrapperVideoUrl) wrapperVideoUrl.classList.remove("hidden");
            } else {
                if (wrapperVideoUrl) wrapperVideoUrl.classList.add("hidden");
            }
        });
    }

    if (toggleShowGallery) {
        toggleShowGallery.addEventListener("change", (e) => {
            const isTrial = appState.invitationData && appState.invitationData.status === "trial";
            if (isTrial) {
                e.target.checked = false;
                showToast("Fitur Galeri Foto Pengantin hanya tersedia untuk pengguna Premium!", "error");
                triggerUpgradeModal();
                return;
            }
            if (e.target.checked) {
                if (wrapperPhotoGallery) wrapperPhotoGallery.classList.remove("hidden");
            } else {
                if (wrapperPhotoGallery) wrapperPhotoGallery.classList.add("hidden");
            }
        });
    }

    // Populate existing gallery elements
    renderGalleryPreview();

    // Handle Local Compression Uploads
    if (galleryInput) {
        galleryInput.addEventListener("change", async (e) => {
            const files = Array.from(e.target.files);
            if (files.length === 0) return;

            showToast(`Mengompresi ${files.length} gambar...`, "success");

            for (const file of files) {
                try {
                    // Compress WebP
                    const webpBase64 = await compressImage(file, 800, 0.7);
                    
                    // Add to client state
                    if (appState.invitationData && appState.invitationData.gallery) {
                        appState.invitationData.gallery.push(webpBase64);
                    }
                    
                    // Trigger client-to-server photo upload if online
                    uploadPhotoToServer(webpBase64, "gallery_" + Date.now());
                    
                } catch (error) {
                    console.error("Gagal kompresi file: ", error);
                    showToast("Satu gambar gagal dikompresi.", "error");
                }
            }
            
            renderGalleryPreview();
            showToast("Gambar berhasil dikompresi ke WebP ultra-ringan!", "success");
        });
    }

    // Save Media State
    const btnSaveMedia = document.getElementById("btn-save-media");
    if (btnSaveMedia) {
        btnSaveMedia.addEventListener("click", () => {
            if (!appState.invitationData) {
                showToast("Data undangan tidak ditemukan!", "error");
                return;
            }
            appState.invitationData.showVideo = toggleShowVideo ? toggleShowVideo.checked : false;
            appState.invitationData.videoUrl = mediaVideoUrl ? mediaVideoUrl.value : "";
            appState.invitationData.showGallery = toggleShowGallery ? toggleShowGallery.checked : false;
            
            const rsvpToggle = document.getElementById("toggle-enable-rsvp");
            appState.invitationData.enableRsvp = rsvpToggle ? rsvpToggle.checked : true;
            
            // Save display order and Adab walimah
            const displayOrderEl = document.getElementById("couple-display-order");
            appState.invitationData.coupleDisplayOrder = displayOrderEl ? displayOrderEl.value : "groom_first";
            const adabToggle = document.getElementById("toggle-show-adab");
            appState.invitationData.showAdabWalimah = adabToggle ? adabToggle.checked : false;

            // Save background music url
            if (mediaMusicPreset) {
                if (mediaMusicPreset.value === "custom") {
                    appState.invitationData.musicUrl = mediaMusicCustomUrl ? (mediaMusicCustomUrl.value || "https://assets.mixkit.co/music/preview/mixkit-beautiful-dream-acoustic-guitar-and-piano-1002.mp3") : "https://assets.mixkit.co/music/preview/mixkit-beautiful-dream-acoustic-guitar-and-piano-1002.mp3";
                } else {
                    appState.invitationData.musicUrl = mediaMusicPreset.value;
                }
            }

            showToast("Pengaturan media berhasil diperbarui!", "success");
            switchView("view-dashboard");
        });
    }
}

function renderGalleryPreview() {
    const container = document.getElementById("gallery-preview-container");
    if (!container) return;
    container.innerHTML = "";

    if (!appState.invitationData || !appState.invitationData.gallery) return;

    appState.invitationData.gallery.forEach((src, idx) => {
        const wrapper = document.createElement("div");
        wrapper.className = "relative aspect-square rounded-xl overflow-hidden border border-terracotta/10 group";
        
        wrapper.innerHTML = `
            <img src="${src}" class="w-full h-full object-cover" />
            <button class="absolute top-1 right-1 w-6 h-6 rounded-full bg-black/60 text-white flex items-center justify-center text-[10px] hover:bg-terracotta transition-colors" data-idx="${idx}">
                <i class="fa-solid fa-trash-can"></i>
            </button>
        `;

        const btn = wrapper.querySelector("button");
        if (btn) {
            btn.addEventListener("click", (e) => {
                e.stopPropagation();
                const removeIdx = parseInt(e.currentTarget.getAttribute("data-idx"));
                if (appState.invitationData && appState.invitationData.gallery) {
                    appState.invitationData.gallery.splice(removeIdx, 1);
                    renderGalleryPreview();
                    showToast("Foto dihapus dari galeri", "success");
                }
            });
        }

        container.appendChild(wrapper);
    });
}

// ----------------------------------------------------
// B. ALARM SYSTEM - NATIVE LOCAL NOTIFICATIONS
// ----------------------------------------------------
async function scheduleNativeAlarms(eventDateISO) {
    if (!eventDateISO) return;

    // Check permissions if capacitor local notifications are available
    if (isAndroid) {
        try {
            // Forward calculation to native android bridge
            AndroidBridge.scheduleWeddingReminders(eventDateISO);
            console.log("Calculated and delegated local notification to AndroidBridge.");
            return;
        } catch (err) {
            console.error("Failed native notification delegation:", err);
        }
    }

    // fallback mock notifications log inside app
    console.log("Scheduling mock local notifications for Wedding Date:", eventDateISO);
    const weddingDate = new Date(eventDateISO + "T09:00:00");
    if (isNaN(weddingDate.getTime())) return;

    const oneDayMs = 24 * 60 * 60 * 1000;
    const h7Time = new Date(weddingDate.getTime() - (7 * oneDayMs));
    const h3Time = new Date(weddingDate.getTime() - (3 * oneDayMs));
    const h1Time = new Date(weddingDate.getTime() - (1 * oneDayMs));

    console.log("Local Alarms scheduled for:", {
        "H-7 Date": h7Time.toLocaleDateString(),
        "H-3 Date": h3Time.toLocaleDateString(),
        "H-1 Date": h1Time.toLocaleDateString()
    });
}

// ----------------------------------------------------
// WHATSAPP & CONTACT SHARE MODULE
// ----------------------------------------------------
function initContactsShare() {
    const templateInput = document.getElementById("wa-template");
    const container = document.getElementById("guest-list-container");

    // Initialize default template message
    if (templateInput) {
        templateInput.value = appState.waMessageTemplate;
    }

    // Reset Template
    const btnResetTemplate = document.getElementById("btn-reset-template");
    if (btnResetTemplate && templateInput) {
        btnResetTemplate.addEventListener("click", () => {
            templateInput.value = appState.waMessageTemplate;
            showToast("Template di-reset ke standar", "success");
        });
    }

    renderGuestList();

    // Import Contacts Native or Mock fallback
    const btnImportContacts = document.getElementById("btn-import-contacts");
    if (btnImportContacts) {
        btnImportContacts.addEventListener("click", () => {
            if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.readPhoneContacts === 'function') {
                showToast("Membaca kontak telepon Anda...", "success");
                AndroidBridge.readPhoneContacts();
            } else {
                const mockGuests = [
                    { name: "Pak De Heri", phone: "6281234567890", status: "Unsent" },
                    { name: "Siti Rahma (Bohemian Friend)", phone: "6289876543210", status: "Unsent" },
                    { name: "Ahmad Subagio", phone: "6285511223344", status: "Unsent" },
                    { name: "Clara Wijaya", phone: "6281988223344", status: "Unsent" }
                ];
                appState.contacts = mockGuests;
                renderGuestList();
                showToast("Kontak contoh berhasil dimuat (Mock mode).", "success");
            }
        });
    }

    // Add Manual Guest
    const btnAddGuest = document.getElementById("btn-add-guest-list");
    if (btnAddGuest) {
        btnAddGuest.addEventListener("click", () => {
            const nameInput = document.getElementById("share-guest-name");
            const phoneInput = document.getElementById("share-guest-phone");
            if (!nameInput || !phoneInput) return;

            const name = nameInput.value.trim();
            let phone = phoneInput.value.trim();

            if (!name || !phone) {
                showToast("Nama & nomor HP wajib diisi", "error");
                return;
            }

            // Clean phone number (prefix 62)
            if (phone.startsWith("0")) {
                phone = "62" + phone.slice(1);
            } else if (phone.startsWith("+")) {
                phone = phone.slice(1);
            }

            appState.contacts.push({ name, phone, status: "Unsent" });
            nameInput.value = "";
            phoneInput.value = "";
            
            renderGuestList();
            showToast("Tamu ditambahkan ke daftar", "success");
        });
    }
}

function renderGuestList() {
    const container = document.getElementById("guest-list-container");
    if (!container) return;
    container.innerHTML = "";

    if (appState.contacts.length === 0) {
        container.innerHTML = `<div class="text-center py-4 text-xs text-dustyrose">Belum ada tamu terdaftar. Silakan import atau tambah manual.</div>`;
        return;
    }

    appState.contacts.forEach((guest, idx) => {
        const card = document.createElement("div");
        card.className = "flex items-center justify-between p-3.5 bg-cream/30 rounded-xl border border-terracotta/5 hover:border-terracotta/15 transition-all";
        
        const badgeColor = guest.status === "Sent" ? "bg-sage/10 text-sage" : "bg-terracotta/10 text-terracotta";
        const badgeLabel = guest.status === "Sent" ? "Terkirim" : "Belum Kirim";

        card.innerHTML = `
            <div>
                <span class="block text-xs font-bold text-chocolate">${guest.name}</span>
                <span class="text-[10px] text-dustyrose font-mono flex items-center mt-0.5">
                    <i class="fa-solid fa-phone mr-1 text-[8px]"></i>+${guest.phone}
                </span>
            </div>
            <div class="flex items-center space-x-2">
                <span class="text-[9px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${badgeColor}">${badgeLabel}</span>
                <button class="w-8 h-8 rounded-lg bg-sage text-white flex items-center justify-center hover:bg-terracotta hover:scale-105 transition-all text-xs" data-idx="${idx}">
                    <i class="fa-solid fa-paper-plane"></i>
                </button>
            </div>
        `;

        card.querySelector("button").addEventListener("click", () => {
            sendWhatsAppMessage(idx);
        });

        container.appendChild(card);
    });
}

// D. WHATSAPP SENDER HELPER
function sendWhatsAppMessage(idx) {
    const guest = appState.contacts[idx];
    const rawTemplate = document.getElementById("wa-template").value;
    
    // Auto-generate customized details
    const guestSlug = appState.clientSlug;
    const personalizedUrl = `https://intim.my.id/${guestSlug}?to=${encodeURIComponent(guest.name)}`;
    const weddingDateStr = appState.invitationData.eventDateISO;
    const venueStr = appState.invitationData.venueName;

    // Replace place-holders
    let personalizedMsg = rawTemplate
        .replace(/{nama}/g, guest.name)
        .replace(/{tanggal}/g, weddingDateStr)
        .replace(/{tempat}/g, venueStr)
        .replace(/{link}/g, personalizedUrl);

    const encodedMsg = encodeURIComponent(personalizedMsg);
    
    // Trigger window.open for WhatsApp
    const waUrl = `https://wa.me/${guest.phone}?text=${encodedMsg}`;
    
    // Set status to Sent
    guest.status = "Sent";
    renderGuestList();
    showToast(`Membuka WhatsApp untuk: ${guest.name}`, "success");
    
    window.open(waUrl, "_blank");
}



// ----------------------------------------------------
// GIFT REGISTRY PANEL
// ----------------------------------------------------
function initGiftRegistry() {
    renderGifts();

    const btnAddGift = document.getElementById("btn-add-gift");
    if (btnAddGift) {
        btnAddGift.addEventListener("click", () => {
            const input = document.getElementById("input-gift-name");
            if (!input) return;
            const val = input.value.trim();

            if (!val) {
                showToast("Tulis nama item kado fisik", "error");
                return;
            }

            const newId = appState.gifts.length > 0 ? Math.max(...appState.gifts.map(g => g.id)) + 1 : 1;
            appState.gifts.push({
                id: newId,
                name: val,
                bookedBy: "",
                status: "Available"
            });

            input.value = "";
            renderGifts();
            showToast("Kado fisik baru berhasil diregistrasi!", "success");
        });
    }
}

function renderGifts() {
    const container = document.getElementById("gift-list-container");
    if (!container) return;
    container.innerHTML = "";

    if (appState.gifts.length === 0) {
        container.innerHTML = `<div class="text-center py-4 text-xs text-dustyrose">Belum ada kado fisik terdaftar.</div>`;
        return;
    }

    appState.gifts.forEach(gift => {
        const item = document.createElement("div");
        item.className = "p-3 bg-cream/30 rounded-xl border border-terracotta/5 flex items-center justify-between";
        
        const isBooked = gift.status === "Booked";
        const badgeColor = isBooked ? "bg-sage/10 text-sage" : "bg-terracotta/10 text-terracotta";
        const badgeLabel = isBooked ? `Sudah Diklaim: ${gift.bookedBy}` : "Belum Diklaim";

        item.innerHTML = `
            <div>
                <span class="block text-xs font-bold text-chocolate">${gift.name}</span>
                <span class="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full inline-block mt-1 ${badgeColor}">${badgeLabel}</span>
            </div>
            <div class="flex items-center space-x-1.5">
                <button class="btn-toggle-gift w-8 h-8 rounded-lg bg-alabaster hover:bg-cream border border-terracotta/10 flex items-center justify-center text-sage text-xs" data-id="${gift.id}">
                    <i class="fa-solid ${isBooked ? 'fa-unlock' : 'fa-hand-holding-gift'}"></i>
                </button>
                <button class="btn-delete-gift w-8 h-8 rounded-lg bg-alabaster hover:bg-terracotta/10 border border-terracotta/10 flex items-center justify-center text-terracotta text-xs" data-id="${gift.id}">
                    <i class="fa-regular fa-trash-can"></i>
                </button>
            </div>
        `;

        // Claim Toggle
        item.querySelector(".btn-toggle-gift").addEventListener("click", () => {
            if (isBooked) {
                gift.status = "Available";
                gift.bookedBy = "";
                showToast("Klaim kado dibatalkan", "success");
            } else {
                gift.status = "Booked";
                gift.bookedBy = "Kel. Hendra Wijaya"; // Sample guest
                showToast("Kado ditandai sudah diklaim oleh tamu", "success");
            }
            renderGifts();
        });

        // Delete
        item.querySelector(".btn-delete-gift").addEventListener("click", () => {
            appState.gifts = appState.gifts.filter(g => g.id !== gift.id);
            renderGifts();
            showToast("Kado dihapus", "success");
        });

        container.appendChild(item);
    });
}

// ----------------------------------------------------
// RESELLER PANEL HUB
// ----------------------------------------------------
function initResellerPanel() {
    renderResellerClients();

    const priceInput = document.getElementById("reseller-custom-price");
    const refLinkEl = document.getElementById("reseller-ref-link");
    const btnCopyRef = document.getElementById("btn-copy-ref-link");
    const btnWithdraw = document.getElementById("btn-withdraw-reseller");
    const btnContactOwner = document.getElementById("btn-contact-owner");

    // Initialize display values if reseller data exists
    if (appState.resellerData) {
        const commissionEl = document.getElementById("reseller-commission");
        if (commissionEl) commissionEl.textContent = `Rp ${appState.resellerData.commission.toLocaleString()}`;
        const pointsEl = document.getElementById("reseller-points");
        if (pointsEl) pointsEl.textContent = `${appState.resellerData.points} PTS`;
    }

    // Dynamic Referral Link generation based on pricing markup
    const updateRefLink = () => {
        if (!priceInput || !refLinkEl || !appState.resellerData) return;
        const val = parseInt(priceInput.value) || 35000;
        const link = `https://intim.my.id/register?ref=${appState.resellerData.affiliateId}&price=${val}`;
        refLinkEl.textContent = link;
    };

    if (priceInput) {
        priceInput.addEventListener("input", () => {
            updateRefLink();
        });
        priceInput.addEventListener("blur", () => {
            const val = parseInt(priceInput.value) || 0;
            if (val < 35000) {
                priceInput.value = 35000;
                showToast("Harga minimal reseller adalah Rp 35.000", "error");
            }
            updateRefLink();
        });
    }

    if (btnCopyRef) {
        btnCopyRef.addEventListener("click", () => {
            if (refLinkEl) {
                navigator.clipboard.writeText(refLinkEl.textContent);
                showToast("Link registrasi klien disalin ke clipboard!", "success");
            }
        });
    }

    // Direct WhatsApp Withdraw Request
    if (btnWithdraw) {
        btnWithdraw.addEventListener("click", () => {
            if (!appState.resellerData) return;
            const balance = appState.resellerData.commission;
            if (balance <= 0) {
                showToast("Saldo komisi Anda kosong. Tidak ada komisi untuk dicairkan.", "error");
                return;
            }
            const waAdminNumber = "6285798644642";
            const waText = `Halo Owner Intim, saya reseller ID: ${appState.resellerData.affiliateId}\n\nSaya mengajukan pencairan komisi reseller sebesar Rp ${balance.toLocaleString()}\n\nMohon informasi data transfer untuk pengiriman manual (Proses 3-5 hari kerja). Terima kasih!`;
            const waUrl = `https://api.whatsapp.com/send?phone=${waAdminNumber}&text=${encodeURIComponent(waText)}`;
            
            if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
                AndroidBridge.openExternalBrowser(waUrl);
            } else {
                window.open(waUrl, '_blank');
            }
        });
    }

    // Direct Contact Admin Support (Request 4)
    if (btnContactOwner) {
        btnContactOwner.addEventListener("click", () => {
            if (!appState.resellerData) return;
            const waAdminNumber = "6285798644642";
            const waText = `Halo Owner Intim, saya mitra reseller ID: ${appState.resellerData.affiliateId}\n\nAda hal yang ingin saya tanyakan terkait kemitraan / program penjualan Undangan Intim.`;
            const waUrl = `https://api.whatsapp.com/send?phone=${waAdminNumber}&text=${encodeURIComponent(waText)}`;
            
            if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
                AndroidBridge.openExternalBrowser(waUrl);
            } else {
                window.open(waUrl, '_blank');
            }
        });
    }
}

function renderResellerClients() {
    if (!appState.resellerData) return;
    const listContainer = document.getElementById("reseller-client-list");
    if (listContainer) listContainer.innerHTML = "";

    const clients = appState.resellerData.clients || [];
    const clientCountEl = document.getElementById("reseller-client-count");
    if (clientCountEl) clientCountEl.textContent = clients.length;

    clients.forEach((c) => {
        const item = document.createElement("div");
        item.className = "p-3.5 bg-cream/30 rounded-xl border border-terracotta/5 flex items-center justify-between";
        
        const statusClass = c.status === "Active" ? "bg-sage/10 text-sage" : "bg-dustyrose/10 text-dustyrose";

        item.innerHTML = `
            <div>
                <span class="block text-xs font-bold text-chocolate">${c.couple}</span>
                <span class="text-[9px] text-dustyrose block mt-0.5">Link: <b>intim.my.id/${c.slug}</b> • Paket: <b>${c.package}</b></span>
            </div>
            <div class="flex items-center space-x-2">
                <span class="text-[9px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${statusClass}">${c.status}</span>
                <button class="w-7 h-7 rounded-lg bg-alabaster hover:bg-cream border border-terracotta/10 flex items-center justify-center text-sage text-xs">
                    <i class="fa-solid fa-pen-to-square"></i>
                </button>
            </div>
        `;

        // Quick edit trigger
        item.querySelector("button").addEventListener("click", () => {
            appState.clientSlug = c.slug;
            appState.userRole = 'client';
            
            // Switch login session context to this selected client
            appState.isLoggedIn = true;
            document.getElementById("dash-couple-title").textContent = `${c.couple} Wedding`;
            document.getElementById("dash-link").textContent = `https://intim.my.id/${c.slug}`;
            
            showToast(`Mengalihkan sesi edit ke: ${c.couple}`, "success");
            switchView("view-dashboard");
        });

        listContainer.appendChild(item);
    });
}

// ----------------------------------------------------
// CANVAS DRAWING PIPELINE - BESPOKE VISITOR ANALYTICS CHART
// ----------------------------------------------------
function drawAnalyticsChart() {
    const canvas = document.getElementById("chart-rsvp");
    if (!canvas || !appState.invitationData) return;

    const ctx = canvas.getContext("2d");
    const dpr = window.devicePixelRatio || 1;
    
    // Set internal resolution based on CSS size
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const height = rect.height;

    // Clear background
    ctx.clearRect(0, 0, width, height);

    // Dynamic visitor calculation based on real appState.invitationData.visits
    const totalVisits = appState.invitationData.visits || 0;
    
    // Distribute total visits over 7 days to form a realistic growth curve
    const coefficients = [0.05, 0.08, 0.12, 0.15, 0.20, 0.25, 0.15]; 
    const visits = coefficients.map(c => Math.round(c * totalVisits));
    
    const maxVal = Math.max(10, ...visits) * 1.2;
    const days = ["Sen", "Sel", "Rab", "Kam", "Jum", "Sab", "Min"];

    const padding = { top: 20, right: 15, bottom: 25, left: 35 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Draw grid lines
    ctx.strokeStyle = "rgba(192, 92, 70, 0.08)";
    ctx.lineWidth = 1;
    const gridRows = 4;
    for (let i = 0; i <= gridRows; i++) {
        const y = padding.top + (chartHeight / gridRows) * i;
        ctx.beginPath();
        ctx.moveTo(padding.left, y);
        ctx.lineTo(width - padding.right, y);
        ctx.stroke();

        // draw Y labels
        ctx.fillStyle = "#8E7A75";
        ctx.font = "bold 9px 'Plus Jakarta Sans'";
        ctx.textAlign = "right";
        ctx.fillText(Math.round(maxVal - (maxVal / gridRows) * i), padding.left - 8, y + 3);
    }

    // Draw X labels & compute points
    const points = [];
    const stepX = chartWidth / (days.length - 1);
    
    days.forEach((day, idx) => {
        const x = padding.left + stepX * idx;
        const val = visits[idx];
        const y = padding.top + chartHeight - (val / maxVal) * chartHeight;
        points.push({ x, y, val });

        // Draw day text
        ctx.fillStyle = "#8E7A75";
        ctx.font = "bold 9px 'Plus Jakarta Sans'";
        ctx.textAlign = "center";
        ctx.fillText(day, x, height - 8);
    });

    // Draw Smooth Area Gradient
    ctx.beginPath();
    ctx.moveTo(points[0].x, padding.top + chartHeight);
    
    // Create curve paths
    for (let i = 0; i < points.length; i++) {
        if (i === 0) {
            ctx.lineTo(points[i].x, points[i].y);
        } else {
            const cpX1 = points[i-1].x + (points[i].x - points[i-1].x) / 2;
            const cpY1 = points[i-1].y;
            const cpX2 = points[i-1].x + (points[i].x - points[i-1].x) / 2;
            const cpY2 = points[i].y;
            ctx.bezierCurveTo(cpX1, cpY1, cpX2, cpY2, points[i].x, points[i].y);
        }
    }
    ctx.lineTo(points[points.length - 1].x, padding.top + chartHeight);
    ctx.closePath();

    const areaGrad = ctx.createLinearGradient(0, padding.top, 0, padding.top + chartHeight);
    areaGrad.addColorStop(0, "rgba(192, 92, 70, 0.25)");
    areaGrad.addColorStop(1, "rgba(192, 92, 70, 0.0)");
    ctx.fillStyle = areaGrad;
    ctx.fill();

    // Draw Curve Line
    ctx.beginPath();
    ctx.lineWidth = 3;
    ctx.strokeStyle = "#C05C46"; // Bohemian Terracotta
    ctx.lineCap = "round";
    
    for (let i = 0; i < points.length; i++) {
        if (i === 0) {
            ctx.moveTo(points[i].x, points[i].y);
        } else {
            const cpX1 = points[i-1].x + (points[i].x - points[i-1].x) / 2;
            const cpY1 = points[i-1].y;
            const cpX2 = points[i-1].x + (points[i].x - points[i-1].x) / 2;
            const cpY2 = points[i].y;
            ctx.bezierCurveTo(cpX1, cpY1, cpX2, cpY2, points[i].x, points[i].y);
        }
    }
    ctx.stroke();

    // Draw points & labels on top of curves
    points.forEach((pt, idx) => {
        // Draw circle point
        ctx.beginPath();
        ctx.arc(pt.x, pt.y, 4, 0, 2 * Math.PI);
        ctx.fillStyle = "#FFFCF7";
        ctx.fill();
        ctx.lineWidth = 2;
        ctx.strokeStyle = "#728C7D"; // Sage Green
        ctx.stroke();

        // draw value hovering if it's the last or highest point
        if (idx === points.length - 1) {
            ctx.fillStyle = "#3C2F2F";
            ctx.font = "black 9px 'Plus Jakarta Sans'";
            ctx.textAlign = "center";
            ctx.fillText(pt.val, pt.x, pt.y - 8);
        }
    });
}

// ----------------------------------------------------
// AJAX CLIENT-SERVER COMMUNICATIONS
// ----------------------------------------------------
function syncDataWithServer() {
    showToast("Sinkronisasi data ke server...", "success");

    const payload = {
        ...appState.invitationData,
        slug: appState.clientSlug
    };

    fetch("https://intim.my.id/server-api.php?action=save-invitation", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(payload)
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === "success") {
            showToast("Sinkronisasi sukses! Data aman di cloud.", "success");
        } else {
            showToast("Sinkronisasi gagal: " + data.message, "error");
        }
    })
    .catch(err => {
        console.error("Failed to upload invitation metadata:", err);
        showToast("Sinkronisasi gagal. Menggunakan penyimpanan offline sementara.", "error");
    });
}

function uploadPhotoToServer(base64Data, filename) {
    const formData = new FormData();
    formData.append("slug", appState.clientSlug);
    formData.append("photo_base64", base64Data);
    formData.append("filename", filename);

    fetch("https://intim.my.id/server-api.php?action=upload-photo", {
        method: "POST",
        body: formData
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === "success") {
            console.log("Photo synced to server: " + data.file_path);
        }
    })
    .catch(err => {
        console.warn("Server file uploading offline or unreachable, saved locally only.", err);
    });
}

// ----------------------------------------------------
// CLIENT SELF-SERVICE ORDERING & PAYMENT CONTROLLERS
// ----------------------------------------------------
let orderState = {
    name: "",
    phone: "",
    email: "",
    theme: "v9",
    slug: "",
    password: "",
    receiptBase64: ""
};

function handleClientOrder(e) {
    e.preventDefault();
    const name = document.getElementById("order-name").value.trim();
    const phone = document.getElementById("order-phone").value.trim();
    const email = document.getElementById("order-email").value.trim();
    const theme = document.getElementById("order-theme").value;
    const slug = document.getElementById("order-slug").value.trim().toLowerCase();
    const password = document.getElementById("order-password").value;

    if (!name || !phone || !slug || !password) {
        showToast("Mohon isi semua field data!", "error");
        return;
    }
    
    const sendBtn = e.submitter || document.querySelector("#form-client-order button[type='submit']");
    const originalText = sendBtn.innerHTML;
    sendBtn.disabled = true;
    sendBtn.innerHTML = '<i class="fa-solid fa-spinner animate-spin"></i><span>Mendaftarkan...</span>';
    
    fetch("https://intim.my.id/server-api.php?action=register-trial", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, phone, email, theme, slug, password })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(err => { throw new Error(err.message || "Gagal daftar trial") });
        }
        return res.json();
    })
    .then(res => {
        appState.clientSlug = slug;
        appState.isLoggedIn = true;
        appState.invitationData = res.data;

        // Store session to localStorage
        localStorage.setItem("session_slug", slug);
        localStorage.setItem("session_password", password);
        localStorage.setItem("session_role", "client");
        
        // Populate inputs
        document.getElementById("groom-name").value = res.data.groomName || "";
        document.getElementById("groom-fullname").value = res.data.groomFullName || "";
        document.getElementById("groom-parents").value = res.data.groomParents || "";
        document.getElementById("groom-instagram").value = res.data.groomInstagram || "";
        document.getElementById("bride-name").value = res.data.brideName || "";
        document.getElementById("bride-fullname").value = res.data.brideFullName || "";
        document.getElementById("bride-parents").value = res.data.brideParents || "";
        document.getElementById("bride-instagram").value = res.data.brideInstagram || "";
        document.getElementById("event-date").value = res.data.eventDateISO || "";
        document.getElementById("akad-time").value = res.data.akadTime || "";
        document.getElementById("resepsi-time").value = res.data.resepsiTime || "";
        document.getElementById("venue-name").value = res.data.venueName || "";
        document.getElementById("venue-address").value = res.data.venueAddress || "";
        document.getElementById("maps-url").value = res.data.mapsUrl || "";
        document.getElementById("bank-name").value = res.data.bankName || "";
        document.getElementById("bank-holder").value = res.data.bankHolder || "";
        document.getElementById("bank-number").value = res.data.bankNumber || "";
        document.getElementById("wallet-name").value = res.data.walletName || "";
        document.getElementById("wallet-number").value = res.data.walletNumber || "";
        
        document.getElementById("dash-couple-title").textContent = `${res.data.brideName} & ${res.data.groomName} Wedding`;
        document.getElementById("dash-link").textContent = `https://intim.my.id/${slug}`;
        applyTrialLockRules("trial");
        
        // Hide reseller menu for clients
        document.getElementById("nav-tab-reseller").classList.add("hidden");
        
        showToast("Pendaftaran Trial Berhasil! Akun aktif selama 24 jam.", "success");
        document.getElementById("btn-logout").classList.remove("hidden");
        
        // Start countdown timer
        startTrialCountdown(res.data.trial_expires_at);
        
        switchView("view-dashboard");
    })
    .catch(err => {
        showToast(err.message, "error");
    })
    .finally(() => {
        sendBtn.disabled = false;
        sendBtn.innerHTML = originalText;
    });
}

function handleReceiptSelect(e) {
    const file = e.target.files[0];
    if (!file) return;
    
    const label = document.getElementById("receipt-label-text");
    label.textContent = "Mengompres bukti bayar...";
    
    // Compress receipt image before upload to avoid STB bandwidth load
    const reader = new FileReader();
    reader.onload = function(event) {
        const img = new Image();
        img.onload = function() {
            const canvas = document.createElement("canvas");
            let width = img.width;
            let height = img.height;
            
            // Max size 800px for receipt
            if (width > 800) {
                height = Math.round(height * 800 / width);
                width = 800;
            }
            canvas.width = width;
            canvas.height = height;
            
            const ctx = canvas.getContext("2d");
            ctx.drawImage(img, 0, 0, width, height);
            
            orderState.receiptBase64 = canvas.toDataURL("image/jpeg", 0.6);
            label.textContent = `Resi Terpilih: ${file.name} ✅`;
            showToast("Bukti pembayaran berhasil dimuat dan dikompresi.", "success");
        };
        img.src = event.target.result;
    };
    reader.readAsDataURL(file);
}

function submitReceiptOrder() {
    if (!orderState.receiptBase64) {
        showToast("Silakan unggah bukti transfer terlebih dahulu!", "error");
        return;
    }
    
    // If upgrading from dashboard
    if (!orderState.slug) {
        orderState.slug = appState.clientSlug;
        orderState.password = appState.invitationData.password;
        orderState.name = appState.invitationData.order_name || appState.invitationData.brideName || "";
        orderState.phone = appState.invitationData.order_phone || appState.invitationData.walletNumber || "";
        orderState.email = appState.invitationData.order_email || "";
        orderState.theme = appState.invitationData.order_theme || "v9";
    }
    
    const sendBtn = document.getElementById("btn-submit-receipt") || document.querySelector("#view-payment button");
    const originalText = sendBtn.innerHTML;
    sendBtn.disabled = true;
    sendBtn.innerHTML = '<i class="fa-solid fa-spinner animate-spin"></i><span>Mengirim...</span>';
    
    fetch("https://intim.my.id/server-api.php?action=order", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(orderState)
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(err => { throw new Error(err.message || "Gagal memproses order") });
        }
        return res.json();
    })
    .then(res => {
        const clientName = orderState.name;
        const clientPhone = orderState.phone;
        const clientSlug = orderState.slug;
        const clientTheme = orderState.theme;

        document.getElementById("pending-slug-display").textContent = orderState.slug;
        document.getElementById("login-slug").value = orderState.slug;
        document.getElementById("login-password").value = orderState.password;
        
        // Reset state
        orderState = { name: "", phone: "", email: "", theme: "v9", slug: "", password: "", receiptBase64: "" };
        document.getElementById("receipt-label-text").textContent = "Pilih Foto Bukti Bayar";
        document.getElementById("form-client-order").reset();
        
        showToast("Bukti pembayaran berhasil dikirim!", "success");
        
        // Update client status locally
        appState.invitationData.status = "pending_payment";
        
        // WhatsApp Redirect to Admin Bima
        const waAdminNumber = "6285798644642";
        const waText = `Halo Admin Undangan Intim, saya sudah mendaftar/melakukan upgrade dan transfer.\n\nDetail Pendaftaran:\n- Nama: ${clientName}\n- WhatsApp Klien: ${clientPhone}\n- Link Slug: intim.my.id/${clientSlug}\n- Pilihan Tema: ${clientTheme.toUpperCase()}\n- Bukti Transfer: https://intim.my.id/assets/uploads/receipt_${clientSlug}.webp\n\nMohon bantuannya untuk verifikasi akun. Terima kasih!`;
        const waUrl = `https://api.whatsapp.com/send?phone=${waAdminNumber}&text=${encodeURIComponent(waText)}`;
        
        if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
            AndroidBridge.openExternalBrowser(waUrl);
        } else {
            window.open(waUrl, '_blank');
        }
        
        switchView("view-activation-pending");
    })
    .catch(err => {
        showToast(err.message, "error");
    })
    .finally(() => {
        sendBtn.disabled = false;
        sendBtn.innerHTML = originalText;
    });
}

// ----------------------------------------------------
// LIVE IFRAME PREVIEW CONTROLLERS
// ----------------------------------------------------
function openLivePreview() {
    // Determine active theme option in selector dropdown based on current selected values
    const currentTheme = appState.invitationData.order_theme || "v9";
    document.getElementById("preview-theme-select").value = currentTheme;
    
    reloadPreviewIframe();
    switchView("view-preview");
}

function reloadPreviewIframe() {
    const iframe = document.getElementById("preview-iframe");
    const selectedTheme = document.getElementById("preview-theme-select").value;
    const slug = appState.clientSlug;
    
    // Set preview source inside iframe wrapper
    iframe.src = `https://intim.my.id/themes/${selectedTheme}/index.html?id=${slug}&preview_mode=1`;
}

// Link visit and copy handlers for dashboard
document.addEventListener("DOMContentLoaded", () => {
    const btnCopy = document.getElementById("btn-copy-link");
    const btnVisit = document.getElementById("btn-visit-link");
    
    if (btnCopy) {
        btnCopy.addEventListener("click", () => {
            const urlText = document.getElementById("dash-link").textContent;
            navigator.clipboard.writeText(urlText);
            showToast("Link undangan berhasil disalin!", "success");
        });
    }
    
    if (btnVisit) {
        btnVisit.addEventListener("click", () => {
            const urlText = document.getElementById("dash-link").textContent;
            if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
                AndroidBridge.openExternalBrowser(urlText);
            } else {
                window.open(urlText, '_blank');
            }
        });
    }
});

// Callback for Native Contacts Bridge loading
function populateNativeContacts(data) {
    try {
        const contacts = typeof data === 'string' ? JSON.parse(data) : data;
        if (contacts && contacts.length > 0) {
            appState.contacts = contacts;
            renderGuestList();
            showToast(`Berhasil memuat ${contacts.length} kontak dari HP!`, "success");
        } else {
            showToast("Tidak ada kontak yang ditemukan.", "error");
        }
    } catch (e) {
        console.error("Gagal parse native contacts: ", e);
        showToast("Gagal membaca kontak telepon.", "error");
    }
}

// Fetch and populate real dynamic visitor & RSVP counts from STB backend
function updateDashboardStats() {
    if (!appState.invitationData) return;
    const visitsCount = appState.invitationData.visits || 0;
    const visitsEl = document.getElementById("stat-views");
    if (visitsEl) {
        visitsEl.textContent = visitsCount.toLocaleString();
    }

    // Fetch dynamic comments/RSVPs count from server
    fetch(`https://intim.my.id/server-api.php?action=comment&slug=${appState.clientSlug}`)
    .then(res => res.json())
    .then(res => {
        const rsvpsEl = document.getElementById("stat-rsvp");
        if (rsvpsEl && res.code === 200 && res.data) {
            rsvpsEl.textContent = res.data.length.toLocaleString();
        }
    })
    .catch(err => console.error("Error fetching RSVPs:", err));
}

// ----------------------------------------------------
// SAAS UPGRADE & TRIAL HELPER FUNCTIONS
// ----------------------------------------------------
let trialInterval = null;

function loadThemesDynamic() {
    fetch("https://intim.my.id/themes.json")
    .then(res => {
        if (!res.ok) throw new Error("Themes load failed");
        return res.json();
    })
    .then(themes => {
        // 1. Swipeable Welcome Theme List
        const welcomeList = document.getElementById("welcome-theme-list");
        if (welcomeList) {
            welcomeList.innerHTML = themes.map(theme => {
                const tag = (theme.tags && theme.tags[0]) || theme.tag || "Premium";
                const price = theme.priceStart || theme.price || 35000;
                const image = theme.image || `https://intim.my.id/themes/${theme.slug}/preview.jpg`;
                const description = theme.demoDescription || theme.style || theme.description || "";
                
                return `
                <div class="bg-alabaster rounded-[28px] border border-[#E8E2DE] p-5 shadow-sm space-y-3">
                    <div class="flex justify-between items-center">
                        <span class="text-xs bg-terracotta/10 text-terracotta font-bold px-3 py-1 rounded-full">${tag}</span>
                        <span class="text-sm font-bold text-chocolate">Rp ${price.toLocaleString("id-ID")}</span>
                    </div>
                    <div class="aspect-[16/9] bg-gradient-to-br from-terracotta/20 to-sage/20 rounded-[20px] overflow-hidden flex items-center justify-center relative border border-[#E8E2DE]">
                        <img src="${image}" class="w-full h-full object-cover opacity-80" alt="${theme.slug}" onerror="this.src='https://intim.my.id/assets/images/pputama.jpg'">
                        <span class="absolute inset-0 bg-black/35 flex items-center justify-center text-white font-bold text-lg" style="font-family:'Caveat',cursive;">${theme.name} (${theme.slug.toUpperCase()})</span>
                    </div>
                    <p class="text-xs text-dustyrose">${description}</p>
                    
                    <div class="grid grid-cols-2 gap-2.5 pt-1.5">
                        <button type="button" onclick="previewThemeDemo('${theme.slug}')" class="py-2.5 bg-cream hover:bg-cream/80 text-chocolate font-bold text-xs rounded-xl focus:outline-none flex items-center justify-center space-x-1 border border-[#E8E2DE]">
                            <i class="fa-solid fa-eye"></i>
                            <span>Lihat Demo</span>
                        </button>
                        <button type="button" onclick="selectCatalogTheme('${theme.slug}')" class="py-2.5 bg-terracotta hover:bg-terracotta/90 text-white font-bold text-xs rounded-xl focus:outline-none flex items-center justify-center space-x-1">
                            <i class="fa-solid fa-circle-check"></i>
                            <span>Pilih Tema</span>
                        </button>
                    </div>
                </div>
                `;
            }).join("");
        }

        // 2. Select Option in client registration form
        const themeSelect = document.getElementById("order-theme");
        if (themeSelect) {
            themeSelect.innerHTML = themes.map(theme => `
                <option value="${theme.slug}">${theme.name} (${theme.slug.toUpperCase()})</option>
            `).join("");
        }

        // 3. Select Option in live preview panel selector
        const previewSelect = document.getElementById("preview-theme-select");
        if (previewSelect) {
            previewSelect.innerHTML = themes.map(theme => `
                <option value="${theme.slug}">${theme.name} (${theme.slug.toUpperCase()})</option>
            `).join("");
        }
    })
    .catch(err => {
        console.error("Gagal mengambil themes.json secara dinamis, catalog offline digunakan:", err);
    });
}

function handleForgotPassword() {
    const slug = prompt("Masukkan Slug Undangan Anda (contoh: alya-fajar):");
    if (!slug) return;
    
    const cleanSlug = slug.trim().toLowerCase().replace(/[^a-z0-9_-]/g, '');
    if (!cleanSlug) {
        showToast("Slug tidak valid.", "error");
        return;
    }
    
    const waAdminNumber = "6285798644642";
    const waText = `Halo Admin Undangan Intim,\n\nSaya lupa kata sandi untuk akun undangan dengan slug: *${cleanSlug}*.\n\nMohon bantuannya untuk melakukan reset sandi. Terima kasih!`;
    const waUrl = `https://api.whatsapp.com/send?phone=${waAdminNumber}&text=${encodeURIComponent(waText)}`;
    
    if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
        AndroidBridge.openExternalBrowser(waUrl);
    } else {
        window.open(waUrl, '_blank');
    }
    showToast("Membuka WhatsApp Admin untuk reset sandi...", "success");
}

function startTrialCountdown(expiresTimestamp) {
    if (trialInterval) clearInterval(trialInterval);
    
    const card = document.getElementById("trial-alert-card");
    const counterText = document.getElementById("trial-countdown");
    const overlay = document.getElementById("dashboard-locked-overlay");
    const saveBtn = document.getElementById("btn-save-editor");
    
    if (!card) return;
    card.classList.remove("hidden");
    
    function updateTimer() {
        const now = Math.floor(Date.now() / 1000);
        const diff = expiresTimestamp - now;
        
        if (diff <= 0) {
            clearInterval(trialInterval);
            counterText.textContent = "Trial Expired / Habis";
            if (overlay) overlay.classList.remove("hidden");
            if (saveBtn) saveBtn.disabled = true;
            return;
        }
        
        const days = Math.floor(diff / 86400);
        const hours = Math.floor((diff % 86400) / 3600);
        const minutes = Math.floor((diff % 3600) / 60);
        const seconds = diff % 60;
        
        counterText.textContent = `${days} Hari : ${hours} Jam : ${minutes} Menit : ${seconds} Detik`;
        if (overlay) overlay.classList.add("hidden");
        if (saveBtn) saveBtn.disabled = false;
    }
    
    updateTimer();
    trialInterval = setInterval(updateTimer, 1000);
}

function openUpgradeView() {
    orderState = {
        name: appState.invitationData.order_name || appState.invitationData.brideName || "",
        phone: appState.invitationData.order_phone || appState.invitationData.walletNumber || "",
        email: appState.invitationData.order_email || "",
        theme: appState.invitationData.order_theme || appState.invitationData.theme || "v9",
        slug: appState.clientSlug,
        password: appState.invitationData.password || "",
        receiptBase64: ""
    };
    
    document.getElementById("receipt-label-text").textContent = "Pilih Foto Bukti Bayar";
    
    // Dynamically adjust back button of payment view to dashboard
    const backBtn = document.querySelector("#view-payment button[onclick*='switchView']");
    if (backBtn) {
        backBtn.setAttribute("onclick", "switchView('view-dashboard')");
    }
    
    switchView("view-payment");
}

function handleLogout() {
    if (trialInterval) clearInterval(trialInterval);
    const trialCard = document.getElementById("trial-alert-card");
    if (trialCard) trialCard.classList.add("hidden");
    const lockOverlay = document.getElementById("dashboard-locked-overlay");
    if (lockOverlay) lockOverlay.classList.add("hidden");
    
    appState.clientSlug = "";
    appState.isLoggedIn = false;
    appState.invitationData = null;
    
    // Clear localStorage session
    localStorage.removeItem("session_slug");
    localStorage.removeItem("session_password");
    localStorage.removeItem("session_role");
    
    document.getElementById("form-login").reset();
    document.getElementById("btn-logout").classList.add("hidden");
    document.getElementById("global-nav").classList.add("hidden");
    document.getElementById("nav-tab-reseller").classList.remove("hidden");
    
    switchView("view-welcome");
    showToast("Anda telah keluar dari sistem", "success");
}

// ----------------------------------------------------
// COMPLEX ACCORDIONS & LOVE STORY & PROFILE HELPERS
// ----------------------------------------------------
let storyList = [];

function renderStoryNodes(stories) {
    storyList = stories;
    const container = document.getElementById("story-editor-container");
    if (!container) return;
    
    if (storyList.length === 0) {
        container.innerHTML = `<p class="text-xs text-dustyrose text-center py-2">Belum ada cerita perjalanan cinta.</p>`;
        return;
    }
    
    container.innerHTML = storyList.map((story, index) => `
        <div class="bg-cream/45 p-4 rounded-2xl border border-[#E8E2DE] space-y-3 relative">
            <div class="flex justify-between items-center">
                <span class="text-xs font-bold text-chocolate">Cerita ${index + 1}</span>
                <button type="button" onclick="deleteStoryNode(${index})" class="text-terracotta hover:text-terracotta/80 focus:outline-none">
                    <i class="fa-solid fa-trash-can text-sm"></i>
                </button>
            </div>
            <div>
                <label class="block text-[10px] font-bold text-[#8E7A75] mb-1">Judul Cerita</label>
                <input type="text" class="story-title w-full px-3 py-2 text-xs rounded-lg bg-white border border-[#E8E2DE] text-chocolate" value="${story.title || ''}" placeholder="e.g. Pertama Bertemu" required>
            </div>
            <div>
                <label class="block text-[10px] font-bold text-[#8E7A75] mb-1">Deskripsi Cerita</label>
                <textarea class="story-desc w-full px-3 py-2 text-xs rounded-lg bg-white border border-[#E8E2DE] text-chocolate" rows="3" placeholder="Ceritakan detail kisahnya..." required>${story.description || ''}</textarea>
            </div>
        </div>
    `).join("");
}

function addNewStoryNode() {
    storyList.push({ title: "Judul Cerita", description: "Ceritakan perjalanan cinta Anda di sini." });
    renderStoryNodes(storyList);
    showToast("Cerita baru berhasil ditambahkan!", "success");
}

function deleteStoryNode(index) {
    if (confirm("Apakah Anda yakin ingin menghapus bagian cerita ini?")) {
        storyList.splice(index, 1);
        renderStoryNodes(storyList);
        showToast("Cerita berhasil dihapus.", "success");
    }
}

function getStoryNodesData() {
    const titles = document.querySelectorAll(".story-title");
    const descs = document.querySelectorAll(".story-desc");
    const result = [];
    for (let i = 0; i < titles.length; i++) {
        result.push({
            title: titles[i].value.trim(),
            description: descs[i].value.trim()
        });
    }
    return result;
}

// Quick Templates Setter
function setQuickTitle(title) {
    const el = document.getElementById("custom-title");
    if (el) {
        el.value = title;
        showToast("Judul diubah ke template cepat!", "success");
    }
}

function setQuickIntro(type) {
    const elIntro = document.getElementById("custom-intro");
    const elQuote = document.getElementById("custom-quote");
    const elSrc = document.getElementById("custom-quote-src");
    
    if (!elIntro) return;
    
    let textIntro = "";
    let textQuote = "";
    let textSrc = "";
    
    if (type === 'muslim') {
        textIntro = "Maha Suci Allah yang telah menciptakan makhluk-Nya berpasang-pasangan. Ya Allah, perkenankanlah kami merangkai cinta kasih yang Kau ridhai dalam ikatan pernikahan suci kami...";
        textQuote = "Dan di antara tanda-tanda (kebesaran)-Nya ialah Dia menciptakan pasangan-pasangan untukmu dari jenismu sendiri, agar kamu cenderung dan merasa tenteram kepadanya, dan Dia menjadikan di antaramu rasa kasih dan sayang. Sungguh, pada yang demikian itu benar-benar terdapat tanda-tanda (kebesaran Allah) bagi kaum yang berpikir.";
        textSrc = "Ar-Rum: 21";
    } else if (type === 'kristen') {
        textIntro = "Kasih itu sabar, kasih itu murah hati, kasih tidak cemburu. Dengan memohon tuntunan Kasih karunia Tuhan Yesus Kristus, kami mengundang Bapak/Ibu sekalian untuk menghadiri pemberkatan pernikahan kami...";
        textQuote = "Demikianlah mereka bukan lagi dua, melainkan satu. Karena itu, apa yang telah dipersatukan Allah, tidak boleh diceraikan manusia.";
        textSrc = "Matius 19:6";
    } else if (type === 'hindu') {
        textIntro = "Om Swastyastu, atas asung kertha wara nugraha Ida Sang Hyang Widhi Wasa, kami bermaksud mengundang Bapak/Ibu/Saudara/i sekalian untuk dapat menghadiri Upacara Pawiwahan (Pernikahan) putra-putri kami...";
        textQuote = "Grbhnami te saubhagatvaya hastam mayapatya jarradastir-yathasah, Bhago Aryama Savita Purandhir-mahyam tva-durgarhapatyaya devah. (Aku memegang tanganmu untuk saling mendukung, semoga kita diberkahi dan berumur panjang dalam membina keluarga sejahtera).";
        textSrc = "Rig Veda X.85.36";
    } else {
        textIntro = "Kisah cinta terindah adalah saat dua hati menyatu dalam janji setia sehidup semati. Kami mengundang bapak/ibu sekalian untuk ikut menjadi saksi ikatan janji suci kami...";
        textQuote = "Kisah cinta terindah adalah saat dua hati menyatu dalam janji setia sehidup semati.";
        textSrc = "Pujangga";
    }
    
    elIntro.value = textIntro;
    if (elQuote) elQuote.value = textQuote;
    if (elSrc) elSrc.value = textSrc;
    
    showToast("Teks & kutipan disesuaikan ke template cepat!", "success");
}

// Profile Menu Action Handlers
function showRiwayatTransaksi() {
    const userStatus = appState.invitationData ? appState.invitationData.status : "trial";
    if (userStatus === "trial") {
        alert("Riwayat Transaksi:\n\n1. Registrasi Akun Trial (Uji Coba 24 Jam) - GRATIS (Aktif)");
    } else {
        alert("Riwayat Transaksi:\n\n1. Registrasi Akun Trial (Uji Coba 24 Jam) - GRATIS\n2. Upgrade Paket Premium - Rp 35.000 (Lunas)");
    }
}

function showBantuanFAQ() {
    alert("Bantuan & FAQ:\n\nQ: Bagaimana cara mengaktifkan undangan selamanya?\nA: Hubungi admin kami di 6285798644642 untuk upgrade ke paket premium.\n\nQ: Berapa lama waktu verifikasi pembayaran?\nA: Estimasi proses verifikasi adalah 5-15 menit setelah upload resi.");
}

function showPrivacyPolicy() {
    alert("Kebijakan Privasi:\n\nKami melindungi seluruh informasi pribadi pengantin, foto-foto galeri, buku tamu, dan detail acara dengan enkripsi SSL. Data Anda tidak akan dibagikan ke pihak ketiga mana pun.");
}

function handleDeleteAccount() {
    if (confirm("PERINGATAN! Apakah Anda yakin ingin menghapus permanen akun Anda dan seluruh data undangan? Tindakan ini TIDAK BISA dibatalkan!")) {
        const passwordConfirm = prompt("Masukkan kata sandi akun Anda untuk konfirmasi penghapusan:");
        if (passwordConfirm === appState.invitationData.password) {
            fetch(`https://intim.my.id/server-api.php?action=delete-client-self&slug=${appState.clientSlug}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ password: passwordConfirm })
            })
            .then(res => res.json())
            .then(res => {
                showToast("Akun Anda berhasil dihapus selamanya.", "success");
                handleLogout();
            })
            .catch(e => {
                showToast("Gagal menghapus akun. Server tidak dapat merespons.", "error");
            });
        } else {
            alert("Sandi konfirmasi salah! Penghapusan akun dibatalkan.");
        }
    }
}

function initKadoManager() {
    const kadoToggle = document.getElementById("toggle-enable-kado");
    const fields = document.getElementById("kado-details-fields");
    if (kadoToggle && fields) {
        kadoToggle.addEventListener("change", (e) => {
            if (e.target.checked) {
                fields.classList.remove("hidden");
            } else {
                fields.classList.add("hidden");
            }
        });
    }
}

// ----------------------------------------------------
// DEFAULT RESETS & MEDIA HELPER ROUTINES
// ----------------------------------------------------
function setQuickQuote() {
    const quoteEl = document.getElementById("custom-quote");
    const srcEl = document.getElementById("custom-quote-src");
    if (quoteEl && srcEl) {
        quoteEl.value = "Dan di antara tanda-tanda (kebesaran)-Nya ialah Dia menciptakan pasangan-pasangan untukmu dari jenismu sendiri, agar kamu cenderung dan merasa tenteram kepadanya, dan Dia menjadikan di antaramu rasa kasih dan sayang. Sungguh, pada yang demikian itu benar-benar terdapat tanda-tanda (kebesaran Allah) bagi kaum yang berpikir.";
        srcEl.value = "Ar-Rum: 21";
        showToast("Quote dikembalikan ke bawaan!", "success");
    }
}

function setQuickOutro() {
    const outroEl = document.getElementById("custom-outro");
    if (outroEl) {
        outroEl.value = "Merupakan suatu kehormatan dan kebahagiaan bagi kami apabila Bapak/Ibu/Saudara/i berkenan hadir dan memberikan doa restu kepada kedua mempelai. Atas kehadiran dan doa restunya kami ucapkan terima kasih.";
        showToast("Teks penutup dikembalikan ke bawaan!", "success");
    }
}

function openThemeSelectorModalDirect() {
    const list = ["v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9"];
    const names = {
        "v1": "Classic Minimalist (V1)",
        "v2": "Vintage Rustic (V2)",
        "v3": "Modern Floral (V3)",
        "v4": "Midnight Glamour (V4)",
        "v5": "Bohemian Chic (V5)",
        "v6": "Sakura Blossom (V6)",
        "v7": "Emerald Gold (V7)",
        "v8": "Royal Garden (V8)",
        "v9": "Scrapbook Polaroid (V9)"
    };
    
    let promptMsg = "Pilih tema baru Anda:\n\n";
    list.forEach(slug => {
        promptMsg += `- ${slug.toUpperCase()} : ${names[slug]}\n`;
    });
    
    const choice = prompt(promptMsg + "\nMasukkan kode tema (e.g. v9):", appState.invitationData.order_theme || "v9");
    if (choice && list.includes(choice.toLowerCase().trim())) {
        const themeSlug = choice.toLowerCase().trim();
        appState.invitationData.order_theme = themeSlug;
        
        const nameEl = document.getElementById("active-theme-name");
        if (nameEl) nameEl.textContent = "Tema Elegan (" + themeSlug.toUpperCase() + ")";
        
        showToast("Tema aktif berhasil diubah!", "success");
    } else if (choice) {
        alert("Kode tema tidak valid! Pilih v1 sampai v9.");
    }
}

function triggerUpgradeModal() {
    const modal = document.getElementById("premium-upgrade-modal");
    if (modal) modal.classList.remove("hidden");
}

function closeUpgradeModal() {
    const modal = document.getElementById("premium-upgrade-modal");
    if (modal) modal.classList.add("hidden");
}

function confirmUpgradeAction() {
    closeUpgradeModal();
    openUpgradeView();
}

function applyTrialLockRules(status) {
    const isTrial = status === "trial";
    
    // 1. Kado & Rekening Digital (Point 5)
    const rekeningAccordion = document.getElementById("acc-rekening");
    if (rekeningAccordion) {
        rekeningAccordion.querySelectorAll("input, textarea, select, button").forEach(el => {
            el.disabled = isTrial;
            if (isTrial) {
                el.classList.add("opacity-50", "pointer-events-none");
            } else {
                el.classList.remove("opacity-50", "pointer-events-none");
            }
        });
    }
    
    const bankName = document.getElementById("bank-name");
    const bankHolder = document.getElementById("bank-holder");
    const bankNumber = document.getElementById("bank-number");
    const walletName = document.getElementById("wallet-name");
    const walletNumber = document.getElementById("wallet-number");
    const toggleEnableKado = document.getElementById("toggle-enable-kado");
    
    const kadoRecipientName = document.getElementById("kado-recipient-name");
    const kadoRecipientAddress = document.getElementById("kado-recipient-address");
    const kadoRecipientPhone = document.getElementById("kado-recipient-phone");
    
    if (bankName) bankName.disabled = isTrial;
    if (bankHolder) bankHolder.disabled = isTrial;
    if (bankNumber) bankNumber.disabled = isTrial;
    if (walletName) walletName.disabled = isTrial;
    if (walletNumber) walletNumber.disabled = isTrial;
    if (toggleEnableKado) toggleEnableKado.disabled = isTrial;
    
    if (kadoRecipientName) kadoRecipientName.disabled = isTrial;
    if (kadoRecipientAddress) kadoRecipientAddress.disabled = isTrial;
    if (kadoRecipientPhone) kadoRecipientPhone.disabled = isTrial;
    
    const kadoLockedBanner = document.getElementById("kado-locked-banner");
    if (kadoLockedBanner) {
        if (isTrial) kadoLockedBanner.classList.remove("hidden");
        else kadoLockedBanner.classList.add("hidden");
    }
    
    // 2. Photo Gallery & Uploads (Point 6)
    const toggleShowGallery = document.getElementById("toggle-show-gallery");
    const galleryInput = document.getElementById("gallery-input");
    const galleryUploadLabel = galleryInput ? galleryInput.parentElement : null;
    
    if (galleryInput) galleryInput.disabled = isTrial;
    
    // Visual opacity block for gallery upload label if trial
    if (galleryUploadLabel) {
        if (isTrial) {
            galleryUploadLabel.classList.add("opacity-50", "pointer-events-none");
            // Add a locked overlay badge
            if (!document.getElementById("gallery-lock-badge")) {
                const badge = document.createElement("span");
                badge.id = "gallery-lock-badge";
                badge.className = "text-[9px] bg-terracotta text-white font-black px-2 py-0.5 rounded mt-1";
                badge.innerHTML = `<i class="fa-solid fa-lock mr-1"></i>Premium Only`;
                galleryUploadLabel.appendChild(badge);
            }
        } else {
            galleryUploadLabel.classList.remove("opacity-50", "pointer-events-none");
            const badge = document.getElementById("gallery-lock-badge");
            if (badge) badge.remove();
        }
    }
    
    // 3. Kirim Undangan WhatsApp guest share tab (Point 7)
    const shareLockedCard = document.getElementById("share-locked-card");
    const shareContentWrapper = document.getElementById("share-content-wrapper");
    if (shareLockedCard && shareContentWrapper) {
        if (isTrial) {
            shareLockedCard.classList.remove("hidden");
            shareContentWrapper.classList.add("hidden");
        } else {
            shareLockedCard.classList.add("hidden");
            shareContentWrapper.classList.remove("hidden");
        }
    }
    
    // 4. Custom Background Music (Point 10)
    const musicCustomUrl = document.getElementById("media-music-custom-url");
    if (musicCustomUrl) musicCustomUrl.disabled = isTrial;
}

function previewThemeDemo(slug) {
    const url = `https://intim.my.id/themes/${slug}/index.html?id=arya-fajar&preview_mode=1`;
    if (isAndroid) {
        try {
            AndroidBridge.openExternalBrowser(url);
        } catch (e) {
            window.open(url, "_blank");
        }
    } else {
        window.open(url, "_blank");
    }
}

function selectCatalogTheme(slug) {
    // 1. Select option in registration form
    const orderTheme = document.getElementById("order-theme");
    if (orderTheme) orderTheme.value = slug;
    
    // 2. Open registration view
    switchView("view-order-form");
    showToast(`Tema ${slug.toUpperCase()} berhasil dipilih! Silakan isi data pendaftaran.`, "success");
}

let paymentPollInterval = null;
let lastPaymentUrl = "";

function startSayaBayarPayment() {
    const slug = orderState.slug || appState.clientSlug;
    if (!slug) {
        showToast("Slug tidak ditemukan. Silakan ulangi pendaftaran.", "error");
        return;
    }

    const payBtn = document.getElementById("btn-pay-sayabayar");
    const statusContainer = document.getElementById("payment-pending-status");

    payBtn.disabled = true;
    payBtn.innerHTML = '<i class="fa-solid fa-spinner animate-spin mr-1.5"></i><span>Menghubungkan...</span>';

    fetch("https://intim.my.id/server-api.php?action=create-sayabayar-invoice", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ slug: slug })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(err => { throw new Error(err.message || "Gagal membuat invoice") });
        }
        return res.json();
    })
    .then(res => {
        lastPaymentUrl = res.payment_url;
        
        // Show status poller
        if (statusContainer) statusContainer.classList.remove("hidden");
        
        // Open payment url in external web browser / browser tab
        if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
            AndroidBridge.openExternalBrowser(res.payment_url);
        } else {
            window.open(res.payment_url, "_blank");
        }
        
        // Start polling payment status from STB every 5 seconds
        if (paymentPollInterval) clearInterval(paymentPollInterval);
        paymentPollInterval = setInterval(checkPaymentStatus, 5000);
        showToast("Halaman pembayaran berhasil dibuka!", "success");
    })
    .catch(err => {
        showToast(err.message, "error");
    })
    .finally(() => {
        payBtn.disabled = false;
        payBtn.innerHTML = '<i class="fa-solid fa-credit-card"></i><span>Bayar Sekarang (Instan & Otomatis)</span>';
    });
}

function checkPaymentStatus() {
    const slug = orderState.slug || appState.clientSlug;
    if (!slug) return;

    fetch(`https://intim.my.id/server-api.php?action=check-trial&slug=${slug}`)
    .then(res => res.json())
    .then(res => {
        if (res.expired === false) {
            // Invoice paid successfully & active!
            if (paymentPollInterval) clearInterval(paymentPollInterval);
            paymentPollInterval = null;

            showToast("Pembayaran Sukses! Selamat, akun Premium Anda aktif.", "success");
            
            // Hide pending indicators
            const statusContainer = document.getElementById("payment-pending-status");
            if (statusContainer) statusContainer.classList.add("hidden");

            // Perform automatic login refresh / sync
            forceRefreshLogin(slug, orderState.password || appState.invitationData.password);
        }
    })
    .catch(err => {
        console.error("Gagal memeriksa status pembayaran:", err);
    });
}

function reopenPaymentUrl() {
    if (lastPaymentUrl) {
        if (typeof AndroidBridge !== 'undefined' && typeof AndroidBridge.openExternalBrowser === 'function') {
            AndroidBridge.openExternalBrowser(lastPaymentUrl);
        } else {
            window.open(lastPaymentUrl, "_blank");
        }
    } else {
        showToast("Invoice pembayaran belum dibuat.", "error");
    }
}

function forceRefreshLogin(slug, password) {
    fetch("https://intim.my.id/server-api.php?action=login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ slug, password })
    })
    .then(res => {
        if (!res.ok) throw new Error("Gagal menyinkronkan data login");
        return res.json();
    })
    .then(res => {
        appState.clientSlug = slug;
        appState.isLoggedIn = true;
        appState.invitationData = res.data;
        
        if (trialInterval) clearInterval(trialInterval);
        const trialCard = document.getElementById("trial-alert-card");
        if (trialCard) trialCard.classList.add("hidden");
        const lockOverlay = document.getElementById("dashboard-locked-overlay");
        if (lockOverlay) lockOverlay.classList.add("hidden");
        const saveBtn = document.getElementById("btn-save-editor");
        if (saveBtn) saveBtn.disabled = false;
        
        applyTrialLockRules(res.data.status);
        
        document.getElementById("dash-couple-title").textContent = `${res.data.brideName} & ${res.data.groomName} Wedding`;
        document.getElementById("profile-name").textContent = res.data.order_name || res.data.brideName || "Klien Undangan";
        document.getElementById("profile-email").textContent = res.data.order_email || "klien@gmail.com";
        document.getElementById("profile-quota").textContent = "1";
        
        document.getElementById("global-nav").classList.remove("hidden");
        document.getElementById("nav-tab-reseller").classList.add("hidden");
        document.getElementById("btn-logout").classList.remove("hidden");
        
        // Reset local order state
        orderState = { name: "", phone: "", email: "", theme: "v9", slug: "", password: "", receiptBase64: "" };
        document.getElementById("form-client-order").reset();

        switchView("view-dashboard");
    })
    .catch(err => {
        console.error("Gagal menyegarkan status login:", err);
    });
}

/**
 * Handles system back presses intercepted by Android Native bridge.
 */
function handleAndroidBackPress() {
    if (typeof currentView === "undefined") return;
    
    // Welcome views or root screens -> Exit App
    const exitViews = ["view-welcome", "view-login", "view-landing", "view-order-form", "view-dashboard", "view-reseller"];
    if (exitViews.includes(currentView)) {
        if (typeof AndroidBridge !== "undefined" && typeof AndroidBridge.exitApp === "function") {
            AndroidBridge.exitApp();
        } else {
            window.history.back();
        }
    } else {
        // Nested views -> Go back to dashboard / main panel
        if (appState.userRole === "reseller") {
            switchView("view-reseller");
        } else {
            switchView("view-dashboard");
        }
    }
}

/**
 * Checks local storage for saved credentials and executes automatic login sync.
 */
function checkAutoLogin() {
    const savedSlug = localStorage.getItem("session_slug");
    const savedPassword = localStorage.getItem("session_password");
    const savedRole = localStorage.getItem("session_role");
    
    if (savedRole === "reseller" && savedSlug) {
        autoLoginReseller(savedSlug);
    } else if (savedRole === "client" && savedSlug && savedPassword) {
        autoLoginClient(savedSlug, savedPassword);
    }
}

function autoLoginClient(slug, password) {
    fetch("https://intim.my.id/server-api.php?action=login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ slug, password })
    })
    .then(res => {
        if (!res.ok) throw new Error("Auto login failed");
        return res.json();
    })
    .then(res => {
        appState.clientSlug = slug;
        appState.isLoggedIn = true;
        appState.invitationData = res.data;
        
        if (trialInterval) clearInterval(trialInterval);
        const trialCard = document.getElementById("trial-alert-card");
        if (trialCard) trialCard.classList.add("hidden");
        const lockOverlay = document.getElementById("dashboard-locked-overlay");
        if (lockOverlay) lockOverlay.classList.add("hidden");
        const saveBtn = document.getElementById("btn-save-editor");
        if (saveBtn) saveBtn.disabled = false;

        if (res.data.status === 'trial') {
            startTrialCountdown(res.data.trial_expires_at);
        }
        
        // Populate inputs
        document.getElementById("groom-name").value = res.data.groomName || "";
        document.getElementById("groom-fullname").value = res.data.groomFullName || "";
        document.getElementById("groom-parents").value = res.data.groomParents || "";
        document.getElementById("groom-instagram").value = res.data.groomInstagram || "";
        document.getElementById("bride-name").value = res.data.brideName || "";
        document.getElementById("bride-fullname").value = res.data.brideFullName || "";
        document.getElementById("bride-parents").value = res.data.brideParents || "";
        document.getElementById("bride-instagram").value = res.data.brideInstagram || "";
        document.getElementById("event-date").value = res.data.eventDateISO || "";
        document.getElementById("akad-time").value = res.data.akadTime || "";
        document.getElementById("resepsi-time").value = res.data.resepsiTime || "";
        document.getElementById("venue-name").value = res.data.venueName || "";
        document.getElementById("venue-address").value = res.data.venueAddress || "";
        document.getElementById("maps-url").value = res.data.mapsUrl || "";
        document.getElementById("bank-name").value = res.data.bankName || "";
        document.getElementById("bank-holder").value = res.data.bankHolder || "";
        document.getElementById("bank-number").value = res.data.bankNumber || "";
        document.getElementById("wallet-name").value = res.data.walletName || "";
        document.getElementById("wallet-number").value = res.data.walletNumber || "";
        
        document.getElementById("custom-title").value = res.data.customTitle || "The Wedding Of";
        document.getElementById("custom-quote").value = res.data.customQuote || "Dan di antara tanda-tanda (kebesaran)-Nya ialah Dia menciptakan pasangan-pasangan untukmu dari jenismu sendiri...";
        document.getElementById("custom-quote-src").value = res.data.customQuoteSrc || "Ar-Rum: 21";
        document.getElementById("custom-intro").value = res.data.customIntro || "Maha Suci Allah yang telah menciptakan makhluk-Nya berpasang-pasangan...";
        document.getElementById("custom-outro").value = res.data.customOutro || "Merupakan suatu kehormatan...";
        
        document.getElementById("couple-display-order").value = res.data.coupleDisplayOrder || "groom_first";
        document.getElementById("toggle-show-adab").checked = res.data.showAdabWalimah === true;
        
        const currentActiveTheme = res.data.order_theme || "v9";
        document.getElementById("active-theme-name").textContent = "Tema Elegan (" + currentActiveTheme.toUpperCase() + ")";
        
        const isKadoEnabled = res.data.enableKado === true;
        document.getElementById("toggle-enable-kado").checked = isKadoEnabled;
        document.getElementById("kado-recipient-name").value = res.data.kadoRecipientName || "";
        document.getElementById("kado-recipient-address").value = res.data.kadoRecipientAddress || "";
        document.getElementById("kado-recipient-phone").value = res.data.kadoRecipientPhone || "";
        
        if (isKadoEnabled) {
            document.getElementById("kado-details-fields").classList.remove("hidden");
        } else {
            document.getElementById("kado-details-fields").classList.add("hidden");
        }
        
        renderStoryNodes(res.data.story || []);
        
        document.getElementById("profile-name").textContent = res.data.order_name || res.data.brideName || "Klien Undangan";
        document.getElementById("profile-email").textContent = res.data.order_email || "klien@gmail.com";
        const userStatus = res.data.status || "active";
        document.getElementById("profile-quota").textContent = userStatus === "trial" ? "0" : "1";
        
        applyTrialLockRules(userStatus);
        
        const firstLetter = (res.data.order_name || "K").charAt(0).toUpperCase();
        const avatarEl = document.getElementById("profile-avatar-container");
        if (avatarEl) {
            avatarEl.innerHTML = `<span class="text-terracotta font-black text-2xl">${firstLetter}</span>`;
        }
        
        document.getElementById("dash-couple-title").textContent = `${res.data.brideName} & ${res.data.groomName} Wedding`;
        document.getElementById("dash-link").textContent = `https://intim.my.id/${slug}`;
        document.getElementById("nav-tab-reseller").classList.add("hidden");
        document.getElementById("btn-logout").classList.remove("hidden");
        document.getElementById("global-nav").classList.remove("hidden");
        
        switchView("view-dashboard");
    })
    .catch(err => {
        localStorage.removeItem("session_slug");
        localStorage.removeItem("session_password");
        localStorage.removeItem("session_role");
    });
}

function autoLoginReseller(email) {
    fetch("https://intim.my.id/server-api.php?action=reseller-login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email })
    })
    .then(res => {
        if (!res.ok) throw new Error("Auto login reseller failed");
        return res.json();
    })
    .then(res => {
        if (res.status !== "success" || !res.data) throw new Error("Reseller not found");
        
        appState.isLoggedIn = true;
        appState.userRole = 'reseller';
        appState.resellerData = res.data;
        
        document.getElementById("reseller-affiliate-id").textContent = `ID: ${res.data.affiliateId}`;
        const commission = res.data.commission || 0;
        const points = res.data.points || 0;
        document.getElementById("reseller-commission").textContent = `Rp ${commission.toLocaleString()}`;
        document.getElementById("reseller-points").textContent = `${points} PTS`;
        
        const customPriceInput = document.getElementById("reseller-custom-price");
        if (customPriceInput) customPriceInput.value = 35000;
        
        const refLinkEl = document.getElementById("reseller-ref-link");
        if (refLinkEl) {
            refLinkEl.textContent = `https://intim.my.id/register?ref=${res.data.affiliateId}&price=35000`;
        }
        
        renderResellerClients();
        document.getElementById("nav-tab-reseller").classList.remove("hidden");
        document.getElementById("btn-logout").classList.remove("hidden");
        document.getElementById("global-nav").classList.remove("hidden");
        
        switchView("view-reseller");
    })
    .catch(err => {
        localStorage.removeItem("session_slug");
        localStorage.removeItem("session_role");
    });
}
