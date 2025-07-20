<?php

declare(strict_types=1);

namespace app\widgets;

use yii\bootstrap5\Widget;
use yii\data\Pagination;
use yii\helpers\Html;

class SearchResultsSummary extends Widget
{
    public Pagination $pagination;
    private int $totalCount;

    public function init(): void
    {
        $this->totalCount = $this->pagination->totalCount;
    }

    public function renderSummary(): string
    {
        $totalCount = $this->pagination->totalCount;
        $start = (($this->pagination->getPage() + 1) * $this->pagination->pageSize - $this->pagination->pageSize) + 1;
        $end = ($this->pagination->getPage() + 1) * $this->pagination->pageSize;
        if ($end > $totalCount) {
            $end = $totalCount;
        }

        $string = \Yii::t(
            'yii',
            'Показано записей {start} – {end} из {n}',
            [
                'n' => number_format($totalCount, 0, '', ' '),
                'start' => number_format($start, 0, '', ' '),
                'end' => number_format($end, 0, '', ' ')
            ]
        );
        return Html::tag('p', $string, ['class' => 'summary']);
    }

    public function run(): string
    {
        return $this->totalCount ? $this->renderSummary() : '';
    }
}
