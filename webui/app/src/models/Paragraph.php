<?php

declare(strict_types=1);

namespace src\models;

use yii\base\Model;

class Paragraph extends Model
{    
    public string $genre;
    public string $author;
    public string $title;
    public string $text;
    public string $position;
    public string $length;
    public array $highlight;
    public string $source_uuid;
    public string $source;
    private int $id;

    public static function create(
        string $text,
        string $position,
        string $length,
        array $highlight,
    ): self {
        $paragraph = new static();

        $paragraph->text = $text;
        $paragraph->position = $position;
        $paragraph->length = $length;
        $paragraph->highlight = $highlight;

        return $paragraph;
    }

    public function setId($id): void
    {
        $this->id = $id;
    }

    public function getId(): int
    {
        return $this->id;
    }
}
