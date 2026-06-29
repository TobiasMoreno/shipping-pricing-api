INSERT INTO shipping_zones (code, name, region, is_active) VALUES
    ('CABA', 'Ciudad de Buenos Aires', 'Centro', true),
    ('CORDOBA_CAPITAL', 'Cordoba Capital', 'Centro', true),
    ('PATAGONIA', 'Patagonia', 'Sur', false);

INSERT INTO pricing_rules
    (shipping_type, origin_zone_code, destination_zone_code, base_price_cents, price_per_km_cents, price_per_kg_cents, express_multiplier, priority_multiplier, max_weight_kg, min_distance_km, max_distance_km, is_active)
VALUES
    ('standard', 'CORDOBA_CAPITAL', 'CABA', 5000, 4, 1000, 1.5, 1.15, 30, 0, 1500, true),
    ('express',  'CORDOBA_CAPITAL', 'CABA', 8000, 6, 1500, 1.5, 1.15, 20, 0, 1500, true),
    ('standard', '*', '*', 9000, 5, 1200, 1.5, 1.2, 40, 0, 5000, true);

INSERT INTO promotions (code, description, discount_type, discount_value, max_discount_cents, is_active) VALUES
    ('SHIP10', '10% off shipping', 'percentage', 10, 5000, true);
