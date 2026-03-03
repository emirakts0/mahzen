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
info "Kullanıcı kaydediliyor: emir@mahzen.dev"
REG=$($CURL -X POST "$BASE/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"emir@mahzen.dev","display_name":"Emir","password":"mahzen123"}')

ACCESS=$(echo "$REG" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4 || true)
REFRESH=$(echo "$REG" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4 || true)

if [ -z "$ACCESS" ]; then
  warn "Kayıt başarısız, giriş deneniyor (kullanıcı zaten var)"
  LOGIN=$($CURL -X POST "$BASE/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"emir@mahzen.dev","password":"mahzen123"}')
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

Cap aşılınca yeni array allocate edilir — büyük sliceler için önceden cap belirle.

Dikkat: Slice'ı fonksiyona geçince backing array paylaşılır. Kopyalamak için copy() kullan.
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
  "PostgreSQL EXPLAIN ANALYZE okuma rehberi" \
  "/notlar/veritabani" \
  "public" \
  "EXPLAIN ANALYZE sorguyu gerçekten çalıştırır ve gerçek süreleri gösterir.

  EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) SELECT ...;

Önemli kavramlar:
- Seq Scan: Tüm tabloyu tarar. Index yoksa veya az satır varsa normal.
- Index Scan: İndex kullanıyor.
- Bitmap Heap Scan: Çok satır için index + heap erişimi.
- Hash Join vs Nested Loop vs Merge Join: Join stratejileri.

Dikkat edilecekler:
- actual rows vs estimated rows arası büyük fark → istatistikler güncel değil (ANALYZE çalıştır)
- cost=0.00..1234.56 rows=1000 width=32 formatı
- Buffers: shared hit=X read=Y → hit cache'den, read diskten

Slow query bulma:
  SELECT query, mean_exec_time, calls
  FROM pg_stat_statements
  ORDER BY mean_exec_time DESC LIMIT 10;" \
  "$TAG_DB"

create_entry \
  "PostgreSQL index stratejileri" \
  "/notlar/veritabani" \
  "public" \
  "B-tree (varsayılan): =, <, >, BETWEEN, LIKE 'foo%' için uygun.
Hash: Sadece = için. PostgreSQL 10+ WAL-safe.
GIN: Array, JSONB, full-text search için.
GiST: Geometrik tipler, range tipler için.
BRIN: Çok büyük, sıralı tablolar için (zaman serisi).

Composite index sırası önemli:
  CREATE INDEX ON orders (user_id, created_at);
  → user_id = ? AND created_at > ? → kullanılır
  → sadece created_at > ? → kullanılmaz

Partial index — koşullu:
  CREATE INDEX ON entries (user_id) WHERE deleted_at IS NULL;

Expression index:
  CREATE INDEX ON users (lower(email));
  → WHERE lower(email) = 'foo@bar.com' kullanır

Covering index (INCLUDE):
  CREATE INDEX ON entries (user_id) INCLUDE (title, created_at);
  → index-only scan mümkün olur" \
  "$TAG_DB"

create_entry \
  "pgx v5 transaction kalıpları" \
  "/notlar/veritabani" \
  "private" \
  "pgx/v5 ile transaction:

  tx, err := pool.Begin(ctx)
  if err != nil { return err }
  defer tx.Rollback(ctx) // commit sonrası no-op

  // ... sorgular ...

  return tx.Commit(ctx)

Savepoint:
  _, err = tx.Exec(ctx, \"SAVEPOINT sp1\")
  // ...
  _, err = tx.Exec(ctx, \"ROLLBACK TO SAVEPOINT sp1\")

Batch (toplu sorgu):
  batch := &pgx.Batch{}
  batch.Queue(\"INSERT INTO ...\", ...)
  batch.Queue(\"UPDATE ...\", ...)
  br := tx.SendBatch(ctx, batch)
  defer br.Close()

pgxpool.Pool thread-safe, connection'ları otomatik yönetir.
Her request için pool.Acquire() yerine doğrudan pool.Query() kullan." \
  "$TAG_DB" "$TAG_GO"

# /notlar/linux
create_entry \
  "Günlük kullandığım Linux komutları" \
  "/notlar/linux" \
  "public" \
  "Disk kullanımı:
  df -h           # filesystem bazında
  du -sh *        # dizin boyutları
  ncdu            # interaktif

Process:
  ps aux | grep foo
  htop            # interaktif
  pgrep -f pattern

Network:
  ss -tlnp        # açık portlar (netstat'ın yerine)
  ip a            # arayüzler
  curl -v         # HTTP debug

Dosya arama:
  find . -name '*.go' -newer go.mod
  fd pattern      # daha hızlı (fd-find)
  rg 'pattern'    # ripgrep, grep'ten hızlı

Log takip:
  journalctl -u myservice -f
  tail -f /var/log/syslog

Sistem bilgisi:
  uname -r        # kernel versiyonu
  lscpu           # CPU
  free -h         # RAM
  lsblk           # diskler" \
  "$TAG_LINUX" "$TAG_SHELL"

create_entry \
  "systemd servis birimi yazmak" \
  "/notlar/linux" \
  "public" \
  "/etc/systemd/system/mahzen.service:

  [Unit]
  Description=Mahzen Knowledge Platform
  After=network.target postgresql.service
  Requires=postgresql.service

  [Service]
  Type=simple
  User=mahzen
  WorkingDirectory=/opt/mahzen
  ExecStart=/opt/mahzen/mahzen
  Restart=on-failure
  RestartSec=5s
  Environment=MAHZEN_DATABASE_HOST=localhost
  EnvironmentFile=/etc/mahzen/env

  # Güvenlik
  NoNewPrivileges=true
  PrivateTmp=true
  ProtectSystem=strict
  ReadWritePaths=/opt/mahzen/data

  [Install]
  WantedBy=multi-user.target

Komutlar:
  systemctl daemon-reload
  systemctl enable --now mahzen
  systemctl status mahzen
  journalctl -u mahzen -f" \
  "$TAG_LINUX" "$TAG_DEVOPS"

create_entry \
  "tmux hızlı başvuru" \
  "/notlar/linux" \
  "public" \
  "Prefix: Ctrl+b (varsayılan)

Session:
  tmux new -s isim      # yeni session
  tmux ls               # listele
  tmux attach -t isim   # bağlan
  prefix + d            # detach

Window (sekme):
  prefix + c    # yeni window
  prefix + ,    # yeniden adlandır
  prefix + n/p  # sonraki/önceki
  prefix + 0-9  # index'e git

Pane (bölme):
  prefix + %    # dikey böl
  prefix + \"    # yatay böl
  prefix + ok   # pane arası geç
  prefix + z    # zoom (büyüt/küçült)
  prefix + x    # pane kapat

Kopyala:
  prefix + [    # copy mode
  space         # seçim başlat
  enter         # kopyala
  prefix + ]    # yapıştır

~/.tmux.conf önerileri:
  set -g mouse on
  set -g history-limit 50000
  set-option -g default-terminal 'tmux-256color'" \
  "$TAG_LINUX" "$TAG_SHELL"

# /notlar/git
create_entry \
  "Git workflow — feature branch" \
  "/notlar/git" \
  "public" \
  "Standart feature branch akışı:

  git checkout -b feat/yeni-ozellik main
  # geliştir...
  git add -p                    # interaktif, dikkatli ekle
  git commit -m 'feat: kısa açıklama'

  # main güncellemek için rebase:
  git fetch origin
  git rebase origin/main

  # PR aç, merge sonrası:
  git checkout main
  git pull
  git branch -d feat/yeni-ozellik

Commit mesajı formatı (Conventional Commits):
  feat:     yeni özellik
  fix:      hata düzeltme
  refactor: davranış değişikliği olmayan refactor
  chore:    build, bağımlılık vb.
  docs:     sadece dokümantasyon
  test:     test ekleme/düzenleme

Altın kural: Her commit çalışır ve tek bir şey yapar." \
  "$TAG_GIT"

create_entry \
  "Git komutları — sık unuttuklanlar" \
  "/notlar/git" \
  "public" \
  "Son commit'i düzelt (push edilmemişse):
  git commit --amend --no-edit

Staged olmayan değişiklikleri geri al:
  git restore dosya.txt

Stage'i geri al:
  git restore --staged dosya.txt

Belirli commit'e kadar yumuşak sıfırla:
  git reset HEAD~3           # commitleri geri al, değişiklikler kalır

Stash:
  git stash push -m 'açıklama'
  git stash list
  git stash pop

Bir commiti başka branch'e uygula:
  git cherry-pick <sha>

Hangi commit bu satırı yazdı:
  git log -S 'aranan kod' --all

Kimin yazdığını bul:
  git blame dosya.go -L 10,20

İki branch farkı:
  git diff main...feature" \
  "$TAG_GIT"

create_entry \
  "Git hooks — pre-commit kurulumu" \
  "/notlar/git" \
  "private" \
  ".git/hooks/pre-commit (chmod +x):

  #!/bin/sh
  set -e

  # Go lint
  golangci-lint run ./...

  # Go test
  go test ./... -race

  # Format kontrolü
  if ! gofmt -l . | grep -q '^'; then
    echo 'gofmt farklılığı var:'
    gofmt -l .
    exit 1
  fi

  echo 'pre-commit geçti'

Veya lefthook / husky ile yönet:
  lefthook.yml:
    pre-commit:
      commands:
        lint:
          run: golangci-lint run ./...
        test:
          run: go test ./... -race

Takım genelinde uygulamak için lefthook install CI'ye ekle." \
  "$TAG_GIT" "$TAG_SHELL"

# /notlar/api
create_entry \
  "REST API tasarım ilkeleri" \
  "/notlar/api" \
  "public" \
  "Kaynak adları çoğul ve isim olmalı:
  GET    /users          → listele
  POST   /users          → oluştur
  GET    /users/{id}     → tek kayıt
  PUT    /users/{id}     → tam güncelle
  PATCH  /users/{id}     → kısmi güncelle
  DELETE /users/{id}     → sil

İlişkili kaynaklar:
  GET /users/{id}/posts  → kullanıcının yazıları

HTTP status kodları:
  200 OK           → başarılı GET, PUT, PATCH
  201 Created      → başarılı POST
  204 No Content   → başarılı DELETE
  400 Bad Request  → geçersiz istek
  401 Unauthorized → kimlik doğrulama gerekli
  403 Forbidden    → yetkisiz
  404 Not Found    → kaynak yok
  409 Conflict     → çakışma (duplicate)
  422 Unprocessable → validasyon hatası
  429 Too Many Requests
  500 Internal Error

Versiyonlama: /v1/, /v2/ — header versiyonlamadan kaçın.
Sayfalama: ?limit=20&offset=0 veya cursor tabanlı." \
  "$TAG_API" "$TAG_ARCH"

create_entry \
  "JWT yapısı ve doğrulama" \
  "/notlar/api" \
  "public" \
  "JWT = Header.Payload.Signature (Base64URL kodlanmış, nokta ile ayrılmış)

Header: {\"alg\":\"HS256\",\"typ\":\"JWT\"}
Payload (claims):
  sub   → subject (user id)
  iat   → issued at
  exp   → expiry
  nbf   → not before (opsiyonel)

İmzalama: HMAC-SHA256(base64(header) + '.' + base64(payload), secret)

Önemli noktalar:
- Payload encrypt edilmez, sadece imzalanır. Hassas veri koyma.
- exp kontrolü kritik — kütüphane otomatik yapıyor mu doğrula.
- Access token kısa (15dk), refresh token uzun (7 gün).
- Refresh token rotation: her kullanımda yenisi ver, eskisini sil.
- JWS vs JWE: imzalama vs şifreleme — çoğu kullanım JWS yeter.

Yaygın hata: alg:none saldırısı. Doğrulayıcı her zaman alg'ı kontrol etmeli." \
  "$TAG_API" "$TAG_SECURITY"

create_entry \
  "HTTP cache başlıkları" \
  "/notlar/api" \
  "public" \
  "Cache-Control direktifleri:
  no-store         → asla cache'leme
  no-cache         → önce sunucuya sor (revalidate)
  private          → sadece tarayıcı cache'ler
  public           → CDN de cache'leyebilir
  max-age=3600     → 1 saat cache'le
  s-maxage=3600    → CDN için max-age
  must-revalidate  → süresi dolunca yeniden doğrula
  immutable        → hiç değişmeyecek (içerik hash'li URL ile)

ETag ve conditional request:
  Sunucu: ETag: \"abc123\"
  İstemci: If-None-Match: \"abc123\"
  Değişmemişse: 304 Not Modified (body yok)

Last-Modified / If-Modified-Since: Tarih bazlı, ETag daha güvenilir.

API'lerde pratik:
  GET /users/{id} → Cache-Control: private, max-age=60
  GET /config     → Cache-Control: public, max-age=3600
  POST/PUT/DELETE → Cache-Control: no-store" \
  "$TAG_API"

# /notlar/docker
create_entry \
  "Docker multi-stage build" \
  "/notlar/docker" \
  "public" \
  "Go uygulaması için örnek:

  # Build aşaması
  FROM golang:1.24-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o /mahzen ./cmd/mahzen

  # Final imaj
  FROM scratch
  COPY --from=builder /mahzen /mahzen
  COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
  EXPOSE 8080
  ENTRYPOINT [\"/mahzen\"]

Faydaları:
- Final imaj sadece binary içerir (~10MB vs ~800MB)
- Build araçları production'a gitmez
- Güvenlik yüzey alanı küçülür

scratch imajı için dikkat:
- ca-certificates kopyalanmazsa TLS çalışmaz
- /etc/passwd olmadığı için USER root olarak çalışır — distroless tercih et" \
  "$TAG_DOCKER" "$TAG_DEVOPS"

create_entry \
  "Docker Compose — geliştirme ortamı" \
  "/notlar/docker" \
  "public" \
  "Mahzen geliştirme ortamı için compose yapısı:

  services:
    postgres:
      image: postgres:16-alpine
      environment:
        POSTGRES_USER: mahzen
        POSTGRES_PASSWORD: mahzen
        POSTGRES_DB: mahzen
      volumes:
        - pg_data:/var/lib/postgresql/data
      healthcheck:
        test: [\"CMD\", \"pg_isready\", \"-U\", \"mahzen\"]
        interval: 5s

    typesense:
      image: typesense/typesense:26.0
      command: --data-dir /data --api-key mahzen-key
      volumes:
        - ts_data:/data

  volumes:
    pg_data:
    ts_data:

Sık kullanılan komutlar:
  docker compose up -d          # arka planda başlat
  docker compose logs -f        # logları takip et
  docker compose exec postgres psql -U mahzen
  docker compose down -v        # container + volume sil" \
  "$TAG_DOCKER" "$TAG_DEVOPS"

# /notlar/guvenik
create_entry \
  "Şifre hashing — bcrypt vs argon2" \
  "/notlar/guvenik" \
  "public" \
  "Asla: MD5, SHA-1, SHA-256 ile şifre hashleme. Bunlar hız için tasarlanmış.
Şifre için: yavaş, tuzlu (salted) algoritmalar.

bcrypt:
  - Maks 72 byte input (uzun şifreler kırpılır)
  - Cost factor: 10-12 genellikle iyi
  - Go: golang.org/x/crypto/bcrypt
  cost := bcrypt.DefaultCost // 10

argon2id (modern tercih):
  - Bellek yoğun → GPU saldırısına dayanıklı
  - OWASP önerisi: m=64MB, t=3, p=4
  - Go: golang.org/x/crypto/argon2
  hash := argon2.IDKey(password, salt, 3, 64*1024, 4, 32)

Tuz (salt): Otomatik dahil edilir, kendin üretmene gerek yok (bcrypt) veya rastgele 16 byte (argon2).

Doğrulama süresi:
  bcrypt cost=12 → ~250ms → iyi
  Çok hızlıysa maliyeti artır" \
  "$TAG_SECURITY"

create_entry \
  "SQL injection önleme" \
  "/notlar/guvenik" \
  "public" \
  "Asla string birleştirme ile sorgu oluşturma:
  // YANLIŞ:
  query := \"SELECT * FROM users WHERE email='\" + email + \"'\"

  // DOĞRU (parametre):
  rows, err := db.Query(\"SELECT * FROM users WHERE email=\$1\", email)

pgx/v5 ile her zaman parametre kullan:
  pool.QueryRow(ctx, \"SELECT id FROM users WHERE email=\$1\", email)

ORM kullanıyorsan bile raw query dikkat:
  // YANLIŞ:
  db.Raw(\"SELECT * FROM users WHERE name = \" + name)
  // DOĞRU:
  db.Raw(\"SELECT * FROM users WHERE name = ?\", name)

Ayrıca:
- Minimum yetki prensibini uygula (DB user sadece gereken izinlere sahip)
- Error mesajlarında SQL detayı sızdırma
- sqlc gibi araçlar type-safe sorgu üretir → injection riski azalır" \
  "$TAG_SECURITY" "$TAG_DB"

create_entry \
  "CORS ve güvenli header'lar" \
  "/notlar/guvenik" \
  "public" \
  "CORS (Cross-Origin Resource Sharing):
  Access-Control-Allow-Origin: https://app.mahzen.dev  // * üretimde kötü
  Access-Control-Allow-Methods: GET, POST, PUT, DELETE
  Access-Control-Allow-Headers: Authorization, Content-Type
  Access-Control-Max-Age: 86400  // preflight cache

Güvenlik header'ları (her response'a ekle):
  Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
  X-Content-Type-Options: nosniff
  X-Frame-Options: DENY
  Content-Security-Policy: default-src 'self'
  Referrer-Policy: strict-origin-when-cross-origin
  Permissions-Policy: camera=(), microphone=()

Cookie güvenliği:
  Set-Cookie: session=...; HttpOnly; Secure; SameSite=Strict; Path=/

HSTS: HTTP'yi HTTPS'e yönlendir, tarayıcı sonraki isteklerde direkt HTTPS kullanır." \
  "$TAG_SECURITY" "$TAG_API"

# /notlar/performans
create_entry \
  "Go profiling — pprof rehberi" \
  "/notlar/performans" \
  "public" \
  "pprof endpoint'i ekle (geliştirme):
  import _ \"net/http/pprof\"
  go http.ListenAndServe(\":6060\", nil)

Profil topla:
  go tool pprof http://localhost:6060/debug/pprof/heap     # bellek
  go tool pprof http://localhost:6060/debug/pprof/profile  # CPU (30sn)
  go tool pprof http://localhost:6060/debug/pprof/goroutine

Etkileşimli analiz:
  (pprof) top10           # en çok kaynak tüketen
  (pprof) web             # SVG grafiği (graphviz gerekli)
  (pprof) list FuncName   # kaynak kodu bazında

Benchmark profili:
  go test -bench=. -cpuprofile cpu.prof -memprofile mem.prof
  go tool pprof cpu.prof

Trace (goroutine zamanlaması):
  curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out
  go tool trace trace.out

Yaygın sorunlar:
- Fazla allocation → sync.Pool veya pre-allocate
- Lock contention → mutex profili
- Sistem çağrısı bekleme → goroutine profili" \
  "$TAG_PERF" "$TAG_GO"

create_entry \
  "Veritabanı bağlantı havuzu boyutlandırma" \
  "/notlar/performans" \
  "public" \
  "Kural: pool büyüklüğü = CPU çekirdeği sayısı * 2 + etkin disk sayısı değil.

Gerçek formül için:
  max_connections = (core_available_per_jvm * 2) + effective_spindle_count

Pratik başlangıç noktası: 10 bağlantı yeterlidir çoğu uygulama için.
HikariCP ölçütleri gösteriyor ki 10 bağlantı 10.000 eşzamanlı isteği kaldırabilir.

pgx pool yapılandırma:
  MaxConns: 25          // max açık bağlantı
  MinConns: 5           // min hazır bağlantı
  MaxConnLifetime: 1h   // eskiyen bağlantıları yenile
  MaxConnIdleTime: 30m  // boşta bekleyen bağlantıları kapat
  HealthCheckPeriod: 1m // sağlık kontrolü

PostgreSQL tarafında:
  max_connections = 200  // toplam
  Her uygulama sunucusu için: max_connections / sunucu_sayısı

PgBouncer ile bağlantı havuzu:
  Transaction mode → ölçeklenebilir, session-level özellikler çalışmaz
  Session mode    → tam uyumluluk, az kazanç" \
  "$TAG_PERF" "$TAG_DB"

# /notlar/mimari
create_entry \
  "Katmanlı mimari — bağımlılık yönü" \
  "/notlar/mimari" \
  "public" \
  "Mahzen'in katman yapısı:
  handler → service → domain ← infra

Domain katmanı:
- Sadece stdlib import eder
- Interface ve entity tanımlar
- Hiçbir şeye bağımlı değil

Service katmanı:
- Domain interface'lerine bağımlı (soyut)
- Business logic burada
- Infra'yı import etmez

Handler katmanı:
- HTTP çerçevesini bilir
- Service'lere bağımlı

Infra katmanı:
- Domain interface'lerini implement eder (DIP)
- Dışa bağımlılıklar burada: DB, cache, S3, AI

Bağlantı noktası: main.go
- Tüm bağımlılıkları oluşturur ve birbirine bağlar
- Constructor injection kullanır
- Global state yok

Bu yapı sayesinde:
- Service testleri infra gerektirmez (mock ile)
- DB değişince handler etkilenmez
- AI sağlayıcı değişince business logic etkilenmez" \
  "$TAG_ARCH"

create_entry \
  "Event-driven vs request-response mimari" \
  "/notlar/mimari" \
  "public" \
  "Request-Response (senkron):
  - Basit, anlaşılması kolay
  - Yüksek gecikme kabul edilemezse tercih et
  - REST, gRPC bu modeli kullanır

Event-driven (asenkron):
  - Üretici ve tüketici ayrışır
  - Yeniden deneme, ölü mesaj kuyruğu (DLQ) gerekir
  - Nihai tutarlılık (eventual consistency)
  - Kafka, RabbitMQ, NATS

Ne zaman event-driven?
  ✓ İşlem uzun sürüyor (email gönderme, video işleme)
  ✓ Birden fazla servis ilgilenecek (fan-out)
  ✓ Kaynak sistemden bağımsız hız kontrolü lazım
  ✗ Basit CRUD uygulamaları
  ✗ Güçlü tutarlılık gereksinimi

Outbox pattern (hybrid):
  - DB'ye yaz + aynı transactionda event kaydet
  - Ayrı worker event'leri message broker'a gönderir
  - At-least-once delivery garantisi" \
  "$TAG_ARCH"

create_entry \
  "Mikroservis mi, monolit mi?" \
  "/notlar/mimari" \
  "public" \
  "Martin Fowler'ın tavsiyesi: Monolith first.

Monolit avantajları:
- Geliştirme hızlı
- Operasyonel karmaşıklık az
- Refactor kolay
- Network latency yok
- Transaction yönetimi basit

Mikroservis ne zaman?
- Bağımsız ölçeklendirme şart (belirli modüller çok daha fazla yük)
- Farklı teknoloji yığınları gerçekten gerekli
- Bağımsız dağıtım zorunluluğu (100+ geliştirici)
- Organizasyon zaten ayrışmış (Conway's Law)

Yanlış mikroservis işaretleri:
- Servisler birbirini senkron çağırıyor (distributed monolith)
- Her değişiklik birden fazla repo gerektiriyor
- Testler çok karmaşık
- Distributed transaction kullanıyorsunuz (2PC vs SAGA)

Mahzen için: Monolit doğru seçim. Yük arttığında modular monolit → servis ayrımı." \
  "$TAG_ARCH"

# /notlar/yapay-zeka
create_entry \
  "Embedding modelleri ve kullanım alanları" \
  "/notlar/yapay-zeka" \
  "public" \
  "Embedding: Metni sayısal vektöre dönüştürme. Anlamsal benzerlik ölçmeye yarar.

OpenAI modelleri:
  text-embedding-3-small: 1536 boyut, ucuz, hızlı (Mahzen kullanıyor)
  text-embedding-3-large: 3072 boyut, daha iyi kalite

Açık kaynak alternatifler:
  nomic-embed-text     → 768 boyut, iyi Türkçe desteği
  BGE-M3               → çok dilli
  all-MiniLM-L6-v2    → küçük, hızlı

Kullanım alanları:
  Anlamsal arama       → sorgu embedding + vektör DB arama
  Öneri sistemi        → benzer içerik bul
  Kümeleme             → k-means, HDBSCAN
  Anomali tespiti      → ortalamadan uzak olanlar
  Çift-kule (bi-encoder) → hızlı benzerlik

Cosine similarity vs dot product vs euclidean:
  Normalize vektörler için hepsi eşdeğer.
  Normalize edilmemişse cosine similarity tercih et." \
  "$TAG_AI"

create_entry \
  "RAG — Retrieval Augmented Generation" \
  "/notlar/yapay-zeka" \
  "public" \
  "RAG: LLM'e bilgi tabanından ilgili bağlam ekleyerek cevap ürettirme.

Akış:
  1. Soru → embedding
  2. Vektör DB'de en yakın N belge → bul
  3. Belgeler + soru → LLM'e gönder
  4. LLM cevap üretir

Neden gerekli:
  - LLM'in bilgisi belirli bir tarihe kadar (cutoff)
  - Özel/gizli bilgi (şirket içi dokümanlar)
  - Hallucination azaltma (kaynağa bağla)

Chunking stratejileri:
  Sabit boyut      → basit, anlam kopmalar olabilir
  Cümle bazlı     → daha anlamlı
  Recursive       → LangChain varsayılanı
  Semantic        → embedding benzerliğine göre kes

Değerlendirme:
  Context recall    → ilgili belgeler bulundu mu?
  Answer faithfulness → cevap bağlama sadık mı?
  Answer relevancy  → soruya cevap veriyor mu?

Mahzen bu pattern'ı destekler: içerikler embedding'leniyor, semantic search ile ilgili entry'ler bulunuyor." \
  "$TAG_AI" "$TAG_ARCH"

# /notlar/frontend
create_entry \
  "React performans optimizasyonu" \
  "/notlar/frontend" \
  "public" \
  "Re-render tetikleyenler: state değişimi, prop değişimi, parent re-render.

useMemo — pahalı hesaplama cache'le:
  const sorted = useMemo(
    () => items.sort(compareFn),
    [items]
  )

useCallback — fonksiyon referansını sabitle:
  const handleClick = useCallback(() => {
    doSomething(id)
  }, [id])

React.memo — prop değişmezse render etme:
  export default React.memo(MyComponent)

Ne zaman optimize et:
  ✓ Profiler gösteriyor ki bu component yavaş
  ✓ Liste içinde sık render olan item'lar
  ✗ Her yere ekleme — okunabilirliği bozar, bazen yavaşlatır

Virtualizasyon — büyük listeler için:
  react-window veya @tanstack/react-virtual

Code splitting:
  const LazyPage = React.lazy(() => import('./HeavyPage'))
  <Suspense fallback={<Spinner />}><LazyPage /></Suspense>" \
  "$TAG_FRONTEND"

create_entry \
  "TypeScript utility types hızlı referans" \
  "/notlar/frontend" \
  "public" \
  "Partial<T>    → tüm property'leri opsiyonel yapar
Required<T>   → tüm property'leri zorunlu yapar
Readonly<T>   → tüm property'leri readonly yapar
Pick<T, K>    → sadece belirli property'leri al
Omit<T, K>    → belirli property'leri çıkar
Record<K, V>  → key-value map tipi
Exclude<T, U> → T'den U'yu çıkar (union için)
Extract<T, U> → T ve U'nun kesişimi
NonNullable<T>→ null ve undefined'ı çıkar
ReturnType<F> → fonksiyon dönüş tipini al
Parameters<F> → fonksiyon parametre tiplerini al
Awaited<T>    → Promise'i çöz

Örnekler:
  type UpdateUser = Partial<Pick<User, 'name' | 'email'>>

  type EventHandler<T> = (event: T) => void

  type ApiResponse<T> = {
    data: T
    error?: string
    total?: number
  }

Template literal types:
  type Route = '/users' | '/entries'
  type GetRoute = \`GET \${Route}\`" \
  "$TAG_FRONTEND"

# /gunluk
create_entry \
  "Haftalık okuma listesi" \
  "/gunluk" \
  "private" \
  "Bu hafta okuyacaklarım:

- [ ] Go 1.24 release notes — range over functions GA oldu
- [ ] 'The Pragmatic Engineer' newsletter — bu haftaki sayı
- [ ] PostgreSQL 17 yeni özellikleri (özellikle MERGE RETURNING)
- [ ] Typesense v26 changelog
- [ ] 'Staff Engineer' kitabı — 3. bölüm

İzleyeceklerim:
- [ ] GopherCon 2024 konuşmaları (YouTube)
- [ ] Kelsey Hightower'ın son Kubernetes konuşması

Dinleyeceklerim:
- [ ] Go Time podcast — context paketi bölümü
- [ ] Software Engineering Daily — LLM infrastructure

Notlar:
- range over func gerçekten faydalı görünüyor, deneyelim
- PostgreSQL MERGE çok güçlü, upsert için doğal" \
  "$TAG_GO" "$TAG_DB"

create_entry \
  "Bugün öğrendiklerim" \
  "/gunluk" \
  "private" \
  "Go'da strings.Builder sıfırlamak:
  var b strings.Builder
  b.Reset()  // yeniden kullanılabilir

pgx'te named arguments:
  pgx henüz named arguments desteklemiyor (\$1, \$2... kullan)
  pgxscan ile struct'a doğrudan map edebilirsin

Typesense multi-search:
  Tek HTTP isteğiyle birden fazla koleksiyonda arama yapılabilir
  Performans açısından avantajlı

tmux'ta bir penceredeki tüm pane'leri senkronize komut:
  prefix + : setw synchronize-panes on
  Hepsine aynı anda yazılır — çoklu sunucu yönetimi için iyi

curl ile HTTP/3 test:
  curl --http3 https://localhost:8080/v1/entries

Bugün çözdüğüm sorun:
  indexEntryAsync timeout'suz bırakılmıştı, OpenAI cevap vermezse
  goroutine sonsuza kadar açık kalıyordu. 2 dakika timeout eklendi." \
  ""

create_entry \
  "Proje fikirleri ve notlar" \
  "/gunluk" \
  "private" \
  "Mahzen için yapılacaklar:
  - [ ] Full-text search için Türkçe tokenizer araştır
  - [ ] Entry versiyonlama — her güncelleme geçmişi sakla
  - [ ] Markdown render desteği (frontend)
  - [ ] Toplu import (JSON/markdown dosyalarından)
  - [ ] Entry'e dosya ekleme (PDF, resim)
  - [ ] Collaborative editing araştır (CRDT? Yjs?)
  - [ ] CLI client — terminal'den hızlı not ekle
  - [ ] Obsidian plugin — mevcut vault'u senkronize et
  - [ ] Webhook desteği — entry oluşununca dış sisteme bildir
  - [ ] Rate limiting — kullanıcı başına limit

Teknik borç:
  - Handler testleri genişletilmeli (infra entegrasyon testleri yok)
  - Config doğrulama mesajları daha açıklayıcı olabilir
  - Graceful shutdown'da açık S3 upload'ları beklemeli mi?

Araştırılacak:
  - pgvector vs Typesense — hangi durumda hangisi?
  - DragonflyDB — Redis uyumlu, daha hızlı
  - Litestream — SQLite replikasyonu (küçük kurulumlar için?)" \
  ""

# /referans
create_entry \
  "HTTP durum kodları tam liste" \
  "/referans" \
  "public" \
  "1xx — Bilgi:
  100 Continue
  101 Switching Protocols

2xx — Başarı:
  200 OK
  201 Created
  202 Accepted (async işlem başladı)
  204 No Content
  206 Partial Content (range request)

3xx — Yönlendirme:
  301 Moved Permanently (kalıcı, cache'lenir)
  302 Found (geçici)
  304 Not Modified (cache geçerli)
  307 Temporary Redirect (metodu değiştirme)
  308 Permanent Redirect (metodu değiştirme)

4xx — İstemci Hatası:
  400 Bad Request
  401 Unauthorized (kimlik doğrulama gerekli)
  403 Forbidden (yetkisiz)
  404 Not Found
  405 Method Not Allowed
  408 Request Timeout
  409 Conflict
  410 Gone (kalıcı olarak silindi)
  422 Unprocessable Entity (validasyon)
  429 Too Many Requests

5xx — Sunucu Hatası:
  500 Internal Server Error
  502 Bad Gateway
  503 Service Unavailable
  504 Gateway Timeout" \
  "$TAG_API"

create_entry \
  "Regex hızlı başvuru" \
  "/referans" \
  "public" \
  "Karakter sınıfları:
  .    → herhangi bir karakter (\\n hariç)
  \\d   → rakam [0-9]
  \\w   → kelime karakteri [a-zA-Z0-9_]
  \\s   → boşluk karakteri
  \\D   → rakam dışı
  \\W   → kelime dışı

Niceleyiciler:
  *    → 0 veya daha fazla (greedy)
  +    → 1 veya daha fazla
  ?    → 0 veya 1
  {n}  → tam n kez
  {n,m}→ n ile m arası
  *?   → lazy (mümkün az)

Çıpalar:
  ^    → satır başı
  $    → satır sonu
  \\b   → kelime sınırı

Gruplar:
  (abc)       → yakalayan grup
  (?:abc)     → yakalamayan grup
  (?P<name>.) → adlandırılmış grup (Go)
  (?=abc)     → lookahead
  (?!abc)     → negatif lookahead

Go'da regex:
  re := regexp.MustCompile(\`^[a-z]+\$\`)
  re.MatchString('hello')
  re.FindAllString(text, -1)
  re.ReplaceAllString(text, replacement)

Sık kullanılan:
  Email:   ^[\\w.+-]+@[\\w-]+\\.[\\w.]+$
  UUID:    ^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$
  Semver:  ^v?(\\d+)\\.(\\d+)\\.(\\d+)$" \
  "$TAG_REGEX"

create_entry \
  "Go standart kütüphane — sık kullanılan paketler" \
  "/referans" \
  "public" \
  "fmt       — formatlı I/O, Printf, Errorf
strings   — dize işlemleri (Builder, Contains, Split, TrimSpace)
strconv   — tip dönüşümleri (Atoi, FormatFloat, ParseBool)
errors    — New, Is, As, Unwrap
context   — Context, WithTimeout, WithCancel, WithValue
io        — Reader, Writer, ReadAll, Copy, NopCloser
os        — Dosya I/O, env vars, exit
path/filepath — platform-ağımsız yol işlemleri
time      — Time, Duration, Since, After, Ticker
sync      — Mutex, RWMutex, WaitGroup, Once
sync/atomic — atomik işlemler (counter vb.)
encoding/json — Marshal, Unmarshal, Decoder, Encoder
net/http  — HTTP client ve server
regexp    — regex
sort      — sliceları sırala
slices    — Go 1.21+ yeni dilim yardımcıları
maps      — Go 1.21+ map yardımcıları
log/slog  — yapılandırılmış loglama (Go 1.21+)
testing   — test ve benchmark
crypto/rand — kriptografik rastgelelik" \
  "$TAG_GO"

create_entry \
  "Makefile kalıpları" \
  "/referans" \
  "public" \
  ".PHONY hedefleri her zaman tanımla (dosya varsa çakışma olmasın):
  .PHONY: build test lint clean

Değişkenler:
  BINARY_NAME := myapp
  BUILD_FLAGS := -ldflags \"-s -w\"

Bağımlı hedefler:
  build: deps
    go build \$(BUILD_FLAGS) -o \$(BINARY_NAME) ./cmd/myapp

Hedef açıklaması (help sistemi):
  ## build: Binary oluştur
  build:
    @echo 'Building...'
    go build ...

  help:  ## Bu yardım mesajını göster
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS=\":.*?## \"}; {printf \"\\033[36m%-20s\\033[0m %s\\n\", $$1, $$2}'

Shell komutu sonucu değişkene:
  VERSION := \$(shell git describe --tags --always)

Hata yoksay:
  -rm -f generated.go  # - öneki hata olsa bile devam et

Paralel yürütme:
  make -j4 target1 target2" \
  "$TAG_SHELL" "$TAG_DEVOPS"

create_entry \
  "JSON işleme — Go" \
  "/referans" \
  "public" \
  "Temel:
  type User struct {
    ID    string \`json:\"id\"\`
    Name  string \`json:\"name,omitempty\"\`
    Age   int    \`json:\"-\"\`  // dahil etme
  }

  data, _ := json.Marshal(user)
  json.Unmarshal(data, &user)

Streaming (büyük dosyalar):
  dec := json.NewDecoder(r)
  for dec.More() {
    var item Item
    dec.Decode(&item)
  }

Dinamik JSON:
  var raw map[string]json.RawMessage
  json.Unmarshal(data, &raw)

json.RawMessage — ertelenmiş parse:
  type Envelope struct {
    Type    string          \`json:\"type\"\`
    Payload json.RawMessage \`json:\"payload\"\`
  }

Özel marshal:
  func (t *MyTime) UnmarshalJSON(b []byte) error {
    // özel format parse
  }

go-json veya sonic — daha hızlı alternatifler:
  Drop-in replacement, büyük payload'larda fark yaratır." \
  "$TAG_GO"

# /projeler
create_entry \
  "Mahzen — proje notları" \
  "/projeler/mahzen" \
  "private" \
  "Mimari kararlar:

1. Neden HTTP/3?
   QUIC, TCP'nin head-of-line blocking sorununu çözer.
   quic-go kütüphanesi Go için olgunlaşmış.
   Alt-Svc header ile HTTP/3 reklam edilir, istemci destekliyorsa geçer.

2. Neden Typesense?
   Elasticsearch'e göre çok daha az kaynak kullanır.
   Tek binary, kolay kurulum.
   Hem full-text hem vector search destekler.
   Go SDK kaliteli.

3. Neden RustFS?
   MinIO-uyumlu S3 API.
   Büyük içerikler (>64KB) object store'a gider, DB şişmez.

4. Neden sqlc?
   Type-safe sorgular, kod üretimi.
   ORM overhead yok.
   Sorguları SQL olarak yazıyorsun, Go kodu otomatik oluşuyor.

5. Neden bcrypt?
   Şifre hashing için endüstri standardı.
   argon2id geçiş planlanıyor.

Bilinen sınırlamalar:
- Entry versiyonlama yok
- Dosya ekleme yok
- Collaborative editing yok
- Rate limiting yok" \
  "$TAG_ARCH" "$TAG_GO"

create_entry \
  "Mahzen — API notları ve curl örnekleri" \
  "/projeler/mahzen" \
  "private" \
  "Kayıt ve giriş:
  curl -sk -X POST https://localhost:8080/v1/auth/register \\
    -H 'Content-Type: application/json' \\
    -d '{\"email\":\"test@test.com\",\"display_name\":\"Test\",\"password\":\"test1234\"}'

Entry oluştur:
  curl -sk -X POST https://localhost:8080/v1/entries \\
    -H 'Authorization: Bearer <token>' \\
    -H 'Content-Type: application/json' \\
    -d '{\"title\":\"Notum\",\"content\":\"İçerik\",\"path\":\"/notlar\",\"visibility\":\"public\"}'

Arama:
  curl -sk 'https://localhost:8080/v1/search/keyword?query=golang' \\
    -H 'Authorization: Bearer <token>'

Tag oluştur ve entry'e ekle:
  TAG=\$(curl -sk -X POST .../v1/tags -d '{\"name\":\"go\"}' | jq -r .tag.id)
  curl -sk -X POST .../v1/entries/{id}/tags \\
    -d \"{\\\"tag_id\\\":\\\"\$TAG\\\"}\"

HTTP/3 test:
  curl --http3 -sk https://localhost:8080/v1/entries

Profil endpoint (geliştirme):
  curl http://localhost:6060/debug/pprof/goroutine?debug=1" \
  "$TAG_API" "$TAG_SHELL"

create_entry \
  "Go bağımlılık yönetimi" \
  "/projeler/mahzen" \
  "public" \
  "go.mod: Modül tanımı, minimum gereksinim versiyonları
go.sum: Kriptografik hash'ler, bütünlük doğrulama

Sık kullanılan komutlar:
  go mod tidy         # gereksiz bağımlılıkları temizle, eksikleri ekle
  go mod download     # önbelleğe al
  go mod verify       # hash'leri doğrula
  go mod vendor       # vendor/ dizinini güncelle

Versiyon:
  go get github.com/foo/bar@v1.2.3    # belirli versiyon
  go get github.com/foo/bar@latest    # son kararlı
  go get github.com/foo/bar@main      # dal

Major versiyon:
  import \"github.com/foo/bar/v2\"  // v2+ için path değişir

Araçlar:
  go list -m all              # tüm bağımlılıklar
  go list -m -u all           # güncellenebilecekler
  govulncheck ./...           # güvenlik açığı tarama

Workspace (çoklu modül geliştirme):
  go work init ./mahzen ./shared
  go work sync" \
  "$TAG_GO"

info "Entryler oluşturuldu: $created"

# ─── Özet ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  Seed tamamlandı!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  Kullanıcı  : emir@mahzen.dev / mahzen123"
echo -e "  Taglar     : 15"
echo -e "  Entry'ler  : $created"
echo -e "  Yollar     : /notlar/golang, /notlar/veritabani,"
echo -e "               /notlar/linux, /notlar/git, /notlar/api,"
echo -e "               /notlar/docker, /notlar/guvenik,"
echo -e "               /notlar/performans, /notlar/mimari,"
echo -e "               /notlar/yapay-zeka, /notlar/frontend,"
echo -e "               /gunluk, /referans, /projeler/mahzen"
echo ""
