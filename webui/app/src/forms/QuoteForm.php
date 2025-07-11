<?php

declare(strict_types=1);

namespace src\forms;

use yii\base\Model;

class QuoteForm extends Model
{
    public string $uuid = '';

    public function rules(): array
    {
        return [
            ['uuid', 'required'],
            ['uuid', 'string'],
        ];
    }
}
