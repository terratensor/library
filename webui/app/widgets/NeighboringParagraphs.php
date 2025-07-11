<?php

declare(strict_types=1);

namespace app\widgets;

use yii\base\Widget;
use yii\helpers\Html;

class NeighboringParagraphs extends Widget
{
    public int $paragraphID;

    public function run()
    {
        $content = "Параграфы рядом: ";
        $content .= Html::a('3', ['site/neighboring', 'id' => $this->paragraphID, 'num' => 3]);
        $content .= ", ";
        $content .= Html::a('5', ['site/neighboring', 'id' => $this->paragraphID, 'num' => 5]);
        $content .= ", ";
        $content .= Html::a('10', ['site/neighboring', 'id' => $this->paragraphID, 'num' => 10]);
        echo Html::tag('div', $content);
        parent::run();
    }
}
