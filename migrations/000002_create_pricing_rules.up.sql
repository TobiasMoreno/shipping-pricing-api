CREATE TABLE pricing_rules (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipping_type         VARCHAR(32)  NOT NULL,
    origin_zone_code      VARCHAR(64)  NOT NULL,
    destination_zone_code VARCHAR(64)  NOT NULL,
    base_price_cents      BIGINT       NOT NULL,
    price_per_km_cents    BIGINT       NOT NULL,
    price_per_kg_cents    BIGINT       NOT NULL,
    express_multiplier    NUMERIC(6,3) NOT NULL DEFAULT 1,
    priority_multiplier   NUMERIC(6,3) NOT NULL DEFAULT 1,
    max_weight_kg         NUMERIC(10,3) NOT NULL,
    min_distance_km       NUMERIC(10,3) NOT NULL DEFAULT 0,
    max_distance_km       NUMERIC(10,3) NOT NULL,
    is_active             BOOLEAN      NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- At most one active rule per exact (shipping_type, origin, destination) combo.
CREATE UNIQUE INDEX uq_pricing_rules_active_combo
    ON pricing_rules (shipping_type, origin_zone_code, destination_zone_code)
    WHERE is_active;
