-- Daily order fulfillment rate.
--
-- KNOWN DATA GAP: the generator currently always emits fulfillment_status
-- = null (generator/internal/gen/gen.go, NewOrderDetail hardcodes
-- FulfillmentStatus: nil), even though the Avro schema documents
-- 'fulfilled' / 'partial' / 'not_eligible' as valid values. Until the
-- generator is updated to vary this field, every row here will land in
-- unknown_count and fulfillment_rate_pct will read 0. Built to spec now
-- so the fix is a generator-only change later, with no dbt changes needed.

select
    cast(created_at as date) as order_date,
    count(*) as line_item_count,
    sum(case when fulfillment_status = 'fulfilled' then 1 else 0 end) as fulfilled_count,
    sum(case when fulfillment_status = 'partial' then 1 else 0 end) as partial_count,
    sum(case when fulfillment_status = 'not_eligible' then 1 else 0 end) as not_eligible_count,
    sum(case when fulfillment_status is null then 1 else 0 end) as unknown_count,
    round(
        100.0 * sum(case when fulfillment_status = 'fulfilled' then 1 else 0 end)
        / nullif(count(*), 0),
        2
    ) as fulfillment_rate_pct
from {{ ref('stg_order_details') }}
where not is_synthetic_fraud
group by cast(created_at as date)
