(function() {
  // Configuration
  const waNumber = '6285798644642';
  const welcomeMsg = 'Halo! Ada yang bisa kami bantu? Tanya aja tentang fitur, harga, atau cara bikin undangan. Kami siap bantu!';
  
  // Local Database of automated FAQ responses
  const faqAnswers = {
    'Berapa harga paketnya?': 'Harga paket undangan kami flat <strong>Rp 35.000</strong> (Promo diskon dari Rp 59.000) kak. Sudah termasuk semua fitur premium tanpa biaya tambahan!',
    'Bisa custom tema sendiri?': 'Bisa banget kak! Kakak bebas ganti foto, musik latar, teks, warna, hingga kustomisasi doa/ucapan. Tim desainer kami juga siap membantu jika ada request khusus.',
    'Cara kirim via WhatsApp?': 'Setelah mengisi form order awal, sistem kami akan langsung mengarahkan kakak ke WhatsApp. Kakak tinggal kirimkan foto-foto prewedding dan detail data acaranya via WA tersebut.',
    'Berapa lama prosesnya?': 'Proses pengerjaannya sangat cepat kak, biasanya berkisar antara <strong>1 hingga 3 jam saja</strong> setelah semua data dan bahan kami terima dengan lengkap.'
  };

  const faqs = Object.keys(faqAnswers);

  // Inject styles automatically
  const style = document.createElement('style');
  style.innerHTML = `
    /* Floating Chat Widget Styles */
    .chat-widget-btn {
      position: fixed;
      bottom: 24px;
      right: 24px;
      width: 60px;
      height: 60px;
      background-color: #0f172a; /* Slate 900 */
      color: #ffffff;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 4px 12px rgba(0,0,0,0.25);
      cursor: pointer;
      z-index: 9999;
      transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1), background-color 0.2s;
    }
    .chat-widget-btn:hover {
      transform: scale(1.05);
      background-color: #1e293b;
    }
    .chat-widget-btn svg {
      width: 28px;
      height: 28px;
      fill: currentColor;
    }
    .chat-widget-badge {
      position: absolute;
      top: -2px;
      right: -2px;
      background-color: #ef4444; /* Red 500 */
      color: white;
      font-size: 11px;
      font-weight: bold;
      width: 20px;
      height: 20px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      border: 2px solid #0f172a;
    }
    .chat-widget-box {
      position: fixed;
      bottom: 96px;
      right: 24px;
      width: 360px;
      max-height: 520px;
      background-color: #ffffff;
      border-radius: 16px;
      box-shadow: 0 8px 30px rgba(0,0,0,0.15);
      display: flex;
      flex-direction: column;
      overflow: hidden;
      z-index: 9999;
      transform: translateY(20px);
      opacity: 0;
      pointer-events: none;
      transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.25s cubic-bezier(0.4, 0, 0.2, 1);
      font-family: sans-serif;
    }
    .chat-widget-box.open {
      transform: translateY(0);
      opacity: 1;
      pointer-events: auto;
    }
    .chat-widget-header {
      background-color: #0f172a;
      color: #ffffff;
      padding: 16px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 1px solid rgba(255,255,255,0.1);
    }
    .chat-widget-profile {
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .chat-widget-avatar {
      width: 36px;
      height: 36px;
      background-color: #ffffff; /* White background for logo */
      border-radius: 50%;
      display: block;
      object-fit: contain;
      padding: 5px;
      box-sizing: border-box;
    }
    .chat-widget-title-container {
      display: flex;
      flex-direction: column;
    }
    .chat-widget-name {
      font-weight: bold;
      font-size: 14px;
      line-height: 1.2;
    }
    .chat-widget-status {
      font-size: 11px;
      opacity: 0.8;
    }
    .chat-widget-close {
      cursor: pointer;
      opacity: 0.7;
      transition: opacity 0.2s;
    }
    .chat-widget-close:hover {
      opacity: 1;
    }
    .chat-widget-close svg {
      width: 20px;
      height: 20px;
      fill: currentColor;
    }
    .chat-widget-body {
      padding: 16px;
      overflow-y: auto;
      flex-grow: 1;
      background-color: #f8fafc;
      display: flex;
      flex-direction: column;
      gap: 12px;
      max-height: 360px;
      box-sizing: border-box;
    }
    .chat-widget-bubble {
      background-color: #ffffff;
      border: 1px solid #e2e8f0;
      border-radius: 0 16px 16px 16px;
      padding: 12px;
      font-size: 13px;
      color: #334155;
      line-height: 1.5;
      max-width: 85%;
      box-shadow: 0 2px 4px rgba(0,0,0,0.02);
      box-sizing: border-box;
    }
    .chat-widget-bubble.user {
      align-self: flex-end;
      background-color: #0f172a;
      color: #ffffff;
      border-radius: 16px 0 16px 16px;
      border: none;
    }
    .chat-widget-time {
      font-size: 10px;
      color: #94a3b8;
      margin-top: -6px;
    }
    .chat-widget-time.user {
      align-self: flex-end;
      margin-right: 2px;
    }
    .chat-widget-label {
      font-size: 11px;
      color: #64748b;
      font-weight: 600;
      margin-top: 8px;
    }
    .chat-widget-faq-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }
    .chat-widget-faq-btn {
      background-color: #ffffff;
      border: 1px solid #e2e8f0;
      color: #0f172a;
      padding: 10px 14px;
      border-radius: 10px;
      font-size: 13px;
      text-align: left;
      cursor: pointer;
      transition: all 0.2s;
      font-weight: 500;
      outline: none;
    }
    .chat-widget-faq-btn:hover {
      background-color: #f1f5f9;
      border-color: #cbd5e1;
      transform: translateY(-1px);
    }
    .chat-widget-footer {
      padding: 12px;
      border-top: 1px solid #e2e8f0;
      background-color: #ffffff;
      display: flex;
      align-items: center;
      gap: 8px;
      box-sizing: border-box;
    }
    .chat-widget-input {
      flex-grow: 1;
      border: 1px solid #cbd5e1;
      border-radius: 20px;
      padding: 8px 16px;
      font-size: 13px;
      outline: none;
      transition: border-color 0.2s;
      color: #0f172a;
      box-sizing: border-box;
      background-color: #ffffff;
    }
    .chat-widget-input:focus {
      border-color: #0f172a;
    }
    .chat-widget-send-btn {
      width: 36px;
      height: 36px;
      border-radius: 50%;
      background-color: #f1f5f9;
      color: #94a3b8;
      border: none;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: default;
      transition: all 0.2s;
      padding: 0;
      outline: none;
    }
    .chat-widget-send-btn.active {
      background-color: #0f172a;
      color: #ffffff;
      cursor: pointer;
    }
    .chat-widget-send-btn svg {
      width: 18px;
      height: 18px;
      fill: currentColor;
    }

    /* Typing indicator anim */
    .chat-widget-typing {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 12px 16px;
      background-color: #ffffff;
      border: 1px solid #e2e8f0;
      border-radius: 0 16px 16px 16px;
      width: fit-content;
      box-shadow: 0 2px 4px rgba(0,0,0,0.02);
    }
    .chat-widget-typing span {
      width: 6px;
      height: 6px;
      background-color: #94a3b8;
      border-radius: 50%;
      display: inline-block;
      animation: chat-widget-bounce 1.4s infinite both;
    }
    .chat-widget-typing span:nth-child(2) {
      animation-delay: .2s;
    }
    .chat-widget-typing span:nth-child(3) {
      animation-delay: .4s;
    }
    @keyframes chat-widget-bounce {
      0%, 80%, 100% { transform: scale(0); }
      40% { transform: scale(1); }
    }

    @media (max-width: 480px) {
      .chat-widget-box {
        bottom: 84px;
        right: 16px;
        left: 16px;
        width: auto;
        max-height: 480px;
      }
      .chat-widget-btn {
        bottom: 16px;
        right: 16px;
      }
    }
  `;
  document.head.appendChild(style);

  // Create Widget Elements
  const widgetContainer = document.createElement('div');
  widgetContainer.id = 'chat-widget-root';
  
  // Floating Button HTML
  widgetContainer.innerHTML = `
    <div class="chat-widget-btn" id="chatWidgetBtn">
      <svg viewBox="0 0 24 24">
        <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
      </svg>
      <div class="chat-widget-badge">1</div>
    </div>
    
    <div class="chat-widget-box" id="chatWidgetBox">
      <div class="chat-widget-header">
        <div class="chat-widget-profile">
          <img src="/assets/images/favicon.png" class="chat-widget-avatar" alt="Logo">
          <div class="chat-widget-title-container">
            <span class="chat-widget-name">Tim Undangan Intim</span>
            <span class="chat-widget-status">Siap bantu kamu</span>
          </div>
        </div>
        <div class="chat-widget-close" id="chatWidgetClose">
          <svg viewBox="0 0 24 24">
            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
          </svg>
        </div>
      </div>
      
      <div class="chat-widget-body" id="chatWidgetBody">
        <div class="chat-widget-bubble">${welcomeMsg}</div>
        <span class="chat-widget-time">Baru saja</span>
        
        <div class="chat-widget-label" id="faqLabel">Yang sering ditanyakan:</div>
        <div class="chat-widget-faq-list" id="chatWidgetFaqList"></div>
      </div>
      
      <div class="chat-widget-footer">
        <input type="text" class="chat-widget-input" id="chatWidgetInput" placeholder="Tulis pesan...">
        <button class="chat-widget-send-btn" id="chatWidgetSendBtn" disabled>
          <svg viewBox="0 0 24 24">
            <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
          </svg>
        </button>
      </div>
    </div>
  `;

  document.body.appendChild(widgetContainer);

  // DOM Elements
  const btn = document.getElementById('chatWidgetBtn');
  const box = document.getElementById('chatWidgetBox');
  const closeBtn = document.getElementById('chatWidgetClose');
  const body = document.getElementById('chatWidgetBody');
  const faqList = document.getElementById('chatWidgetFaqList');
  const faqLabel = document.getElementById('faqLabel');
  const input = document.getElementById('chatWidgetInput');
  const sendBtn = document.getElementById('chatWidgetSendBtn');
  const badge = btn.querySelector('.chat-widget-badge');

  // Populate FAQs
  function renderFaqs() {
    faqList.innerHTML = '';
    faqs.forEach(faq => {
      const faqBtn = document.createElement('button');
      faqBtn.className = 'chat-widget-faq-btn';
      faqBtn.innerText = faq;
      faqBtn.addEventListener('click', () => handleFaqClick(faq));
      faqList.appendChild(faqBtn);
    });
    faqLabel.style.display = 'block';
    faqList.style.display = 'flex';
    scrollToBottom();
  }

  // Handle FAQ local automated responses
  function handleFaqClick(faq) {
    // 1. Hide FAQ list temporarily during answering
    faqLabel.style.display = 'none';
    faqList.style.display = 'none';

    // 2. Append user bubble
    appendMessage(faq, true);
    scrollToBottom();

    // 3. Show typing indicator
    const typing = showTypingIndicator();
    scrollToBottom();

    // 4. Delay and reply
    setTimeout(() => {
      typing.remove();
      const reply = faqAnswers[faq] || 'Maaf, saya tidak mengerti pertanyaan tersebut.';
      appendMessage(reply, false);
      
      // 5. Restore FAQ buttons so they can ask more questions
      renderFaqs();
    }, 1000);
  }

  // Append a bubble helper
  function appendMessage(text, isUser) {
    const bubble = document.createElement('div');
    bubble.className = `chat-widget-bubble${isUser ? ' user' : ''}`;
    bubble.innerHTML = text;
    
    const time = document.createElement('span');
    time.className = `chat-widget-time${isUser ? ' user' : ''}`;
    time.innerText = getFormattedTime();

    body.appendChild(bubble);
    body.appendChild(time);
  }

  // Show typing helper
  function showTypingIndicator() {
    const typingDiv = document.createElement('div');
    typingDiv.className = 'chat-widget-typing';
    typingDiv.innerHTML = '<span></span><span></span><span></span>';
    body.appendChild(typingDiv);
    return typingDiv;
  }

  // Helper scroll
  function scrollToBottom() {
    body.scrollTop = body.scrollHeight;
  }

  // Helper formatted time
  function getFormattedTime() {
    const now = new Date();
    const hrs = String(now.getHours()).padStart(2, '0');
    const mins = String(now.getMinutes()).padStart(2, '0');
    return `${hrs}:${mins}`;
  }

  // Initial FAQ load
  renderFaqs();

  // Show/Hide Box
  btn.addEventListener('click', () => {
    box.classList.toggle('open');
    if (badge) badge.style.display = 'none';
    scrollToBottom();
  });

  closeBtn.addEventListener('click', () => {
    box.classList.remove('open');
  });

  // Input check
  input.addEventListener('input', () => {
    const text = input.value.trim();
    if (text.length > 0) {
      sendBtn.disabled = false;
      sendBtn.classList.add('active');
    } else {
      sendBtn.disabled = true;
      sendBtn.classList.remove('active');
    }
  });

  // Send trigger (Custom user typing redirects to WhatsApp!)
  const handleSend = () => {
    const text = input.value.trim();
    if (text.length > 0) {
      // Show user typed bubble
      appendMessage(text, true);
      input.value = '';
      sendBtn.disabled = true;
      sendBtn.classList.remove('active');
      scrollToBottom();

      // Show typing indicator
      const typing = showTypingIndicator();
      scrollToBottom();

      // System notification: Connecting to WhatsApp
      setTimeout(() => {
        typing.remove();
        appendMessage('Menghubungkan kamu ke WhatsApp Tim Undangan Intim...', false);
        scrollToBottom();

        // Redirect to WhatsApp after short delay
        setTimeout(() => {
          sendWhatsAppMessage(text);
          renderFaqs();
        }, 1200);
      }, 1000);
    }
  };

  sendBtn.addEventListener('click', handleSend);
  input.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      handleSend();
    }
  });

  // Helper: Redirect to WA
  function sendWhatsAppMessage(msg) {
    const encodedText = encodeURIComponent(msg);
    const url = `https://wa.me/${waNumber}?text=${encodedText}`;
    window.open(url, '_blank');
  }
})();
