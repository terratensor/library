<?php

declare(strict_types=1);

namespace src\forms;

use yii\base\Model;

class SearchForm extends Model
{
    public string $query = '';
    public string $genre = '';
    public string $author = '';
    public string $title = '';
    public string $text = '';
    public string $source_uuid = '';
    public string $matching = 'query_string';
    public bool $singleLineMode = false;

    public function rules(): array
    {
        return [
            ['query', 'string'],
            ['genre', 'string'],
            ['author', 'string'],
            ['title', 'string'],
            ['text', 'string'],
            ['source_uuid', 'string'],
            ['matching', 'in', 'range' => array_keys($this->getMatching())],
            ['singleLineMode', 'boolean'],
        ];
    }

    public function getMatching(): array
    {
        return [
            'query_string' => 'Обычный поиск',
            'match_phrase' => 'Точное соответствие',
            'match' => 'Любое слово',
        ];
    }

    public function formName(): string
    {
        return 'search';
    }
}
