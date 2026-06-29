CREATE TABLE shipping_quotes (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_hash            VARCHAR(64)  NOT NULL,
    origin_zip_code         VARCHAR(32)  NOT NULL DEFAULT '',
    destination_zip_code    VARCHAR(32)  NOT NULL DEFAULT '',
    origin_zone_code        VARCHAR(64)  NOT NULL,
    destination_zone_code   VARCHAR(64)  NOT NULL,
    shipping_type           VARCHAR(32)  NOT NULL,
    final_price_cents       BIGINT       NOT NULL,
    currency                VARCHAR(8)   NOT NULL,
    estimated_delivery_days INT          NOT NULL,
    was_cached              BOOLEAN      NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_shipping_quotes_request_hash ON shipping_quotes (request_hash);
