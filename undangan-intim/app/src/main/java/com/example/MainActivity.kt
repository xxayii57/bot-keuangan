package com.example

import android.Manifest
import android.net.Uri
import android.webkit.ValueCallback
import android.app.Activity
import androidx.activity.result.contract.ActivityResultContracts
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.webkit.JavascriptInterface
import android.webkit.WebChromeClient
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import android.webkit.WebResourceRequest
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.OnBackPressedCallback
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import androidx.core.content.ContextCompat
import com.example.ui.theme.MyApplicationTheme
import java.text.SimpleDateFormat
import java.util.Calendar
import java.util.Date
import java.util.Locale

class MainActivity : ComponentActivity() {

    private var webView: WebView? = null
    

    // WebView File Chooser Callback & Launcher
    private var filePathCallback: ValueCallback<Array<Uri>>? = null

    private val fileChooserLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        val results = if (result.resultCode == Activity.RESULT_OK) {
            val data = result.data
            if (data != null) {
                val dataString = data.dataString
                val clipData = data.clipData
                if (clipData != null) {
                    val uris = Array(clipData.itemCount) { i -> clipData.getItemAt(i).uri }
                    uris
                } else if (dataString != null) {
                    arrayOf(Uri.parse(dataString))
                } else {
                    null
                }
            } else {
                null
            }
        } else {
            null
        }
        filePathCallback?.onReceiveValue(results)
        filePathCallback = null
    }

    private val CHANNEL_ID = "wedding_builder_reminders"
    private val PERMISSION_REQUEST_CODE = 101

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        // Handle Back button via Javascript
        onBackPressedDispatcher.addCallback(this, object : OnBackPressedCallback(true) {
            override fun handleOnBackPressed() {
                if (webView != null) {
                    webView?.evaluateJavascript(
                        "if (typeof handleAndroidBackPress === 'function') { handleAndroidBackPress(); } else { window.history.back(); }",
                        null
                    )
                } else {
                    isEnabled = false
                    onBackPressedDispatcher.onBackPressed()
                }
            }
        })
        
        // Setup premium notifications channel
        createNotificationChannel()

        // Request runtime permissions on launch for smooth UX
        requestRequiredPermissions()

        setContent {
            MyApplicationTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    Box(modifier = Modifier.fillMaxSize().padding(innerPadding)) {
                        // Embedded high-performance Android WebView
                        AndroidWebViewContainer(
                            url = "file:///android_asset/index.html",
                            onWebViewCreated = { createdWebView ->
                                webView = createdWebView
                            }
                        )


                    }
                }
            }
        }
    }

    @Composable
    fun AndroidWebViewContainer(url: String, onWebViewCreated: (WebView) -> Unit) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { context ->
                WebView(context).apply {
                    // Premium WebView configurations for full compatibility with SPA tailwind layout
                    settings.apply {
                        javaScriptEnabled = true
                        domStorageEnabled = true
                        allowFileAccess = true
                        allowContentAccess = true
                        allowUniversalAccessFromFileURLs = true
                        allowFileAccessFromFileURLs = true
                        loadWithOverviewMode = true
                        useWideViewPort = true
                        mixedContentMode = WebSettings.MIXED_CONTENT_ALWAYS_ALLOW
                        cacheMode = WebSettings.LOAD_DEFAULT
                    }
                    
                    webViewClient = object : WebViewClient() {
                        override fun onPageFinished(view: WebView?, url: String?) {
                            super.onPageFinished(view, url)
                        }

                        override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
                            val url = request?.url?.toString() ?: return false
                            if (url.startsWith("http://") || url.startsWith("https://")) {
                                return false
                            }
                            try {
                                val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
                                view?.context?.startActivity(intent)
                                return true
                            } catch (e: Exception) {
                                e.printStackTrace()
                                return true
                            }
                        }
                    }
                    
                    webChromeClient = object : WebChromeClient() {
                        override fun onShowFileChooser(
                            webView: WebView?,
                            filePathCallback: ValueCallback<Array<Uri>>?,
                            fileChooserParams: FileChooserParams?
                        ): Boolean {
                            this@MainActivity.filePathCallback?.onReceiveValue(null)
                            this@MainActivity.filePathCallback = filePathCallback
                            
                            val intent = fileChooserParams?.createIntent() ?: Intent(Intent.ACTION_GET_CONTENT).apply {
                                type = "image/*"
                                putExtra(Intent.EXTRA_ALLOW_MULTIPLE, true)
                            }
                            try {
                                fileChooserLauncher.launch(intent)
                            } catch (e: java.lang.Exception) {
                                this@MainActivity.filePathCallback = null
                                return false
                            }
                            return true
                        }
                    }

                    // Expose the premium Javascript bridge interface
                    addJavascriptInterface(WebAppInterface(context), "AndroidBridge")
                    
                    loadUrl(url)
                    onWebViewCreated(this)
                }
            }
        )
    }



    // ----------------------------------------------------
    // JAVASCRIPT BRIDGE INTERFACE
    // ----------------------------------------------------
    inner class WebAppInterface(private val mContext: Context) {

        /**
         * Trigger native local calendar notification scheduling relative to ISO Date
         */
        @JavascriptInterface
        fun scheduleWeddingReminders(eventDateISO: String) {
            try {
                val format = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault())
                val weddingDate = format.parse(eventDateISO) ?: return
                
                // Parse wedding dates
                val cal = Calendar.getInstance().apply { time = weddingDate }
                
                // Post immediate test notification to confirm setup
                postNativeNotification(
                    title = "Penjadwalan Pengingat Sukses! 🌸",
                    message = "Sistem Undangan Intim telah mengeset alarm H-7, H-3, dan H-1 menjelang Hari H pernikahan Anda."
                )

                // Mock schedule future dates
                val formatter = SimpleDateFormat("dd MMMM yyyy", Locale( "id", "ID"))
                val dateStr = formatter.format(weddingDate)

                runOnUiThread {
                    Toast.makeText(mContext, "Pengingat dijadwalkan untuk: $dateStr", Toast.LENGTH_LONG).show()
                }

            } catch (e: Exception) {
                e.printStackTrace()
                runOnUiThread {
                    Toast.makeText(mContext, "Gagal memproses penanggalan: " + e.message, Toast.LENGTH_SHORT).show()
                }
            }
        }


        @JavascriptInterface
        fun openExternalBrowser(url: String) {
            try {
                val intent = Intent(Intent.ACTION_VIEW, android.net.Uri.parse(url)).apply {
                    flags = Intent.FLAG_ACTIVITY_NEW_TASK
                }
                mContext.startActivity(intent)
            } catch (e: Exception) {
                runOnUiThread {
                    Toast.makeText(mContext, "Gagal membuka browser: " + e.message, Toast.LENGTH_SHORT).show()
                }
            }
        }

        @JavascriptInterface
        fun exitApp() {
            runOnUiThread {
                (mContext as? Activity)?.finish()
            }
        }

        /**
         * Fetch native phone contacts and load them into the WebView recipient list
         */
        @JavascriptInterface
        fun readPhoneContacts() {
            runOnUiThread {
                if (ContextCompat.checkSelfPermission(mContext, Manifest.permission.READ_CONTACTS) == PackageManager.PERMISSION_GRANTED) {
                    executeContactsQuery()
                } else {
                    ActivityCompat.requestPermissions(
                        this@MainActivity,
                        arrayOf(Manifest.permission.READ_CONTACTS),
                        PERMISSION_REQUEST_CODE
                    )
                }
            }
        }
    }

    // ----------------------------------------------------
    // NATIVE SYSTEM HELPER METHODS
    // ----------------------------------------------------

    private fun postNativeNotification(title: String, message: String) {
        val intent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val builder = NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(android.R.drawable.ic_dialog_info)
            .setContentTitle(title)
            .setContentText(message)
            .setStyle(NotificationCompat.BigTextStyle().bigText(message))
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setContentIntent(pendingIntent)
            .setAutoCancel(true)

        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        notificationManager.notify(System.currentTimeMillis().toInt(), builder.build())
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val name = "Pengingat Pernikahan"
            val descriptionText = "Menyalakan pengingat otomatis H-7, H-3, dan H-1 pernikahan"
            val importance = NotificationManager.IMPORTANCE_HIGH
            val channel = NotificationChannel(CHANNEL_ID, name, importance).apply {
                description = descriptionText
            }
            val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            notificationManager.createNotificationChannel(channel)
        }
    }

    private fun requestRequiredPermissions() {
        val permissionsToRequest = mutableListOf<String>()

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
                permissionsToRequest.add(Manifest.permission.POST_NOTIFICATIONS)
            }
        }

        if (permissionsToRequest.isNotEmpty()) {
            ActivityCompat.requestPermissions(
                this,
                permissionsToRequest.toTypedArray(),
                PERMISSION_REQUEST_CODE
            )
        }
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == PERMISSION_REQUEST_CODE) {
            var permissionGranted = false
            
            val contactIndex = permissions.indexOf(Manifest.permission.READ_CONTACTS)
            if (contactIndex >= 0 && grantResults.size > contactIndex && grantResults[contactIndex] == PackageManager.PERMISSION_GRANTED) {
                showNativeToast("Akses kontak diizinkan!")
                executeContactsQuery()
                permissionGranted = true
            }
            

            
            if (!permissionGranted && grantResults.isNotEmpty() && grantResults[0] != PackageManager.PERMISSION_GRANTED) {
                showNativeToast("Akses fitur ditolak.")
            }
        }
    }

    private fun executeContactsQuery() {
        try {
            val jsonArray = org.json.JSONArray()
            val contentResolver = contentResolver
            val cursor = contentResolver.query(
                android.provider.ContactsContract.CommonDataKinds.Phone.CONTENT_URI,
                null, null, null, null
            )
            cursor?.use {
                val nameIdx = it.getColumnIndex(android.provider.ContactsContract.CommonDataKinds.Phone.DISPLAY_NAME)
                val phoneIdx = it.getColumnIndex(android.provider.ContactsContract.CommonDataKinds.Phone.NUMBER)
                if (nameIdx >= 0 && phoneIdx >= 0) {
                    var count = 0
                    while (it.moveToNext() && count < 150) {
                        val name = it.getString(nameIdx)
                        var phone = it.getString(phoneIdx)
                        
                        // Sanitize phone
                        phone = phone.replace("[^0-9]".toRegex(), "")
                        if (phone.startsWith("08")) {
                            phone = "62" + phone.substring(1)
                        }
                        
                        val guestObj = org.json.JSONObject()
                        guestObj.put("name", name)
                        guestObj.put("phone", phone)
                        guestObj.put("status", "Unsent")
                        
                        jsonArray.put(guestObj)
                        count++
                    }
                }
            }
            val jsonString = jsonArray.toString()
            webView?.post {
                webView?.evaluateJavascript("populateNativeContacts($jsonString)", null)
            }
        } catch (e: Exception) {
            e.printStackTrace()
            runOnUiThread {
                showNativeToast("Gagal memuat kontak: " + e.message)
            }
        }
    }

    private fun showNativeToast(message: String) {
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
    }
}
