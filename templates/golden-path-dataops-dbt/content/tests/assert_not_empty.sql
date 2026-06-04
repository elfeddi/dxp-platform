-- Test : la table ne doit pas être vide
select count(*) as total
from {{ ref('staging') }}
having count(*) = 0
