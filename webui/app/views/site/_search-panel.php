<?php

use yii\bootstrap5\Html;
use yii\bootstrap5\ActiveForm;

/** Search settings form block */
echo Html::beginForm(['/site/search-settings'], 'post', ['name' => 'searchSettingsForm', 'class' => 'd-flex']);
echo Html::hiddenInput('value', 'toggle');
echo Html::endForm();
$inputTemplate = '<div class="input-group mb-2">
          {input}
          <button class="btn btn-primary px-3" type="submit" id="button-search"><i class="bi bi-search"></i></button>
          <button class="btn btn-outline-secondary ' .
    (Yii::$app->session->get('show_search_settings') ? 'active' : "") . '" id="button-search-settings">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-sliders" viewBox="0 0 16 16">
              <path fill-rule="evenodd" d="M11.5 2a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zM9.05 3a2.5 2.5 0 0 1 4.9 0H16v1h-2.05a2.5 2.5 0 0 1-4.9 0H0V3h9.05zM4.5 7a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zM2.05 8a2.5 2.5 0 0 1 4.9 0H16v1H6.95a2.5 2.5 0 0 1-4.9 0H0V8h2.05zm9.45 4a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm-2.45 1a2.5 2.5 0 0 1 4.9 0H16v1h-2.05a2.5 2.5 0 0 1-4.9 0H0v-1h9.05z"/>
            </svg>
          </button>
          </div>';
?>
<div class="search-block">
    <div class="container-fluid">
        <?php $form = ActiveForm::begin(
            [
                'method' => 'GET',
                'action' => ['site/search'],
                'options' => ['class' => 'pb-1 mb-2 pt-3', 'autocomplete' => 'off'],
            ]
        ); ?>
        <?= Html::hiddenInput('page', Yii::$app->request->get('page', 1)) ?>
        <!-- Добавляем скрытые поля для сохранения фильтров -->
        <?= Html::hiddenInput('search[genre]', $model->genre) ?>
        <?= Html::hiddenInput('search[author]', $model->author) ?>
        <?= Html::hiddenInput('search[title]', $model->title) ?>
        <?= Html::hiddenInput('search[singleLineMode]', $model->singleLineMode ? '1' : '0') ?>
        <div class="d-flex align-items-center">
            <?= $form->field($model, 'query', [
                'inputTemplate' => $inputTemplate,
                'options' => [
                    'class' => 'w-100',
                    'role' => 'search'
                ]
            ])->textInput(
                [
                    'type' => 'search',
                    'class' => 'form-control form-control-lg',
                    'placeholder' => "Поиск",
                    'autocomplete' => 'off',
                ]
            )->label(false); ?>
        </div>
        <?php if (!empty($model->genre) || !empty($model->author) || !empty($model->title)): ?>
            <!-- Добавляем кнопку сброса -->
            <div class="d-flex align-items-center mb-2 flex-wrap">
                <?= Html::a('Сбросить все', ['site/search'], [
                    'class' => 'btn btn-outline-danger btn-sm me-2' .
                        (empty($model->genre) && empty($model->author) && empty($model->title) ? ' d-none' : ''),
                    'id' => 'reset-filters'
                ]) ?>
                <div id="active-filters-container" class="d-flex flex-wrap">
                    <?php if (!empty($model->genre)): ?>
                        <span class="filter-badge genre-badge" data-bs-toggle="tooltip" title="<?= Html::encode($model->genre) ?>">
                            <span class="text"><?= Html::encode(mb_substr($model->genre, 0, 30) . (mb_strlen($model->genre) > 30 ? '...' : '')) ?></span>
                            <a href="<?= \yii\helpers\Url::to(\src\helpers\SearchHelper::getFilterUrl('genre', '')) ?>"
                                class="text-reset close" aria-label="Удалить">&times;</a>
                        </span>
                    <?php endif; ?>

                    <?php if (!empty($model->author)): ?>
                        <span class="filter-badge author-badge" data-bs-toggle="tooltip" title="<?= Html::encode($model->author) ?>">
                            <span class="text"><?= Html::encode(mb_substr($model->author, 0, 30) . (mb_strlen($model->author) > 30 ? '...' : '')) ?></span>
                            <a href="<?= \yii\helpers\Url::to(\src\helpers\SearchHelper::getFilterUrl('author', '')) ?>"
                                class="text-reset close" aria-label="Удалить">&times;</a>
                        </span>
                    <?php endif; ?>

                    <?php if (!empty($model->title)): ?>
                        <span class="filter-badge title-badge" data-bs-toggle="tooltip" title="<?= Html::encode($model->title) ?>">
                            <span class="text"><?= Html::encode(mb_substr($model->title, 0, 30) . (mb_strlen($model->title) > 30 ? '...' : '')) ?></span>
                            <a href="<?= \yii\helpers\Url::to(\src\helpers\SearchHelper::getFilterUrl('title', '')) ?>"
                                class="text-reset close" aria-label="Удалить">&times;</a>
                        </span>
                    <?php endif; ?>
                </div>
            </div>
        <?php endif; ?>
        <div id="search-setting-panel"
            class="search-setting-panel <?= Yii::$app->session->get('show_search_settings') ? 'show-search-settings' : '' ?>">

            <!-- Чекбокс для включения/выключения нечёткого поиска -->
            <?= $form->field($model, 'fuzzy', ['options' => ['class' => '']])
                ->checkbox()
                ->label('Нечёткий поиск'); ?>
            <!-- Чекбокс для включения/выключения однострочного режима -->
            <?= $form->field($model, 'singleLineMode', [
                'options' => ['class' => 'pb-2 single-line-mode'],
                'template' => "<div class=\"form-check form-switch\">\n{input}\n{label}\n</div>",
                'labelOptions' => ['class' => 'form-check-label'],
            ])->checkbox([
                'class' => 'form-check-input',
                'id' => 'single-line-mode',
                'uncheck' => null,
                'data-scroll' => 'true', // Добавляем атрибут для обработки скролла
            ], false)->label('Однострочный режим (убрать переносы строк)');
            ?>
        </div>

        <?php ActiveForm::end(); ?>
    </div>
</div>