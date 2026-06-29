DELETE FROM promotions WHERE code = 'SHIP10';
DELETE FROM pricing_rules WHERE origin_zone_code IN ('CORDOBA_CAPITAL', '*');
DELETE FROM shipping_zones WHERE code IN ('CABA', 'CORDOBA_CAPITAL', 'PATAGONIA');
