-- name: GetAuctionDataForPublic :many
SELECT
    COALESCE(auction_packages.id, 0)::bigint                  AS package_id,
    COALESCE(auction_packages.name, '')::text                 AS package_name,
    COALESCE(auction_items.name, '')::text                    AS item_name,
    COALESCE(auction_items.description, '')::text             AS item_description,
    COALESCE(auction_items.unit, '')::text                    AS item_unit,
    COALESCE(auction_items.quantity, 0)::float8               AS item_quantity,
    COALESCE(auction_items.note, '')::text                    AS item_note,
    COALESCE(auction_participant_prices.unit_price, '0')::text AS participant_price,
    COALESCE(workers.job_title_in_project, '')::text          AS participant_title
FROM auction_items
FULL JOIN auction_packages ON auction_packages.id = auction_items.auction_package_id
FULL JOIN auctions ON auctions.id = auction_packages.auction_id
LEFT JOIN auction_participant_prices ON auction_participant_prices.auction_item_id = auction_items.id
LEFT JOIN users ON users.id = auction_participant_prices.user_id
LEFT JOIN workers ON workers.id = users.worker_id
WHERE auctions.id = $1
ORDER BY auction_packages.id, auction_items.id, auction_participant_prices.user_id;

-- name: GetAuctionDataForPrivate :many
SELECT
    COALESCE(auction_packages.id, 0)::bigint                  AS package_id,
    COALESCE(auction_packages.name, '')::text                 AS package_name,
    COALESCE(auction_items.id, 0)::bigint                     AS item_id,
    COALESCE(auction_items.name, '')::text                    AS item_name,
    COALESCE(auction_items.description, '')::text             AS item_description,
    COALESCE(auction_items.unit, '')::text                    AS item_unit,
    COALESCE(auction_items.quantity, 0)::float8               AS item_quantity,
    COALESCE(auction_items.note, '')::text                    AS item_note,
    COALESCE(auction_participant_prices.comments, '')::text   AS participant_comment,
    COALESCE(users.id, 0)::bigint                             AS participant_user_id,
    COALESCE(auction_participant_prices.unit_price, '0')::text AS participant_price,
    COALESCE(workers.job_title_in_project, '')::text          AS participant_title
FROM auction_items
FULL JOIN auction_packages ON auction_packages.id = auction_items.auction_package_id
FULL JOIN auctions ON auctions.id = auction_packages.auction_id
LEFT JOIN auction_participant_prices ON auction_participant_prices.auction_item_id = auction_items.id
LEFT JOIN users ON users.id = auction_participant_prices.user_id
LEFT JOIN workers ON workers.id = users.worker_id
WHERE auctions.id = $1
ORDER BY auction_packages.id;

-- name: UpsertAuctionParticipantPrice :exec
-- Migration 00002 added a unique index on (auction_item_id, user_id),
-- so the lookup-then-update-or-insert dance collapses into a single
-- atomic upsert that's safe under concurrent writes.
INSERT INTO auction_participant_prices (auction_item_id, user_id, unit_price, comments)
VALUES ($1, $2, $3, $4)
ON CONFLICT (auction_item_id, user_id)
DO UPDATE SET unit_price = EXCLUDED.unit_price, comments = EXCLUDED.comments;
