'use strict';

const INVITATION_KEY = new URLSearchParams(window.location.search).get("id") || "alya-fajar";
const GUEST_NAME_PARAM = new URLSearchParams(window.location.search).get("to") || "Tamu Undangan";

let INVITATION_DATA = null;
let wishesList = [];
let audioContext = null;
let bgMusic = null;
let isPlaying = false;

const $ = (selector, root = document) => root.querySelector(selector);
const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

function sanitize(text) {
    if (!text) return '';
    const temp = document.createElement('div');
    temp.textContent = text;
    return temp.innerHTML;
}

function formatDate(isoString) {
    try {
        const date = new Date(isoString);
        return date.toLocaleDateString('id-ID', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    } catch (e) {
        return isoString;
    }
}

// Open Sheets / Modals
function openSheet(id) {
    const el = document.getElementById(id);
    if (!el) return;
    el.style.display = 'flex';
    setTimeout(() => {
        el.style.transform = 'translateY(0)';
        el.style.opacity = '1';
    }, 10);
    
    // Disable main page scrolling
    document.body.style.overflow = 'hidden';
}

function closeSheet(id) {
    const el = document.getElementById(id);
    if (!el) return;
    el.style.transform = 'translateY(20px)';
    el.style.opacity = '0';
    setTimeout(() => {
        el.style.display = 'none';
    }, 300);
    
    document.body.style.overflow = '';
}

// 1. Initial Configuration Load
async function loadConfig() {
    try {
        const response = await fetch(`/data/${INVITATION_KEY}.json`);
        if (!response.ok) throw new Error('Failed to load invitation data');
        INVITATION_DATA = await response.json();
        
        console.log('✅ Invitation data loaded:', INVITATION_DATA);
        
        populateThemeData();
        initCountdown();
        initMusic();
        loadWishes();
    } catch (e) {
        console.error('Error loading config:', e);
        alert('Gagal memuat data undangan. Silakan periksa file JSON.');
    }
}

// 2. Populate DOM Data
function populateThemeData() {
    const coupleNameText = `${INVITATION_DATA.brideName} & ${INVITATION_DATA.groomName}`;
    document.title = `The Wedding of ${INVITATION_DATA.brideFullName} & ${INVITATION_DATA.groomFullName}`;
    
    // Sticky demo bar hiding
    const params = new URLSearchParams(window.location.search);
    const isRealCust = params.has('id') && params.get('id') !== 'alya-fajar';
    if (isRealCust) {
        const bar = document.getElementById('sticky-demo-bar');
        if (bar) bar.style.display = 'none';
        document.querySelector('main').style.paddingBottom = '80px';
    } else {
        document.querySelector('main').style.paddingBottom = '160px';
    }

    // Cover Page
    $('#cover-bride-name').textContent = INVITATION_DATA.brideName;
    $('#cover-groom-name').textContent = INVITATION_DATA.groomName;
    
    // Main Headers
    $('#hero-bride-name').textContent = INVITATION_DATA.brideName;
    $('#hero-groom-name').textContent = INVITATION_DATA.groomName;
    $('#closing-names').textContent = coupleNameText;

    // Guest Cover
    $('#guestNameCover').textContent = GUEST_NAME_PARAM;
    $('#rsvp-name').value = GUEST_NAME_PARAM !== 'Tamu Undangan' ? GUEST_NAME_PARAM : '';

    // Photos
    if (INVITATION_DATA.bridePhoto) {
        $('#cover-bride-photo').src = INVITATION_DATA.bridePhoto;
        $('#bride-photo').src = INVITATION_DATA.bridePhoto;
    }
    if (INVITATION_DATA.groomPhoto) {
        $('#cover-groom-photo').src = INVITATION_DATA.groomPhoto;
        $('#groom-photo').src = INVITATION_DATA.groomPhoto;
    }
    if (INVITATION_DATA.coverPhoto || INVITATION_DATA.couplePhoto) {
        const mainP = INVITATION_DATA.coverPhoto || INVITATION_DATA.couplePhoto;
        $('#cover-main-photo').src = mainP;
        $('#hero-photo').src = mainP;
    }

    // Mempelai profiles
    $('#bride-short-name').textContent = INVITATION_DATA.brideName;
    $('#bride-fullname').textContent = INVITATION_DATA.brideFullName;
    $('#bride-parents').textContent = INVITATION_DATA.brideParents;

    $('#groom-short-name').textContent = INVITATION_DATA.groomName;
    $('#groom-fullname').textContent = INVITATION_DATA.groomFullName;
    $('#groom-parents').textContent = INVITATION_DATA.groomParents;

    // Quote / Verse
    if (INVITATION_DATA.openingQuote) {
        $('#quote-text').textContent = INVITATION_DATA.openingQuote;
    }

    // Year
    try {
        const yr = new Date(INVITATION_DATA.eventDateISO).getFullYear();
        if (yr) $('#hero-year').textContent = yr;
    } catch(e) {}

    // Events Details
    const cleanDate = formatDate(INVITATION_DATA.eventDateISO || INVITATION_DATA.eventDateText);
    
    // Akad
    $('#akad-date').textContent = cleanDate;
    $('#akad-time').textContent = `Pukul ${INVITATION_DATA.akadTime || '08.00 WIB'} - selesai`;
    $('#akad-venue').textContent = INVITATION_DATA.venueName;
    $('#akad-address').textContent = INVITATION_DATA.venueAddress;
    $('#akad-maps').href = INVITATION_DATA.mapsUrl;

    // Resepsi
    $('#resepsi-date').textContent = cleanDate;
    $('#resepsi-time').textContent = `Pukul ${INVITATION_DATA.resepsiTime || '11.00 WIB'} - selesai`;
    $('#resepsi-venue').textContent = INVITATION_DATA.venueName;
    $('#resepsi-address').textContent = INVITATION_DATA.venueAddress;
    $('#resepsi-maps').href = INVITATION_DATA.mapsUrl;

    // Timeline stories
    const storyContainer = $('#story-timeline');
    if (storyContainer && INVITATION_DATA.story) {
        storyContainer.innerHTML = '';
        INVITATION_DATA.story.forEach(item => {
            storyContainer.innerHTML += `
                <div class="mb-5 last:mb-0">
                    <p class="pl-marker text-pl-accent text-lg mb-1">${sanitize(item.title)} (${sanitize(item.date)})</p>
                    <p class="font-body text-pl-text text-sm leading-[1.6]">${sanitize(item.description)}</p>
                </div>
            `;
        });
    }

    // Polaroid Gallery
    const galleryContainer = $('#gallery-grid');
    if (galleryContainer && INVITATION_DATA.gallery) {
        galleryContainer.innerHTML = '';
        const captions = ['us ♡', 'candid', 'love', 'happy day', 'forever', 'together', 'magic', 'smile'];
        const tapes = ['pl-tape-pink', 'pl-tape-yellow', 'pl-tape-teal', 'pl-tape-cream'];
        const tilts = ['pl-tilt-l1', 'pl-tilt-r1', 'pl-tilt-l2', 'pl-tilt-r2', 'pl-tilt-0'];

        INVITATION_DATA.gallery.forEach((url, idx) => {
            const cap = captions[idx % captions.length];
            const tape = tapes[idx % tapes.length];
            const tilt = tilts[idx % tilts.length];
            
            galleryContainer.innerHTML += `
                <div class="pl-polaroid ${tilt} relative cursor-pointer" onclick="openGalleryModal(${idx})">
                    <span class="pl-tape ${tape}" style="top:-8px; left:50%; transform:translateX(-50%) rotate(3deg); width:50px;"></span>
                    <img src="${url}" class="w-full aspect-square object-cover pl-photo-sepia" alt="Gallery ${idx+1}">
                    <div class="pl-polaroid-caption pl-handwritten">🌸 ${cap}</div>
                </div>
            `;
        });
    }

    // Banks Angpao info modal
    const giftModalContent = $('#direct-gift-modal .overflow-y-auto');
    if (giftModalContent && INVITATION_DATA.bankName) {
        let banksHtml = `
            <p class="pl-handwritten text-pl-muted text-center text-[18px] mb-4">Silakan transfer tanda kasih secara langsung melalui rekening berikut:</p>
            <div class="space-y-4">
        `;
        
        banksHtml += `
            <div class="pl-sticky pl-sticky-pink text-center">
                <p class="pl-marker text-pl-text text-lg mb-1">${INVITATION_DATA.bankName}</p>
                <p class="pl-handwritten text-pl-text text-2xl font-bold tracking-wider my-2">${INVITATION_DATA.bankNumber}</p>
                <p class="pl-handwritten text-pl-muted text-sm">A/N ${INVITATION_DATA.bankHolder || INVITATION_DATA.brideFullName}</p>
                <button onclick="copyToClipboard('${INVITATION_DATA.bankNumber}', 'Nomor Rekening')" class="pl-handwritten bg-pl-text text-pl-paper text-xs px-4 py-1 mt-3 font-bold transition active:scale-95 hover:bg-pl-accent">copy nomor →</button>
            </div>
        `;
        
        if (INVITATION_DATA.walletName && INVITATION_DATA.walletNumber) {
            banksHtml += `
                <div class="pl-sticky pl-sticky-blue text-center">
                    <p class="pl-marker text-pl-text text-lg mb-1">${INVITATION_DATA.walletName}</p>
                    <p class="pl-handwritten text-pl-text text-2xl font-bold tracking-wider my-2">${INVITATION_DATA.walletNumber}</p>
                    <p class="pl-handwritten text-pl-muted text-sm">A/N ${INVITATION_DATA.bankHolder || INVITATION_DATA.brideFullName}</p>
                    <button onclick="copyToClipboard('${INVITATION_DATA.walletNumber}', 'Nomor E-Wallet')" class="pl-handwritten bg-pl-text text-pl-paper text-xs px-4 py-1 mt-3 font-bold transition active:scale-95 hover:bg-pl-accent">copy nomor →</button>
                </div>
            `;
        }
        
        banksHtml += '</div>';
        giftModalContent.innerHTML = banksHtml;
    }
}

// 3. Countdown Timer
function initCountdown() {
    const targetDate = new Date(INVITATION_DATA.eventDateISO || '2026-12-15T08:00:00+07:00').getTime();
    
    function updateCountdown() {
        const now = new Date().getTime();
        const diff = targetDate - now;
        
        if (diff <= 0) {
            $('#countdown-days').textContent = '00';
            $('#countdown-hours').textContent = '00';
            $('#countdown-minutes').textContent = '00';
            $('#countdown-seconds').textContent = '00';
            clearInterval(timerInterval);
            return;
        }
        
        const d = Math.floor(diff / (1000 * 60 * 60 * 24));
        const h = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const m = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        const s = Math.floor((diff % (1000 * 60)) / 1000);
        
        $('#countdown-days').textContent = String(d).padStart(2, '0');
        $('#countdown-hours').textContent = String(h).padStart(2, '0');
        $('#countdown-minutes').textContent = String(m).padStart(2, '0');
        $('#countdown-seconds').textContent = String(s).padStart(2, '0');
    }
    
    updateCountdown();
    const timerInterval = setInterval(updateCountdown, 1000);
}

// 4. Music Player
function initMusic() {
    bgMusic = document.getElementById('bgMusic');
    if (!bgMusic) return;
    
    bgMusic.src = INVITATION_DATA.musicUrl || "https://media.indoinvite.com/2db3bf1e16cd47a08843bb881e39cce7:indoinvite-staging/indoinvite-staging/indoinvite-staging/nikah/theme/music/1659512405.mp3";
    bgMusic.load();
    
    // Audio Context unlock for mobile autoplay
    window.addEventListener('click', startAudioContext, { once: true });
}

function startAudioContext() {
    if (!audioContext) {
        audioContext = new (window.AudioContext || window.webkitAudioContext)();
    }
    if (audioContext && audioContext.state === 'suspended') {
        audioContext.resume();
    }
}

function toggleMusic() {
    if (!bgMusic) return;
    
    if (isPlaying) {
        bgMusic.pause();
        $('#music-icon').className = 'fa-solid fa-play text-[16px]';
        isPlaying = false;
    } else {
        bgMusic.play().then(() => {
            $('#music-icon').className = 'fa-solid fa-pause text-[16px]';
            isPlaying = true;
        }).catch(e => console.log('Autoplay blocked:', e));
    }
}

function openInvitation() {
    const cover = document.getElementById('welcome-cover');
    cover.style.opacity = '0';
    setTimeout(() => {
        cover.style.display = 'none';
        if (!isPlaying) {
            toggleMusic();
        }
    }, 500);
}

// 5. Load and Submit Comments / Doa
async function loadWishes() {
    try {
        const response = await fetch(`/api/comment`, {
            headers: {
                'Authorization': `Bearer ${INVITATION_KEY}`
            }
        });
        const res = await response.json();
        
        if (res.code === 200) {
            wishesList = res.data.map(c => ({
                name: c.name,
                presence: c.presence,
                text: c.comment,
                timestamp: c.created_at
            }));
            
            renderWishes();
        }
    } catch (e) {
        console.error('Error loading wishes:', e);
    }
}

function renderWishes() {
    const wishesWrap = $('#wishes-list');
    if (!wishesWrap) return;
    
    $('#wishes-count').textContent = wishesList.length;
    wishesWrap.innerHTML = '';
    
    if (wishesList.length === 0) {
        wishesWrap.innerHTML = `
            <div class="text-center py-10 text-pl-muted">
                <i class="fa-regular fa-comments text-4xl mb-3 text-pl-accent/40"></i>
                <p class="pl-handwritten text-pl-text text-[24px]">belum ada pesan ♡</p>
                <p class="pl-handwritten text-[18px] mt-1">jadi yang pertama kasih doa restu!</p>
            </div>
        `;
        return;
    }
    
    wishesList.forEach(w => {
        const presenceBadge = w.presence ? 
            `<span class="pl-handwritten text-[16px] text-emerald-700 bg-emerald-50 px-2.5 py-0.5 border border-emerald-300 font-bold">Hadir ♡</span>` :
            `<span class="pl-handwritten text-[16px] text-pl-muted bg-gray-50 px-2.5 py-0.5 border border-gray-300">Absen</span>`;
            
        wishesWrap.innerHTML += `
            <div class="bg-pl-paper border border-pl-text/10 p-4 shadow-sm hover:shadow-md transition">
                <div class="flex justify-between items-center mb-1">
                    <strong class="pl-marker text-pl-text text-lg">${sanitize(w.name)}</strong>
                    ${presenceBadge}
                </div>
                <p class="pl-handwritten text-pl-text text-[18px] mt-2 leading-[1.4] bg-pl-cream/40 p-2.5 border border-dashed border-pl-text/20">
                    "${sanitize(w.text)}"
                </p>
                <div class="text-right text-[10px] text-pl-muted mt-2">
                    ${new Date(w.timestamp).toLocaleDateString('id-ID', {hour: '2-digit', minute: '2-digit'})}
                </div>
            </div>
        `;
    });
}

// RSVP Stepper / Buttons
function selectPresence(isPresent) {
    $('#rsvp-presence').value = isPresent ? '1' : '0';
    $('#rsvp-btn-hadir').classList.toggle('selected', isPresent);
    $('#rsvp-btn-absen').classList.toggle('selected', !isPresent);
    
    // Hide or show pax container
    $('#rsvp-pax-container').style.display = isPresent ? 'block' : 'none';
}

function adjustPax(val) {
    const paxInput = $('#rsvp-pax');
    let currentVal = parseInt(paxInput.value) || 1;
    currentVal = Math.max(1, Math.min(10, currentVal + val));
    paxInput.value = currentVal;
}

async function submitRsvp(event) {
    event.preventDefault();
    
    const name = $('#rsvp-name').value.trim();
    const presence = $('#rsvp-presence').value === '1';
    const pax = presence ? parseInt($('#rsvp-pax').value) || 1 : 0;
    const comment = $('#rsvp-comment').value.trim();
    
    if (!name || !comment) {
        alert('Mohon isi nama dan ucapan Anda!');
        return;
    }
    
    const submitBtn = $('#rsvp-submit');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Mengirim...';
    
    try {
        const response = await fetch(`/api/comment`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${INVITATION_KEY}`
            },
            body: JSON.stringify({
                name,
                presence,
                pax,
                comment
            })
        });
        
        const res = await response.json();
        
        if (response.ok && res.code === 201) {
            alert('Terima kasih! RSVP dan ucapan Anda berhasil terkirim.');
            $('#rsvp-name').value = '';
            $('#rsvp-comment').value = '';
            closeSheet('rsvp-sheet');
            
            // Reload wishes
            loadWishes();
        } else {
            throw new Error(res.error ? res.error[0] : 'Failed to save RSVP');
        }
    } catch (e) {
        console.error('Error submitting RSVP:', e);
        alert('Gagal mengirim ucapan, silakan coba lagi.');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Kirim Ucapan';
    }
}

// Lightbox Modal for Gallery
let activeGalleryIdx = 0;
function openGalleryModal(idx) {
    if (!INVITATION_DATA || !INVITATION_DATA.gallery) return;
    activeGalleryIdx = idx;
    
    const modal = $('#gallery-modal');
    const img = $('#gallery-modal-img');
    
    img.src = INVITATION_DATA.gallery[idx];
    modal.classList.remove('hidden');
    modal.classList.add('flex');
}

function closeGalleryModal() {
    const modal = $('#gallery-modal');
    modal.classList.add('hidden');
    modal.classList.remove('flex');
}

function prevGalleryImage(event) {
    event.stopPropagation();
    if (!INVITATION_DATA || !INVITATION_DATA.gallery) return;
    activeGalleryIdx = (activeGalleryIdx - 1 + INVITATION_DATA.gallery.length) % INVITATION_DATA.gallery.length;
    $('#gallery-modal-img').src = INVITATION_DATA.gallery[activeGalleryIdx];
}

function nextGalleryImage(event) {
    event.stopPropagation();
    if (!INVITATION_DATA || !INVITATION_DATA.gallery) return;
    activeGalleryIdx = (activeGalleryIdx + 1) % INVITATION_DATA.gallery.length;
    $('#gallery-modal-img').src = INVITATION_DATA.gallery[activeGalleryIdx];
}

// Clipboard copy helper
function copyToClipboard(text, message) {
    navigator.clipboard.writeText(text).then(() => {
        alert(`${message} berhasil disalin!`);
    }).catch(e => {
        console.error('Copy failed:', e);
    });
}

// Scroll Intersection Observer for wiggles/effects
document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
            }
        });
    }, { threshold: 0.1 });
    
    // Watch all pl-reveal and pl-polaroid items for wiggles
    $$('.pl-polaroid, .pl-sticky, .pl-notebook, .pl-reveal').forEach(el => {
        el.classList.add('transition-all', 'duration-700', 'translate-y-6', 'opacity-0');
        observer.observe(el);
    });
    
    // Overwrite class on elements when they become visible
    setInterval(() => {
        $$('.visible').forEach(el => {
            el.style.opacity = '1';
            el.style.transform = 'translateY(0) rotate(var(--tw-rotate, 0deg))';
        });
    }, 100);
});


// Automatically increment visit counter
(function() {
    const urlParams = new URLSearchParams(window.location.search);
    const slug = urlParams.get('id') || window.location.pathname.split('/').pop().replace('.html', '');
    if (slug) {
        fetch('/server-api.php?action=track-visit&slug=' + slug).catch(e => console.log('visit track error', e));
    }
})();
