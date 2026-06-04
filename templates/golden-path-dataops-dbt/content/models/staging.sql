-- ${{ values.name }} — staging
-- ${{ values.description }}
with source as (
    select * from {{ source('raw', '${{ values.name }}') }}
),
cleaned as (
    select
        id,
        created_at,
        updated_at
    from source
    where id is not null
)
select * from cleaned
