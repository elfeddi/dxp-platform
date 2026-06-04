<?php
header('Content-Type: application/json');
echo json_encode([
    'service'  => '${{ values.name }}',
    'status'   => 'ok',
    'platform' => 'DxP',
]);
