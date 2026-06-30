const INVITATION_CONFIG = {
  bride:{role:"Bride",name:"Alya Putri Ramadhani",shortName:"Alya",order:"Putri pertama",parents:"Bapak Ahmad & Ibu Siti",photo:"assets/images/bride.jpg",tagline:"A warm melody wrapped in patience and grace."},
  groom:{role:"Groom",name:"Fajar Pratama",shortName:"Fajar",order:"Putra pertama",parents:"Bapak Hasan & Ibu Nur",photo:"assets/images/groom.jpg",tagline:"A steady rhythm that finally found home."},
  coupleName:"Alya & Fajar",
  coupleTagline:"Our Love Story, Now Playing",
  eventDate:"2026-12-20T10:00:00+07:00",
  displayDate:"Sabtu, 20 Desember 2026",
  akad:{title:"Akad Nikah",time:"10.00 WIB - selesai",location:"Masjid Agung",address:"Jl. Contoh Alamat Akad No. 1",maps:"https://maps.google.com/?q=Masjid%20Agung"},
  resepsi:{title:"Resepsi",time:"11.30 WIB - 14.30 WIB",location:"Gedung Serbaguna",address:"Jl. Contoh Alamat Resepsi No. 2",maps:"https://maps.google.com/?q=Gedung%20Serbaguna"},
  locations:{short:"Masjid Agung & Gedung Serbaguna",mainMaps:"https://maps.google.com/?q=Gedung%20Serbaguna"},
  bankAccounts:[{bank:"Bank BCA",number:"1234567890",name:"Alya"},{bank:"Bank Mandiri",number:"9876543210",name:"Fajar"},{bank:"DANA / OVO",number:"08xxxxxxxxxx",name:"Alya / Fajar"}],
  guestParamKeys:["to","kpd","kepada"],
  playlistStory:[
    {title:"First Meet",date:"2021",desc:"Pertemuan sederhana yang tanpa sadar menjadi intro dari cerita panjang.",duration:"03:12"},
    {title:"First Date",date:"2022",desc:"Obrolan kecil, tawa canggung, dan satu langkah awal yang manis.",duration:"04:05"},
    {title:"Falling in Love",date:"2023",desc:"Rasa yang tumbuh pelan, seperti lagu favorit yang diputar berulang.",duration:"03:45"},
    {title:"Engagement",date:"2025",desc:"Janji baik disampaikan, keluarga dipertemukan, doa mulai disatukan.",duration:"04:30"},
    {title:"Wedding Day",date:"2026",desc:"Track terbaik dari album kehidupan kami dimulai hari ini.",duration:"∞"}
  ],
  gallery:[
    {src:"assets/images/gallery-1.jpg",caption:"Album Cover I"},{src:"assets/images/gallery-2.jpg",caption:"Single Moment II"},{src:"assets/images/gallery-3.jpg",caption:"Acoustic Memory III"},{src:"assets/images/gallery-4.jpg",caption:"Final Track IV"}
  ],
  quotes:[
    {title:"QS. Adh-Dhariyat: 49",text:"Dan segala sesuatu Kami ciptakan berpasang-pasangan agar kamu mengingat kebesaran Allah.",source:"QS. Adh-Dhariyat: 49"},
    {title:"Love Note",text:"Cinta terbaik bukan yang paling berisik, tetapi yang tetap tinggal saat dunia kehilangan nada.",source:"Alya & Fajar"}
  ],
  favoriteSong:{title:"Lagu Kenangan Kita",caption:"Satu lagu yang selalu mengingatkan kami bahwa rumah bisa berbentuk seseorang.",cover:"assets/images/cover.jpg"},
  audio:{src:"assets/music.mp3",title:"The Wedding of Alya & Fajar"},
  socialShare:{title:"The Wedding of Alya & Fajar",text:"Our Love Story, Now Playing"},
  theme:{default:"dark",storageKey:"weddingMusicTheme"},
  heroButtons:{calendar:true,maps:true,share:true},
  assets:{cover:"assets/images/cover.jpg",qris:"assets/images/qris.jpg",share:"assets/images/share.jpg"}
};

const $ = (selector, parent=document) => parent.querySelector(selector);
const $$ = (selector, parent=document) => Array.from(parent.querySelectorAll(selector));

const bodyEl = document.body;
const INVITATION_KEY = bodyEl.dataset.key
    || new URLSearchParams(location.search).get("id")
    || "alya-fajar";
const API_BASE = bodyEl.dataset.url
    || "https://intim.my.id";

const STORAGE_KEY = "weddingMusicRSVP";
let audio, isPlaying=false, guestName="Kepada Yth. Bapak/Ibu/Saudara/i", comments=[], commentPage=1, activeGalleryIndex=0, toastReady=false;

window.addEventListener("DOMContentLoaded", async () => {
  const params = new URLSearchParams(window.location.search);
  const invitationId = params.get('id') || 'alya-fajar';
  try {
      const res = await fetch(`/data/${invitationId}.json`);
      if (res.ok) {
          const data = await res.json();
          INVITATION_CONFIG.coupleName = (data.brideName || "Alya") + " & " + (data.groomName || "Fajar");
          INVITATION_CONFIG.coupleTagline = data.coupleTagline || INVITATION_CONFIG.coupleTagline;
          
          INVITATION_CONFIG.bride.name = data.brideFullName || INVITATION_CONFIG.bride.name;
          INVITATION_CONFIG.bride.shortName = data.brideName || INVITATION_CONFIG.bride.shortName;
          INVITATION_CONFIG.bride.parents = data.brideParents || INVITATION_CONFIG.bride.parents;
          INVITATION_CONFIG.bride.photo = data.bridePhoto || INVITATION_CONFIG.bride.photo;
          
          INVITATION_CONFIG.groom.name = data.groomFullName || INVITATION_CONFIG.groom.name;
          INVITATION_CONFIG.groom.shortName = data.groomName || INVITATION_CONFIG.groom.shortName;
          INVITATION_CONFIG.groom.parents = data.groomParents || INVITATION_CONFIG.groom.parents;
          INVITATION_CONFIG.groom.photo = data.groomPhoto || INVITATION_CONFIG.groom.photo;
          
          INVITATION_CONFIG.eventDate = data.eventDateISO || INVITATION_CONFIG.eventDate;
          INVITATION_CONFIG.displayDate = data.eventDateText || INVITATION_CONFIG.displayDate;
          
          INVITATION_CONFIG.akad.time = data.akadTime || INVITATION_CONFIG.akad.time;
          INVITATION_CONFIG.akad.location = data.venueName || INVITATION_CONFIG.akad.location;
          INVITATION_CONFIG.akad.address = data.venueAddress || INVITATION_CONFIG.akad.address;
          INVITATION_CONFIG.akad.maps = data.mapsUrl || INVITATION_CONFIG.akad.maps;
          
          INVITATION_CONFIG.resepsi.time = data.resepsiTime || INVITATION_CONFIG.resepsi.time;
          INVITATION_CONFIG.resepsi.location = data.venueName || INVITATION_CONFIG.resepsi.location;
          INVITATION_CONFIG.resepsi.address = data.venueAddress || INVITATION_CONFIG.resepsi.address;
          INVITATION_CONFIG.resepsi.maps = data.mapsUrl || INVITATION_CONFIG.resepsi.maps;
          
          INVITATION_CONFIG.locations.short = (data.venueName || "Masjid Agung") + " & " + (data.venueName || "Gedung Serbaguna");
          INVITATION_CONFIG.locations.mainMaps = data.mapsUrl || INVITATION_CONFIG.locations.mainMaps;
          
          if (data.story && Array.isArray(data.story)) {
              const defaultDurations = ["03:12", "04:05", "03:45", "04:30", "∞"];
              INVITATION_CONFIG.playlistStory = data.story.map((s, idx) => ({
                  title: s.title,
                  date: s.date,
                  desc: s.description,
                  duration: defaultDurations[idx % defaultDurations.length]
              }));
          }
          if (data.bankName && data.bankNumber) {
              INVITATION_CONFIG.bankAccounts = [
                  { bank: data.bankName, number: data.bankNumber, name: data.bankHolder || data.brideName }
              ];
              if (data.walletName && data.walletNumber) {
                  INVITATION_CONFIG.bankAccounts.push({ bank: data.walletName, number: data.walletNumber, name: data.bankHolder || data.brideName });
              }
          }
          if (INVITATION_CONFIG.quotes && INVITATION_CONFIG.quotes[0]) {
              INVITATION_CONFIG.quotes[0].text = data.openingQuote || INVITATION_CONFIG.quotes[0].text;
          }
          INVITATION_CONFIG.audio.src = data.musicUrl || INVITATION_CONFIG.audio.src;
          INVITATION_CONFIG.audio.title = "The Wedding of " + INVITATION_CONFIG.coupleName;
          INVITATION_CONFIG.socialShare.title = "The Wedding of " + INVITATION_CONFIG.coupleName;
          INVITATION_CONFIG.assets.cover = data.couplePhoto || INVITATION_CONFIG.assets.cover;
          if (data.gallery && Array.isArray(data.gallery)) {
              INVITATION_CONFIG.gallery = data.gallery.map((g, idx) => ({
                  src: typeof g === 'string' ? g : g.src,
                  caption: typeof g === 'string' ? `Album Cover ${idx + 1}` : g.caption
              }));
          }
      }
  } catch (e) {
      console.error("Gagal meload dynamic config:", e);
  }

  hydrateStaticContent();
  initToast();
  initTheme();
  initGuestName();
  initBackgroundFallbacks();
  initOpeningScreen();
  initMusic();
  initCountdown();
  initCoupleProfiles();
  initQuotes();
  initPlaylistStory();
  initGallery();
  initRSVP();
  initGift();
  initScrollSpy();
  initAnimations();
  initParallax();
  initButtons();
  setTimeout(() => $("#loadingScreen")?.classList.add("hide"), 650);
});

function hydrateStaticContent(){
  document.title = `The Wedding of ${INVITATION_CONFIG.coupleName}`;
  $$('[data-couple-name]').forEach(el => el.textContent = INVITATION_CONFIG.coupleName);
  $("#openingDate").textContent = INVITATION_CONFIG.displayDate;
  $("#heroTitle").textContent = `The Wedding of ${INVITATION_CONFIG.coupleName}`;
  $("#heroMeta").textContent = `${INVITATION_CONFIG.displayDate} • ${INVITATION_CONFIG.locations.short}`;
  $("#coupleTagline").textContent = INVITATION_CONFIG.coupleTagline;
  const metaDesc = $('meta[name="description"]'); if(metaDesc) metaDesc.content = `${INVITATION_CONFIG.socialShare.text} - ${INVITATION_CONFIG.coupleName}`;
}

function initGuestName(){
  const params = new URLSearchParams(window.location.search);
  const raw = INVITATION_CONFIG.guestParamKeys.map(k => params.get(k)).find(Boolean);
  const prefix = params.get("prefix") || "Kepada Yth.";
  const clean = raw ? decodeURIComponent(raw).replace(/\+/g," ").trim() : "Bapak/Ibu/Saudara/i";
  guestName = raw ? clean : "Bapak/Ibu/Saudara/i";
  $("#openingGuest").textContent = raw ? `${prefix} ${clean}` : "Kepada Yth. Bapak/Ibu/Saudara/i";
  $("#heroGuest").textContent = `Dear, ${guestName}`;
  $("#rsvpGuestText").textContent = `Dear, ${guestName}. Terima kasih sudah menjadi bagian dari playlist hidup kami.`;
  const input = $("#guestInput"); if(input && raw) input.value = clean;
}

function initOpeningScreen(){
  $("#openInvitationBtn")?.addEventListener("click", openInvitation);
  $("#playStoryBtn")?.addEventListener("click", () => { openInvitation(); setTimeout(() => document.getElementById("story")?.scrollIntoView({behavior:"smooth"}), 550); });
}

function openInvitation(){
  document.body.classList.remove("no-scroll");
  $("#openingScreen")?.classList.add("closed");
  ["#app","#mainNav","#musicPlayer"].forEach(id => { const el=$(id); el?.classList.remove("hidden-ui"); el?.classList.add("visible-ui"); });
  toggleMusic(true);
  launchSparkles();
  showToast("Undangan dibuka. Kisah cinta mulai diputar.", "success");
}

function initMusic(){
  audio = $("#mainAudio");
  if(!audio) return;
  audio.src = INVITATION_CONFIG.audio.src;
  audio.loop = true;
  audio.volume = .72;
  $("#musicToggleBtn")?.addEventListener("click", () => toggleMusic());
  $("#muteBtn")?.addEventListener("click", toggleMute);
  $$('[data-music-toggle]').forEach(btn => btn.addEventListener("click", () => toggleMusic()));
  audio.addEventListener("timeupdate", updatePlayerUI);
  audio.addEventListener("play", () => { isPlaying=true; updatePlayerUI(); });
  audio.addEventListener("pause", () => { isPlaying=false; updatePlayerUI(); });
  audio.addEventListener("error", () => {
    const fallbackUrl = "https://media.indoinvite.com/2db3bf1e16cd47a08843bb881e39cce7:indoinvite-staging/indoinvite-staging/indoinvite-staging/nikah/theme/music/1659512405.mp3";
    console.warn("Local music failed, using online fallback:", fallbackUrl);
    audio.src = fallbackUrl;
    audio.load();
  });
  updatePlayerUI();
}

function toggleMusic(forcePlay){
  if(!audio) return;
  if(forcePlay === true || !isPlaying){
    audio.play().catch(() => showToast("Tap tombol play kalau browser ngeblok autoplay. Teknologi kadang sok suci.", "info"));
  } else audio.pause();
  setTimeout(updatePlayerUI, 80);
}

function toggleMute(){
  if(!audio) return;
  audio.muted = !audio.muted;
  $("#muteBtn").textContent = audio.muted ? "×" : "♪";
  showToast(audio.muted ? "Musik dimatikan" : "Musik dinyalakan", "success");
}

function updatePlayerUI(){
  const playIcon = isPlaying ? "⏸" : "▶";
  $("#musicToggleBtn") && ($("#musicToggleBtn").textContent = playIcon);
  $$('[data-music-toggle]').forEach(btn => btn.textContent = playIcon);
  $("#playerStatus") && ($("#playerStatus").textContent = isPlaying ? "Now Playing" : "Paused");
  const progress = audio && audio.duration ? (audio.currentTime / audio.duration) * 100 : isPlaying ? 42 : 0;
  $("#playerProgress") && ($("#playerProgress").style.width = `${Math.min(progress,100)}%`);
}

function initCountdown(){
  const target = new Date(INVITATION_CONFIG.eventDate).getTime();
  const box = $("#countdown");
  const render = () => {
    const diff = target - Date.now();
    if(diff <= 0){
      box.innerHTML = ["0","0","0","0"].map((v,i)=>`<div class="count-box"><strong>${v}</strong><span>${["Hari","Jam","Menit","Detik"][i]}</span></div>`).join("");
      $("#countdownStatus").textContent = "Acara telah berlangsung • This track has already been played";
      return;
    }
    const d = Math.floor(diff/86400000), h = Math.floor(diff/3600000)%24, m = Math.floor(diff/60000)%60, s = Math.floor(diff/1000)%60;
    box.innerHTML = [[d,"Hari"],[h,"Jam"],[m,"Menit"],[s,"Detik"]].map(([v,l])=>`<div class="count-box"><strong>${String(v).padStart(2,"0")}</strong><span>${l}</span></div>`).join("");
    $("#countdownStatus").textContent = "Menuju track terbaik kami.";
  };
  render(); setInterval(render,1000);
  const tracks = [{no:"01",...INVITATION_CONFIG.akad},{no:"02",...INVITATION_CONFIG.resepsi}];
  $("#eventTracks").innerHTML = tracks.map(t => `<article class="track-row"><span class="track-no">Track ${t.no}</span><div class="track-main"><h3>${t.title}</h3><p>${t.time} • ${t.location}<br>${t.address}</p></div><a class="btn btn-ghost" target="_blank" rel="noopener" href="${t.maps}">Maps</a></article>`).join("");
}

function addToCalendar(){
  const start = new Date(INVITATION_CONFIG.eventDate);
  const end = new Date(start.getTime()+4.5*60*60*1000);
  const fmt = d => d.toISOString().replace(/[-:]/g,"").split(".")[0]+"Z";
  const title = encodeURIComponent(`The Wedding of ${INVITATION_CONFIG.coupleName}`);
  const details = encodeURIComponent(`Undangan pernikahan ${INVITATION_CONFIG.coupleName}. ${INVITATION_CONFIG.coupleTagline}`);
  const location = encodeURIComponent(INVITATION_CONFIG.locations.short);
  const url = `https://calendar.google.com/calendar/render?action=TEMPLATE&text=${title}&dates=${fmt(start)}/${fmt(end)}&details=${details}&location=${location}`;
  window.open(url,"_blank","noopener");
}

function initCoupleProfiles(){
  const card = person => `<div class="artist-photo image-fallback" data-bg="${person.photo}"></div><div><span class="eyebrow">${person.role}</span><h3>${person.name}</h3><p>${person.order} dari ${person.parents}</p><p>${person.tagline}</p></div>`;
  $("#brideCard").innerHTML = card(INVITATION_CONFIG.bride);
  $("#groomCard").innerHTML = card(INVITATION_CONFIG.groom);
  initBackgroundFallbacks();
}

function initQuotes(){
  $("#quotesGrid").innerHTML = INVITATION_CONFIG.quotes.map(q => `<article class="quote-card glass-card"><span class="eyebrow">${q.title}</span><blockquote>“${q.text}”</blockquote><cite>${q.source}</cite></article>`).join("");
}

function initPlaylistStory(){
  $("#playlistStory").innerHTML = INVITATION_CONFIG.playlistStory.map((item,i)=>`<article class="playlist-row"><span class="track-no">${String(i+1).padStart(2,"0")}</span><div class="playlist-main"><h3>Track ${String(i+1).padStart(2,"0")} — ${item.title}</h3><p>${item.date} • ${item.desc}</p></div><span class="track-duration">${item.duration}</span></article>`).join("");
  const song = INVITATION_CONFIG.favoriteSong;
  $("#favoriteSong").innerHTML = `<div class="favorite-cover image-fallback" data-bg="${song.cover}"></div><div><span class="eyebrow">Our Favorite Song</span><h3>${song.title}</h3><p>${song.caption}</p><button class="btn btn-primary" data-music-toggle>Putar / Hentikan Musik</button></div>`;
  initBackgroundFallbacks();
  $$('[data-music-toggle]').forEach(btn => btn.onclick = () => toggleMusic());
}

function initGallery(){
  const gallerySection = $("#gallery");
  if ((!INVITATION_CONFIG.gallery || INVITATION_CONFIG.gallery.length === 0) && gallerySection) {
    gallerySection.style.display = "none";
    const navBtn = document.querySelector('[href="#gallery"]');
    if (navBtn) navBtn.style.display = 'none';
    return;
  }
  const items = INVITATION_CONFIG.gallery.map((g,i)=>`<button class="gallery-item image-fallback" data-bg="${g.src}" data-gallery-index="${i}" aria-label="${g.caption}" loading="lazy"></button>`).join("");
  $("#galleryGrid").innerHTML = items;
  $("#galleryCarousel").innerHTML = items;
  $$('[data-gallery-index]').forEach(btn => btn.addEventListener("click", () => openImageModal(Number(btn.dataset.galleryIndex))));
  $("#viewGalleryBtn")?.addEventListener("click", () => openImageModal(0));
  initBackgroundFallbacks();
}

function openImageModal(index=0, customSrc=null, caption=""){
  activeGalleryIndex = index;
  const data = customSrc ? {src:customSrc, caption} : INVITATION_CONFIG.gallery[index];
  if(!data) return;
  $("#modalImage").src = data.src;
  $("#modalCaption").textContent = data.caption || "Preview";
  $("#modal").classList.add("open");
  $("#modal").setAttribute("aria-hidden","false");
}

function closeModal(){
  $("#modal").classList.remove("open");
  $("#modal").setAttribute("aria-hidden","true");
}

async function loadComments() {
  const wrapper = $("#commentsList");
  if (!wrapper) return;

  wrapper.innerHTML = `<article class="comment-card"><p>Memuat ucapan dari playlist...</p></article>`;

  try {
    const res = await fetch(`${API_BASE}/api/comment`, {
      headers: {
        "Authorization": "Bearer " + INVITATION_KEY
      }
    });
    if (!res.ok) throw new Error("API error " + res.status);
    const json = await res.json();
    const commentsList = (json.data && json.data.lists) || json.data || [];
    comments = commentsList;
    renderComments();
  } catch (err) {
    console.warn("API load failed, using local fallback:", err);
    comments = JSON.parse(localStorage.getItem(STORAGE_KEY) || "[]");
    renderComments();
  }
}

function initRSVP() {
  $("#rsvpForm")?.addEventListener("submit", saveComment);
  $("#commentSearch")?.addEventListener("input", () => { commentPage = 1; renderComments(); });
  $("#commentFilter")?.addEventListener("change", filterComments);
  loadComments();
}

async function saveComment(e) {
  e.preventDefault();
  const name = $("#guestInput").value.trim();
  const status = $("#statusInput").value;
  const total = $("#totalGuestInput").value.trim();
  const message = $("#messageInput").value.trim();
  if (!name || !message) return showToast("Nama dan ucapan wajib diisi.", "info");

  const submitBtn = $("#rsvpForm button[type=submit]");
  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.textContent = "Mengirim...";
  }

  const payload = {
    name,
    presence: status === "Datang",
    comment: message + (total ? ` (Membawa ${total} tamu)` : "")
  };

  try {
    const res = await fetch(`${API_BASE}/api/comment`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": "Bearer " + INVITATION_KEY
      },
      body: JSON.stringify(payload)
    });
    if (!res.ok) throw new Error("API error " + res.status);
    showToast("Ucapan berhasil disimpan ke playlist", "success");
    e.target.reset();
    $("#guestInput").value = guestName !== "Bapak/Ibu/Saudara/i" ? guestName : "";
    commentPage = 1;
    loadComments();
  } catch (err) {
    console.warn("API post failed, saving locally:", err);
    comments.unshift({
      uuid: "local_" + Date.now(),
      name: payload.name,
      presence: payload.presence,
      comment: payload.comment,
      like_count: 0,
      created_at: new Date().toISOString()
    });
    localStorage.setItem(STORAGE_KEY, JSON.stringify(comments));
    e.target.reset();
    $("#guestInput").value = guestName !== "Bapak/Ibu/Saudara/i" ? guestName : "";
    commentPage = 1;
    renderComments();
    showToast("💾 Ucapan tersimpan di local storage.");
  } finally {
    if (submitBtn) {
      submitBtn.disabled = false;
      submitBtn.textContent = "Kirim Ucapan";
    }
  }
}

function renderComments() {
  const query = ($("#commentSearch")?.value || "").toLowerCase();
  const filter = $("#commentFilter")?.value || "all";
  
  let filteredData = comments.filter(c => {
    const isDatang = c.presence === true || String(c.presence).toLowerCase() === "true" || String(c.status || "").toLowerCase() === "datang";
    const statusText = isDatang ? "Datang" : "Berhalangan";
    
    const matchesFilter = filter === "all" || statusText === filter || (filter === "Konfirmasi Presensi" && statusText !== "Datang" && statusText !== "Berhalangan");
    const msg = c.comment || c.message || c.text || "";
    const matchesQuery = `${c.name} ${msg}`.toLowerCase().includes(query);
    return matchesFilter && matchesQuery;
  });

  const perPage = 5, pages = Math.max(1, Math.ceil(filteredData.length / perPage));
  commentPage = Math.min(commentPage, pages);
  const start = (commentPage - 1) * perPage;
  const pageItems = filteredData.slice(start, start + perPage);

  $("#commentsList").innerHTML = pageItems.length ? pageItems.map(c => {
    const isDatang = c.presence === true || String(c.presence).toLowerCase() === "true" || String(c.status || "").toLowerCase() === "datang";
    const statusText = isDatang ? "Datang" : "Berhalangan";
    const badgeClass = isDatang ? "datang" : "berhalangan";
    const id = c.uuid || c.id || "";
    const likes = c.like_count !== undefined ? c.like_count : (c.likes || 0);
    const dateText = c.created_at || c.createdAt || c.timestamp;
    const formattedDate = dateText ? new Date(dateText).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" }) : c.date || "Baru saja";
    const msg = formatMessage(c.comment || c.message || c.text || "");
    const likedClass = c.liked ? "liked" : "";

    return `
      <article class="comment-card" data-comment-id="${id}">
        <div>
          <strong>${escapeHTML(c.name || "Tamu")}</strong> 
          <span class="badge ${badgeClass}">${statusText}</span>
          <p>${msg}</p>
          <small>${formattedDate}</small>
        </div>
        <button class="like-btn ${likedClass}" data-like="${id}">♥ <span class="like-count-num">${likes}</span></button>
      </article>
    `;
  }).join("") : `<article class="comment-card"><p>Belum ada ucapan. Sepi amat, kayak grup keluarga pas diminta patungan.</p></article>`;

  $("#commentPagination").innerHTML = Array.from({ length: pages }, (_, i) => `<button class="${commentPage === i + 1 ? "active" : ""}" data-page="${i + 1}">${i + 1}</button>`).join("");
  
  $$('[data-page]').forEach(b => b.onclick = () => { commentPage = Number(b.dataset.page); renderComments(); });
  $$('[data-like]').forEach(b => b.onclick = () => likeComment(b.dataset.like, b));
  $$('.comment-card[data-comment-id]').forEach(card => card.ondblclick = () => {
    const likeBtn = card.querySelector('[data-like]');
    if (likeBtn) likeComment(card.dataset.commentId, likeBtn);
  });
  renderStats();
}

function filterComments() { commentPage = 1; renderComments(); }

async function likeComment(id, buttonEl) {
  if (!id) return;
  try {
    await fetch(`${API_BASE}/api/comment/${id}`, {
      method: "POST",
      headers: {
        "Authorization": "Bearer " + INVITATION_KEY
      }
    });
    if (buttonEl) {
      buttonEl.classList.add("liked");
      const countEl = buttonEl.querySelector(".like-count-num") || buttonEl;
      countEl.textContent = parseInt(countEl.textContent || 0) + 1;
    }
  } catch {
    if (buttonEl) buttonEl.classList.toggle("liked");
  }
}

function renderStats() {
  const total = comments.length;
  const datang = comments.filter(c => c.presence === true || String(c.presence).toLowerCase() === "true" || String(c.status || "").toLowerCase() === "datang").length;
  const off = total - datang;
  $("#rsvpStats").innerHTML = [[total, "Total"], [datang, "Datang"], [off, "Berhalangan"]].map(([n, l]) => `<div class="stat-card"><strong>${n}</strong><span>${l}</span></div>`).join("");
}

function formatMessage(text){
  return escapeHTML(text)
    .replace(/\*([^*]+)\*/g,"<strong>$1</strong>")
    .replace(/_([^_]+)_/g,"<em>$1</em>")
    .replace(/\/([^/]+)\//g,"<s>$1</s>");
}
function escapeHTML(s){ return s.replace(/[&<>"]/g, m => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"}[m])); }

function initGift(){
  $("#giftGrid").innerHTML = INVITATION_CONFIG.bankAccounts.map(a => `<article class="gift-card glass-card"><span class="eyebrow">${a.bank}</span><strong>${a.number}</strong><p>a.n. ${a.name}</p><button class="btn btn-ghost copy-btn" data-copy="${a.number}">Copy</button></article>`).join("");
  $$('[data-copy]').forEach(btn => btn.onclick = () => copyText(btn.dataset.copy));
  $("#copyAllGiftBtn")?.addEventListener("click", () => copyText(INVITATION_CONFIG.bankAccounts.map(a=>`${a.bank}: ${a.number} a.n. ${a.name}`).join("\n")));
  $("#qrisBtn")?.addEventListener("click", () => openImageModal(0, INVITATION_CONFIG.assets.qris, "QRIS Wedding Gift"));
}

function copyText(text){
  navigator.clipboard?.writeText(text).then(()=>showToast("Berhasil disalin", "success")).catch(()=>{
    const area=document.createElement("textarea"); area.value=text; document.body.appendChild(area); area.select(); document.execCommand("copy"); area.remove(); showToast("Berhasil disalin", "success");
  });
}

function shareInvitation(){
  const data = {title:INVITATION_CONFIG.socialShare.title,text:INVITATION_CONFIG.socialShare.text,url:location.href};
  if(navigator.share) navigator.share(data).catch(()=>copyText(location.href)); else copyText(location.href);
}

function initTheme(){
  const saved = localStorage.getItem(INVITATION_CONFIG.theme.storageKey) || INVITATION_CONFIG.theme.default;
  if(saved === "soft") document.body.classList.add("soft-theme");
  updateThemeButton();
  $("#themeToggle")?.addEventListener("click", () => {
    document.body.classList.toggle("soft-theme");
    localStorage.setItem(INVITATION_CONFIG.theme.storageKey, document.body.classList.contains("soft-theme") ? "soft" : "dark");
    updateThemeButton();
  });
}
function updateThemeButton(){ const btn=$("#themeToggle"); if(btn) btn.textContent = document.body.classList.contains("soft-theme") ? "Mode: Soft Romantic" : "Mode: Dark Player"; }

function initScrollSpy(){
  const links = $$(".nav-link");
  const sections = links.map(a => $(a.getAttribute("href"))).filter(Boolean);
  const obs = new IntersectionObserver(entries => entries.forEach(entry => {
    if(entry.isIntersecting){ links.forEach(l=>l.classList.toggle("active", l.getAttribute("href") === `#${entry.target.id}`)); }
  }), {threshold:.35});
  sections.forEach(s => obs.observe(s));
  $("#backToTop")?.addEventListener("click", () => scrollTo({top:0,behavior:"smooth"}));
  addEventListener("scroll", () => $("#backToTop")?.classList.toggle("show", scrollY > 700), {passive:true});
}

function initAnimations(){
  const obs = new IntersectionObserver(entries => entries.forEach(e => { if(e.isIntersecting) e.target.classList.add("is-visible"); }), {threshold:.14});
  $$(".reveal").forEach(el => obs.observe(el));
}

function initToast(){ toastReady = true; }
function showToast(message, type="success"){
  const box = $("#toastContainer"); if(!box || !toastReady) return;
  const toast = document.createElement("div"); toast.className = `toast ${type}`; toast.textContent = message; box.appendChild(toast);
  setTimeout(()=>{toast.style.opacity="0"; toast.style.transform="translateY(8px)"; setTimeout(()=>toast.remove(),250);}, 2800);
}

function initParallax(){
  addEventListener("scroll", () => {
    $$(".bg-layer").forEach(el => {
      const rect = el.getBoundingClientRect();
      if(rect.bottom > 0 && rect.top < innerHeight) el.style.backgroundPosition = `center ${50 + rect.top * -0.015}%`;
    });
  }, {passive:true});
}

function initBackgroundFallbacks(){
  $$('[data-bg]').forEach(el => {
    const src = el.dataset.bg;
    const img = new Image();
    img.onload = () => { el.style.backgroundImage = `url("${src}")`; };
    img.onerror = () => el.classList.add("fallback-active");
    img.src = src;
  });
}

function initButtons(){
  $("#calendarBtn")?.addEventListener("click", addToCalendar);
  $("#locationBtn")?.addEventListener("click", () => window.open(INVITATION_CONFIG.locations.mainMaps,"_blank","noopener"));
  $("#shareBtn")?.addEventListener("click", shareInvitation);
  $("#coverPreviewBtn")?.addEventListener("click", () => openImageModal(0, INVITATION_CONFIG.assets.cover, "Wedding Album Cover"));
  $$('[data-modal-close]').forEach(el => el.addEventListener("click", closeModal));
  $("#modalNext")?.addEventListener("click", () => openImageModal((activeGalleryIndex+1)%INVITATION_CONFIG.gallery.length));
  $("#modalPrev")?.addEventListener("click", () => openImageModal((activeGalleryIndex-1+INVITATION_CONFIG.gallery.length)%INVITATION_CONFIG.gallery.length));
  addEventListener("keydown", e => { if(e.key === "Escape") closeModal(); if(e.key === "ArrowRight") $("#modalNext")?.click(); if(e.key === "ArrowLeft") $("#modalPrev")?.click(); });
  let sx=0; $("#modal")?.addEventListener("touchstart", e => sx=e.touches[0].clientX, {passive:true});
  $("#modal")?.addEventListener("touchend", e => { const dx=e.changedTouches[0].clientX-sx; if(Math.abs(dx)>50) (dx<0?$("#modalNext"):$("#modalPrev"))?.click(); }, {passive:true});
}

function launchSparkles(){
  const canvas = $("#sparkleCanvas"), ctx = canvas.getContext("2d");
  canvas.width = innerWidth; canvas.height = innerHeight;
  const parts = Array.from({length:70},()=>({x:innerWidth/2,y:innerHeight*.45,vx:(Math.random()-.5)*8,vy:(Math.random()-.8)*8,life:70+Math.random()*40,size:2+Math.random()*4}));
  let frame=0;
  function draw(){
    ctx.clearRect(0,0,canvas.width,canvas.height);
    parts.forEach(p=>{p.x+=p.vx;p.y+=p.vy;p.vy+=.09;p.life-=1;ctx.globalAlpha=Math.max(p.life/90,0);ctx.fillStyle=Math.random()>.45?"#1DB954":"#F1D7D7";ctx.beginPath();ctx.arc(p.x,p.y,p.size,0,Math.PI*2);ctx.fill();});
    frame++; if(frame<100) requestAnimationFrame(draw); else ctx.clearRect(0,0,canvas.width,canvas.height);
  }
  draw();
}


// Automatically increment visit counter
(function() {
    const urlParams = new URLSearchParams(window.location.search);
    const slug = urlParams.get('id') || window.location.pathname.split('/').pop().replace('.html', '');
    if (slug) {
        fetch('/server-api.php?action=track-visit&slug=' + slug).catch(e => console.log('visit track error', e));
    }
})();
