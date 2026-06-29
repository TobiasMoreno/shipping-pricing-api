CREATE TABLE promotions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code               VARCHAR(64)   NOT NULL UNIQUE,
    description        TEXT          NOT NULL DEFAULT '',
    discount_type      VARCHAR(32)   NOT NULL,
    discount_value     NUMERIC(12,3) NOT NULL,
    max_discount_cents BIGINT        NOT NULL DEFAULT 0,
    starts_at          TIMESTAMPTZ,
    ends_at            TIMESTAMPTZ,
    is_active          BOOLEAN       NOT NULL DEFAULT true,
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT now()
);
