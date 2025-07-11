<?php

declare(strict_types=1);

namespace app\widgets;

use src\models\Paragraph;
use yii\base\Widget;
use yii\data\Pagination;
use yii\helpers\Html;
use yii\helpers\Url;

class FollowParagraph extends Widget
{
    /**
     * @var string
     */
    public string $title = 'Перейти в книгу';
    /**
     * @var Paragraph
     */
    public Paragraph $paragraph;
    /**
     * @var array|mixed
     */
    protected mixed $paragraph_id;
    /**
     * @var array|mixed
     */
    private mixed $position;

    public Pagination $pagination;

    public function init(): void
    {
        $this->paragraph_id = $this->paragraph->getId();
        $this->position = $this->paragraph->position;
    }

    public function getUrl(): string
    {
        $total = ceil($this->paragraph->position /$this->pagination->pageSize);
        return Url::to(
            [
                'book/view',
                'id' => $this->paragraph->book_id,
                'page' => $total,
                'c' => $this->position,
                '#' => $this->position
            ]
        );

    }

    public function run(): string
    {
        return Html::a(
            $this->title,
            $this->getUrl(),
        );
    }
}
