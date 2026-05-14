-- +goose Up
-- The frontend RequirePermission gate (AuthProvider.hasPermission) matches
-- the URL last segment of permission.resourceUrl against the route's last
-- segment. The correction route is /invoice/correction (last segment
-- "correction"), but the seeded resource URL was /invoice-correction (last
-- segment "invoice-correction"), so the gate could never match — leaving
-- the correction page unreachable. Align the URL with the route.
--
-- Idempotent: only updates if the URL is still the legacy value.

-- +goose StatementBegin
UPDATE resources
SET url = '/correction'
WHERE name = 'Корректировка оператора'
  AND url = '/invoice-correction';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE resources
SET url = '/invoice-correction'
WHERE name = 'Корректировка оператора'
  AND url = '/correction';
-- +goose StatementEnd
