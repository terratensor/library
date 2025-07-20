<?php

/** @var yii\web\View $this */
/** @var string $name */
/** @var string $message */
/** @var Exception$exception */

use yii\helpers\Html;

$this->title = $name;
?>
<div class="site-error pt-5">
  <div class="search-block">
    <div class="container-fluid">

      <h1><?= Html::encode($this->title) ?></h1>

      <div class="alert alert-danger">
          <?= nl2br(Html::encode($message)) ?>
      </div>

      <p>
        The above error occurred while the Web server was processing your request.
      </p>
      <p>
        Please contact us if you think this is a server error. Thank you.
      </p>

    </div>
  </div>
  <div class="container-fluid search-results">
  </div>
</div>

<div class="site-error">



</div>
