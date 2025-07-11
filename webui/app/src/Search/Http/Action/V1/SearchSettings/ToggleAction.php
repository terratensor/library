<?php

declare(strict_types=1);

namespace src\Search\Http\Action\V1\SearchSettings;

use src\Search\Command\ToggleSearchSettings\Request\Command;
use src\Search\Command\ToggleSearchSettings\Request\Handler;
use src\Search\Form\SearchSettings\ToggleForm;
use Yii;
use yii\base\Action;

class ToggleAction extends Action
{
    private Handler $handler;

    public function __construct($id, $controller, Handler $handler, $config = [])
    {
        parent::__construct($id, $controller, $config);
        $this->handler = $handler;
    }

    public function run(): void
    {
        $form = new ToggleForm();
        $session = Yii::$app->session;
        var_dump($form->load(Yii::$app->request->post()));
        try {
            if ($form->load(Yii::$app->request->post(), '') && $form->validate()) {
                $command = new Command();
                $command->value = $form->value;
                $command->session = $session;
                $this->handler->handle($command);
            }
        } catch (\DomainException $e) {
            Yii::$app->errorHandler->logException($e);
        }
    }
}
