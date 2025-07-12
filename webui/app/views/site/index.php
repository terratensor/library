<?php

declare(strict_types=1);

use yii\helpers\Url;
use yii\bootstrap5\Html;
use src\forms\SearchForm;

/** @var yii\web\View $this 
 * @var SearchForm $model
 * @var string $errorQueryMessage
 * @var array $aggs
 */

?>
<div class="site-index">
    <?= $this->render('_search-panel', ['model' => $model, 'aggs' => $aggs]) ?>

    <div class="container-fluid search-results">
        <div class="row">
            <div class="col-md-12">
                <div class="total-count mb-4">
                    <h4>Всего документов в базе: <?= number_format($aggs['hits']['total'] ?? 0, 0, '', ' ') ?></h4>
                </div>
            </div>
        </div>

        <div class="row">
            <!-- Жанры -->
            <div class="col-md-4">
                <div class="card facet-card mb-4">
                    <div class="card-header">
                        <h5>Жанры</h5>
                    </div>
                    <div class="card-body">
                        <ul class="facet-list">
                            <?php foreach ($aggs['aggregations']['genre_group']['buckets'] as $genre): ?>
                                <?php if (!empty($genre['key'])): ?>
                                    <li>
                                        <a href="<?= Url::to(['site/search', 'search' => ['genre' => $genre['key']]]) ?>">
                                            <?= Html::encode($genre['key']) ?>
                                            <span class="badge bg-secondary float-end"><?= number_format($genre['doc_count'], 0, '', ' ') ?></span>
                                        </a>
                                    </li>
                                <?php endif; ?>
                            <?php endforeach; ?>
                        </ul>
                    </div>
                </div>
            </div>

            <!-- Авторы -->
            <div class="col-md-4">
                <div class="card facet-card mb-4">
                    <div class="card-header">
                        <h5>Авторы</h5>
                    </div>
                    <div class="card-body">
                        <ul class="facet-list">
                            <?php foreach ($aggs['aggregations']['author_group']['buckets'] as $author): ?>
                                <?php if (!empty($author['key'])): ?>
                                    <li>
                                        <a href="<?= Url::to(['site/search', 'search' => ['author' => $author['key']]]) ?>">
                                            <?= Html::encode($author['key']) ?>
                                            <span class="badge bg-secondary float-end"><?= number_format($author['doc_count'], 0, '', ' ') ?></span>
                                        </a>
                                    </li>
                                <?php endif; ?>
                            <?php endforeach; ?>
                        </ul>
                    </div>
                </div>
            </div>

            <!-- Названия -->
            <div class="col-md-4">
                <div class="card facet-card mb-4">
                    <div class="card-header">
                        <h5>Названия</h5>
                    </div>
                    <div class="card-body">
                        <ul class="facet-list">
                            <?php foreach ($aggs['aggregations']['title_group']['buckets'] as $title): ?>
                                <?php if (!empty($title['key'])): ?>
                                    <li>
                                        <a href="<?= Url::to(['site/search', 'search' => ['title' => $title['key']]]) ?>">
                                            <?= Html::encode(mb_substr($title['key'], 0, 50) . (mb_strlen($title['key']) > 50 ? '...' : '')) ?>
                                            <span class="badge bg-secondary float-end"><?= number_format($title['doc_count'], 0, '', ' ') ?></span>
                                        </a>
                                    </li>
                                <?php endif; ?>
                            <?php endforeach; ?>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<?php $js = <<<JS
$(document).ready(function() {
    // Поиск внутри фасетов
    $('.facet-card .card-header').each(function() {
        var header = $(this);
        var facetType = header.find('h5').text().trim();
        header.append('<div class="facet-search mb-2"><input type="text" class="form-control form-control-sm" placeholder="Поиск в ' + facetType + '..."></div>');
        
        header.find('input').on('keyup', function() {
            var searchText = $(this).val().toLowerCase();
            var list = header.next().find('.facet-list li');
            
            list.each(function() {
                var text = $(this).text().toLowerCase();
                if (text.indexOf(searchText) === -1) {
                    $(this).hide();
                } else {
                    $(this).show();
                }
            });
        });
    });
});
JS;
$this->registerJs($js);
