<?php

declare(strict_types=1);

namespace app\commands;

use PhpParser\Node\Stmt\Const_;
use SQLite3;
use yii\console\Controller;
use yii\console\ExitCode;
use yii\helpers\Console;

class InitialController extends \yii\console\Controller
{
    public function actionIndex()
    {
        $path = __DIR__ . '/../data/library.db';
        $resource = fopen($path, 'a');
        if ($resource) {
            $this->stdout('Checking the database file is successful' . PHP_EOL, Console::FG_GREEN);
        } else {
            $this->stdout('Checking the database file is an error' . PHP_EOL, Console::FG_RED);
        }
        fclose($resource);
    }
}
