<?php

return [
    'class' => \yii\db\Connection::class,
    'dsn' => 'sqlite:'.realpath(__DIR__.'/../data/library.db'),
    'charset' => 'utf8',
    'enableSchemaCache' => true,

    // Schema cache options (for production environment)
    //'enableSchemaCache' => true,
    //'schemaCacheDuration' => 60,
    //'schemaCache' => 'cache',
];
