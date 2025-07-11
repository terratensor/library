<?php

use yii\bootstrap5\Html;

?>
<footer class="footer mt-auto py-3 text-muted">
  <div class="container-fluid">
    <div class="d-flex align-items-baseline justify-content-between">
      <span><?= Yii::$app->name; ?>
          </span>
      <span><?= Html::a('Обратная связь', 'https://svodd.ru/contact'); ?></span>
    </div>
  </div>
</footer>
