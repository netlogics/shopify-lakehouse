-- Repeat purchase rate by acquisition cohort (month of each customer's
-- first order). Excludes synthetic-fraud-flagged order lines (see
-- shopify-lakehouse-92x.2). A customer counts as "repeat" if they placed
-- more than one distinct order, regardless of how many months later.

with customer_orders as (
    select
        customer_id,
        order_id,
        cast(created_at as date) as order_date
    from {{ ref('stg_order_details') }}
    where not is_synthetic_fraud and customer_id is not null
),

customer_summary as (
    select
        customer_id,
        min(order_date) as first_order_date,
        count(distinct order_id) as order_count
    from customer_orders
    group by customer_id
),

cohorts as (
    select
        date_trunc('month', first_order_date) as cohort_month,
        customer_id,
        case when order_count > 1 then 1 else 0 end as is_repeat_customer
    from customer_summary
)

select
    cohort_month,
    count(*) as new_customers,
    sum(is_repeat_customer) as repeat_customers,
    round(100.0 * sum(is_repeat_customer) / nullif(count(*), 0), 2) as repeat_purchase_rate_pct
from cohorts
group by cohort_month
