package delivery

import (
	"net/http"
	"strings"
	"time"
)

// Endpoint sistem: halaman root dan health check.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveRoot(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CleanNest Backend API</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 40px 20px; line-height: 1.6; }
        .container { max-width: 800px; margin: 0 auto; background: #1e293b; border-radius: 12px; padding: 32px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); }
        h1 { color: #38bdf8; margin-top: 0; display: flex; align-items: center; gap: 10px; }
        .badge { background: #0284c7; color: white; font-size: 14px; padding: 4px 10px; border-radius: 20px; font-weight: normal; }
        .status-pill { display: inline-block; background: #10b981; color: white; padding: 4px 12px; border-radius: 9999px; font-size: 13px; font-weight: 600; }
        .card { background: #334155; border-radius: 8px; padding: 16px; margin: 20px 0; }
        a { color: #38bdf8; text-decoration: none; font-weight: 500; }
        a:hover { text-decoration: underline; }
        ul { padding-left: 20px; }
        li { margin-bottom: 8px; }
        code { background: #0f172a; color: #a5f3fc; padding: 2px 6px; border-radius: 4px; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🧺 CleanNest API <span class="badge">v1.0.0</span></h1>
        <p><span class="status-pill">● Backend Server Aktif</span> &nbsp; Clean Architecture Golang (Tanpa Controller/Handler)</p>
        
        <div class="card">
            <h3>🔗 Endpoint Publik Cepat:</h3>
            <ul>
                <li><strong>Health Check:</strong> <a href="/api/v1/health" target="_blank">/api/v1/health</a></li>
                <li><strong>Katalog Layanan:</strong> <a href="/api/v1/services" target="_blank">/api/v1/services</a></li>
                <li><strong>Pelacakan Pesanan:</strong> <code>GET /api/v1/orders/track?code=CN-...</code></li>
            </ul>
        </div>

        <div class="card">
            <h3>🔑 Akun Pengujian Default:</h3>
            <ul>
                <li><strong>Admin:</strong> <code>admin@cleannest.com</code> | Password: <code>password123</code></li>
                <li><strong>Kasir:</strong> <code>kasir@cleannest.com</code> | Password: <code>password123</code></li>
            </ul>
        </div>

        <p style="color: #94a3b8; font-size: 14px;">Gunakan Postman, Insomnia, atau cURL untuk menguji seluruh rute RESTful API.</p>
    </div>
</body>
</html>`
		w.Write([]byte(html))
		return
	}

	SendJSON(w, http.StatusOK, "CleanNest Backend Laundry API", map[string]interface{}{
		"status":       "online",
		"architecture": "Clean Architecture (No Controller/Handler)",
		"endpoints": map[string]string{
			"health":   "/api/v1/health",
			"services": "/api/v1/services",
			"track":    "/api/v1/orders/track?code={order_code}",
			"login":    "POST /api/v1/auth/login",
			"register": "POST /api/v1/auth/register",
		},
	})
}

func (rt *Router) serveHealthCheck(w http.ResponseWriter, r *http.Request) {
	SendJSON(w, http.StatusOK, "CleanNest Backend Laundry API running smoothly", map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}
