/**
 * ==========================================================================
 * CINEMATIC WEDDING INVITATION — CENTRAL CONFIG OBJECT
 * ==========================================================================
 * Edit bagian ini untuk menyesuaikan data pengantin, jadwal, dan aset.
 * data-key & data-url bisa juga diset via atribut HTML atau URL param.
 */
const INVITATION_CONFIG = {
    couple: {
        bride: {
            shortName: "Alya",
            fullName: "Alya Khairunnisa, S.E",
            parents: "Putri dari Bapak Ridwan & Ibu Aminah",
            photo: "assets/images/bride.jpg"
        },
        groom: {
            shortName: "Fajar",
            fullName: "Fajar Shidiq, S.Kom",
            parents: "Putra dari Bapak Ahmad & Ibu Siti",
            photo: "assets/images/groom.jpg"
        },
        metaTitle: "The Wedding of Alya & Fajar"
    },
    schedule: {
        targetDate: "2026-12-20T10:00:00",
        stringDate: "Sabtu, 20 Desember 2026",
        akad: {
            time: "10.00 WIB - Selesai",
            venue: "Masjid Agung",
            address: "Jl. Alun-Alun No. 1, Pusat Kota",
            mapsLink: "https://maps.google.com"
        },
        resepsi: {
            time: "11.30 WIB - 14.30 WIB",
            venue: "Gedung Serbaguna",
            address: "Jl. Pemuda No. 45, Samping Alun-Alun",
            mapsLink: "https://maps.google.com"
        }
    },
    loveStory: [
        { date: "Januari 2024", title: "First Meet", desc: "Pertemuan pertama di sebuah project kolaborasi kerja sinematografi yang tidak terduga." },
        { date: "Juni 2024", title: "First Date", desc: "Mulai menyatukan visi dan komitmen bersama, saling mengenal keluarga masing-masing." },
        { date: "Agustus 2025", title: "Engagement", desc: "Ikatan janji formal di hadapan kedua belah pihak keluarga besar dengan penuh haru." },
        { date: "Desember 2026", title: "Wedding Day ✨", desc: "Hari penentuan komitmen suci seumur hidup. Bab baru yang paling dinantikan." }
    ],
    gallery: [
        "https://images.unsplash.com/photo-1529636798458-92182e662485?auto=format&fit=crop&w=800&q=80",
        "https://images.unsplash.com/photo-1519225421980-715cb0215aed?auto=format&fit=crop&w=800&q=80",
        "https://images.unsplash.com/photo-1519741497674-611481863552?auto=format&fit=crop&w=800&q=80",
        "https://images.unsplash.com/photo-1511285560929-80b456fea0bc?auto=format&fit=crop&w=800&q=80",
        "https://images.unsplash.com/photo-1520854221256-17451cc331bf?auto=format&fit=crop&w=800&q=80"
    ],
    digitalWallets: [
        { type: "bank", provider: "BCA", accountNum: "1234567890", owner: "Alya Khairunnisa" },
        { type: "bank", provider: "Mandiri", accountNum: "9876543210", owner: "Fajar Shidiq" }
    ],
    musicUrl: "https://media.indoinvite.com/2db3bf1e16cd47a08843bb881e39cce7:indoinvite-staging/indoinvite-staging/indoinvite-staging/nikah/theme/music/1659512405.mp3",
    defaultGuestPlaceholder: "Bapak/Ibu/Saudara/i"
};

/* ============================================================
   GLOBAL STATE VARS
   ============================================================ */
let audioInstance = null;
let isAudioPlaying = false;
let currentPage = 1;
const PER_PAGE = 5;

// Ambil invitation key & API url dari atribut data- pada <body> atau URL param
const bodyEl = document.body;
const INVITATION_KEY = bodyEl.dataset.key
    || new URLSearchParams(location.search).get("id")
    || "alya-fajar";
const API_BASE = bodyEl.dataset.url
    || "https://intim.my.id";

/* ============================================================
   BOOT: DOMContentLoaded
   ============================================================ */
document.addEventListener("DOMContentLoaded", async () => {
    const params = new URLSearchParams(window.location.search);
    const invitationId = params.get('id') || 'alya-fajar';
    try {
        const res = await fetch(`/data/${invitationId}.json`);
        if (res.ok) {
            const data = await res.json();
            INVITATION_CONFIG.couple.bride.fullName = data.brideFullName || INVITATION_CONFIG.couple.bride.fullName;
            INVITATION_CONFIG.couple.bride.shortName = data.brideName || INVITATION_CONFIG.couple.bride.shortName;
            INVITATION_CONFIG.couple.bride.parents = data.brideParents || INVITATION_CONFIG.couple.bride.parents;
            INVITATION_CONFIG.couple.bride.photo = data.bridePhoto || INVITATION_CONFIG.couple.bride.photo;
            
            INVITATION_CONFIG.couple.groom.fullName = data.groomFullName || INVITATION_CONFIG.couple.groom.fullName;
            INVITATION_CONFIG.couple.groom.shortName = data.groomName || INVITATION_CONFIG.couple.groom.shortName;
            INVITATION_CONFIG.couple.groom.parents = data.groomParents || INVITATION_CONFIG.couple.groom.parents;
            INVITATION_CONFIG.couple.groom.photo = data.groomPhoto || INVITATION_CONFIG.couple.groom.photo;
            
            INVITATION_CONFIG.couple.metaTitle = "The Wedding of " + (data.brideName || "Alya") + " & " + (data.groomName || "Fajar");
            
            INVITATION_CONFIG.schedule.targetDate = data.eventDateISO || INVITATION_CONFIG.schedule.targetDate;
            INVITATION_CONFIG.schedule.stringDate = data.eventDateText || INVITATION_CONFIG.schedule.stringDate;
            
            INVITATION_CONFIG.schedule.akad.time = data.akadTime || INVITATION_CONFIG.schedule.akad.time;
            INVITATION_CONFIG.schedule.akad.venue = data.venueName || INVITATION_CONFIG.schedule.akad.venue;
            INVITATION_CONFIG.schedule.akad.address = data.venueAddress || INVITATION_CONFIG.schedule.akad.address;
            INVITATION_CONFIG.schedule.akad.mapsLink = data.mapsUrl || INVITATION_CONFIG.schedule.akad.mapsLink;
            
            INVITATION_CONFIG.schedule.resepsi.time = data.resepsiTime || INVITATION_CONFIG.schedule.resepsi.time;
            INVITATION_CONFIG.schedule.resepsi.venue = data.venueName || INVITATION_CONFIG.schedule.resepsi.venue;
            INVITATION_CONFIG.schedule.resepsi.address = data.venueAddress || INVITATION_CONFIG.schedule.resepsi.address;
            INVITATION_CONFIG.schedule.resepsi.mapsLink = data.mapsUrl || INVITATION_CONFIG.schedule.resepsi.mapsLink;
            
            if (data.story && Array.isArray(data.story)) {
                INVITATION_CONFIG.loveStory = data.story.map(s => ({
                    date: s.date,
                    title: s.title,
                    desc: s.description
                }));
            }
            if (data.bankName && data.bankNumber) {
                INVITATION_CONFIG.digitalWallets = [
                    { type: "bank", provider: data.bankName, accountNum: data.bankNumber, owner: data.bankHolder || data.brideFullName }
                ];
                if (data.walletName && data.walletNumber) {
                    INVITATION_CONFIG.digitalWallets.push({ type: "bank", provider: data.walletName, accountNum: data.walletNumber, owner: data.bankHolder || data.brideFullName });
                }
            }
            INVITATION_CONFIG.musicUrl = data.musicUrl || INVITATION_CONFIG.musicUrl;
            if (data.gallery && Array.isArray(data.gallery)) {
                INVITATION_CONFIG.gallery = data.gallery;
            }
        }
    } catch (e) {
        console.error("Gagal meload dynamic config:", e);
    }

    injectStaticData();
    extractGuestName();
    startCountdown();
    renderTimeline();
    renderGallery();
    renderGiftCards();
    loadComments();
    setupScrollReveal();
    setupNavActiveScroll();
    setupEventBinders();
    restoreTheme();
});

/* ============================================================
   1. INJECT DATA CONFIG → DOM
   ============================================================ */
function injectStaticData() {
    const set = (id, val) => { const el = document.getElementById(id); if (el) el.innerText = val; };

    set("coverDate", INVITATION_CONFIG.schedule.stringDate);
    set("groomFullName", INVITATION_CONFIG.couple.groom.fullName);
    set("groomParents", INVITATION_CONFIG.couple.groom.parents);
    set("brideFullName", INVITATION_CONFIG.couple.bride.fullName);
    set("brideParents", INVITATION_CONFIG.couple.bride.parents);

    set("timeAkad", INVITATION_CONFIG.schedule.akad.time);
    set("venueAkad", INVITATION_CONFIG.schedule.akad.venue);
    set("addrAkad", INVITATION_CONFIG.schedule.akad.address);

    set("timeResepsi", INVITATION_CONFIG.schedule.resepsi.time);
    set("venueResepsi", INVITATION_CONFIG.schedule.resepsi.venue);
    set("addrResepsi", INVITATION_CONFIG.schedule.resepsi.address);

    const linkAkad = document.getElementById("linkAkad");
    const linkResepsi = document.getElementById("linkResepsi");
    if (linkAkad) linkAkad.href = INVITATION_CONFIG.schedule.akad.mapsLink;
    if (linkResepsi) linkResepsi.href = INVITATION_CONFIG.schedule.resepsi.mapsLink;
}

/* ============================================================
   2. GUEST NAME FROM URL PARAM (?to=Nama atau ?kpd=Nama)
   ============================================================ */
function extractGuestName() {
    const params = new URLSearchParams(location.search);
    const guest = params.get("to") || params.get("kpd") || params.get("guest");
    const prefix = params.get("prefix");
    const el = document.getElementById("guestName");
    const labelEl = document.querySelector(".invitation-to p");
    
    if (prefix && labelEl) {
        labelEl.innerText = decodeURIComponent(prefix).replace(/\+/g, " ");
    }
    if (el) {
        el.innerText = guest ? decodeURIComponent(guest).replace(/\+/g, " ") : INVITATION_CONFIG.defaultGuestPlaceholder;
    }
}

/* ============================================================
   3. COUNTDOWN TIMER ENGINE
   ============================================================ */
function startCountdown() {
    const target = new Date(INVITATION_CONFIG.schedule.targetDate).getTime();

    const tick = () => {
        const diff = target - Date.now();

        if (diff <= 0) {
            const mini = document.getElementById("miniCountdown");
            if (mini) mini.innerText = "Acara telah berlangsung!";
            const main = document.getElementById("mainCountdownBox");
            const over = document.getElementById("countdownOverText");
            if (main) main.style.display = "none";
            if (over) over.style.display = "block";
            clearInterval(interval);
            return;
        }

        const d = Math.floor(diff / 86400000);
        const h = Math.floor((diff % 86400000) / 3600000);
        const m = Math.floor((diff % 3600000) / 60000);
        const s = Math.floor((diff % 60000) / 1000);

        const mini = document.getElementById("miniCountdown");
        if (mini) mini.innerText = `${d}d : ${String(h).padStart(2,"0")}h : ${String(m).padStart(2,"0")}m : ${String(s).padStart(2,"0")}s`;

        const pad = (id, v) => { const el = document.getElementById(id); if (el) el.innerText = String(v).padStart(2,"0"); };
        pad("days", d);
        pad("hours", h);
        pad("minutes", m);
        pad("seconds", s);
    };

    tick();
    const interval = setInterval(tick, 1000);
}

/* ============================================================
   4. LOVE STORY TIMELINE
   ============================================================ */
function renderTimeline() {
    const container = document.getElementById("storyTimeline");
    if (!container) return;
    container.innerHTML = INVITATION_CONFIG.loveStory.map(item => `
        <div class="timeline-node reveal">
            <div class="timeline-date">${item.date}</div>
            <h3 class="timeline-header">${item.title}</h3>
            <p class="timeline-desc">${item.desc}</p>
        </div>
    `).join("");
}

/* ============================================================
   5. GALLERY CAROUSEL
   ============================================================ */
function renderGallery() {
    const gallerySection = document.getElementById("gallery");
    if ((!INVITATION_CONFIG.gallery || INVITATION_CONFIG.gallery.length === 0) && gallerySection) {
        gallerySection.style.display = "none";
        const navBtn = document.querySelector('[href="#gallery"]');
        if (navBtn) navBtn.style.display = 'none';
        return;
    }
    const track = document.getElementById("galleryTrack");
    if (!track) return;
    track.innerHTML = INVITATION_CONFIG.gallery.map(src => `
        <div class="gallery-thumbnail-card" data-src="${src}">
            <img src="${src}" alt="Scene" loading="lazy" onerror="this.parentElement.style.display='none'">
        </div>
    `).join("");

    track.querySelectorAll(".gallery-thumbnail-card").forEach(card => {
        card.addEventListener("click", () => openModal(card.dataset.src));
    });
}

/* ============================================================
   6. GIFT CARDS DIGITAL
   ============================================================ */
function renderGiftCards() {
    const container = document.getElementById("giftCardsContainer");
    if (!container) return;
    container.innerHTML = INVITATION_CONFIG.digitalWallets.map(w => {
        if (w.type === "bank") {
            return `
            <div class="payment-card reveal">
                <div class="card-chip"></div>
                <div class="card-bank-name">${w.provider}</div>
                <span class="card-number-string">${w.accountNum}</span>
                <div class="card-holder-name">a.n. ${w.owner}</div>
                <div class="card-action-row">
                    <button class="btn-secondary btn-sm btn-copy-action" data-copy="${w.accountNum}">
                        📋 Copy Nomor
                    </button>
                </div>
            </div>`;
        } else {
            return `
            <div class="payment-card reveal">
                <div class="card-chip" style="background:#52b788;"></div>
                <div class="card-bank-name">${w.provider}</div>
                <span class="card-number-string">QRIS CODE</span>
                <div class="card-holder-name">${w.owner}</div>
                <div class="card-action-row">
                    <button class="btn-primary btn-sm btn-view-qris" data-qr="${w.qrPath}">
                        📱 Scan QRIS
                    </button>
                </div>
            </div>`;
        }
    }).join("");

    container.addEventListener("click", e => {
        const copyBtn = e.target.closest(".btn-copy-action");
        const qrisBtn = e.target.closest(".btn-view-qris");
        if (copyBtn) {
            navigator.clipboard.writeText(copyBtn.dataset.copy)
                .then(() => showToast("✅ Nomor rekening berhasil disalin!"))
                .catch(() => showToast("Gagal menyalin, coba manual."));
        }
        if (qrisBtn) openModal(qrisBtn.dataset.qr);
    });
}

/* ============================================================
   7. COMMENTS / UCAPAN via API
   ============================================================ */
async function loadComments() {
    const container = document.getElementById("reviewsContainer");
    const countEl = document.getElementById("totalReviewsCount");
    if (!container) return;

    container.innerHTML = `<p style="text-align:center; color:var(--text-muted); font-size:0.85rem; padding:20px 0;">Memuat ucapan...</p>`;

    try {
        const res = await fetch(`${API_BASE}/api/comment`, {
            headers: {
                "Authorization": "Bearer " + INVITATION_KEY
            }
        });
        if (!res.ok) throw new Error("API error " + res.status);
        const json = await res.json();
        const comments = (json.data && json.data.lists) || json.data || [];

        if (countEl) countEl.innerText = comments.length;
        renderCommentPage(comments, 1);
    } catch (err) {
        console.warn("API comment fetch failed, using localStorage fallback:", err);
        const local = JSON.parse(localStorage.getItem("v3_reviews_" + INVITATION_KEY) || "[]");
        if (countEl) countEl.innerText = local.length;
        renderCommentPage(local, 1);
    }
}

function renderCommentPage(comments, page) {
    const container = document.getElementById("reviewsContainer");
    const pagination = document.getElementById("reviewsPagination");
    const pageIndicator = document.getElementById("pageIndicator");
    if (!container) return;

    currentPage = page;
    const totalPages = Math.max(1, Math.ceil(comments.length / PER_PAGE));
    const start = (page - 1) * PER_PAGE;
    const slice = comments.slice(start, start + PER_PAGE);

    if (comments.length === 0) {
        container.innerHTML = `<p style="text-align:center; color:var(--text-muted); font-size:0.9rem; padding:30px 0;">Belum ada ucapan. Jadilah yang pertama! 💌</p>`;
        if (pagination) pagination.style.display = "none";
        return;
    }

    if (pagination) pagination.style.display = totalPages > 1 ? "flex" : "none";
    if (pageIndicator) pageIndicator.innerText = `${page}/${totalPages}`;

    container.innerHTML = slice.map(item => {
        const isDatang = item.presence === true || String(item.presence).toLowerCase() === "true" || String(item.status || "").toLowerCase() === "datang";
        const badgeClass = isDatang ? "datang" : "berhalangan";
        const statusText = isDatang ? "Datang" : "Berhalangan";
        const msg = parseMarkdown(escapeHtml(item.comment || item.message || item.text || ""));
        const likes = item.like_count !== undefined ? item.like_count : (item.likes || 0);
        const timeAgo = relativeTime(item.created_at || item.createdAt || item.timestamp || new Date().toISOString());
        const id = item.uuid || item.id || item._id || "";

        return `
        <div class="review-card" data-id="${id}">
            <div class="review-top-meta">
                <span class="review-author">👤 ${escapeHtml(item.name || "Tamu")}</span>
                <span class="badge-status ${badgeClass}">${statusText}</span>
            </div>
            <p class="review-msg-body">${msg}</p>
            <div class="review-footer-row">
                <span>${timeAgo}</span>
                <button class="btn-like-interaction" data-id="${id}">
                    <svg viewBox="0 0 24 24"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
                    <span class="like-count-num">${likes}</span> Like
                </button>
            </div>
        </div>`;
    }).join("");

    // Pagination buttons
    const btnPrev = document.getElementById("btnPrevPage");
    const btnNext = document.getElementById("btnNextPage");
    if (btnPrev) {
        btnPrev.disabled = page <= 1;
        btnPrev.onclick = () => renderCommentPage(comments, page - 1);
    }
    if (btnNext) {
        btnNext.disabled = page >= totalPages;
        btnNext.onclick = () => renderCommentPage(comments, page + 1);
    }

    // Like buttons
    container.querySelectorAll(".btn-like-interaction").forEach(btn => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const commentId = btn.dataset.id;
            if (!commentId) return;
            try {
                await fetch(`${API_BASE}/api/comment/${commentId}`, {
                    method: "POST",
                    headers: { 
                        "Authorization": "Bearer " + INVITATION_KEY
                    }
                });
                const countEl = btn.querySelector(".like-count-num");
                if (countEl) countEl.innerText = parseInt(countEl.innerText || 0) + 1;
                btn.classList.add("liked");
            } catch {
                // toggle visual saja jika API gagal
                btn.classList.toggle("liked");
            }
        });
    });

    // Double-tap card untuk like (mobile)
    container.querySelectorAll(".review-card").forEach(card => {
        let lastTap = 0;
        card.addEventListener("touchend", e => {
            const now = Date.now();
            if (now - lastTap < 300) {
                const likeBtn = card.querySelector(".btn-like-interaction");
                if (likeBtn) likeBtn.click();
                const touch = e.changedTouches[0];
                burstHearts(touch.clientX, touch.clientY);
                e.preventDefault();
            }
            lastTap = now;
        });
    });
}

/* ============================================================
   8. RSVP FORM SUBMIT
   ============================================================ */
function setupRsvpForm() {
    const form = document.getElementById("rsvpForm");
    if (!form) return;

    form.addEventListener("submit", async e => {
        e.preventDefault();
        const submitBtn = form.querySelector("button[type=submit]");
        submitBtn.disabled = true;
        submitBtn.innerText = "Mengirim...";

        const statusVal = form.querySelector("#rsvpStatus").value;
        const payload = {
            name: form.querySelector("#rsvpName").value.trim(),
            presence: statusVal === "Datang",
            comment: form.querySelector("#rsvpMessage").value.trim()
        };

        if (!payload.name || !payload.comment) {
            showToast("⚠️ Semua kolom wajib diisi.");
            submitBtn.disabled = false;
            submitBtn.innerText = "Kirim Ucapan";
            return;
        }

        try {
            const res = await fetch(`${API_BASE}/api/comment`, {
                method: "POST",
                headers: { 
                    "Content-Type": "application/json",
                    "Authorization": "Bearer " + INVITATION_KEY
                },
                body: JSON.stringify(payload)
            });

            if (!res.ok) throw new Error("Server error " + res.status);
            showToast("🎉 Ucapan berhasil dikirim! Terima kasih.");
            form.reset();
            loadComments();

        } catch (err) {
            console.warn("API post failed, saving to localStorage:", err);
            // Fallback localStorage
            const local = JSON.parse(localStorage.getItem("v3_reviews_" + INVITATION_KEY) || "[]");
            local.unshift({
                uuid: "local_" + Date.now(),
                name: payload.name,
                presence: payload.presence,
                comment: payload.comment,
                like_count: 0,
                created_at: new Date().toISOString()
            });
            localStorage.setItem("v3_reviews_" + INVITATION_KEY, JSON.stringify(local));
            showToast("💾 Ucapan tersimpan secara lokal.");
            form.reset();
            loadComments();
        } finally {
            submitBtn.disabled = false;
            submitBtn.innerText = "Kirim Ucapan";
        }
    });
}

/* ============================================================
   9. AUDIO SYSTEM
   ============================================================ */
function initAudio() {
    if (audioInstance) return;
    const musicSrc = INVITATION_CONFIG.musicUrl || INVITATION_CONFIG.musicPath;
    if (!musicSrc) return;

    audioInstance = new Audio(musicSrc);
    audioInstance.loop = true;
    audioInstance.volume = 0.5;

    audioInstance.onerror = () => {
        const btn = document.getElementById("musicToggle");
        if (btn) btn.style.display = "none";
    };

    audioInstance.play().then(() => {
        isAudioPlaying = true;
        const btn = document.getElementById("musicToggle");
        if (btn) { btn.style.display = "flex"; btn.classList.remove("paused"); }
    }).catch(() => {
        const btn = document.getElementById("musicToggle");
        if (btn) btn.style.display = "flex";
    });
}

/* ============================================================
   10. SCROLL REVEAL — IntersectionObserver
   ============================================================ */
function setupScrollReveal() {
    const observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add("active-reveal");
                observer.unobserve(entry.target);
            }
        });
    }, { threshold: 0.12 });

    document.querySelectorAll(".reveal, .reveal-left, .reveal-right").forEach(el => {
        observer.observe(el);
    });
}

/* ============================================================
   11. BOTTOM NAV ACTIVE ON SCROLL
   ============================================================ */
function setupNavActiveScroll() {
    const sections = document.querySelectorAll("section[id]");
    const navItems = document.querySelectorAll(".nav-item");
    const scrollEl = document.getElementById("scrollContainer") || window;

    const activate = () => {
        let currentId = "";
        sections.forEach(sec => {
            const scrollTop = scrollEl === window ? window.scrollY : scrollEl.scrollTop;
            if (sec.offsetTop - 150 <= scrollTop) currentId = sec.id;
        });
        navItems.forEach(item => {
            const href = item.getAttribute("href") || "";
            item.classList.toggle("active", href === "#" + currentId);
        });
    };

    scrollEl.addEventListener("scroll", activate, { passive: true });
    activate();

    // Anchor links smooth scroll
    document.querySelectorAll(".anchor-link, .nav-item").forEach(link => {
        link.addEventListener("click", e => {
            const href = link.getAttribute("href");
            if (href && href.startsWith("#")) {
                e.preventDefault();
                const target = document.getElementById(href.slice(1));
                if (target) {
                    const panel = document.getElementById("scrollContainer");
                    if (panel) {
                        panel.scrollTo({ top: target.offsetTop - 10, behavior: "smooth" });
                    } else {
                        target.scrollIntoView({ behavior: "smooth" });
                    }
                }
            }
        });
    });
}

/* ============================================================
   12. THEME SWITCHER (Dark / Light)
   ============================================================ */
function restoreTheme() {
    const saved = localStorage.getItem("v3_theme") || "dark";
    applyTheme(saved);
}

function applyTheme(theme) {
    document.body.classList.toggle("theme-dark", theme === "dark");
    document.body.classList.toggle("theme-light", theme === "light");

    const btnDark = document.getElementById("btnThemeDark");
    const btnLight = document.getElementById("btnThemeLight");
    if (btnDark) btnDark.classList.toggle("active", theme === "dark");
    if (btnLight) btnLight.classList.toggle("active", theme === "light");

    localStorage.setItem("v3_theme", theme);
}

/* ============================================================
   13. EVENT BINDERS — Master Setup
   ============================================================ */
function setupEventBinders() {
    // Buka Undangan (opening cover)
    const btnOpen = document.getElementById("btnOpenInvitation");
    if (btnOpen) {
        btnOpen.addEventListener("click", () => {
            const cover = document.getElementById("openingCover");
            const main = document.getElementById("mainContainer");
            if (cover) {
                cover.classList.add("fade-out-cinematic");
                setTimeout(() => { cover.style.display = "none"; }, 1200);
            }
            if (main) main.style.display = "flex";
            initAudio();
            confettiBurst();
            setupScrollReveal();
        });
    }

    // Lihat Trailer (buka ke story section)
    const btnTrailer = document.getElementById("btnViewTrailer");
    if (btnTrailer) {
        btnTrailer.addEventListener("click", () => {
            if (btnOpen) btnOpen.click();
            setTimeout(() => {
                const story = document.getElementById("story");
                if (story) story.scrollIntoView({ behavior: "smooth" });
            }, 400);
        });
    }

    // Music Toggle
    const musicBtn = document.getElementById("musicToggle");
    if (musicBtn) {
        musicBtn.addEventListener("click", () => {
            if (!audioInstance) { initAudio(); return; }
            if (audioInstance.paused) {
                audioInstance.play();
                musicBtn.classList.remove("paused");
            } else {
                audioInstance.pause();
                musicBtn.classList.add("paused");
            }
        });
    }

    // Modal close
    const modalClose = document.getElementById("btnModalClose");
    const modal = document.getElementById("mediaModal");
    if (modalClose) modalClose.addEventListener("click", () => modal && modal.classList.remove("open"));
    if (modal) modal.addEventListener("click", e => { if (e.target === modal) modal.classList.remove("open"); });
    document.addEventListener("keydown", e => { if (e.key === "Escape" && modal) modal.classList.remove("open"); });

    // Copy link btn
    document.querySelectorAll(".btn-copy-link").forEach(btn => {
        btn.addEventListener("click", () => {
            navigator.clipboard.writeText(location.href)
                .then(() => showToast("🔗 Link undangan berhasil disalin!"))
                .catch(() => showToast("Gagal menyalin link."));
        });
    });

    // Add to Calendar
    const btnCalendar = document.getElementById("btnAddToCalendar");
    if (btnCalendar) {
        btnCalendar.addEventListener("click", () => {
            const title = encodeURIComponent(INVITATION_CONFIG.couple.metaTitle);
            const loc = encodeURIComponent(INVITATION_CONFIG.schedule.akad.venue + ", " + INVITATION_CONFIG.schedule.akad.address);
            const details = encodeURIComponent("Undangan pernikahan resmi dari " + INVITATION_CONFIG.couple.bride.shortName + " & " + INVITATION_CONFIG.couple.groom.shortName);
            window.open(`https://calendar.google.com/calendar/render?action=TEMPLATE&text=${title}&dates=20261220T030000Z/20261220T083000Z&details=${details}&location=${loc}`, "_blank");
        });
    }

    // Gallery carousel nav
    const track = document.getElementById("galleryTrack");
    const btnPrev = document.getElementById("btnGalleryPrev");
    const btnNext = document.getElementById("btnGalleryNext");
    if (track && btnPrev) btnPrev.addEventListener("click", () => track.scrollBy({ left: -180, behavior: "smooth" }));
    if (track && btnNext) btnNext.addEventListener("click", () => track.scrollBy({ left: 180, behavior: "smooth" }));

    // Theme toggle buttons
    const btnDark = document.getElementById("btnThemeDark");
    const btnLight = document.getElementById("btnThemeLight");
    if (btnDark) btnDark.addEventListener("click", () => applyTheme("dark"));
    if (btnLight) btnLight.addEventListener("click", () => applyTheme("light"));

    // RSVP form
    setupRsvpForm();
}

/* ============================================================
   14. MODAL OPEN
   ============================================================ */
function openModal(src) {
    const modal = document.getElementById("mediaModal");
    const img = document.getElementById("modalTargetImg");
    if (!modal || !img || !src) return;
    img.src = src;
    modal.classList.add("open");
}

/* ============================================================
   15. TOAST NOTIFICATION
   ============================================================ */
function showToast(msg) {
    const toast = document.getElementById("toastNotification");
    if (!toast) return;
    toast.innerText = msg;
    toast.classList.add("show");
    clearTimeout(toast._timer);
    toast._timer = setTimeout(() => toast.classList.remove("show"), 3000);
}

/* ============================================================
   16. CONFETTI BURST (saat buka undangan)
   ============================================================ */
function confettiBurst() {
    const colors = ["#E50914", "#FFD700", "#ffffff", "#46d369", "#00b4d8"];
    const count = 80;
    for (let i = 0; i < count; i++) {
        const el = document.createElement("div");
        el.style.cssText = `
            position:fixed;
            left:${Math.random() * 100}vw;
            top:-10px;
            width:${6 + Math.random() * 8}px;
            height:${6 + Math.random() * 8}px;
            background:${colors[Math.floor(Math.random() * colors.length)]};
            border-radius:${Math.random() > 0.5 ? "50%" : "0"};
            opacity:1;
            z-index:99998;
            pointer-events:none;
            animation:confettiFall ${1.5 + Math.random() * 2}s ease forwards;
            animation-delay:${Math.random() * 0.5}s;
        `;
        document.body.appendChild(el);
        setTimeout(() => el.remove(), 4000);
    }

    if (!document.getElementById("confettiStyle")) {
        const style = document.createElement("style");
        style.id = "confettiStyle";
        style.textContent = `
            @keyframes confettiFall {
                0%   { transform: translateY(0) rotate(0deg); opacity:1; }
                100% { transform: translateY(100vh) rotate(720deg); opacity:0; }
            }
        `;
        document.head.appendChild(style);
    }
}

/* ============================================================
   17. FLOATING HEARTS (double-tap)
   ============================================================ */
function burstHearts(x, y) {
    for (let i = 0; i < 5; i++) {
        const heart = document.createElement("span");
        heart.className = "heart-particle";
        heart.innerText = "❤️";
        heart.style.cssText = `
            left:${x + (Math.random() * 40 - 20)}px;
            top:${y}px;
            font-size:${16 + Math.random() * 14}px;
            animation-duration:${0.8 + Math.random() * 0.5}s;
        `;
        document.body.appendChild(heart);
        setTimeout(() => heart.remove(), 1500);
    }
}

/* ============================================================
   18. UTILITY HELPERS
   ============================================================ */
function escapeHtml(str) {
    return String(str).replace(/[&<>"']/g, c => ({ "&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;" }[c]));
}

function parseMarkdown(str) {
    return str
        .replace(/\*(.*?)\*/g, "<strong>$1</strong>")
        .replace(/_(.*?)_/g, "<em>$1</em>")
        .replace(/~(.*?)~/g, "<del>$1</del>");
}

function relativeTime(ts) {
    const diff = Date.now() - new Date(ts).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return "Baru saja";
    if (mins < 60) return `${mins} menit lalu`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours} jam lalu`;
    const days = Math.floor(hours / 24);
    if (days < 30) return `${days} hari lalu`;
    return new Date(ts).toLocaleDateString("id-ID", { day:"numeric", month:"short", year:"numeric" });
}

// Automatically increment visit counter
(function() {
    const urlParams = new URLSearchParams(window.location.search);
    const slug = urlParams.get('id') || window.location.pathname.split('/').pop().replace('.html', '');
    if (slug) {
        fetch('/server-api.php?action=track-visit&slug=' + slug).catch(e => console.log('visit track error', e));
    }
})();
