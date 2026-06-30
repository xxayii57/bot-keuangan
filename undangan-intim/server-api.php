<?php
/**
 * Undangan Intim - SaaS Wedding Invitation Builder
 * server-api.php - Secure Backend REST API
 * Handles file uploads, invitation data persistent, and RSVP guest comments.
 */

// Global config and Security settings
header("Access-Control-Allow-Origin: *");
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
        handleSaveInvitation($baseDataDir);
        break;

    case 'upload-photo':
        handleUploadPhoto($baseUploadsDir);
        break;

    case 'login':
        handleLogin($baseDataDir);
        break;

    case 'order':
        handleOrder($baseDataDir, $baseUploadsDir);
        break;

    case 'register-trial':
        handleRegisterTrial($baseDataDir);
        break;

    case 'check-trial':
        handleCheckTrial($baseDataDir);
        break;

    case 'reseller-login':
        handleResellerLogin($baseDataDir);
        break;

    case 'checkin':
        handleCheckin($baseDataDir);
        break;

    case 'comment':
        handleComment($baseDataDir);
        break;

    case 'track-visit':
        handleTrackVisit($baseDataDir);
        break;

    case 'create-sayabayar-invoice':
        handleCreateSayaBayarInvoice($baseDataDir);
        break;

    case 'sayabayar-callback':
        handleSayaBayarCallback($baseDataDir);
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
 * Accepts wedding invitation JSON payload and persists to disk.
 */
function handleSaveInvitation($dataDir) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }

    // Capture payload
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);

    if (json_last_error() !== JSON_ERROR_NONE || !$data) {
        sendError(400, "Malformed JSON Payload.");
    }

    // Extract & Validate Slug
    $slug = isset($data['slug']) ? $data['slug'] : '';
    $slug = sanitizeSlug($slug);

    if (empty($slug)) {
        sendError(400, "Missing or invalid invitation slug identifier.");
    }

    // Validate Full JSON Structure explicitly
    $requiredFields = [
        'brideName' => 'string',
        'groomName' => 'string',
        'brideFullName' => 'string',
        'groomFullName' => 'string',
        'brideParents' => 'string',
        'groomParents' => 'string',
        'eventDateISO' => 'string',
        'akadTime' => 'string',
        'resepsiTime' => 'string',
        'venueName' => 'string',
        'venueAddress' => 'string',
        'mapsUrl' => 'string',
        'musicUrl' => 'string',
        'videoUrl' => 'string',
        'bankName' => 'string',
        'bankNumber' => 'string',
        'bankHolder' => 'string',
        'walletName' => 'string',
        'walletNumber' => 'string',
        'showVideo' => 'boolean',
        'showGallery' => 'boolean',
        'enableRsvp' => 'boolean',
        'gallery' => 'array'
    ];

    foreach ($requiredFields as $field => $type) {
        if (!isset($data[$field])) {
            sendError(400, "Missing required payload field: {$field}");
        }
        $actualType = gettype($data[$field]);
        if ($type === 'boolean' && $actualType !== 'boolean') {
            // Permit string or integer conversion for loose forms
            $data[$field] = filter_var($data[$field], FILTER_VALIDATE_BOOLEAN);
        } elseif ($type === 'array' && $actualType !== 'array') {
            sendError(400, "Field '{$field}' must be of type array.");
        }
    }

    // Prevent Directory Traversal & Write JSON
    $safeFilePath = $dataDir . "/" . $slug . ".json";
    
    // Format JSON pretty print for high readability
    $jsonOutput = json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
    
    if (file_put_contents($safeFilePath, $jsonOutput) !== false) {
        sendResponse([
            "status" => "success",
            "message" => "Invitation settings saved successfully.",
            "slug" => $slug,
            "file_path" => $safeFilePath
        ]);
    } else {
        sendError(500, "Failed writing data file. Check storage write permissions.");
    }
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
 * Scans QR attendance and logs timestamp to JSON file.
 */
function handleCheckin($dataDir) {
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

    $attendancePath = $dataDir . "/" . $slug . "_attendance.json";

    // Load existing log
    $logs = [];
    if (file_exists($attendancePath)) {
        $rawLogs = file_get_contents($attendancePath);
        $logs = json_decode($rawLogs, true) ?: [];
    }

    // Create timestamp
    $timestamp = date('Y-m-d H:i:s');
    $logs[] = [
        "guest_id" => $guestId,
        "timestamp" => $timestamp,
        "agent" => isset($_SERVER['HTTP_USER_AGENT']) ? $_SERVER['HTTP_USER_AGENT'] : 'Capacitor Android Client'
    ];

    if (file_put_contents($attendancePath, json_encode($logs, JSON_PRETTY_PRINT)) !== false) {
        sendResponse([
            "status" => "success",
            "message" => "Guest attendance checked in successfully.",
            "guest_id" => $guestId,
            "timestamp" => $timestamp
        ]);
    } else {
        sendError(500, "Failed updating attendance files.");
    }
}

/**
 * 4. GET & POST /api/comment
 * Interface with the wishes guestbook comments logs.
 */
function handleComment($dataDir) {
    $method = $_SERVER['REQUEST_METHOD'];
    $slug = isset($_REQUEST['slug']) ? $_REQUEST['slug'] : '';
    $slug = sanitizeSlug($slug);

    if (empty($slug)) {
        sendError(400, "Missing client invitation slug identifier.");
    }

    $commentPath = $dataDir . "/" . $slug . "_comments.json";

    if ($method === 'GET') {
        // Load wishes guestbook
        if (file_exists($commentPath)) {
            $comments = json_decode(file_get_contents($commentPath), true) ?: [];
        } else {
            // Seed a default warm greeting comment
            $comments = [
                [
                    "name" => "Riana & Roni",
                    "comment" => "Selamat menempuh hidup baru Aria & Bima! Semoga sakinah mawaddah warahmah, dilancarkan sampai hari H yaa! 🤍✨",
                    "attendance" => "Hadir",
                    "timestamp" => date('Y-m-d H:i:s', time() - 3600 * 2)
                ]
            ];
            file_put_contents($commentPath, json_encode($comments, JSON_PRETTY_PRINT));
        }

        sendResponse([
            "status" => "success",
            "comments" => $comments
        ]);

    } elseif ($method === 'POST') {
        // Post new guest wish
        $name = isset($_POST['name']) ? trim(strip_tags($_POST['name'])) : '';
        $comment = isset($_POST['comment']) ? trim(strip_tags($_POST['comment'])) : '';
        $attendance = isset($_POST['attendance']) ? trim(strip_tags($_POST['attendance'])) : 'Hadir'; // Hadir, Tidak Hadir, Ragu-ragu

        if (empty($name) || empty($comment)) {
            sendError(400, "Name and comment message are required.");
        }

        $comments = [];
        if (file_exists($commentPath)) {
            $comments = json_decode(file_get_contents($commentPath), true) ?: [];
        }

        $comments[] = [
            "name" => $name,
            "comment" => $comment,
            "attendance" => $attendance,
            "timestamp" => date('Y-m-d H:i:s')
        ];

        if (file_put_contents($commentPath, json_encode($comments, JSON_PRETTY_PRINT)) !== false) {
            sendResponse([
                "status" => "success",
                "message" => "Wish comment added to guestbook.",
                "comment" => $comments[count($comments) - 1]
            ]);
        } else {
            sendError(500, "Failed saving your wish to the guestbook.");
        }
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
 * 5. POST /api/login
 * Verifies slug credentials against static client JSON files.
 */
function handleLogin($dataDir) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }
    
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);
    
    $slug = isset($data['slug']) ? sanitizeSlug($data['slug']) : '';
    $password = isset($data['password']) ? $data['password'] : '';
    $ref = isset($data['ref']) ? trim($data['ref']) : '';
    $price = isset($data['price']) ? intval($data['price']) : 35000;
    
    if (empty($slug) || empty($password)) {
        sendError(400, "Username (Slug) dan Sandi harus diisi.");
    }
    
    $filePath = $dataDir . "/" . $slug . ".json";
    if (!file_exists($filePath)) {
        sendError(404, "Undangan dengan slug '{$slug}' tidak ditemukan.");
    }
    
    $fileData = json_decode(file_get_contents($filePath), true);
    
    // Validate stored password
    $storedPassword = isset($fileData['password']) ? $fileData['password'] : '';
    
    // Support legacy invitations which don't have password attribute
    if (!empty($storedPassword) && $storedPassword !== $password) {
        sendError(401, "Sandi yang Anda masukkan salah.");
    }
    
    // Validate billing verification status
    $status = isset($fileData['status']) ? $fileData['status'] : 'active';
    if ($status === 'pending_payment') {
        sendError(403, "Pembayaran belum diverifikasi. Akun Anda sedang dalam proses aktivasi.");
    }
    
    sendResponse([
        "status" => "success",
        "message" => "Login berhasil.",
        "slug" => $slug,
        "role" => "client",
        "data" => $fileData
    ]);
}

/**
 * 6. POST /api/order
 * Receives new orders, receipt screenshots, and creates inactive client accounts.
 */
function handleOrder($dataDir, $uploadsDir) {
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
    
    $filePath = $dataDir . "/" . $slug . ".json";
    $existingData = [];
    if (file_exists($filePath)) {
        $existingData = json_decode(file_get_contents($filePath), true);
        $existingStatus = isset($existingData['status']) ? $existingData['status'] : 'active';
        if ($existingStatus !== 'trial') {
            sendError(409, "Slug '{$slug}' sudah digunakan. Silakan pilih nama slug lain.");
        }
    }
    
    // Decode receipt image
    if (preg_match('/^data:image\/(\w+);base64,/', $receiptBase64, $matches)) {
        $receiptBase64 = substr($receiptBase64, strpos($receiptBase64, ',') + 1);
    }
    $decodedReceipt = base64_decode($receiptBase64);
    if ($decodedReceipt === false) {
        sendError(400, "Bukti transfer tidak valid.");
    }
    
    $receiptFileName = "receipt_{$slug}.webp";
    $receiptPath = $uploadsDir . "/" . $receiptFileName;
    file_put_contents($receiptPath, $decodedReceipt);
    
    // Default template data values for the newly created invitation
    $newInvitation = array_merge($existingData, [
        "slug" => $slug,
        "password" => $password,
        "status" => "pending_payment",
        "visits" => isset($existingData['visits']) ? $existingData['visits'] : 0,
        "order_name" => $name,
        "order_phone" => $phone,
        "order_email" => $email,
        "order_theme" => $theme,
        "ref" => $ref,
        "price" => $price,
    ]);
    
    $defaults = [
        "brideName" => "Siti",
        "groomName" => "Budi",
        "brideFullName" => "Siti Rahmaawati, S.Psi",
        "groomFullName" => "Budi Santoso, S.Kom",
        "brideParents" => "Putri tercinta dari Bapak H. Ahmad & Ibu Hj. Siti",
        "groomParents" => "Putra tercinta dari Bapak Rahmat & Ibu Lina",
        "eventDateISO" => date('Y-m-d', strtotime('+30 days')),
        "akadTime" => "08:00 WIB - selesai",
        "resepsiTime" => "11:00 WIB - selesai",
        "venueName" => "Gedung Serbaguna Harmoni",
        "venueAddress" => "Jl. Harmoni Indah No. 12, Jakarta",
        "mapsUrl" => "https://maps.google.com",
        "musicUrl" => getThemeMusicUrl($theme),
        "videoUrl" => "",
        "bankName" => "BCA",
        "bankNumber" => "1234567890",
        "bankHolder" => "Budi Santoso",
        "walletName" => "DANA",
        "walletNumber" => $phone,
        "showVideo" => false,
        "showGallery" => true,
        "enableRsvp" => true,
        "gallery" => [
            "https://images.unsplash.com/photo-1519741497674-611481863552?q=80&w=300&auto=format&fit=crop",
            "https://images.unsplash.com/photo-1583939003579-730e3918a45a?q=80&w=300&auto=format&fit=crop",
            "https://images.unsplash.com/photo-1511285560929-80b456fea0bc?q=80&w=300&auto=format&fit=crop"
        ],
        "story" => [
            ["title" => "Pertemuan Pertama", "date" => "2024", "description" => "Awal mula cerita kami dimulai dari sini."]
        ]
    ];
    
    foreach ($defaults as $key => $val) {
        if (!isset($newInvitation[$key])) {
            $newInvitation[$key] = $val;
        }
    }
    
    $jsonOutput = json_encode($newInvitation, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
    if (file_put_contents($filePath, $jsonOutput) !== false) {
        sendResponse([
            "status" => "success",
            "message" => "Pendaftaran berhasil, menunggu aktivasi pembayaran.",
            "slug" => $slug
        ]);
    } else {
        sendError(500, "Gagal memproses pendaftaran di server.");
    }
}

/**
 * Handles checking if trial has expired or invitation is inactive.
 */
function handleCheckTrial($dataDir) {
    $slug = isset($_GET['slug']) ? sanitizeSlug($_GET['slug']) : '';
    if (empty($slug)) {
        sendResponse(["expired" => true, "reason" => "missing_slug"]);
    }
    
    $filePath = $dataDir . "/" . $slug . ".json";
    if (!file_exists($filePath)) {
        sendResponse(["expired" => true, "reason" => "not_found"]);
    }
    
    $fileData = json_decode(file_get_contents($filePath), true);
    $status = isset($fileData['status']) ? $fileData['status'] : 'active';
    
    if ($status === 'trial') {
        $expiresAt = isset($fileData['trial_expires_at']) ? intval($fileData['trial_expires_at']) : 0;
        if (time() > $expiresAt) {
            sendResponse(["expired" => true, "reason" => "trial_expired", "trial_expires_at" => $expiresAt]);
        }
    }
    
    $active = isset($fileData['active']) ? (bool)$fileData['active'] : true;
    if (!$active || $status === 'pending_payment') {
        sendResponse(["expired" => true, "reason" => "inactive"]);
    }
    
    sendResponse(["expired" => false]);
}

/**
 * Handles registering a 1-day trial client.
 */
function handleRegisterTrial($dataDir) {
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
    
    if (empty($name) || empty($phone) || empty($slug) || empty($password)) {
        sendError(400, "Semua data wajib diisi untuk daftar trial.");
    }
    
    $filePath = $dataDir . "/" . $slug . ".json";
    $existingData = [];
    if (file_exists($filePath)) {
        $existingData = json_decode(file_get_contents($filePath), true);
        $existingStatus = isset($existingData['status']) ? $existingData['status'] : 'active';
        if ($existingStatus !== 'trial') {
            sendError(409, "Slug '{$slug}' sudah digunakan. Silakan pilih nama slug lain.");
        }
    }
    
    // Parse groom and bride from slug (e.g. arya-fajar -> Arya & Fajar)
    $parts = explode('-', $slug);
    $parsedGroom = "Budi";
    $parsedBride = "Siti";
    if (count($parts) >= 2) {
        $parsedGroom = ucfirst(trim($parts[0]));
        $parsedBride = ucfirst(trim($parts[1]));
    } else if (count($parts) === 1 && !empty($parts[0])) {
        $parsedGroom = ucfirst(trim($parts[0]));
    }

    $newInvitation = [
        "slug" => $slug,
        "password" => $password,
        "status" => "trial",
        "trial_expires_at" => time() + 86400, // 24 Hours
        "active" => true,
        "visits" => 0,
        "order_name" => $name,
        "order_phone" => $phone,
        "order_email" => $email,
        "order_theme" => $theme,
        "ref" => $ref,
        "price" => $price,
        
        "brideName" => $parsedBride,
        "groomName" => $parsedGroom,
        "brideFullName" => $parsedBride,
        "groomFullName" => $parsedGroom,
        "brideParents" => "Putri tercinta dari Orang Tua Mempelai Wanita",
        "groomParents" => "Putra tercinta dari Orang Tua Mempelai Pria",
        "eventDateISO" => date('Y-m-d', strtotime('+30 days')),
        "akadTime" => "08:00 WIB - selesai",
        "resepsiTime" => "11:00 WIB - selesai",
        "venueName" => "Gedung Serbaguna Harmoni",
        "venueAddress" => "Jl. Harmoni Indah No. 12, Jakarta",
        "mapsUrl" => "https://maps.google.com",
        "musicUrl" => getThemeMusicUrl($theme),
        "videoUrl" => "",
        "bankName" => "BCA",
        "bankNumber" => "1234567890",
        "bankHolder" => "Budi Santoso",
        "walletName" => "DANA",
        "walletNumber" => $phone,
        "showVideo" => false,
        "showGallery" => true,
        "enableRsvp" => true,
        "gallery" => [
            "https://images.unsplash.com/photo-1519741497674-611481863552?q=80&w=300&auto=format&fit=crop",
            "https://images.unsplash.com/photo-1583939003579-730e3918a45a?q=80&w=300&auto=format&fit=crop",
            "https://images.unsplash.com/photo-1511285560929-80b456fea0bc?q=80&w=300&auto=format&fit=crop"
        ],
        "story" => [
            ["title" => "Pertemuan Pertama", "date" => "2024", "description" => "Awal mula cerita kami dimulai dari sini."]
        ]
    ];
    
    $jsonOutput = json_encode($newInvitation, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
    if (file_put_contents($filePath, $jsonOutput) !== false) {
        sendResponse([
            "status" => "success",
            "message" => "Pendaftaran trial berhasil.",
            "slug" => $slug,
            "data" => $newInvitation
        ]);
    } else {
        sendError(500, "Gagal memproses pendaftaran trial di server.");
    }
}

/**
 * Processes visit tracking by incrementing the 'visits' counter in slug JSON data.
 */
function handleTrackVisit($dataDir) {
    $slug = isset($_GET['slug']) ? sanitizeSlug($_GET['slug']) : '';
    if (empty($slug)) {
        sendError(400, "Missing slug query parameter.");
    }
    
    $filePath = $dataDir . "/" . $slug . ".json";
    if (file_exists($filePath)) {
        $json = file_get_contents($filePath);
        $data = json_decode($json, true);
        if (!isset($data['visits'])) {
            $data['visits'] = 0;
        }
        $data['visits'] += 1;
        file_put_contents($filePath, json_encode($data, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES));
        sendResponse(["status" => "success", "visits" => $data['visits']]);
    } else {
        sendError(404, "Invitation not found.");
    }
}

/**
 * Dynamic Reseller Login and Commission calculation
 */
function handleResellerLogin($dataDir) {
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        sendError(405, "Method Not Allowed. Use POST.");
    }
    $rawInput = file_get_contents("php://input");
    $data = json_decode($rawInput, true);
    $email = isset($data['email']) ? trim(strtolower($data['email'])) : '';
    if (empty($email)) {
        sendError(400, "Email wajib diisi.");
    }
    
    // Sanitise email for safe filename
    $safeEmail = preg_replace('/[^a-z0-9@._-]/', '', $email);
    $filePath = $dataDir . "/reseller_" . $safeEmail . ".json";
    
    if (!file_exists($filePath)) {
        // Create new reseller profile
        $randId = "IND-" . rand(10000, 99999);
        $newReseller = [
            "email" => $email,
            "affiliateId" => $randId,
            "points" => 0,
            "withdrawals" => 0
        ];
        file_put_contents($filePath, json_encode($newReseller, JSON_PRETTY_PRINT));
    }
    
    $reseller = json_decode(file_get_contents($filePath), true);
    $affiliateId = $reseller['affiliateId'];
    
    // Scan all client files to gather dynamic referrals
    $clients = [];
    $totalCommission = 0;
    
    $files = glob($dataDir . "/*.json");
    foreach ($files as $file) {
        $filename = basename($file);
        if (strpos($filename, 'reseller_') === 0 || strpos($filename, '-wishes') !== false || strpos($filename, '-rsvp') !== false || strpos($filename, '_attendance') !== false || strpos($filename, '_comments') !== false) {
            continue;
        }
        
        $clientData = json_decode(file_get_contents($file), true);
        if (isset($clientData['ref']) && $clientData['ref'] === $affiliateId) {
            $slug = str_replace('.json', '', $filename);
            $bride = isset($clientData['brideName']) ? $clientData['brideName'] : 'Siti';
            $groom = isset($clientData['groomName']) ? $clientData['groomName'] : 'Budi';
            $status = isset($clientData['status']) ? $clientData['status'] : 'pending_payment';
            $price = isset($clientData['price']) ? intval($clientData['price']) : 35000;
            
            $isPaid = ($status === 'active' || (isset($clientData['active']) && $clientData['active'] === true));
            $clientStatus = $isPaid ? 'Active' : 'Pending';
            
            $margin = $price - 35000;
            if ($margin < 0) $margin = 0;
            
            if ($isPaid) {
                $totalCommission += $margin;
            }
            
            $clients[] = [
                "slug" => $slug,
                "couple" => $bride . " & " . $groom,
                "package" => "All-in-One",
                "price" => $price,
                "status" => $clientStatus
            ];
        }
    }
    
    $netCommission = $totalCommission - (isset($reseller['withdrawals']) ? intval($reseller['withdrawals']) : 0);
    if ($netCommission < 0) $netCommission = 0;
    
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

/**
 * Creates a dynamic Saya Bayar payment invoice.
 */
function handleCreateSayaBayarInvoice($dataDir) {
    // Read input request
    $input = file_get_contents("php://input");
    $req = json_decode($input, true);
    
    $slug = isset($req['slug']) ? sanitizeSlug($req['slug']) : '';
    if (empty($slug)) {
        sendError(400, "Slug wajib diisi");
    }
    
    $filePath = $dataDir . "/" . $slug . ".json";
    if (!file_exists($filePath)) {
        sendError(404, "Data undangan tidak ditemukan");
    }
    
    $fileData = json_decode(file_get_contents($filePath), true);
    $price = isset($fileData['price']) ? intval($fileData['price']) : 35000;
    
    $url = "https://api.sayabayar.com/v1/invoices";
    $apiKey = "sk_live_c5f4880a1279a1287ab9531d2b661916ab05662d3a2e9ecc";
    
    $postData = [
        "amount" => $price,
        "reference" => $slug,
        "description" => "Aktivasi Premium Undangan Intim - " . $slug,
        "redirect_url" => "https://intim.my.id/app/index.html"
    ];
    
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        "Content-Type: application/json",
        "x-api-key: " . $apiKey
    ]);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($postData));
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    if ($httpCode !== 201) {
        sendError(500, "Gagal membuat invoice pembayaran. Error: " . $response);
    }
    
    $resData = json_decode($response, true);
    if (!isset($resData['success']) || !$resData['success']) {
        sendError(500, "Respon error dari provider pembayaran.");
    }
    
    sendResponse([
        "status" => "success",
        "payment_url" => $resData['data']['payment_url'],
        "amount" => $resData['data']['amount']
    ]);
}

/**
 * Handles payment callbacks sent from Saya Bayar webhook.
 */
function handleSayaBayarCallback($dataDir) {
    $input = file_get_contents("php://input");
    
    // Log for audit purposes
    file_put_contents($dataDir . "/sayabayar_log.txt", "[" . date('Y-m-d H:i:s') . "] Webhook Received: " . $input . "\n", FILE_APPEND);
    
    $payload = json_decode($input, true);
    if (!$payload) {
        sendError(400, "Invalid payload");
    }
    
    $invoiceId = $payload['data']['id'] ?? '';
    $slug = $payload['data']['reference'] ?? '';
    $event = $payload['event'] ?? '';
    
    if (empty($invoiceId) || empty($slug)) {
        sendError(400, "Missing parameters");
    }
    
    // Make a secure GET verification call to sayabayar API
    $url = "https://api.sayabayar.com/v1/invoices/" . $invoiceId;
    $apiKey = "sk_live_c5f4880a1279a1287ab9531d2b661916ab05662d3a2e9ecc";
    
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        "x-api-key: " . $apiKey
    ]);
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    if ($httpCode !== 200) {
        sendError(500, "Gagal memverifikasi status pembayaran.");
    }
    
    $resData = json_decode($response, true);
    $status = $resData['data']['status'] ?? '';
    
    if ($status === 'paid') {
        $filePath = $dataDir . "/" . $slug . ".json";
        if (file_exists($filePath)) {
            $fileData = json_decode(file_get_contents($filePath), true);
            $fileData['status'] = 'active';
            $fileData['active'] = true;
            file_put_contents($filePath, json_encode($fileData, JSON_PRETTY_PRINT));
            
            file_put_contents($dataDir . "/sayabayar_log.txt", "[" . date('Y-m-d H:i:s') . "] Activated slug: " . $slug . "\n", FILE_APPEND);
        }
    }
    
    sendResponse(["status" => "success"]);
}
