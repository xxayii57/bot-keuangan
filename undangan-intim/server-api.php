<?php
/**
 * Undangan Intim - SaaS Wedding Invitation Builder
 * server-api.php - Secure Backend REST API
 * Handles file uploads, invitation data persistent, and RSVP guest comments.
 */

// Load .env file if it exists
$envFile = __DIR__ . '/.env';
if (file_exists($envFile)) {
    $lines = file($envFile, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);
    foreach ($lines as $line) {
        $line = trim($line);
        if ($line === '' || $line[0] === '#') continue;
        if (strpos($line, '=') === false) continue;
        list($key, $value) = explode('=', $line, 2);
        $key = trim($key);
        $value = trim($value);
        if (!array_key_exists($key, $_ENV)) {
            $_ENV[$key] = $value;
            putenv("$key=$value");
        }
    }
}

// CORS: restrict to allowed origins
$allowedOrigins = [
    'https://intim.my.id',
    'https://www.intim.my.id',
    'http://localhost:8080',
    'http://localhost:8088',
    'http://localhost:3000',
    'http://127.0.0.1:8080'
];
$origin = isset($_SERVER['HTTP_ORIGIN']) ? $_SERVER['HTTP_ORIGIN'] : '';
if (in_array($origin, $allowedOrigins)) {
    header("Access-Control-Allow-Origin: " . $origin);
}
header("Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With");
header("Access-Control-Allow-Methods: GET, POST, OPTIONS");
header("Content-Type: application/json; charset=UTF-8");

// Handle preflight OPTIONS request
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit();
}

// Environment base directory setup
$baseDataDir = "/var/www/html/data";
$baseUploadsDir = "/var/www/html/assets/uploads";

// Graceful fallback to relative directories if running in local sandbox/localhost
if (!is_dir($baseDataDir)) {
    $baseDataDir = __DIR__ . "/data";
}
if (!is_dir($baseUploadsDir)) {
    $baseUploadsDir = __DIR__ . "/assets/uploads";
}

// Ensure storage directories exist securely
if (!is_dir($baseDataDir)) {
    mkdir($baseDataDir, 0755, true);
}
if (!is_dir($baseUploadsDir)) {
    mkdir($baseUploadsDir, 0755, true);
}

// Initialize SQLite database (auto-migrates from JSON on first run)
$db = initDatabase($baseDataDir);

// Helper to get local music path based on theme slug (v1 -> song 1, v2 -> song 2 ... v11 -> song 1)
function getThemeMusicUrl($themeSlug) {
    preg_match('/\\d+/', $themeSlug, $matches);
    $num = isset($matches[0]) ? (int)$matches[0] : 1;
    $idx = ($num - 1) % 10;
    if ($idx < 0) $idx = 0;
    
    $musicList = [
        "https://intim.my.id/assets/music/sampai_tua_nanti.mp3",
        "https://intim.my.id/assets/music/aku_memilihmu.mp3",
        "https://intim.my.id/assets/music/cinta_terakhir.mp3",
        "https://intim.my.id/assets/music/on_this_day.mp3",
        "https://intim.my.id/assets/music/bersamamu.mp3",
        "https://intim.my.id/assets/music/ketika_cinta_bertasbih.mp3",
        "https://intim.my.id/assets/music/for_the_rest_of_my_life.mp3",
        "https://intim.my.id/assets/music/penjaga_hati.mp3",
        "https://intim.my.id/assets/music/akad.mp3",
        "https://intim.my.id/assets/music/nahalal_kawin.mp3"
    ];
    return $musicList[$idx];
}

// Router Logic matching ?action=X and /api/X URL structures
$requestUri = $_SERVER['REQUEST_URI'];
$action = isset($_GET['action']) ? $_GET['action'] : '';

if (empty($action)) {
    // Attempt path-based routing (e.g. /api/save-invitation)
    $path = parse_url($requestUri, PHP_URL_PATH);
    if (strpos($path, '/api/save-invitation') !== false) {
        $action = 'save-invitation';
    } elseif (strpos($path, '/api/upload-photo') !== false) {
        $action = 'upload-photo';
    } elseif (strpos($path, '/api/checkin') !== false) {
        $action = 'checkin';
    } elseif (strpos($path, '/api/comment') !== false) {
        $action = 'comment';
    } elseif (strpos($path, '/api/sayabayar-callback') !== false) {
        $action = 'sayabayar-callback';
    }
}

// Process routes securely
switch ($action) {
    case 'save-invitation':
        handleSaveInvitation($db);
        break;

    case 'upload-photo':
        handleUploadPhoto($baseUploadsDir);
        break;

    case 'login':
        checkRateLimit('login');
        handleLogin($db);
        break;

    case 'order':
        checkRateLimit('order');
        handleOrder($db, $baseUploadsDir);
        break;

    case 'register-trial':
        checkRateLimit('register');
        handleRegisterTrial($db);
        break;

    case 'check-trial':
        handleCheckTrial($db);
        break;

    case 'reseller-login':
        handleResellerLogin($db);
        break;

    case 'checkin':
        handleCheckin($db);
        break;

    case 'comment':
        handleComment($db);
        break;

    case 'track-visit':
        handleTrackVisit($db);
        break;

    case 'create-sayabayar-invoice':
        handleCreateSayaBayarInvoice($db);
        break;

    case 'sayabayar-callback':
        handleSayaBayarCallback($db);
        break;

    default:
        sendError(404, "Endpoint not found or invalid action query specified.");
        break;
}

// ----------------------------------------------------
// ROUTE IMPLEMENTATIONS
// ----------------------------------------------------

/**
 * 1. POST /api/save-invitation
 * Accepts wedding invitation JSON payload and persists to SQLite.
 */
function handleSaveInvitation($db) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }

    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);

    if (json_last_error() !== JSON_ERROR_NONE || !$data) {
        sendError(400, "Malformed JSON Payload.");
    }

    $slug = isset($data['slug']) ? $data['slug'] : '';
    $slug = sanitizeSlug($slug);

    if (empty($slug)) {
        sendError(400, "Missing or invalid invitation slug identifier.");
    }

    $requiredFields = [
        'brideName', 'groomName', 'brideFullName', 'groomFullName',
        'brideParents', 'groomParents', 'eventDateISO', 'akadTime',
        'resepsiTime', 'venueName', 'venueAddress', 'mapsUrl',
        'musicUrl', 'videoUrl', 'bankName', 'bankNumber', 'bankHolder',
        'walletName', 'walletNumber'
    ];
    foreach ($requiredFields as $field) {
        if (!isset($data[$field])) {
            sendError(400, "Missing required payload field: {$field}");
        }
    }

    $stmt = $db->prepare("INSERT OR REPLACE INTO invitations (slug, password, status, trial_expires_at, active, visits, order_name, order_phone, order_email, order_theme, ref, price, brideName, groomName, brideFullName, groomFullName, brideParents, groomParents, eventDateISO, akadTime, resepsiTime, venueName, venueAddress, mapsUrl, musicUrl, videoUrl, bankName, bankNumber, bankHolder, walletName, walletNumber, showVideo, showGallery, enableRsvp, gallery, story, updated_at) VALUES (:slug, :password, :status, :trial_expires_at, :active, :visits, :order_name, :order_phone, :order_email, :order_theme, :ref, :price, :brideName, :groomName, :brideFullName, :groomFullName, :brideParents, :groomParents, :eventDateISO, :akadTime, :resepsiTime, :venueName, :venueAddress, :mapsUrl, :musicUrl, :videoUrl, :bankName, :bankNumber, :bankHolder, :walletName, :walletNumber, :showVideo, :showGallery, :enableRsvp, :gallery, :story, datetime('now'))");

    $stmt->execute([
        ':slug' => $slug,
        ':password' => $data['password'] ?? '',
        ':status' => $data['status'] ?? 'active',
        ':trial_expires_at' => $data['trial_expires_at'] ?? null,
        ':active' => isset($data['active']) ? (int)$data['active'] : 1,
        ':visits' => $data['visits'] ?? 0,
        ':order_name' => $data['order_name'] ?? '',
        ':order_phone' => $data['order_phone'] ?? '',
        ':order_email' => $data['order_email'] ?? '',
        ':order_theme' => $data['order_theme'] ?? 'v9',
        ':ref' => $data['ref'] ?? '',
        ':price' => $data['price'] ?? 35000,
        ':brideName' => $data['brideName'] ?? '',
        ':groomName' => $data['groomName'] ?? '',
        ':brideFullName' => $data['brideFullName'] ?? '',
        ':groomFullName' => $data['groomFullName'] ?? '',
        ':brideParents' => $data['brideParents'] ?? '',
        ':groomParents' => $data['groomParents'] ?? '',
        ':eventDateISO' => $data['eventDateISO'] ?? '',
        ':akadTime' => $data['akadTime'] ?? '',
        ':resepsiTime' => $data['resepsiTime'] ?? '',
        ':venueName' => $data['venueName'] ?? '',
        ':venueAddress' => $data['venueAddress'] ?? '',
        ':mapsUrl' => $data['mapsUrl'] ?? '',
        ':musicUrl' => $data['musicUrl'] ?? '',
        ':videoUrl' => $data['videoUrl'] ?? '',
        ':bankName' => $data['bankName'] ?? '',
        ':bankNumber' => $data['bankNumber'] ?? '',
        ':bankHolder' => $data['bankHolder'] ?? '',
        ':walletName' => $data['walletName'] ?? '',
        ':walletNumber' => $data['walletNumber'] ?? '',
        ':showVideo' => isset($data['showVideo']) ? (int)$data['showVideo'] : 0,
        ':showGallery' => isset($data['showGallery']) ? (int)$data['showGallery'] : 1,
        ':enableRsvp' => isset($data['enableRsvp']) ? (int)$data['enableRsvp'] : 1,
        ':gallery' => json_encode($data['gallery'] ?? []),
        ':story' => json_encode($data['story'] ?? [])
    ]);

    sendResponse([
        "status" => "success",
        "message" => "Invitation settings saved successfully.",
        "slug" => $slug
    ]);
}

/**
 * 2. POST /api/upload-photo
 * Processes base64 WebP image payload and writes to assets folder.
 */
function handleUploadPhoto($uploadsDir) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }

    $slug = isset($_POST['slug']) ? $_POST['slug'] : '';
    $slug = sanitizeSlug($slug);

    if (empty($slug)) {
        sendError(400, "Missing or invalid client slug.");
    }

    $photoBase64 = isset($_POST['photo_base64']) ? $_POST['photo_base64'] : '';
    $filename = isset($_POST['filename']) ? $_POST['filename'] : 'gallery_' . time();
    
    // Sanitize filename
    $filename = preg_replace('/[^a-zA-Z0-9_-]/', '', $filename);

    if (empty($photoBase64)) {
        sendError(400, "Missing base64 photo data.");
    }

    // Parse base64 header (e.g. data:image/webp;base64,...)
    if (preg_match('/^data:image\/(\w+);base64,/', $photoBase64, $matches)) {
        $photoBase64 = substr($photoBase64, strpos($photoBase64, ',') + 1);
    }

    $decodedData = base64_decode($photoBase64);
    if ($decodedData === false) {
        sendError(400, "Invalid base64 payload.");
    }

    // Limit maximum size to 2MB for security & premium SaaS constraints
    if (strlen($decodedData) > 2048 * 1024) {
        sendError(400, "Image size exceeds SaaS premium limit of 2MB.");
    }

    // Force secure webp extension
    $safeFileName = "{$slug}_{$filename}.webp";
    $targetPath = $uploadsDir . "/" . $safeFileName;

    if (file_put_contents($targetPath, $decodedData) !== false) {
        sendResponse([
            "status" => "success",
            "message" => "Photo uploaded & compressed side-client successfully.",
            "file_name" => $safeFileName,
            "file_path" => $targetPath
        ]);
    } else {
        sendError(500, "Failed saving binary asset.");
    }
}

/**
 * 3. POST /api/checkin
 * Scans QR attendance and logs timestamp to SQLite.
 */
function handleCheckin($db) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }

    $slug = isset($_POST['slug']) ? $_POST['slug'] : '';
    $slug = sanitizeSlug($slug);

    $guestId = isset($_POST['guest_id']) ? $_POST['guest_id'] : '';
    $guestId = preg_replace('/[^a-zA-Z0-9_-]/', '', $guestId);

    if (empty($slug) || empty($guestId)) {
        sendError(400, "Missing slug or guest ID parameter.");
    }

    $timestamp = date('Y-m-d H:i:s');
    $agent = $_SERVER['HTTP_USER_AGENT'] ?? 'Capacitor Android Client';

    $stmt = $db->prepare("INSERT INTO attendance (slug, guest_id, timestamp, agent) VALUES (:slug, :guest_id, :timestamp, :agent)");
    $stmt->execute([
        ':slug' => $slug,
        ':guest_id' => $guestId,
        ':timestamp' => $timestamp,
        ':agent' => $agent
    ]);

    sendResponse([
        "status" => "success",
        "message" => "Guest attendance checked in successfully.",
        "guest_id" => $guestId,
        "timestamp" => $timestamp
    ]);
}

/**
 * 4. GET & POST /api/comment
 * Interface with the wishes guestbook comments via SQLite.
 */
function handleComment($db) {
    $method = $_SERVER['REQUEST_METHOD'];
    $slug = isset($_REQUEST['slug']) ? $_REQUEST['slug'] : '';
    $slug = sanitizeSlug($slug);

    if (empty($slug)) {
        sendError(400, "Missing client invitation slug identifier.");
    }

    if ($method === 'GET') {
        $stmt = $db->prepare("SELECT name, comment, attendance, timestamp FROM comments WHERE slug = :slug ORDER BY id ASC");
        $stmt->execute([':slug' => $slug]);
        $comments = $stmt->fetchAll(PDO::FETCH_ASSOC);

        if (empty($comments)) {
            $default = [
                "name" => "Riana & Roni",
                "comment" => "Selamat menempuh hidup baru Aria & Bima! Semoga sakinah mawaddah warahmah, dilancarkan sampai hari H yaa!",
                "attendance" => "Hadir",
                "timestamp" => date('Y-m-d H:i:s', time() - 3600 * 2)
            ];
            $ins = $db->prepare("INSERT INTO comments (slug, name, comment, attendance, timestamp) VALUES (:slug, :name, :comment, :attendance, :timestamp)");
            $ins->execute([':slug' => $slug, ':name' => $default['name'], ':comment' => $default['comment'], ':attendance' => $default['attendance'], ':timestamp' => $default['timestamp']]);
            $comments[] = $default;
        }

        sendResponse(["status" => "success", "comments" => $comments]);

    } elseif ($method === 'POST') {
        $name = isset($_POST['name']) ? trim(strip_tags($_POST['name'])) : '';
        $comment = isset($_POST['comment']) ? trim(strip_tags($_POST['comment'])) : '';
        $attendance = isset($_POST['attendance']) ? trim(strip_tags($_POST['attendance'])) : 'Hadir';

        if (empty($name) || empty($comment)) {
            sendError(400, "Name and comment message are required.");
        }

        $timestamp = date('Y-m-d H:i:s');
        $stmt = $db->prepare("INSERT INTO comments (slug, name, comment, attendance, timestamp) VALUES (:slug, :name, :comment, :attendance, :timestamp)");
        $stmt->execute([
            ':slug' => $slug,
            ':name' => $name,
            ':comment' => $comment,
            ':attendance' => $attendance,
            ':timestamp' => $timestamp
        ]);

        sendResponse([
            "status" => "success",
            "message" => "Wish comment added to guestbook.",
            "comment" => ["name" => $name, "comment" => $comment, "attendance" => $attendance, "timestamp" => $timestamp]
        ]);
    } else {
        sendError(405, "Method Not Allowed.");
    }
}

// ----------------------------------------------------
// UTILITY SANITIZATIONS
// ----------------------------------------------------

/**
 * Securely sanitizes the slug to block directory traversal.
 * Permits only lowercase alphanumeric letters, hyphens, and underscores.
 */
function sanitizeSlug($slug) {
    if (is_array($slug)) return '';
    $slug = trim($slug);
    $slug = preg_replace('/[^a-z0-9_-]/', '', strtolower($slug));
    return $slug;
}

/**
 * Sends a structured JSON error response.
 */
function sendError($code, $message) {
    http_response_code($code);
    echo json_encode([
        "status" => "error",
        "code" => $code,
        "message" => $message
    ], JSON_UNESCAPED_SLASHES);
    exit();
}

/**
 * Sends a successful JSON response.
 */
function sendResponse($payload) {
    http_response_code(200);
    echo json_encode($payload, JSON_UNESCAPED_SLASHES);
    exit();
}

/**
 * Rate limiter — file-based throttle per IP + action.
 * Limits: login 5/min, order 10/hour, general 30/min
 */
function checkRateLimit($action) {
    $limits = [
        'login'    => ['max' => 5,  'window' => 60],
        'order'    => ['max' => 10, 'window' => 3600],
        'register' => ['max' => 5,  'window' => 60],
        'general'  => ['max' => 30, 'window' => 60],
    ];
    
    if (!isset($limits[$action])) $action = 'general';
    $cfg = $limits[$action];
    
    $ip = $_SERVER['REMOTE_ADDR'] ?? '127.0.0.1';
    $key = md5($ip . '_' . $action);
    $rateDir = sys_get_temp_dir() . '/undangan_rates';
    if (!is_dir($rateDir)) mkdir($rateDir, 0755, true);
    
    $rateFile = $rateDir . '/' . $key . '.json';
    $now = time();
    
    $data = ['attempts' => [], 'blocked_until' => 0];
    if (file_exists($rateFile)) {
        $raw = file_get_contents($rateFile);
        $data = json_decode($raw, true) ?: $data;
    }
    
    // Check if currently blocked
    if (isset($data['blocked_until']) && $data['blocked_until'] > $now) {
        $retryAfter = $data['blocked_until'] - $now;
        header("Retry-After: $retryAfter");
        sendError(429, "Terlalu banyak percobaan. Coba lagi dalam {$retryAfter} detik.");
    }
    
    // Purge expired attempts
    $data['attempts'] = array_values(array_filter($data['attempts'], function($t) use ($now, $cfg) {
        return $t > ($now - $cfg['window']);
    }));
    
    // Check limit
    if (count($data['attempts']) >= $cfg['max']) {
        $data['blocked_until'] = $now + $cfg['window'];
        file_put_contents($rateFile, json_encode($data));
        header("Retry-After: {$cfg['window']}");
        sendError(429, "Batas percobaan tercapai. Coba lagi dalam {$cfg['window']} detik.");
    }
    
    // Record this attempt
    $data['attempts'][] = $now;
    file_put_contents($rateFile, json_encode($data));
}

/**
 * Initialize SQLite database and migrate from JSON if first run.
 */
function initDatabase($dataDir) {
    $dbPath = $dataDir . '/undangan.db';
    $db = new PDO('sqlite:' . $dbPath);
    $db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $db->exec('PRAGMA journal_mode=WAL');
    $db->exec('PRAGMA foreign_keys=ON');
    
    // Create tables
    $db->exec("CREATE TABLE IF NOT EXISTS invitations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT UNIQUE NOT NULL,
        password TEXT,
        status TEXT DEFAULT 'active',
        trial_expires_at INTEGER,
        active INTEGER DEFAULT 1,
        visits INTEGER DEFAULT 0,
        order_name TEXT,
        order_phone TEXT,
        order_email TEXT,
        order_theme TEXT,
        ref TEXT,
        price INTEGER DEFAULT 35000,
        brideName TEXT,
        groomName TEXT,
        brideFullName TEXT,
        groomFullName TEXT,
        brideParents TEXT,
        groomParents TEXT,
        eventDateISO TEXT,
        akadTime TEXT,
        resepsiTime TEXT,
        venueName TEXT,
        venueAddress TEXT,
        mapsUrl TEXT,
        musicUrl TEXT,
        videoUrl TEXT,
        bankName TEXT,
        bankNumber TEXT,
        bankHolder TEXT,
        walletName TEXT,
        walletNumber TEXT,
        showVideo INTEGER DEFAULT 0,
        showGallery INTEGER DEFAULT 1,
        enableRsvp INTEGER DEFAULT 1,
        gallery TEXT DEFAULT '[]',
        story TEXT DEFAULT '[]',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )");
    
    $db->exec("CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL,
        name TEXT,
        comment TEXT,
        attendance TEXT DEFAULT 'Hadir',
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    )");
    
    $db->exec("CREATE TABLE IF NOT EXISTS attendance (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL,
        guest_id TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        agent TEXT
    )");
    
    $db->exec("CREATE TABLE IF NOT EXISTS resellers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        affiliate_id TEXT,
        points INTEGER DEFAULT 0,
        withdrawals INTEGER DEFAULT 0
    )");
    
    $db->exec("CREATE INDEX IF NOT EXISTS idx_inv_slug ON invitations(slug)");
    $db->exec("CREATE INDEX IF NOT EXISTS idx_inv_status ON invitations(status)");
    $db->exec("CREATE INDEX IF NOT EXISTS idx_cmt_slug ON comments(slug)");
    $db->exec("CREATE INDEX IF NOT EXISTS idx_att_slug ON attendance(slug)");
    $db->exec("CREATE INDEX IF NOT EXISTS idx_res_email ON resellers(email)");
    
    // Auto-migrate from JSON if migration not done yet
    $migrationFlag = $dataDir . '/.sqlite_migrated';
    if (!file_exists($migrationFlag)) {
        migrateJsonToSqlite($db, $dataDir);
        file_put_contents($migrationFlag, date('Y-m-d H:i:s'));
    }
    
    return $db;
}

/**
 * Migrate all JSON files to SQLite tables.
 */
function migrateJsonToSqlite($db, $dataDir) {
    $files = glob($dataDir . '/*.json');
    if (!$files) return;
    
    $invStmt = $db->prepare("INSERT OR IGNORE INTO invitations (slug, password, status, trial_expires_at, active, visits, order_name, order_phone, order_email, order_theme, ref, price, brideName, groomName, brideFullName, groomFullName, brideParents, groomParents, eventDateISO, akadTime, resepsiTime, venueName, venueAddress, mapsUrl, musicUrl, videoUrl, bankName, bankNumber, bankHolder, walletName, walletNumber, showVideo, showGallery, enableRsvp, gallery, story) VALUES (:slug, :password, :status, :trial_expires_at, :active, :visits, :order_name, :order_phone, :order_email, :order_theme, :ref, :price, :brideName, :groomName, :brideFullName, :groomFullName, :brideParents, :groomParents, :eventDateISO, :akadTime, :resepsiTime, :venueName, :venueAddress, :mapsUrl, :musicUrl, :videoUrl, :bankName, :bankNumber, :bankHolder, :walletName, :walletNumber, :showVideo, :showGallery, :enableRsvp, :gallery, :story)");
    
    $cmtStmt = $db->prepare("INSERT INTO comments (slug, name, comment, attendance, timestamp) VALUES (:slug, :name, :comment, :attendance, :timestamp)");
    $attStmt = $db->prepare("INSERT INTO attendance (slug, guest_id, timestamp, agent) VALUES (:slug, :guest_id, :timestamp, :agent)");
    $resStmt = $db->prepare("INSERT OR IGNORE INTO resellers (email, affiliate_id, points, withdrawals) VALUES (:email, :affiliate_id, :points, :withdrawals)");
    
    $db->beginTransaction();
    try {
        foreach ($files as $file) {
            $basename = basename($file, '.json');
            $data = json_decode(file_get_contents($file), true);
            if (!$data) continue;
            
            // Comments file
            if (substr($basename, -9) === '_comments') {
                $slug = substr($basename, 0, -9);
                foreach ($data as $c) {
                    $cmtStmt->execute([
                        ':slug' => $slug,
                        ':name' => $c['name'] ?? '',
                        ':comment' => $c['comment'] ?? '',
                        ':attendance' => $c['attendance'] ?? 'Hadir',
                        ':timestamp' => $c['timestamp'] ?? date('Y-m-d H:i:s')
                    ]);
                }
                continue;
            }
            
            // Attendance file
            if (substr($basename, -11) === '_attendance') {
                $slug = substr($basename, 0, -11);
                foreach ($data as $a) {
                    $attStmt->execute([
                        ':slug' => $slug,
                        ':guest_id' => $a['guest_id'] ?? '',
                        ':timestamp' => $a['timestamp'] ?? date('Y-m-d H:i:s'),
                        ':agent' => $a['agent'] ?? ''
                    ]);
                }
                continue;
            }
            
            // Reseller file
            if (strpos($basename, 'reseller_') === 0) {
                $resStmt->execute([
                    ':email' => $data['email'] ?? '',
                    ':affiliate_id' => $data['affiliateId'] ?? '',
                    ':points' => $data['points'] ?? 0,
                    ':withdrawals' => $data['withdrawals'] ?? 0
                ]);
                continue;
            }
            
            // Invitation file
            if (!isset($data['slug'])) continue;
            $invStmt->execute([
                ':slug' => $data['slug'],
                ':password' => $data['password'] ?? '',
                ':status' => $data['status'] ?? 'active',
                ':trial_expires_at' => $data['trial_expires_at'] ?? null,
                ':active' => isset($data['active']) ? (int)$data['active'] : 1,
                ':visits' => $data['visits'] ?? 0,
                ':order_name' => $data['order_name'] ?? '',
                ':order_phone' => $data['order_phone'] ?? '',
                ':order_email' => $data['order_email'] ?? '',
                ':order_theme' => $data['order_theme'] ?? 'v9',
                ':ref' => $data['ref'] ?? '',
                ':price' => $data['price'] ?? 35000,
                ':brideName' => $data['brideName'] ?? '',
                ':groomName' => $data['groomName'] ?? '',
                ':brideFullName' => $data['brideFullName'] ?? '',
                ':groomFullName' => $data['groomFullName'] ?? '',
                ':brideParents' => $data['brideParents'] ?? '',
                ':groomParents' => $data['groomParents'] ?? '',
                ':eventDateISO' => $data['eventDateISO'] ?? '',
                ':akadTime' => $data['akadTime'] ?? '',
                ':resepsiTime' => $data['resepsiTime'] ?? '',
                ':venueName' => $data['venueName'] ?? '',
                ':venueAddress' => $data['venueAddress'] ?? '',
                ':mapsUrl' => $data['mapsUrl'] ?? '',
                ':musicUrl' => $data['musicUrl'] ?? '',
                ':videoUrl' => $data['videoUrl'] ?? '',
                ':bankName' => $data['bankName'] ?? '',
                ':bankNumber' => $data['bankNumber'] ?? '',
                ':bankHolder' => $data['bankHolder'] ?? '',
                ':walletName' => $data['walletName'] ?? '',
                ':walletNumber' => $data['walletNumber'] ?? '',
                ':showVideo' => isset($data['showVideo']) ? (int)$data['showVideo'] : 0,
                ':showGallery' => isset($data['showGallery']) ? (int)$data['showGallery'] : 1,
                ':enableRsvp' => isset($data['enableRsvp']) ? (int)$data['enableRsvp'] : 1,
                ':gallery' => json_encode($data['gallery'] ?? []),
                ':story' => json_encode($data['story'] ?? [])
            ]);
        }
        $db->commit();
    } catch (Exception $e) {
        $db->rollBack();
        error_log("Migration failed: " . $e->getMessage());
    }
}

/**
 * 5. POST /api/login
 * Verifies slug credentials against SQLite.
 */
function handleLogin($db) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }
    
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);
    
    $slug = isset($data['slug']) ? sanitizeSlug($data['slug']) : '';
    $password = isset($data['password']) ? $data['password'] : '';
    
    if (empty($slug) || empty($password)) {
        sendError(400, "Username (Slug) dan Sandi harus diisi.");
    }
    
    $stmt = $db->prepare("SELECT * FROM invitations WHERE slug = :slug");
    $stmt->execute([':slug' => $slug]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if (!$row) {
        sendError(404, "Undangan dengan slug '{$slug}' tidak ditemukan.");
    }
    
    $storedPassword = $row['password'] ?? '';
    
    if (!empty($storedPassword)) {
        if (password_verify($password, $storedPassword)) {
            if (password_needs_rehash($storedPassword, PASSWORD_DEFAULT)) {
                $newHash = password_hash($password, PASSWORD_DEFAULT);
                $upd = $db->prepare("UPDATE invitations SET password = :pw WHERE slug = :slug");
                $upd->execute([':pw' => $newHash, ':slug' => $slug]);
            }
        } else {
            sendError(401, "Sandi yang Anda masukkan salah.");
        }
    }
    
    if ($row['status'] === 'pending_payment') {
        sendError(403, "Pembayaran belum diverifikasi. Akun Anda sedang dalam proses aktivasi.");
    }
    
    sendResponse([
        "status" => "success",
        "message" => "Login berhasil.",
        "slug" => $slug,
        "role" => "client",
        "data" => $row
    ]);
}

/**
 * 6. POST /api/order
 * Receives new orders, receipt screenshots, and creates inactive client accounts.
 */
function handleOrder($db, $uploadsDir) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }
    
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);
    
    $name = isset($data['name']) ? trim($data['name']) : '';
    $phone = isset($data['phone']) ? trim($data['phone']) : '';
    $email = isset($data['email']) ? trim($data['email']) : '';
    $theme = isset($data['theme']) ? trim($data['theme']) : 'v9';
    $slug = isset($data['slug']) ? sanitizeSlug($data['slug']) : '';
    $password = isset($data['password']) ? $data['password'] : '';
    $ref = isset($data['ref']) ? trim($data['ref']) : '';
    $price = isset($data['price']) ? intval($data['price']) : 35000;
    $receiptBase64 = isset($data['receiptBase64']) ? $data['receiptBase64'] : (isset($data['receipt_base64']) ? $data['receipt_base64'] : '');
    
    if (empty($name) || empty($phone) || empty($slug) || empty($password) || empty($receiptBase64)) {
        sendError(400, "Semua data wajib diisi, termasuk bukti transfer.");
    }
    
    $check = $db->prepare("SELECT status FROM invitations WHERE slug = :slug");
    $check->execute([':slug' => $slug]);
    $existing = $check->fetch(PDO::FETCH_ASSOC);
    if ($existing && $existing['status'] !== 'trial') {
        sendError(409, "Slug '{$slug}' sudah digunakan. Silakan pilih nama slug lain.");
    }
    
    if (preg_match('/^data:image\/(\w+);base64,/', $receiptBase64, $matches)) {
        $receiptBase64 = substr($receiptBase64, strpos($receiptBase64, ',') + 1);
    }
    $decodedReceipt = base64_decode($receiptBase64);
    if ($decodedReceipt === false) {
        sendError(400, "Bukti transfer tidak valid.");
    }
    
    file_put_contents($uploadsDir . "/receipt_{$slug}.webp", $decodedReceipt);
    
    $parts = explode('-', $slug);
    $parsedGroom = "Budi";
    $parsedBride = "Siti";
    if (count($parts) >= 2) { $parsedGroom = ucfirst(trim($parts[0])); $parsedBride = ucfirst(trim($parts[1])); }
    elseif (count($parts) === 1) { $parsedGroom = ucfirst(trim($parts[0])); }

    $stmt = $db->prepare("INSERT OR REPLACE INTO invitations (slug, password, status, active, visits, order_name, order_phone, order_email, order_theme, ref, price, brideName, groomName, brideFullName, groomFullName, brideParents, groomParents, eventDateISO, akadTime, resepsiTime, venueName, venueAddress, mapsUrl, musicUrl, videoUrl, bankName, bankNumber, bankHolder, walletName, walletNumber, showVideo, showGallery, enableRsvp, gallery, story, updated_at) VALUES (:slug, :password, 'pending_payment', 1, 0, :order_name, :order_phone, :order_email, :order_theme, :ref, :price, :brideName, :groomName, :brideFullName, :groomFullName, :brideParents, :groomParents, :eventDateISO, :akadTime, :resepsiTime, :venueName, :venueAddress, :mapsUrl, :musicUrl, :videoUrl, :bankName, :bankNumber, :bankHolder, :walletName, :walletNumber, 0, 1, 1, :gallery, :story, datetime('now'))");
    
    $gallery = json_encode(["https://images.unsplash.com/photo-1519741497674-611481863552?q=80&w=300&auto=format&fit=crop","https://images.unsplash.com/photo-1583939003579-730e3918a45a?q=80&w=300&auto=format&fit=crop","https://images.unsplash.com/photo-1511285560929-80b456fea0bc?q=80&w=300&auto=format&fit=crop"]);
    $story = json_encode([["title" => "Pertemuan Pertama", "date" => "2024", "description" => "Awal mula cerita kami dimulai dari sini."]]);
    
    $stmt->execute([
        ':slug' => $slug, ':password' => password_hash($password, PASSWORD_DEFAULT),
        ':order_name' => $name, ':order_phone' => $phone, ':order_email' => $email,
        ':order_theme' => $theme, ':ref' => $ref, ':price' => $price,
        ':brideName' => $parsedBride, ':groomName' => $parsedGroom,
        ':brideFullName' => $parsedBride, ':groomFullName' => $parsedGroom,
        ':brideParents' => "Putri tercinta dari Bapak H. Ahmad & Ibu Hj. Siti",
        ':groomParents' => "Putra tercinta dari Bapak Rahmat & Ibu Lina",
        ':eventDateISO' => date('Y-m-d', strtotime('+30 days')),
        ':akadTime' => "08:00 WIB - selesai", ':resepsiTime' => "11:00 WIB - selesai",
        ':venueName' => "Gedung Serbaguna Harmoni", ':venueAddress' => "Jl. Harmoni Indah No. 12, Jakarta",
        ':mapsUrl' => "https://maps.google.com", ':musicUrl' => getThemeMusicUrl($theme),
        ':videoUrl' => "", ':bankName' => "BCA", ':bankNumber' => "1234567890",
        ':bankHolder' => "Budi Santoso", ':walletName' => "DANA", ':walletNumber' => $phone,
        ':gallery' => $gallery, ':story' => $story
    ]);
    
    sendResponse(["status" => "success", "message" => "Pendaftaran berhasil, menunggu aktivasi pembayaran.", "slug" => $slug]);
}

function handleCheckTrial($db) {
    $slug = isset($_GET['slug']) ? sanitizeSlug($_GET['slug']) : '';
    if (empty($slug)) { sendResponse(["expired" => true, "reason" => "missing_slug"]); }
    
    $stmt = $db->prepare("SELECT status, trial_expires_at, active FROM invitations WHERE slug = :slug");
    $stmt->execute([':slug' => $slug]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if (!$row) { sendResponse(["expired" => true, "reason" => "not_found"]); }
    
    if ($row['status'] === 'trial' && time() > intval($row['trial_expires_at'])) {
        sendResponse(["expired" => true, "reason" => "trial_expired", "trial_expires_at" => intval($row['trial_expires_at'])]);
    }
    if (!$row['active'] || $row['status'] === 'pending_payment') {
        sendResponse(["expired" => true, "reason" => "inactive"]);
    }
    
    sendResponse(["expired" => false]);
}

function handleRegisterTrial($db) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') { sendError(405, "Method Not Allowed. Use POST."); }
    
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);
    
    $name = isset($data['name']) ? trim($data['name']) : '';
    $phone = isset($data['phone']) ? trim($data['phone']) : '';
    $email = isset($data['email']) ? trim($data['email']) : '';
    $theme = isset($data['theme']) ? trim($data['theme']) : 'v9';
    $slug = isset($data['slug']) ? sanitizeSlug($data['slug']) : '';
    $password = isset($data['password']) ? $data['password'] : '';
    $ref = isset($data['ref']) ? trim($data['ref']) : '';
    $price = isset($data['price']) ? intval($data['price']) : 35000;
    
    if (empty($name) || empty($phone) || empty($slug) || empty($password)) { sendError(400, "Semua data wajib diisi untuk daftar trial."); }
    
    $check = $db->prepare("SELECT status FROM invitations WHERE slug = :slug");
    $check->execute([':slug' => $slug]);
    $existing = $check->fetch(PDO::FETCH_ASSOC);
    if ($existing && $existing['status'] !== 'trial') {
        sendError(409, "Slug '{$slug}' sudah digunakan. Silakan pilih nama slug lain.");
    }
    
    $parts = explode('-', $slug);
    $parsedGroom = "Budi"; $parsedBride = "Siti";
    if (count($parts) >= 2) { $parsedGroom = ucfirst(trim($parts[0])); $parsedBride = ucfirst(trim($parts[1])); }
    elseif (count($parts) === 1) { $parsedGroom = ucfirst(trim($parts[0])); }

    $gallery = json_encode(["https://images.unsplash.com/photo-1519741497674-611481863552?q=80&w=300&auto=format&fit=crop","https://images.unsplash.com/photo-1583939003579-730e3918a45a?q=80&w=300&auto=format&fit=crop","https://images.unsplash.com/photo-1511285560929-80b456fea0bc?q=80&w=300&auto=format&fit=crop"]);
    $story = json_encode([["title" => "Pertemuan Pertama", "date" => "2024", "description" => "Awal mula cerita kami dimulai dari sini."]]);

    $stmt = $db->prepare("INSERT OR REPLACE INTO invitations (slug, password, status, trial_expires_at, active, visits, order_name, order_phone, order_email, order_theme, ref, price, brideName, groomName, brideFullName, groomFullName, brideParents, groomParents, eventDateISO, akadTime, resepsiTime, venueName, venueAddress, mapsUrl, musicUrl, videoUrl, bankName, bankNumber, bankHolder, walletName, walletNumber, showVideo, showGallery, enableRsvp, gallery, story, updated_at) VALUES (:slug, :password, 'trial', :trial_expires_at, 1, 0, :order_name, :order_phone, :order_email, :order_theme, :ref, :price, :brideName, :groomName, :brideFullName, :groomFullName, :brideParents, :groomParents, :eventDateISO, :akadTime, :resepsiTime, :venueName, :venueAddress, :mapsUrl, :musicUrl, :videoUrl, :bankName, :bankNumber, :bankHolder, :walletName, :walletNumber, 0, 1, 1, :gallery, :story, datetime('now'))");
    
    $stmt->execute([
        ':slug' => $slug, ':password' => password_hash($password, PASSWORD_DEFAULT),
        ':trial_expires_at' => time() + 86400,
        ':order_name' => $name, ':order_phone' => $phone, ':order_email' => $email,
        ':order_theme' => $theme, ':ref' => $ref, ':price' => $price,
        ':brideName' => $parsedBride, ':groomName' => $parsedGroom,
        ':brideFullName' => $parsedBride, ':groomFullName' => $parsedGroom,
        ':brideParents' => "Putri tercinta dari Orang Tua Mempelai Wanita",
        ':groomParents' => "Putra tercinta dari Orang Tua Mempelai Pria",
        ':eventDateISO' => date('Y-m-d', strtotime('+30 days')),
        ':akadTime' => "08:00 WIB - selesai", ':resepsiTime' => "11:00 WIB - selesai",
        ':venueName' => "Gedung Serbaguna Harmoni", ':venueAddress' => "Jl. Harmoni Indah No. 12, Jakarta",
        ':mapsUrl' => "https://maps.google.com", ':musicUrl' => getThemeMusicUrl($theme),
        ':videoUrl' => "", ':bankName' => "BCA", ':bankNumber' => "1234567890",
        ':bankHolder' => "Budi Santoso", ':walletName' => "DANA", ':walletNumber' => $phone,
        ':gallery' => $gallery, ':story' => $story
    ]);
    
    sendResponse(["status" => "success", "message" => "Pendaftaran trial berhasil.", "slug" => $slug]);
}

function handleTrackVisit($db) {
    $slug = isset($_GET['slug']) ? sanitizeSlug($_GET['slug']) : '';
    if (empty($slug)) { sendError(400, "Missing slug query parameter."); }
    
    $db->exec("UPDATE invitations SET visits = visits + 1 WHERE slug = " . $db->quote($slug));
    $stmt = $db->prepare("SELECT visits FROM invitations WHERE slug = :slug");
    $stmt->execute([':slug' => $slug]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if ($row) { sendResponse(["status" => "success", "visits" => $row['visits']]); }
    else { sendError(404, "Invitation not found."); }
}

function handleResellerLogin($db) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') { sendError(405, "Method Not Allowed. Use POST."); }
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);
    $email = isset($data['email']) ? trim(strtolower($data['email'])) : '';
    if (empty($email)) { sendError(400, "Email wajib diisi."); }
    
    $stmt = $db->prepare("SELECT * FROM resellers WHERE email = :email");
    $stmt->execute([':email' => $email]);
    $reseller = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if (!$reseller) {
        $randId = "IND-" . rand(10000, 99999);
        $ins = $db->prepare("INSERT INTO resellers (email, affiliate_id, points, withdrawals) VALUES (:email, :aid, 0, 0)");
        $ins->execute([':email' => $email, ':aid' => $randId]);
        $reseller = ['email' => $email, 'affiliate_id' => $randId, 'points' => 0, 'withdrawals' => 0];
    }
    
    $affiliateId = $reseller['affiliate_id'];
    
    $cstmt = $db->prepare("SELECT slug, brideName, groomName, status, price FROM invitations WHERE ref = :ref");
    $cstmt->execute([':ref' => $affiliateId]);
    $clients = [];
    $totalCommission = 0;
    
    while ($c = $cstmt->fetch(PDO::FETCH_ASSOC)) {
        $isPaid = ($c['status'] === 'active');
        $margin = max(0, intval($c['price']) - 35000);
        if ($isPaid) $totalCommission += $margin;
        $clients[] = [
            "slug" => $c['slug'],
            "couple" => ($c['brideName'] ?? 'Siti') . " & " . ($c['groomName'] ?? 'Budi'),
            "package" => "All-in-One",
            "price" => intval($c['price']),
            "status" => $isPaid ? 'Active' : 'Pending'
        ];
    }
    
    $netCommission = max(0, $totalCommission - intval($reseller['withdrawals']));
    
    sendResponse([
        "status" => "success",
        "data" => [
            "affiliateId" => $affiliateId,
            "commission" => $netCommission,
            "points" => count($clients) * 10,
            "clients" => $clients
        ]
    ]);
}

function handleCreateSayaBayarInvoice($db) {
    $input = file_get_contents("php://input");
    $req = json_decode($input, true);
    
    $slug = isset($req['slug']) ? sanitizeSlug($req['slug']) : '';
    if (empty($slug)) { sendError(400, "Slug wajib diisi"); }
    
    $stmt = $db->prepare("SELECT price FROM invitations WHERE slug = :slug");
    $stmt->execute([':slug' => $slug]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    if (!$row) { sendError(404, "Data undangan tidak ditemukan"); }
    
    $price = intval($row['price']);
    
    $apiKey = getenv('SAYABAYAR_API_KEY');
    if (empty($apiKey)) { sendError(500, "Payment gateway configuration error."); }
    
    $ch = curl_init("https://api.sayabayar.com/v1/invoices");
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, ["Content-Type: application/json", "x-api-key: " . $apiKey]);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
        "amount" => $price, "reference" => $slug,
        "description" => "Aktivasi Premium Undangan Intim - " . $slug,
        "redirect_url" => "https://intim.my.id/app/index.html"
    ]));
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    if ($httpCode !== 201) { sendError(500, "Gagal membuat invoice pembayaran. Error: " . $response); }
    
    $resData = json_decode($response, true);
    if (!isset($resData['success']) || !$resData['success']) { sendError(500, "Respon error dari provider pembayaran."); }
    
    sendResponse(["status" => "success", "payment_url" => $resData['data']['payment_url'], "amount" => $resData['data']['amount']]);
}

function handleSayaBayarCallback($db) {
    $input = file_get_contents("php://input");
    
    $payload = json_decode($input, true);
    if (!$payload) { sendError(400, "Invalid payload"); }
    
    $invoiceId = $payload['data']['id'] ?? '';
    $slug = $payload['data']['reference'] ?? '';
    
    if (empty($invoiceId) || empty($slug)) { sendError(400, "Missing parameters"); }
    
    $apiKey = getenv('SAYABAYAR_API_KEY');
    if (empty($apiKey)) { sendError(500, "Payment gateway configuration error."); }
    
    $ch = curl_init("https://api.sayabayar.com/v1/invoices/" . $invoiceId);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, ["x-api-key: " . $apiKey]);
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    if ($httpCode !== 200) { sendError(500, "Gagal memverifikasi status pembayaran."); }
    
    $resData = json_decode($response, true);
    $status = $resData['data']['status'] ?? '';
    
    if ($status === 'paid') {
        $upd = $db->prepare("UPDATE invitations SET status = 'active', active = 1, updated_at = datetime('now') WHERE slug = :slug");
        $upd->execute([':slug' => $slug]);
    }
    
    sendResponse(["status" => "success"]);
}
