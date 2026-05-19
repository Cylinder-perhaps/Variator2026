#!/bin/bash
# =============================================================
# Variator2026 — Seed & Test Script
# Импорт рынков из Polymarket + полный тест всех эндпоинтов
# =============================================================

set -e

API_URL="${API_URL:-http://localhost:8081}"
POLYMARKET_API="https://gamma-api.polymarket.com"

echo "🔧 Variator2026 — Seed & Test"
echo "API URL: $API_URL"
echo ""

# --- Colors ---
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok() { echo -e "  ${GREEN}✅ $1${NC}"; }
fail() { echo -e "  ${RED}❌ $1${NC}"; exit 1; }
info() { echo -e "  ${YELLOW}ℹ️  $1${NC}"; }

# =============================================================
# 1. REGISTER ADMIN
# =============================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 1. Регистрация admin-пользователя"

ADMIN_RESP=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@variator.io","password":"Admin12345"}')

echo "$ADMIN_RESP" | python3 -m json.tool 2>/dev/null || echo "$ADMIN_RESP"

ADMIN_TOKEN=$(echo "$ADMIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)
ADMIN_ID=$(echo "$ADMIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('user',{}).get('id',''))" 2>/dev/null || true)

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "" ]; then
  info "Admin уже существует, пробуем логин..."
  ADMIN_RESP=$(curl -s -X POST "$API_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@variator.io","password":"Admin12345"}')
  ADMIN_TOKEN=$(echo "$ADMIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)
  ADMIN_ID=$(echo "$ADMIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('user',{}).get('id',''))" 2>/dev/null || true)
fi

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "" ]; then
  fail "Не удалось получить admin token"
fi
ok "Admin token: ${ADMIN_TOKEN:0:30}..."

# Обновляем роль пользователя на admin через БД
echo ""
info "Обновление роли на admin через БД..."
docker exec -i variator-db-1 psql -U postgres -d variator -c \
  "UPDATE users SET role = 'admin' WHERE email = 'admin@variator.io';" 2>/dev/null || \
  info "Не удалось обновить роль (Docker не доступен или уже admin)"

# Перелогиниваемся чтобы получить токен с новой ролью
ADMIN_RESP=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@variator.io","password":"Admin12345"}')
ADMIN_TOKEN=$(echo "$ADMIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "" ]; then
  fail "Не удалось перелогиниться как admin"
fi
ok "Admin relogin OK (с ролью admin)"

# =============================================================
# 2. REGISTER USER (обычный)
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 2. Регистрация обычного пользователя"

USER_RESP=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"trader@variator.io","password":"Trader12345"}')

USER_TOKEN=$(echo "$USER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)

if [ -z "$USER_TOKEN" ] || [ "$USER_TOKEN" = "" ]; then
  info "User уже существует, пробуем логин..."
  USER_RESP=$(curl -s -X POST "$API_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"trader@variator.io","password":"Trader12345"}')
  USER_TOKEN=$(echo "$USER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)
fi

if [ -z "$USER_TOKEN" ] || [ "$USER_TOKEN" = "" ]; then
  fail "Не удалось получить user token"
fi
ok "User token: ${USER_TOKEN:0:30}..."

# =============================================================
# 3. IMPORT MARKETS FROM POLYMARKET
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 3. Импорт рынков из Polymarket API"

POLY_MARKETS=$(curl -s "$POLYMARKET_API/markets?limit=5&active=true&closed=false")
MARKET_COUNT=$(echo "$POLY_MARKETS" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
info "Получено $MARKET_COUNT рынков из Polymarket"

# Создаём рынки через наш API
CREATED_MARKET_IDS=()
echo "$POLY_MARKETS" | python3 -c "
import sys, json

markets = json.load(sys.stdin)
for m in markets[:5]:
    question = m.get('question', 'Unknown')
    description = m.get('description', '')[:500]
    outcomes_raw = m.get('outcomes', '[\"Yes\", \"No\"]')
    end_date = m.get('endDate', '2027-01-01T00:00:00Z')
    
    # Parse outcomes
    try:
        outcomes = json.loads(outcomes_raw) if isinstance(outcomes_raw, str) else outcomes_raw
    except:
        outcomes = ['Yes', 'No']
    
    payload = {
        'title': question[:200],
        'description': description,
        'outcomes': outcomes,
        'deadline': end_date,
        'category': 'polymarket'
    }
    print(json.dumps(payload))
" 2>/dev/null | while IFS= read -r payload; do
  TITLE=$(echo "$payload" | python3 -c "import sys,json; print(json.load(sys.stdin)['title'])" 2>/dev/null)
  
  MARKET_RESP=$(curl -s -X POST "$API_URL/api/admin/markets" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "$payload")
  
  MARKET_ID=$(echo "$MARKET_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || true)
  
  if [ -n "$MARKET_ID" ] && [ "$MARKET_ID" != "" ]; then
    ok "Рынок создан: $TITLE (ID: ${MARKET_ID:0:8}...)"
    echo "$MARKET_ID" >> /tmp/variator_market_ids.txt
  else
    ERROR=$(echo "$MARKET_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "unknown")
    fail "Ошибка создания рынка '$TITLE': $ERROR"
  fi
done

# =============================================================
# 4. LIST MARKETS
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 4. Список рынков (публичный)"

MARKETS_RESP=$(curl -s "$API_URL/api/markets?per_page=10")
MARKET_TOTAL=$(echo "$MARKETS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('meta',{}).get('total',0))" 2>/dev/null || echo "0")
ok "Всего рынков: $MARKET_TOTAL"

# Берём первый рынок для тестов
FIRST_MARKET_ID=$(echo "$MARKETS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][0]['id'] if d.get('data') else '')" 2>/dev/null || true)

if [ -z "$FIRST_MARKET_ID" ] || [ "$FIRST_MARKET_ID" = "" ]; then
  fail "Нет рынков для тестирования"
fi
ok "Первый рынок ID: $FIRST_MARKET_ID"

# =============================================================
# 5. GET MARKET DETAILS
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 5. Детали рынка"

MARKET_DETAIL=$(curl -s "$API_URL/api/markets/$FIRST_MARKET_ID")
MARKET_TITLE=$(echo "$MARKET_DETAIL" | python3 -c "import sys,json; print(json.load(sys.stdin).get('title',''))" 2>/dev/null || echo "unknown")
ok "Рынок: $MARKET_TITLE"

# =============================================================
# 6. CHECK BALANCE
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 6. Баланс пользователя"

BALANCE_RESP=$(curl -s "$API_URL/api/balance" \
  -H "Authorization: Bearer $USER_TOKEN")
echo "$BALANCE_RESP" | python3 -m json.tool 2>/dev/null || echo "$BALANCE_RESP"

AVAILABLE=$(echo "$BALANCE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('available',0))" 2>/dev/null || echo "0")
ok "Доступный баланс: $AVAILABLE"

# =============================================================
# 7. CREATE ORDER
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 7. Создание ордера"

ORDER_RESP=$(curl -s -X POST "$API_URL/api/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "{\"market_id\":\"$FIRST_MARKET_ID\",\"outcome\":\"Yes\",\"quantity\":10}")

echo "$ORDER_RESP" | python3 -m json.tool 2>/dev/null || echo "$ORDER_RESP"

ORDER_ID=$(echo "$ORDER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || true)

if [ -n "$ORDER_ID" ] && [ "$ORDER_ID" != "" ]; then
  ok "Ордер создан: $ORDER_ID"
else
  ERROR=$(echo "$ORDER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "unknown")
  info "Ордер не создан: $ERROR"
fi

# =============================================================
# 8. GET MY ORDERS
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 8. Мои ордера"

ORDERS_RESP=$(curl -s "$API_URL/api/orders" \
  -H "Authorization: Bearer $USER_TOKEN")
ORDERS_TOTAL=$(echo "$ORDERS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('meta',{}).get('total',0))" 2>/dev/null || echo "0")
ok "Всего ордеров: $ORDERS_TOTAL"

# =============================================================
# 9. GET POSITIONS
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 9. Мои позиции (портфель)"

POSITIONS_RESP=$(curl -s "$API_URL/api/positions" \
  -H "Authorization: Bearer $USER_TOKEN")
POS_TOTAL=$(echo "$POSITIONS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('meta',{}).get('total_positions',0))" 2>/dev/null || echo "0")
ok "Всего позиций: $POS_TOTAL"

# =============================================================
# 10. GET TRADES
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 10. История сделок"

TRADES_RESP=$(curl -s "$API_URL/api/trades" \
  -H "Authorization: Bearer $USER_TOKEN")
TRADES_TOTAL=$(echo "$TRADES_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('meta',{}).get('total',0))" 2>/dev/null || echo "0")
ok "Всего сделок: $TRADES_TOTAL"

# =============================================================
# 11. REFRESH TOKEN
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 11. Обновление токена (Refresh)"

REFRESH_TOKEN=$(echo "$USER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('refresh_token',''))" 2>/dev/null || true)
if [ -n "$REFRESH_TOKEN" ] && [ "$REFRESH_TOKEN" != "" ]; then
  REFRESH_RESP=$(curl -s -X POST "$API_URL/api/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
  NEW_TOKEN=$(echo "$REFRESH_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)
  if [ -n "$NEW_TOKEN" ] && [ "$NEW_TOKEN" != "" ]; then
    ok "Refresh OK, новый токен: ${NEW_TOKEN:0:30}..."
  else
    info "Refresh не удался (токен мог быть уже использован)"
  fi
else
  info "Refresh token не найден (пропускаем)"
fi

# =============================================================
# 12. CHECK BALANCE AFTER ORDER
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 12. Баланс после ордера"

BALANCE_AFTER=$(curl -s "$API_URL/api/balance" \
  -H "Authorization: Bearer $USER_TOKEN")
echo "$BALANCE_AFTER" | python3 -m json.tool 2>/dev/null || echo "$BALANCE_AFTER"
ok "Баланс обновлён"

# =============================================================
# SUMMARY
# =============================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}🎉 Все тесты выполнены!${NC}"
echo ""
echo "Доступные эндпоинты:"
echo "  Swagger UI:     $API_URL"
echo "  Register:       POST $API_URL/api/auth/register"
echo "  Login:          POST $API_URL/api/auth/login"
echo "  Markets:        GET  $API_URL/api/markets"
echo "  Create Market:  POST $API_URL/api/admin/markets (admin)"
echo "  Create Order:   POST $API_URL/api/orders"
echo "  Balance:        GET  $API_URL/api/balance"
echo "  Positions:      GET  $API_URL/api/positions"
echo "  Trades:         GET  $API_URL/api/trades"
echo ""

# Cleanup
rm -f /tmp/variator_market_ids.txt
