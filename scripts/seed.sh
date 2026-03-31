#!/usr/bin/env bash
# seed.sh — Mahzen test verisi yükleyici
# Kullanım: bash scripts/seed.sh
# Sistem çalışır durumda olmalı: make run

set -euo pipefail

BASE="https://localhost:8080"
CURL="curl -sk"

# ─── Renkler ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()    { echo -e "${CYAN}[seed]${NC} $*"; }
ok()      { echo -e "${GREEN}[✓]${NC} $*"; }
warn()    { echo -e "${YELLOW}[!]${NC} $*"; }
die()     { echo -e "${RED}[✗]${NC} $*" >&2; exit 1; }

# ─── Register & login ─────────────────────────────────────────────────────────
info "Kullanıcı kaydediliyor: emir@emir.com"
REG=$($CURL -X POST "$BASE/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"emir@emir.com","display_name":"Emir","password":"emir1234"}')

ACCESS=$(echo "$REG" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4 || true)
REFRESH=$(echo "$REG" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4 || true)

if [ -z "$ACCESS" ]; then
  warn "Kayıt başarısız, giriş deneniyor (kullanıcı zaten var)"
  LOGIN=$($CURL -X POST "$BASE/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"emir@emir.com","password":"emir1234"}')
  ACCESS=$(echo "$LOGIN" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4 || true)
  REFRESH=$(echo "$LOGIN" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4 || true)
fi

[ -z "$ACCESS" ] && die "Token alınamadı. Sistem çalışıyor mu? (make run)"
ok "Giriş yapıldı"

AUTH="-H \"Authorization: Bearer $ACCESS\""

# Kısayol fonksiyonu
api() {
  local method=$1; shift
  local path=$1; shift
  $CURL -X "$method" "$BASE$path" \
    -H "Authorization: Bearer $ACCESS" \
    -H "Content-Type: application/json" \
    "$@"
}

# ─── Taglar ───────────────────────────────────────────────────────────────────
info "Taglar oluşturuluyor..."

create_tag() {
  local name=$1
  local res
  res=$(api POST /v1/tags -d "{\"name\":\"$name\"}")
  echo "$res" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || true
}

TAG_GO=$(create_tag "golang")
TAG_PYTHON=$(create_tag "python")
TAG_DB=$(create_tag "veritabanı")
TAG_LINUX=$(create_tag "linux")
TAG_GIT=$(create_tag "git")
TAG_API=$(create_tag "api")
TAG_DOCKER=$(create_tag "docker")
TAG_SECURITY=$(create_tag "güvenlik")
TAG_PERF=$(create_tag "performans")
TAG_ARCH=$(create_tag "mimari")
TAG_AI=$(create_tag "yapay-zeka")
TAG_FRONTEND=$(create_tag "frontend")
TAG_DEVOPS=$(create_tag "devops")
TAG_REGEX=$(create_tag "regex")
TAG_SHELL=$(create_tag "shell")

ok "15 tag oluşturuldu"

# ─── Entry oluşturma fonksiyonu ───────────────────────────────────────────────
created=0

create_entry() {
  local title=$1
  local path=$2
  local visibility=$3
  local content=$4

  shift 4
  local tag_ids_json="[]"

  if [ $# -gt 0 ] && [ -n "$1" ]; then
    tag_ids_json=$(printf '"%s",' "$@" | sed 's/,$//' | sed 's/^/[/' | sed 's/$/]/')
  fi

  local payload
  payload=$(printf '{"title":%s,"path":%s,"visibility":%s,"content":%s,"tag_ids":%s}' \
    "$(echo "$title" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read().strip()))')" \
    "$(echo "$path"  | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read().strip()))')" \
    "$(echo "$visibility" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read().strip()))')" \
    "$(echo "$content" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read().strip()))')" \
    "$tag_ids_json")

  local res
  res=$(api POST /v1/entries -d "$payload")
  local id
  id=$(echo "$res" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || true)
  if [ -n "$id" ]; then
    created=$((created + 1))
    echo "$id"
  else
    warn "Entry oluşturulamadı: $title | $res"
    echo ""
  fi
}

# ─── Entryler ─────────────────────────────────────────────────────────────────
info "Entry'ler oluşturuluyor..."

# /notlar/golang
create_entry \
  "Go'da context kullanımı" \
  "/notlar/golang" \
  "public" \
  "context paketi, Go programlarında deadline, iptal sinyali ve request-scoped değerleri fonksiyonlar arasında taşımak için kullanılır.

Temel kullanım:
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

WithTimeout: Belirli bir süre sonra otomatik iptal eder.
WithCancel: Manuel iptal için kullanılır.
WithDeadline: Belirli bir zamanda iptal eder.
WithValue: Key-value taşımak için (sadece request-scoped metadata için kullanın, business logic için değil).

Önemli kurallar:
- context.Background() sadece main, init ve testlerde kullanılır.
- context.TODO() ne yapacağını bilmiyorsan geçici olarak koy.
- Context'i struct'a gömme; fonksiyonun ilk parametresi olarak geç.
- ctx.Done() channel'ını dinleyerek iptal sinyalini yakala." \
  "$TAG_GO"

create_entry \
  "Go slice vs array farkları" \
  "/notlar/golang" \
  "public" \
  "Array: Sabit boyutlu, değer tipi. [3]int ile [4]int farklı tiplerdir.
Slice: Dinamik boyutlu, bir array'e referans. len ve cap vardır.

Slice oluşturma:
  s := make([]int, 0, 10)  // len=0, cap=10

Append:
  s = append(s, 1, 2, 3)

Cap aşılınca yeni array allocate edilir — büyük slicelar için önceden cap belirle.

Dikkat: Slice'ı fonksiyona geçince backing array paylaşılır. Kopyalamak için copy() kullan:
  dst := make([]int, len(src))
  copy(dst, src)

nil slice vs empty slice: make([]int, 0) ile var s []int davranış olarak neredeyse aynı ama json marshal'da fark var:
  nil   → null
  empty → []" \
  "$TAG_GO"

create_entry \
  "Go interface tasarımı — küçük tutun" \
  "/notlar/golang" \
  "public" \
  "Go'da interface'ler küçük olmalı. Standart kütüphaneden örnekler:
  io.Reader    → tek metod
  io.Writer    → tek metod
  fmt.Stringer → tek metod

Büyük interface'ler bağımlılığı artırır, test yazmayı zorlaştırır.

Kural: Tüketici interface'i tanımlar, sağlayıcı değil.
  // Kötü: servis paketi kendi büyük interface'ini export eder
  // İyi: handler paketi ihtiyacı kadar metod tanımlar

Implicit implementation: Go'da implements anahtar kelimesi yok. Tip interface'i otomatik sağlar.

Boş interface (any/interface{}): Kaçın. Type assertion gerektirir, derleme zamanı güvenliği kaybedilir." \
  "$TAG_GO" "$TAG_ARCH"

create_entry \
  "Goroutine sızıntısı nasıl bulunur" \
  "/notlar/golang" \
  "private" \
  "Goroutine sızıntısı: Başlatılan goroutine hiç sonlanmıyor.

Yaygın sebepler:
1. Channel'a hiç yazılmıyor, goroutine blokta kalıyor
2. context iptal edilmiyor (defer cancel() unutulmuş)
3. Sonsuz for döngüsü, çıkış koşulu yok

Tespit:
  runtime.NumGoroutine()  // anlık sayı
  pprof /debug/pprof/goroutine?debug=1

Test sırasında goleak kullanımı:
  func TestFoo(t *testing.T) {
    defer goleak.VerifyNone(t)
    // ...
  }

Çözüm: Her go func() çağrısında net bir çıkış yolu olduğundan emin ol.
select { case <-ctx.Done(): return } ekle." \
  "$TAG_GO" "$TAG_PERF"

create_entry \
  "Go'da hata yönetimi kalıpları" \
  "/notlar/golang" \
  "public" \
  "Standart kalıp — wrap et:
  if err != nil {
    return fmt.Errorf(\"kullanıcı oluşturulurken: %w\", err)
  }

errors.Is: Hata zincirinde belirli bir değeri ara
errors.As: Hata zincirinde belirli bir tipi ara

Sentinel error: Belirli hataları karşılaştırmak için
  var ErrNotFound = errors.New(\"kayıt bulunamadı\")

Custom error type: Ek bilgi taşımak için
  type ValidationError struct { Field string; Msg string }
  func (e *ValidationError) Error() string { ... }

Kural: Hata mesajları küçük harfle başlar, nokta ile bitmez.
  \"kullanıcı bulunamadı\" ✓
  \"Kullanıcı bulunamadı.\" ✗" \
  "$TAG_GO"

# /notlar/veritabani
create_entry \
  "PostgreSQL index türleri" \
  "/notlar/veritabani" \
  "public" \
  "B-tree: Varsayılan, eşitlik ve range sorguları için. <, >, BETWEEN, ORDER BY.
Hash: Sadece eşitlik (=) için, B-tree'ten hızlı ama range desteklemez.
GIN: Array, JSONB, full-text search için. Yavaş insert, hızlı sorgu.
GiST: Geometrik veri, full-text search. GIN'den daha hızlı insert.

Kıyasla:
  CREATE INDEX idx_users_email ON users(email);           -- B-tree (implicit)
  CREATE INDEX idx_users_email_hash ON users USING hash(email);  -- Hash
  CREATE INDEX idx_docs_content ON docs USING gin(to_tsvector('english', content));  -- GIN

Index sadece gerektiğinde ekle. Write amplification ve disk kullanımını artırır." \
  "$TAG_DB" "$TAG_PERF"

create_entry \
  "PostgreSQL EXPLAIN ANALYZE okuma" \
  "/notlar/veritabani" \
  "public" \
  "EXPLAIN ANALYZE: Sorgu planını gerçek çalıştırma süreleriyle birlikte gösterir.

Önemli metrikler:
- Seq Scan: Full table scan, genelde kötü (büyük tablolarda)
- Index Scan: Index kullanılıyor, iyi
- Bitmap Heap Scan: Index'ten sonra disk okuma, orta
- Nested Loop: Join için, küçük setlerde iyi
- Hash Join: Büyük setlerde iyi
- Merge Join: Sıralı veride iyi

Maliyet (cost):
  cost=0.00..1.01 rows=1 width=100
  - İlk sayı: başlangıç maliyeti
  - İkinci sayı: toplam maliyet
  - rows: tahmin edilen satır sayısı
  - width: ortalama satır boyutu

Dikkat: rows tahmini ile gerçek farkı büyükse ANALYZE çalıştır:
  ANALYZE users;" \
  "$TAG_DB" "$TAG_PERF"

create_entry \
  "SQL injection önleme" \
  "/notlar/veritabani" \
  "public" \
  "Asla string concatenation ile SQL oluşturma:
  // YANLIŞ
  query := fmt.Sprintf('SELECT * FROM users WHERE id = %s', userID)

  // DOĞRU - Prepared statement
  SELECT * FROM users WHERE id = \$1

Go'da pgx/sqlc ile:
  - sqlc parametreleri otomatik escape eder
  - pgx'in QueryRow/Exec metodları prepared statement kullanır

Ekstra önlemler:
- En az privilege principle: Uygulama kullanıcısına sadece gerekli yetkileri ver
- Input validation: UUID formatı, sayı aralıkları kontrol et
- WAF: SQL pattern'lerini engelleyen firewall" \
  "$TAG_DB" "$TAG_SECURITY"

create_entry \
  "Connection pool tuning" \
  "/notlar/veritabani" \
  "private" \
  "pgxpool ayarları:

MaxConns: CPU core sayısı * 2 + disk sayısı (genelde)
  - Çok yüksek: Bağlantı yönetim overhead'i, DB tarafında kaynak tüketimi
  - Çok düşük: Bağlantı bekleme süresi, throughput düşer

MinConns: Idle bağlantı sayısı, soğuk başlangıç gecikmesini azaltır

MaxConnLifetime: Uzun süre açık kalan bağlantılar DB tarafında sorun yaratabilir
  - Aurora/RDS: 30 dakika önerilir (server side timeout)

MaxConnIdleTime: Idle bağlantının kapatılma süresi
  - Düşük trafikte kaynak tasarrufu

HealthCheckPeriod: Bağlantı sağlık kontrolü
  - Network partition sonrası hızlı keşif için önemli

Monitor: Pool metriklerini izle (wait_count, wait_duration)" \
  "$TAG_DB" "$TAG_PERF"

create_entry \
  "JSONB vs JSON PostgreSQL" \
  "/notlar/veritabani" \
  "public" \
  "JSON: Metin olarak saklanır, her sorguda parse edilir
JSONB: Binary format, indexlenebilir, hızlı sorgu

Ne zaman JSONB:
- Sorgulama gerekiyorsa (WHERE data->>'key' = 'value')
- Index ihtiyacı varsa
- Sık okuma yapılıyorsa

Ne zaman JSON:
- Sadece insert/read (sorgulama yok)
- JSON sırası önemliyse (JSONB sırayı garanti etmez)
- Daha az storage (bazen)

JSONB operatörleri:
  ->   : JSON object/array'den değer al (JSON döner)
  ->>  : JSON object/array'den değer al (text döner)
  @>   : Contains (sol sağın alt kümesi mi)
  ?    : Key var mı

Index:
  CREATE INDEX idx_data_gin ON users USING gin(data);
  CREATE INDEX idx_data_key ON users USING btree((data->>'email'));" \
  "$TAG_DB"

# /notlar/linux
create_entry \
  "systemd service yazma" \
  "/notlar/linux" \
  "public" \
  "/etc/systemd/system/mahzen.service:

[Unit]
Description=Mahzen API Server
After=network.target postgresql.service

[Service]
Type=simple
User=mahzen
Group=mahzen
WorkingDirectory=/opt/mahzen
ExecStart=/opt/mahzen/mahzen -config /etc/mahzen/config.yaml
Restart=on-failure
RestartSec=5
Environment=MAHZEN_DATABASE_PASSWORD=secret

[Install]
WantedBy=multi-user.target

Komutlar:
  systemctl daemon-reload
  systemctl enable mahzen
  systemctl start mahzen
  journalctl -u mahzen -f  # log izleme

Restart=always vs on-failure:
  always: Her exit'te restart (0 dahil)
  on-failure: Sadece hatalı exit'te restart" \
  "$TAG_LINUX" "$TAG_DEVOPS"

create_entry \
  "Linux process debugging" \
  "/notlar/linux" \
  "public" \
  "Süreç durumlarını görme:
  ps aux | grep mahzen
  pstree -p | grep mahzen

Dosya descriptor'ları:
  ls -la /proc/<pid>/fd
  lsof -p <pid>

Memory kullanımı:
  cat /proc/<pid>/status | grep -E 'VmRSS|VmSize'
  pmap -x <pid>

CPU kullanımı:
  top -p <pid>
  pidstat -p <pid> 1

Network bağlantıları:
  netstat -tulpn | grep <pid>
  ss -tulpn | grep <pid>

Strace ile syscall izleme:
  strace -p <pid> -f -e trace=network

Core dump alma:
  gcore <pid>
  # veya crash anında otomatik için: ulimit -c unlimited" \
  "$TAG_LINUX" "$TAG_PERF"

create_entry \
  "journalctl kullanımı" \
  "/notlar/linux" \
  "public" \
  "Servis logları:
  journalctl -u mahzen           # belirli servis
  journalctl -u mahzen -f        # follow (tail -f gibi)
  journalctl -u mahzen --since \"1 hour ago\"
  journalctl -u mahzen --since \"2024-01-15\"
  journalctl -u mahzen -p err    # sadece error ve üstü

Tüm loglar:
  journalctl --no-pager           # pager olmadan
  journalctl -b                   # son boot'tan beri
  journalctl -b -1                # önceki boot

Disk kullanımı:
  journalctl --disk-usage
  sudo journalctl --vacuum-size=100M  # max 100MB tut

Output formatları:
  journalctl -o json
  journalctl -o verbose" \
  "$TAG_LINUX"

# /notlar/git
create_entry \
  "Git branch stratejileri" \
  "/notlar/git" \
  "public" \
  "Git Flow:
  - main: Production
  - develop: Development
  - feature/*: Yeni özellikler
  - release/*: Release hazırlığı
  - hotfix/*: Acil düzeltmeler

Trunk-Based (modern):
  - main: Her şey
  - short-lived feature branches (< 1 gün)
  - Feature flags ile deploy

GitHub Flow (basitleştirilmiş):
  - main: Her zaman deployable
  - feature branches → PR → merge

Karşılaştırma:
  Git Flow: Enterprise, scheduled releases
  Trunk-Based: CI/CD mature, continuous deployment
  GitHub Flow: Startup, basit projeler" \
  "$TAG_GIT"

create_entry \
  "Git rebase vs merge" \
  "/notlar/git" \
  "public" \
  "Merge:
  - Yeni commit oluşturur (merge commit)
  - Tarih korunur, gerçek zaman çizgisi
  - git merge feature-branch

Rebase:
  - Commits'i yeniden yazar, düz çizgi
  - git rebase main (feature branch'te)
  - DİKKAT: Pushed commit'leri rebase etme!

Ne zaman ne:
  - Feature branch'te çalışırken: rebase (temiz tarih)
  - main'e merge: merge (gerçek kayıt)
  - Pull request: squash and merge (tek commit)

Interactive rebase:
  git rebase -i HEAD~3  # son 3 commit'i düzenle
  # pick, squash, drop, reword, edit" \
  "$TAG_GIT"

create_entry \
  "Git bisect ile bug bulma" \
  "/notlar/git" \
  "private" \
  "Bug'ın hangi commit'te olduğunu bulmak için binary search:

1. Başlat:
   git bisect start

2. Kötü ve iyi commit'leri işaretle:
   git bisect bad HEAD           # şu an hatalı
   git bisect good v1.2.0        # bu versiyon iyi

3. Git otomatik checkout yapar, test et:
   make test

4. Sonucu bildir:
   git bisect good   # bu commit iyi
   git bisect bad    # bu commit hatalı

5. Bulunca:
   git bisect reset  # bitir

Otomatik:
   git bisect run make test  # test otomatik çalışır

Tespit edilen commit'te ne değiştiğini gör:
   git show <commit-hash>" \
  "$TAG_GIT"

# /notlar/api
create_entry \
  "REST API tasarım prensipleri" \
  "/notlar/api" \
  "public" \
  "Resource-based URL'ler:
  GET    /users           # liste
  GET    /users/123       # tek kayıt
  POST   /users           # oluştur
  PUT    /users/123       # tam güncelleme
  PATCH  /users/123       # kısmi güncelleme
  DELETE /users/123       # sil

Query parameters:
  /users?limit=20&offset=40
  /users?sort=created_at:desc
  /users?filter=status:active

Status codes:
  200 OK, 201 Created, 204 No Content
  400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found
  422 Unprocessable Entity (validation error)
  500 Internal Server Error

Pagination:
  Link header (RFC 5988) veya
  Response body'de { data, total, next_cursor }

Versioning:
  URL path: /v1/users (basit, yaygın)
  Header: Accept: application/vnd.api+json;version=1" \
  "$TAG_API"

create_entry \
  "API rate limiting stratejileri" \
  "/notlar/api" \
  "public" \
  "Token Bucket:
  - Tokens eklenir (refill rate)
  - Her istek bir token tüketir
  - Bucket dolunca istek reddedilir
  - Burst'a izin verir

Leaky Bucket:
  - İstekler kuyruğa alınır
  - Sabit hızla işlenir
  - Queue dolunca reddedilir
  - Smooth traffic

Fixed Window:
  - Örneğin her dakika 100 istek
  - Basit ama window başında burst problemi

Sliding Window:
  - Son N saniye içindeki istekleri say
  - Daha adil ama hesaplama maliyeti var

Implementation:
  - Redis INCR + EXPIRE
  - Token bucket için: Redis clerk veya lua script

Headers:
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 95
  X-RateLimit-Reset: 1640000000" \
  "$TAG_API" "$TAG_PERF"

create_entry \
  "JWT best practices" \
  "/notlar/api" \
  "public" \
  "JWT yapısı: header.payload.signature (base64)

Access Token:
  - Kısa ömür (5-15 dakika)
  - Stateless, DB sorgusu gerektirmez
  - İçinde: user_id, roles, exp, iat

Refresh Token:
  - Uzun ömür (7-30 gün)
  - DB'de saklanır (rotate edilebilir, revoke edilebilir)
  - Sadece yeni access token almak için

Güvenlik:
  - https only (cookie için)
  - HttpOnly cookie (XSS koruması)
  - Secure flag (HTTPS)
  - SameSite=Strict veya Lax (CSRF koruması)
  - Kısa access token ömrü

Blacklist (token revoke):
  - JWT stateless olduğu için anında revoke zor
  - Redis blacklist: logout olan token'ları sakla
  - Her istekte blacklist kontrolü" \
  "$TAG_API" "$TAG_SECURITY"

# /notlar/docker
create_entry \
  "Docker multi-stage build" \
  "/notlar/docker" \
  "public" \
  "# Go için multi-stage Dockerfile

# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mahzen ./cmd/mahzen

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/mahzen .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
USER nobody
ENTRYPOINT [\"./mahzen\"]

Avantajları:
  - Küçük image (alpine + binary ~20MB vs golang image ~800MB)
  - Build araçları runtime'da yok (güvenlik)
  - Daha hızlı deploy

Best practices:
  - .dockerignore kullan
  - Specific tag kullan (alpine:3.19, latest değil)
  - Non-root user ile çalıştır" \
  "$TAG_DOCKER" "$TAG_DEVOPS"

create_entry \
  "Docker compose healthcheck" \
  "/notlar/docker" \
  "public" \
  "services:
  postgres:
    image: postgres:18-alpine
    healthcheck:
      test: [\"CMD-SHELL\", \"pg_isready -U mahzen -d mahzen\"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 10s

  api:
    image: mahzen:latest
    depends_on:
      postgres:
        condition: service_healthy

Healthcheck alanları:
  test: Komut (exit 0 = healthy)
  interval: Kontrol sıklığı
  timeout: Komut zaman aşımı
  retries: Kaç kez başarısız olunca unhealthy
  start_period: Başlangıç grace period

depends_on condition:
  service_started: Container çalışıyor (varsayılan)
  service_healthy: Healthcheck geçti
  service_completed_successfully: Container 0 ile çıktı" \
  "$TAG_DOCKER"

create_entry \
  "Docker resource limits" \
  "/notlar/docker" \
  "private" \
  "Container resource sınırlama:

docker run:
  --memory=\"512m\"       # RAM limiti
  --memory-swap=\"1g\"    # RAM + swap
  --cpus=\"1.5\"          # CPU limiti
  --cpu-shares=512       # CPU ağırlığı (öncelik)

docker-compose:
  services:
    api:
      deploy:
        resources:
          limits:
            cpus: '1.0'
            memory: 512M
          reservations:
            cpus: '0.5'
            memory: 256M

Memory limit aşılırsa: OOM Kill (exit code 137)
CPU limit aşılırsa: Throttling (yavaşlama)

Monitor:
  docker stats
  docker inspect --format '{{.HostConfig.Memory}}' container_name" \
  "$TAG_DOCKER" "$TAG_PERF"

# /notlar/guvenlik
create_entry \
  "OWASP Top 10 2024" \
  "/notlar/guvenlik" \
  "public" \
  "1. Broken Access Control: Yetki kontrolü eksik
2. Cryptographic Failures: Zayıf şifreleme, hardcode key
3. Injection: SQL, command, LDAP injection
4. Insecure Design: Güvenlik tasarım dışı
5. Security Misconfiguration: Default config, açık port
6. Vulnerable Components: Eski kütüphaneler
7. Auth Failures: Zayıf parola, session yönetimi
8. Software/Data Integrity: CI/CD güvenliği, unsigned code
9. Logging/Monitoring: Güvenlik log'u yok
10. SSRF: Server-side request forgery

Her birine karşı:
  1. RBAC, resource-level auth
  2. AES-256, key rotation, env vars
  3. Prepared statements, input validation
  4. Threat modeling, security patterns
  5. Hardening, remove defaults
  6. Dependa bot,定期 audit
  7. MFA, secure session
  8. Signed commits, dependency pinning
  9. Centralized logging, alerting
  10. Allowlist, network segmentation" \
  "$TAG_SECURITY"

create_entry \
  "HTTPS ve TLS sertifikaları" \
  "/notlar/guvenlik" \
  "public" \
  "TLS 1.3 kullan (1.2 deprecated olacak)

Sertifika türleri:
  DV (Domain Validation): Otomatik, sadece domain sahipliği
  OV (Organization Validation): Firma doğrulaması
  EV (Extended Validation): En sıkı, yeşil bar

Let's Encrypt (ücretsiz):
  certbot ile sertifika alma
  Auto-renewal aktif et

Self-signed (development):
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

HTTP to HTTPS redirect:
  nginx config ile 301 redirect

HSTS header:
  Strict-Transport-Security: max-age=31536000; includeSubDomains" \
  "$TAG_SECURITY"

# /notlar/performans
create_entry \
  "Profiling Go uygulamaları" \
  "/notlar/performans" \
  "public" \
  "pprof kullanımı:

import _ \"net/http/pprof\"
// http://localhost:6060/debug/pprof/

CPU profile:
  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
  (pprof) top10
  (pprof) list functionName

Memory profile:
  go tool pprof http://localhost:6060/debug/pprof/heap
  (pprof) top
  (pprof) web  # graphviz gerekli

Goroutine:
  go tool pprof http://localhost:6060/debug/pprof/goroutine

Trace:
  curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5
  go tool trace trace.out

Benchmark:
  go test -bench=. -benchmem ./..." \
  "$TAG_GO" "$TAG_PERF"

create_entry \
  "Caching stratejileri" \
  "/notlar/performans" \
  "public" \
  "Cache-aside (lazy loading):
  1. İstek gelince cache kontrol et
  2. Cache miss → DB'den oku → cache'e yaz
  3. Write-through: DB'ye yazarken cache'i de güncelle

Read-through:
  Cache provider okuma işini yapar

Write-behind (write-back):
  Cache'e yaz, async olarak DB'ye yaz
  - Risk: Crash'te veri kaybı

Invalidation:
  - TTL: Basit, eventual consistency
  - Event-based: Değişince invalidate
  - Versioning: Cache key'e version ekle

Redis için:
  SET key value EX 3600  # 1 saat TTL
  GET key
  DEL key
  MGET key1 key2 key3  # batch okuma

Cache stampede önleme:
  - Probabilistic early expiration
  - Lock (sadece bir istek DB'ye gitsin)" \
  "$TAG_PERF"

# /notlar/mimari
create_entry \
  "Clean Architecture prensipleri" \
  "/notlar/mimari" \
  "public" \
  "Katmanlar (içten dışa):
  1. Entities: Domain objects, business rules
  2. Use Cases: Application business rules
  3. Interface Adapters: Controllers, gateways
  4. Frameworks: DB, Web, UI

Bağımlılık kuralı:
  - Sadece iç katmana bağımlılık
  - Dış katmanlar iç katmanları bilir, tersi değil
  - Dependency Inversion ile (interface'ler iç katmanda)

Avantajları:
  - Framework independence
  - Testable business rules
  - UI independence
  - Database independence

Go'da:
  domain/      → Entities, interfaces (stdlib only)
  service/     → Use cases
  handler/     → Interface adapters
  infra/       → Framework implementations" \
  "$TAG_ARCH"

create_entry \
  "Microservices vs Monolith" \
  "/notlar/mimari" \
  "public" \
  "Monolith avantajları:
  - Basit deploy, debug
  - Transaction kolay
  - Network latency yok
  - Startup hızlı

Monolith dezavantajları:
  - Scale edememe (tüm app scale olmak zorunda)
  - Teknoloji kilidi
  - Büyük codebase = yavaş CI/CD

Microservices avantajları:
  - Independent scaling
  - Teknoloji özgürlüğü
  - Team autonomy
  - Fault isolation

Microservices dezavantajları:
  - Distributed system complexity
  - Network latency, failure
  - Distributed transactions (saga pattern)
  - Operational overhead

Kural: Monolith'ten başla, ihtiyaç olunca böl
  \"First make it work, then make it right, then make it fast\"
  Premature microservices = premature optimization" \
  "$TAG_ARCH"

# /notlar/yapay-zeka
create_entry \
  "Embedding modelleri karşılaştırması" \
  "/notlar/yapay-zeka" \
  "public" \
  "OpenAI text-embedding-3-small:
  - 1536 boyut
  - $0.02/1M tokens
  - Genel amaçlı, iyi kalite

OpenAI text-embedding-3-large:
  - 3072 boyut
  - $0.13/1M tokens
  - Daha yüksek kalite, büyük veri setleri

Cohere embed-english-v3.0:
  - 1024 boyut
  - İyi retrieval performansı

Open source:
  - sentence-transformers/all-MiniLM-L6-v2 (384 boyut, hızlı)
  - BGE-large-en (1024 boyut, yüksek kalite)

Karşılaştırma kriterleri:
  - MTEB benchmark
  - Latency
  - Cost
  - Dimensionalite (depolama maliyeti)" \
  "$TAG_AI"

create_entry \
  "Vector search temelleri" \
  "/notlar/yapay-zeka" \
  "public" \
  "Embedding: Metin → sayı vektörü (örn. 1536 boyut)

Similarity metrics:
  - Cosine similarity: -1 ile 1 arası, yön önemli
  - Euclidean distance: Mutlak mesafe
  - Dot product: Cosine * magnitude

Approximate Nearest Neighbor (ANN):
  - Exact search O(n) → ANN O(log n)
  - HNSW: Graph-based, yüksek recall
  - IVF: Clustering-based
  - LSH: Hash-based

Meilisearch vector search:
  - HNSW kullanır
  - Cosine similarity
  - Filter ile combine edilebilir

Index parametreleri:
  - M: Graph connectivity (16-64)
  - ef_construction: Build quality (100-400)
  - Yüksek değer = daha iyi recall, yavaş index" \
  "$TAG_AI" "$TAG_PERF"

# /notlar/frontend
create_entry \
  "React performance optimization" \
  "/notlar/frontend" \
  "public" \
  "Re-render önleme:
  - React.memo: Component memoization
  - useMemo: Value memoization
  - useCallback: Function memoization

Code splitting:
  const LazyComponent = React.lazy(() => import('./Heavy'))
  <Suspense fallback={<Loading />}>
    <LazyComponent />
  </Suspense>

Virtual lists (büyük listeler):
  import { FixedSizeList } from 'react-window'

Bundle optimization:
  - Tree shaking (named exports)
  - Dynamic imports
  - Analyze: rollup-plugin-visualizer

React DevTools Profiler:
  - Record bir interaction
  - Flame graph'ta yavaş component'ları bul
  - Why did this render? analiz et" \
  "$TAG_FRONTEND" "$TAG_PERF"

create_entry \
  "React Query best practices" \
  "/notlar/frontend" \
  "public" \
  "Query key yapısı:
  ['users']                    # tüm kullanıcılar
  ['users', userId]            # tek kullanıcı
  ['users', { page: 1 }]       # filtreli

Stale-while-revalidate:
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['users'],
    queryFn: fetchUsers,
    staleTime: 5 * 60 * 1000,  // 5 dk fresh
    gcTime: 30 * 60 * 1000,    // 30 dk cache
  })

Mutations:
  const mutation = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

Optimistic update:
  onMutate: Cache'i hemen güncelle
  onError: Rollback yap
  onSettled: Refetch trigger

Parallel queries:
  useQueries({ queries: [...] })" \
  "$TAG_FRONTEND" "$TAG_API"

# /gunluk
create_entry \
  "2024-01-15: API redesign kararları" \
  "/gunluk" \
  "private" \
  "Bugün API tasarımını gözden geçirdik:

1. /v1 prefix'i eklendi - future versioning için
2. Pagination: offset/limit yerine cursor-based düşünüldü ama şu an için offset yeterli
3. Error response standardize edildi: { \"error\": \"message\" }
4. Visibility enum: public/private - daha fazlasına gerek yok (team vs user discussion ertelendi)

Teknik borç:
- OpenAPI spec güncellenecek
- Rate limiting eklemek lazım (token bucket)
- Response compression (gzip)

Notlar:
- Tanstack Query frontend'te çok iyi çalışıyor
- Cache invalidation stratejisi netleştirilecek" \
  "$TAG_API" "$TAG_ARCH"

# /referans
create_entry \
  "Git komutları quick reference" \
  "/referans" \
  "public" \
  "# Temel
git status                    # durumu gör
git add -A                    # tüm değişiklikleri stage'e al
git commit -m \"msg\"         # commit et
git push origin main          # push et

# Branch
git checkout -b feature/x     # yeni branch
git branch -a                 # tüm branch'ler
git branch -d feature/x       # branch sil

# History
git log --oneline -20         # son 20 commit
git log --graph --all         # görsel tarih
git show <commit>             # commit detayı

# Undo
git reset HEAD~1              # son commit'i geri al (staged kalır)
git reset --hard HEAD~1       # son commit'i tamamen sil
git revert <commit>           # yeni bir commit ile geri al

# Stash
git stash                     # değişiklikleri sakla
git stash pop                 # geri al

# Remote
git remote -v                 # remote'ları gör
git fetch --all               # tüm remote'ları çek" \
  "$TAG_GIT"

create_entry \
  "PostgreSQL quick reference" \
  "/referans" \
  "public" \
  "-- Bağlantı
psql -h localhost -U mahzen -d mahzen

-- Temel
\\dt                          # tabloları listele
\\d table_name                # tablo yapısı
\\x                          # genişletilmiş görünüm

-- Sorgular
SELECT * FROM entries LIMIT 10;
EXPLAIN ANALYZE SELECT * FROM entries WHERE path = '/notes';

-- Index'ler
SELECT * FROM pg_indexes WHERE tablename = 'entries';
CREATE INDEX CONCURRENTLY idx_x ON table(col);

-- Backup
pg_dump -h localhost -U mahzen mahzen > backup.sql
psql -h localhost -U mahzen mahzen < backup.sql

-- Maintenance
VACUUM ANALYZE entries;
REINDEX TABLE entries;

-- Monitoring
SELECT * FROM pg_stat_activity;
SELECT * FROM pg_stat_user_tables;" \
  "$TAG_DB"

# /notlar/regex
create_entry \
  "Regex temel desenleri" \
  "/notlar/regex" \
  "public" \
  "Karakter sınıfları:
  .       # herhangi bir karakter
  \\d      # rakam [0-9]
  \\w      # kelime karakteri [a-zA-Z0-9_]
  \\s      # boşluk
  [abc]    # a, b veya c
  [^abc]   # a, b, c dışında

Miktarlar:
  *       # 0 veya daha fazla
  +       # 1 veya daha fazla
  ?       # 0 veya 1
  {n}     # tam n tane
  {n,m}   # n ile m arası

Gruplar:
  (abc)   # yakalama grubu
  (?:abc) # yakalamayan grup
  a|b     # a veya b

Yaygın desenler:
  Email: [\\w.-]+@[\\w.-]+\\.\\w+
  UUID: [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}
  URL: https?://[\\w.-]+(?:/[\\w.-]*)*

Lookahead/lookbehind:
  a(?=b)  # a'dan sonra b varsa
  a(?!b)  # a'dan sonra b yoksa" \
  "$TAG_REGEX"

# /notlar/shell
create_entry \
  "Bash scripting best practices" \
  "/notlar/shell" \
  "public" \
  "Shebang ve strict mode:
  #!/usr/bin/env bash
  set -euo pipefail

Değişkenler:
  readonly MAX_RETRIES=3
  local temp_file

Fonksiyonlar:
  log() { echo timestamp ve mesaj; }

Hata yönetimi:
  trap ile ERR sinyalini yakala

Safe quoting:
  Degiskenleri her zaman cift tirnak icine al
  Varsayilan deger icin syntax kullan

Array ornegi:
  items=(a b c)
  for item in items; do echo item; done

Command substitution:
  result=$(some_command)
  if result bos degilse; then islem yap

Temp dosyalar:
  temp=$(mktemp)
  trap ile cikista sil

Exit codes:
  exit 0: success
  exit 1: failure" \
  "$TAG_SHELL"

# /notlar/docker
create_entry \
  "Docker compose override" \
  "/notlar/docker" \
  "public" \
  "# docker-compose.override.yml automatically loaded

# Development override
services:
  api:
    environment:
      - LOG_LEVEL=debug
    volumes:
      - ./:/app  # hot reload için
    command: go run ./cmd/mahzen

# Production (docker-compose.prod.yml)
services:
  api:
    environment:
      - LOG_LEVEL=info
    image: mahzen:v1.0.0
    restart: always

# Usage
docker compose up                    # docker-compose.yml + override
docker compose -f docker-compose.yml -f docker-compose.prod.yml up

# Override only specific values
services:
  api:
    environment:
      LOG_LEVEL: debug  # just override this" \
  "$TAG_DOCKER"

# /notlar/devops
create_entry \
  "CI/CD pipeline yapısı" \
  "/notlar/devops" \
  "public" \
  "GitHub Actions example:

name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }
      - run: go test ./... -race

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go build -o mahzen ./cmd/mahzen
      - uses: actions/upload-artifact@v4
        with:
          name: mahzen
          path: mahzen

  deploy:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - run: echo \"Deploy to production\"

Best practices:
  - Fail fast (test before build)
  - Cache dependencies
  - Parallel jobs where possible
  - Secrets via GitHub Secrets" \
  "$TAG_DEVOPS"

# /projeler/mahzen
create_entry \
  "Mahzen mimari kararları" \
  "/projeler/mahzen" \
  "private" \
  "ADR-001: Clean Architecture
  - handler → service → domain ← infra
  - Dependency inversion with interfaces
  - Status: Accepted

ADR-002: PostgreSQL for all storage
  - All content stored in content column
  - No S3/object storage
  - Rationale: Simplicity, single source of truth
  - Status: Accepted

ADR-003: Meilisearch for search
  - Keyword + semantic (vector) search
  - Status: Accepted

ADR-004: JWT for authentication
  - Access + refresh token pattern
  - HttpOnly cookies
  - Status: Accepted

ADR-005: Go 1.26
  - Latest stable features
  - Improved generics support
  - Status: Accepted" \
  "$TAG_ARCH"

create_entry \
  "Mahzen development roadmap" \
  "/projeler/mahzen" \
  "private" \
  "v0.1 MVP:
  - [x] Entry CRUD
  - [x] Tag system
  - [x] Keyword search
  - [x] JWT auth

v0.2 AI Integration:
  - [x] OpenAI embeddings
  - [x] Semantic search
  - [x] Auto-summary
  - [ ] Auto-tagging (WIP)

v0.3 Polish:
  - [ ] Better error messages
  - [ ] Rate limiting
  - [ ] Audit logging
  - [ ] Batch operations

v1.0 Production:
  - [ ] Multi-tenancy (teams)
  - [ ] Role-based access
  - [ ] Import/export
  - [ ] API documentation

Future:
  - Self-hosted embeddings
  - Offline mode
  - Mobile app" \
  "$TAG_ARCH" "$TAG_AI"

info "Entry'ler oluşturuldu: $created"

# ─── Özet ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  Seed tamamlandı!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  Kullanıcı  : emir@emir.com / emir1234"
echo -e "  Taglar     : 15"
echo -e "  Entry'ler  : $created"
echo -e "  Yollar     : /notlar/golang, /notlar/veritabani,"
echo -e "               /notlar/linux, /notlar/git, /notlar/api,"
echo -e "               /notlar/docker, /notlar/guvenik,"
echo -e "               /notlar/performans, /notlar/mimari,"
echo -e "               /notlar/yapay-zeka, /notlar/frontend,"
echo -e "               /notlar/regex, /notlar/shell, /notlar/devops,"
echo -e "               /gunluk, /referans, /projeler/mahzen"
echo ""
