<?php

return [
    'adminEmail' => 'admin@example.com',
    'senderEmail' => 'noreply@example.com',
    'senderName' => 'Example.com mailer',
    'cookieDomain' => getenv('COOKIE_DOMAIN') ?: 'localhost',
    'frontendHostInfo' => getenv('FRONTEND_URL') ?: 'http://localhost:8030',
    'githubRepositoryUrl' => getenv('GH_REPO_URL') ?: 'https://github.com/terratensor/regular-library',
    'cleanDesign' => getenv('CLEAN_DESIGN') ?: 1, // default 0 не показывать меню сайта СВОДД, ничего лишнего
    'manticore' => [
        'host' => 'manticore',
        'port' => 9308,
        'max_matches' => getenv('MANTICORE_MAX_MATCHES') ?: 0, // Maximum amount of matches that the server keeps in RAM for each table and can return to the client. Default is unlimited.
    ],
    'searchResults' => [
        'pageSize' => (int)getenv('PAGE_SIZE') ?: 50,
    ],
    'indexes' => [
        'common' => getenv('MANTICORE_DB_NAME_COMMON') ?: 'library',
        'concept' => getenv('MANTICORE_DB_NAME_COMMON') . '_concept' ?: 'library_concept',
    ],
    'bsVersion' => '5.x'
];
