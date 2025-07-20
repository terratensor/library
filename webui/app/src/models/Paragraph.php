<?php

declare(strict_types=1);

namespace src\models;

use yii\base\Model;

class Paragraph extends Model
{
    public string $genre;
    public string $genre_attr;
    public string $author;
    public string $author_attr;
    public string $title;
    public string $title_attr;
    public string $content;
    public string $chunk;
    public string $char_count;
    public string $word_count;
    public string $language;
    public string $ocr_quality; 
    public array $highlight;
    public string $source_uuid;
    public string $source;
    public int $datetime;
    public int $created_at;
    public int $updated_at;
    private int $id;

    public static function create(
        string $text,
        string $position,
        string $length,
        array $highlight,
    ): self {
        $paragraph = new static();

        $paragraph->content = $text;
        $paragraph->chunk = $position;
        $paragraph->cahr_count = $length;
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
