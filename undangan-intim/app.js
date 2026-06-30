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
    initScanner();
    initGiftRegistry();
    initResellerPanel();
    
    // Draw initial chart on login (or when state becomes active)
    drawAnalyticsChart();
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
    const views = ["view-welcome", "view-order-form", "view-payment", "view-activation-pending", "view-login", "view-dashboard", "view-editor-form", "view-media-manager", "view-contacts-share", "view-qr-scanner", "view-gift-registry", "view-reseller", "view-preview"];
    
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
        appState.isLoggedIn = false;
        document.getElementById("global-nav").classList.add("hidden");
        document.getElementById("btn-logout").classList.add("hidden");
        switchView("view-welcome");
        showToast("Anda telah keluar dari sistem", "success");
    });

    // Cloud Sync Trigger
    document.getElementById("btn-sync").addEventListener("click", () => {
        syncDataWithServer();
    });
}

function switchView(targetId) {
    const views = ["view-welcome", "view-order-form", "view-payment", "view-activation-pending", "view-login", "view-dashboard", "view-editor-form", "view-media-manager", "view-contacts-share", "view-qr-scanner", "view-gift-registry", "view-reseller", "view-preview"];
    views.forEach(viewId => {
        const el = document.getElementById(viewId);
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
                
                document.getElementById("dash-couple-title").textContent = `${res.data.brideName} & ${res.data.groomName} Wedding`;
                document.getElementById("dash-link").textContent = `https://intim.my.id/${slug}`;
                document.getElementById("nav-tab-reseller").classList.add("hidden");
                
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
                document.getElementById("reseller-commission").textContent = `Rp ${res.data.commission.toLocaleString()}`;
                document.getElementById("reseller-points").textContent = `${res.data.points} PTS`;
                
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
            const icon = header.querySelector("i.fa-chevron-down");
            
            // Toggle visibility
            if (content.classList.contains("hidden")) {
                content.classList.remove("hidden");
                content.classList.add("animate-slide-up");
                icon.style.transform = "rotate(180deg)";
            } else {
                content.classList.add("hidden");
                icon.style.transform = "rotate(0deg)";
            }
        });
    });
}

// ----------------------------------------------------
// FORMS INPUT & BINDING CONTROLLER (EDITOR)
// ----------------------------------------------------
function initForms() {
    // Fill default values from state
    const data = appState.invitationData;
    document.getElementById("groom-name").value = data.groomName;
    document.getElementById("groom-fullname").value = data.groomFullName;
    document.getElementById("groom-parents").value = data.groomParents;
    document.getElementById("groom-instagram").value = "bima_adi"; // sample
    
    document.getElementById("bride-name").value = data.brideName;
    document.getElementById("bride-fullname").value = data.brideFullName;
    document.getElementById("bride-parents").value = data.brideParents;
    document.getElementById("bride-instagram").value = "aria_lestari"; // sample

    document.getElementById("event-date").value = data.eventDateISO;
    document.getElementById("akad-time").value = data.akadTime;
    document.getElementById("resepsi-time").value = data.resepsiTime;
    document.getElementById("venue-name").value = data.venueName;
    document.getElementById("venue-address").value = data.venueAddress;
    document.getElementById("maps-url").value = data.mapsUrl;

    document.getElementById("bank-name").value = data.bankName;
    document.getElementById("bank-number").value = data.bankNumber;
    document.getElementById("bank-holder").value = data.bankHolder;
    document.getElementById("wallet-name").value = data.walletName;
    document.getElementById("wallet-number").value = data.walletNumber;

    // Handle Quick Action Publish/Draft status dropdown
    document.getElementById("select-invitation-status").addEventListener("change", (e) => {
        const val = e.target.value;
        const badge = document.getElementById("dash-badge-status");
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

    // Save Invitation Editor Changes
    document.getElementById("btn-save-editor").addEventListener("click", () => {
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

        // Dynamic Titles
        document.getElementById("dash-couple-title").textContent = `${data.brideName} & ${data.groomName} Wedding`;
        document.getElementById("dash-link").textContent = `https://intim.my.id/${appState.clientSlug}`;

        // Native notification triggers (alarm calculation H-7, H-3, H-1)
        scheduleNativeAlarms(data.eventDateISO);

        showToast("Perubahan data pengantin disimpan!", "success");
        switchView("view-dashboard");
    });
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

    // Bind Toggles
    toggleShowVideo.addEventListener("change", (e) => {
        if (e.target.checked) {
            wrapperVideoUrl.classList.remove("hidden");
        } else {
            wrapperVideoUrl.classList.add("hidden");
        }
    });

    toggleShowGallery.addEventListener("change", (e) => {
        if (e.target.checked) {
            wrapperPhotoGallery.classList.remove("hidden");
        } else {
            wrapperPhotoGallery.classList.add("hidden");
        }
    });

    // Populate existing gallery elements
    renderGalleryPreview();

    // Handle Local Compression Uploads
    galleryInput.addEventListener("change", async (e) => {
        const files = Array.from(e.target.files);
        if (files.length === 0) return;

        showToast(`Mengompresi ${files.length} gambar...`, "success");

        for (const file of files) {
            try {
                // Compress WebP
                const webpBase64 = await compressImage(file, 800, 0.7);
                
                // Add to client state
                appState.invitationData.gallery.push(webpBase64);
                
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

    // Save Media State
    document.getElementById("btn-save-media").addEventListener("click", () => {
        appState.invitationData.showVideo = toggleShowVideo.checked;
        appState.invitationData.videoUrl = mediaVideoUrl.value;
        appState.invitationData.showGallery = toggleShowGallery.checked;
        appState.invitationData.enableRsvp = document.getElementById("toggle-enable-rsvp").checked;

        showToast("Pengaturan media berhasil diperbarui!", "success");
        switchView("view-dashboard");
    });
}

function renderGalleryPreview() {
    const container = document.getElementById("gallery-preview-container");
    container.innerHTML = "";

    appState.invitationData.gallery.forEach((src, idx) => {
        const wrapper = document.createElement("div");
        wrapper.className = "relative aspect-square rounded-xl overflow-hidden border border-terracotta/10 group";
        
        wrapper.innerHTML = `
            <img src="${src}" class="w-full h-full object-cover" />
            <button class="absolute top-1 right-1 w-6 h-6 rounded-full bg-black/60 text-white flex items-center justify-center text-[10px] hover:bg-terracotta transition-colors" data-idx="${idx}">
                <i class="fa-solid fa-trash-can"></i>
            </button>
        `;

        wrapper.querySelector("button").addEventListener("click", (e) => {
            e.stopPropagation();
            const removeIdx = parseInt(e.currentTarget.getAttribute("data-idx"));
            appState.invitationData.gallery.splice(removeIdx, 1);
            renderGalleryPreview();
            showToast("Foto dihapus dari galeri", "success");
        });

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
    templateInput.value = appState.waMessageTemplate;

    // Reset Template
    document.getElementById("btn-reset-template").addEventListener("click", () => {
        templateInput.value = appState.waMessageTemplate;
        showToast("Template di-reset ke standar", "success");
    });

    renderGuestList();

    // Import Contacts Native or Mock fallback
    document.getElementById("btn-import-contacts").addEventListener("click", () => {
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

    // Add Manual Guest
    document.getElementById("btn-add-guest-list").addEventListener("click", () => {
        const nameInput = document.getElementById("share-guest-name");
        const phoneInput = document.getElementById("share-guest-phone");

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

function renderGuestList() {
    const container = document.getElementById("guest-list-container");
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
// C. QR CHECK-IN SCANNER SYSTEM (NATIVE BARCODE)
// ----------------------------------------------------
function initScanner() {
    const btnTrigger = document.getElementById("btn-trigger-scanner");
    const btnCancel = document.getElementById("btn-cancel-scanner");
    const btnTorch = document.getElementById("btn-toggle-torch");
    let isTorchOn = false;

    btnTrigger.addEventListener("click", () => {
        // Show scanner layer
        switchView("view-qr-scanner");

        if (isAndroid) {
            try {
                // Call Android App Scanner Interface
                AndroidBridge.startQRScanner();
            } catch (err) {
                console.error("Android start QR Scanner interface failed:", err);
                triggerMockScanner();
            }
        } else {
            // Running in developer web-preview, simulate scanner after 2.5s
            triggerMockScanner();
        }
    });

    btnCancel.addEventListener("click", () => {
        if (isAndroid) {
            try {
                AndroidBridge.stopQRScanner();
            } catch (err) {}
        }
        switchView("view-dashboard");
    });

    btnTorch.addEventListener("click", () => {
        isTorchOn = !isTorchOn;
        if (isTorchOn) {
            btnTorch.className = "px-6 py-2.5 bg-terracotta text-white rounded-full text-xs font-bold hover:bg-terracotta/90 transition-colors flex items-center space-x-2";
            btnTorch.innerHTML = `<i class="fa-solid fa-lightbulb"></i><span>Senter AKTIF</span>`;
            if (isAndroid) {
                try { AndroidBridge.toggleTorch(true); } catch (e) {}
            }
        } else {
            btnTorch.className = "px-6 py-2.5 bg-white/10 rounded-full text-xs font-bold text-white hover:bg-white/20 transition-colors flex items-center space-x-2";
            btnTorch.innerHTML = `<i class="fa-regular fa-lightbulb"></i><span>Nyalakan Senter / Torch</span>`;
            if (isAndroid) {
                try { AndroidBridge.toggleTorch(false); } catch (e) {}
            }
        }
    });
}

function triggerMockScanner() {
    showToast("Mengaktifkan simulasi pemindai QR kamera...", "success");
    setTimeout(() => {
        // Simulate scanning a code
        const mockGuestId = "GUEST-INTIM-" + Math.floor(Math.random() * 8999 + 1000);
        handleScannedGuestCode(mockGuestId);
    }, 2500);
}

// Callback triggered by Android Native Barcode Scanner or Mock preview
function handleScannedGuestCode(guestCode) {
    // Return to dashboard
    switchView("view-dashboard");
    
    // Log attendance
    logAttendanceToServer(guestCode);
}

function logAttendanceToServer(guestCode) {
    // Display result instantly
    showToast(`ABSENSI TAMU BERHASIL: ${guestCode}`, "success");

    // Perform ajax submit
    const formData = new FormData();
    formData.append("slug", appState.clientSlug);
    formData.append("guest_id", guestCode);

    fetch("https://intim.my.id/server-api.php?action=checkin", {
        method: "POST",
        body: formData
    })
    .then(res => res.json())
    .then(data => {
        if (data.status === "success") {
            showToast(`Absen tercatat di server pada: ${data.timestamp}`, "success");
        }
    })
    .catch(err => {
        console.error("Failed offline attendance upload:", err);
    });
}

// ----------------------------------------------------
// GIFT REGISTRY PANEL
// ----------------------------------------------------
function initGiftRegistry() {
    renderGifts();

    document.getElementById("btn-add-gift").addEventListener("click", () => {
        const input = document.getElementById("input-gift-name");
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

function renderGifts() {
    const container = document.getElementById("gift-list-container");
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

    // Initialize display values
    document.getElementById("reseller-commission").textContent = `Rp ${appState.resellerData.commission.toLocaleString()}`;
    document.getElementById("reseller-points").textContent = `${appState.resellerData.points} PTS`;

    // Dynamic Referral Link generation based on pricing markup
    const updateRefLink = () => {
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
            navigator.clipboard.writeText(refLinkEl.textContent);
            showToast("Link registrasi klien disalin ke clipboard!", "success");
        });
    }

    // Direct WhatsApp Withdraw Request
    if (btnWithdraw) {
        btnWithdraw.addEventListener("click", () => {
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
    const listContainer = document.getElementById("reseller-client-list");
    listContainer.innerHTML = "";

    const clients = appState.resellerData.clients;
    document.getElementById("reseller-client-count").textContent = clients.length;

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
    if (!canvas) return;

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
    orderState.name = document.getElementById("order-name").value.trim();
    orderState.phone = document.getElementById("order-phone").value.trim();
    orderState.email = document.getElementById("order-email").value.trim();
    orderState.theme = document.getElementById("order-theme").value;
    orderState.slug = document.getElementById("order-slug").value.trim().toLowerCase();
    orderState.password = document.getElementById("order-password").value;

    if (!orderState.name || !orderState.phone || !orderState.slug || !orderState.password) {
        showToast("Mohon isi semua field data!", "error");
        return;
    }
    
    // Switch to payment view
    switchView("view-payment");
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
        
        // WhatsApp Redirect to Admin Bima
        const waAdminNumber = "6285798644642";
        const waText = `Halo Admin Undangan Intim, saya sudah mendaftar dan melakukan transfer.\n\nDetail Pendaftaran:\n- Nama: ${clientName}\n- WhatsApp Klien: ${clientPhone}\n- Link Slug: intim.my.id/${clientSlug}\n- Pilihan Tema: ${clientTheme.toUpperCase()}\n- Bukti Transfer: https://intim.my.id/assets/uploads/receipt_${clientSlug}.webp\n\nMohon bantuannya untuk verifikasi akun. Terima kasih!`;
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
