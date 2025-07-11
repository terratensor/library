<?php

declare(strict_types=1);

/** @var yii\web\View $this */
/** @var ParagraphDataProvider $results */

/** @var string $bookName */

use app\widgets\ScrollWidget;
use src\models\Paragraph;
use src\repositories\ParagraphDataProvider;


$this->title = "Контекст — $bookName";
?>
<div class="container-fluid quote-results">

    <?php
    // Property totalCount пусто пока не вызваны данные модели getModels(),
    // сначала получаем массив моделей, потом получаем общее их количество
    /** @var Paragraph[] $paragraphs */
    $paragraphs = $results->getModels();
    ?>
  <div class="row">
    <div class="col-md-12">
      <div class="card pt-3">
        <div class="card-body">
          <div class="px-xl-5 px-lg-5 px-md-5 px-sm-3 paragraph">
              <?php foreach ($paragraphs as $paragraph): ?>
                <div data-entity-id="<?= $paragraph->uuid; ?>">
                  <div class="paragraph-text">
                      <?php if (!$paragraph->highlight['text'] || !$paragraph->highlight['text'][0]): ?>
                          <?php echo Yii::$app->formatter->asRaw(htmlspecialchars_decode($paragraph->text)); ?>
                      <?php else: ?>
                          <?php echo Yii::$app->formatter->asRaw(htmlspecialchars_decode($paragraph->highlight['text'][0])); ?>
                      <?php endif; ?>
                  </div>
                </div>

              <?php endforeach; ?>
            <div class="d-flex justify-content-start book-name pt-4">
              <div><strong><i><?= $bookName; ?></i></strong></div>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>
</div>
<?= ScrollWidget::widget(['data_entity_id' => isset($paragraph) ? $paragraph->uuid : 0]); ?>
