<?php

declare(strict_types=1);

namespace app\bootstrap;

use Manticoresearch\Client;
use src\repositories\ParagraphRepository;
use Yii;
use yii\base\BootstrapInterface;

class SetUp implements BootstrapInterface
{
    public function bootstrap($app): void
    {
        $container = Yii::$container;

        $container->setSingleton(ParagraphRepository::class, [], [
            new Client($app->params['manticore']),
            $app->params['searchResults']['pageSize'],
        ]);
    }
}
